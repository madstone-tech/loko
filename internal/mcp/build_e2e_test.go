package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

const twoSystems = "../../testdata/projects/two-systems"

// rpcClient drives a real Server over in-memory pipes.
type rpcClient struct {
	t    *testing.T
	in   *io.PipeWriter
	out  *bufio.Reader
	next int
}

func startServer(t *testing.T, root string) *rpcClient {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	srv := NewServer(root, inR, outW)
	svc := usecases.NewAuthoringService(usecases.AuthoringDeps{Root: root, Source: hclsource.New(), Editor: hclsource.NewEditor()})
	enc := encoding.NewEncoder()
	for _, tl := range []Tool{tools.NewDescribeTool(svc, enc), tools.NewQueryTool(svc, enc),
		tools.NewValidateTool(svc, enc), tools.NewApplyEditTool(svc, enc)} {
		if err := srv.RegisterTool(tl); err != nil {
			t.Fatal(err)
		}
	}
	done := make(chan error, 1)
	go func() { done <- srv.Run(context.Background()); _ = outW.Close() }()
	t.Cleanup(func() {
		_ = inW.Close()
		go func() { _, _ = io.Copy(io.Discard, outR) }()
		if err := <-done; err != nil {
			t.Errorf("server: %v", err)
		}
	})
	return &rpcClient{t: t, in: inW, out: bufio.NewReader(outR)}
}

// call sends tools/call and returns the tool's JSON result.
func (c *rpcClient) call(tool string, args map[string]any) map[string]any {
	c.t.Helper()
	c.next++
	args["format"] = "json"
	req, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": c.next, "method": "tools/call",
		"params": map[string]any{"name": tool, "arguments": args}})
	if _, err := c.in.Write(append(req, '\n')); err != nil {
		c.t.Fatal(err)
	}
	line, err := c.out.ReadBytes('\n')
	if err != nil {
		c.t.Fatal(err)
	}
	var resp struct {
		Result struct {
			Content []struct{ Text string } `json:"content"`
		} `json:"result"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(line, &resp); err != nil || resp.Error != nil {
		c.t.Fatalf("%s: %v %v", tool, err, resp.Error)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(resp.Result.Content[0].Text), &out); err != nil {
		c.t.Fatalf("%s result: %v", tool, err)
	}
	return out
}

// apply sends one apply_edit with the current revision and returns the next one.
func (c *rpcClient) apply(rev string, edits []any) string {
	c.t.Helper()
	res := c.call("apply_edit", map[string]any{"base_revision": rev, "edits": edits})
	if res["ok"] != true {
		c.t.Fatalf("apply_edit refused: %v\n%v", res["refusal"], res["diagnostics"])
	}
	return res["revision"].(string)
}

// exportIR compiles a project and returns its IR as generic JSON, ranges removed.
func exportIR(t *testing.T, root string) map[string]any {
	t.Helper()
	res, err := usecases.ExportIR(t.Context(), hclsource.New(), encoding.NewEncoder(), usecases.ExportRequest{Root: root, Format: "json"})
	if err != nil || res.IR == nil {
		t.Fatalf("export %s: %v %v", root, err, res.Diags)
	}
	b, _ := json.Marshal(res.IR)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return stripRanges(m).(map[string]any)
}

func stripRanges(v any) any {
	switch x := v.(type) {
	case map[string]any:
		delete(x, "range")
		for k, e := range x {
			x[k] = stripRanges(e)
		}
	case []any:
		for i, e := range x {
			x[i] = stripRanges(e)
		}
	}
	return v
}

// TestBuildArchitectureEndToEnd is SC-001: starting from nothing but a project
// block, an assistant builds the two-systems architecture over MCP alone.
func TestBuildArchitectureEndToEnd(t *testing.T) {
	t.Parallel()
	want := exportIR(t, twoSystems)
	if len(list(want["elements"])) < 14 || len(list(want["environments"])) != 2 || len(list(want["relationships"])) < 10 {
		t.Fatalf("fixture IR looks empty: %d elements", len(list(want["elements"])))
	}
	root := t.TempDir()
	src, _ := os.ReadFile(filepath.Join(twoSystems, "main.loko.hcl"))
	project := string(src)[:strings.Index(string(src), "\n}\n")+3]
	if err := os.WriteFile(filepath.Join(root, "main.loko.hcl"), []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}

	c := startServer(t, root)
	rev := c.call("validate", map[string]any{})["revision"].(string)
	rev = c.apply(rev, elementEdits(want)) // one batch: every element
	for _, e := range relationshipEdits(want) {
		rev = c.apply(rev, []any{e}) // one at a time, each on the last revision
	}
	c.apply(rev, deploymentEdits(want))

	if v := c.call("validate", map[string]any{}); v["ok"] != true || v["errors"].(float64) != 0 {
		t.Fatalf("built project does not compile: %v", v["diagnostics"])
	}
	if got := exportIR(t, root); !reflect.DeepEqual(got, want) {
		g, _ := json.MarshalIndent(got, "", " ")
		w, _ := json.MarshalIndent(want, "", " ")
		t.Errorf("built IR differs from two-systems:\n--- got\n%s\n--- want\n%s", g, w)
	}
}

func list(v any) []any {
	l, _ := v.([]any)
	return l
}

func str(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

func elementEdits(ir map[string]any) []any {
	rank := map[string]int{"person": 0, "external": 0, "system": 0, "container": 1, "component": 2}
	els := slices.Clone(list(ir["elements"]))
	slices.SortStableFunc(els, func(a, b any) int {
		return rank[str(a.(map[string]any), "kind")] - rank[str(b.(map[string]any), "kind")]
	})
	var out []any
	for _, e := range els {
		el := e.(map[string]any)
		set := map[string]any{}
		for _, k := range []string{"description", "owner", "technology", "docs", "tags"} {
			if v, ok := el[k]; ok {
				set[k] = v
			}
		}
		if p := str(el, "parent"); p != "" {
			set[map[string]string{"container": "system", "component": "container"}[str(el, "kind")]] = p
		}
		out = append(out, map[string]any{"op": "add", "target": "element", "address": el["address"], "set": set})
	}
	return out
}

func relationshipEdits(ir map[string]any) []any {
	var out []any
	for _, r := range list(ir["relationships"]) {
		rel := r.(map[string]any)
		set := map[string]any{"target": rel["target"]}
		for _, k := range []string{"description", "technology"} {
			if v, ok := rel[k]; ok {
				set[k] = v
			}
		}
		out = append(out, map[string]any{"op": "add", "target": "relationship", "address": rel["address"], "set": set})
	}
	return out
}

func deploymentEdits(ir map[string]any) []any {
	var out []any
	for _, e := range list(ir["environments"]) {
		env := e.(map[string]any)
		set := map[string]any{}
		for _, k := range []string{"provider", "account", "region"} {
			if v, ok := env[k]; ok {
				set[k] = v
			}
		}
		out = append(out, map[string]any{"op": "add", "target": "environment", "address": env["address"], "set": set})
		out = append(out, groupEdits(list(env["groups"]))...)
		for _, i := range list(env["instances"]) {
			out = append(out, instanceEdits(str(env, "address"), i.(map[string]any))...)
		}
	}
	return out
}

func groupEdits(groups []any) []any {
	var out []any
	for _, g := range groups {
		grp := g.(map[string]any)
		out = append(out, map[string]any{"op": "add", "target": "group", "address": grp["address"]})
		out = append(out, groupEdits(list(grp["groups"]))...)
	}
	return out
}

func instanceEdits(env string, in map[string]any) []any {
	at := env
	if p := str(in, "placedIn"); p != "" {
		at = p
	}
	set := map[string]any{"of": in["of"]}
	if attrs := list(in["attributes"]); len(attrs) > 0 {
		set["attributes"] = kvMap(attrs)
	}
	out := []any{map[string]any{"op": "add", "target": "instance", "address": at + ".instance." + str(in, "name"), "set": set}}
	for _, c := range list(in["claims"]) {
		claim := c.(map[string]any)
		cs := map[string]any{}
		for _, k := range []string{"address", "addresses"} {
			if v, ok := claim[k]; ok {
				cs[k] = v
			}
		}
		if tags := list(claim["tags"]); len(tags) > 0 {
			cs["tags"] = kvMap(tags)
		}
		out = append(out, map[string]any{"op": "add", "target": "binding", "address": in["address"],
			"binding": map[string]any{"kind": claim["kind"]}, "set": cs})
	}
	return out
}

func kvMap(attrs []any) map[string]any {
	m := map[string]any{}
	for _, a := range attrs {
		kv := a.(map[string]any)
		m[fmt.Sprint(kv["key"])] = kv["value"]
	}
	return m
}
