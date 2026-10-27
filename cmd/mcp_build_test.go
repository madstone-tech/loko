package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

func callBuild(t *testing.T, root string, args map[string]any) map[string]any {
	t.Helper()
	tool := tools.NewBuildTool(usecases.NewAuthoringService(newAuthoringDeps(root)), newBuildDeps(), encoding.NewEncoder())
	args["format"] = "json"
	out, err := tool.Call(t.Context(), args)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out.(string)), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// TestMCPBuildMatchesCLI is FR-015: the build tool writes what loko build
// writes, byte for byte, and returns a view inline exactly as on disk.
func TestMCPBuildMatchesCLI(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "two-systems")
	if err := os.MkdirAll(filepath.Join(root, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	mine := filepath.Join(root, "dist", "NOTES.txt")
	if err := os.WriteFile(mine, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := callBuild(t, root, map[string]any{"formats": []any{"d2", "svg"}, "inline": "system-shop"})
	if res["ok"] != true || res["inline"] == nil {
		t.Fatalf("build: %v", res)
	}
	cli := t.TempDir()
	var out, errBuf bytes.Buffer
	if code, err := runBuildWith(context.Background(), BuildOptions{Root: root, Out: cli,
		Formats: []string{"d2", "svg"}, Stdout: &out, Stderr: &errBuf}); err != nil || code != 0 {
		t.Fatalf("cli build: %d %v", code, err)
	}
	got, want := hashTree(t, filepath.Join(root, "dist")), hashTree(t, cli)
	delete(got, "NOTES.txt")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MCP and CLI builds differ:\n%v\n%v", got, want)
	}
	disk, _ := os.ReadFile(filepath.Join(root, "dist", "diagrams", "system-shop.d2"))
	if res["inline"].(map[string]any)["content"] != string(disk) {
		t.Error("inline content differs from the file on disk")
	}
	if _, err := os.Stat(mine); err != nil {
		t.Errorf("a file the user put in dist was removed: %v", err)
	}
}

func TestMCPBuildRefusals(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "two-systems")
	if res := callBuild(t, root, map[string]any{"out": "../elsewhere"}); res["refusal"].(map[string]any)["reason"] != "path_refused" {
		t.Errorf("out outside the project: %v", res)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "elsewhere")); err == nil {
		t.Error("a refused build wrote outside the project")
	}
	unknown := callBuild(t, root, map[string]any{"formats": []any{"d2"}, "inline": "nope"})
	if e := unknown["inline_error"].(map[string]any); e["reason"] != "unknown_view" ||
		!strings.Contains(strings.Join(anyStrings(e["suggestions"]), " "), "system-shop") {
		t.Errorf("unknown view: %v", unknown)
	}
	strict := callBuild(t, copyFixture(t, "first-run"), map[string]any{"formats": []any{"d2"}, "strict": true})
	if strict["ok"] != false || strict["written"].(float64) == 0 {
		t.Errorf("strict with warnings writes and is not ok: %v", strict["ok"])
	}

	broken := copyFixture(t, "two-systems")
	main := filepath.Join(broken, "main.loko.hcl")
	src, _ := os.ReadFile(main)
	if err := os.WriteFile(main, []byte(strings.Replace(string(src), "container.api", "container.apii", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if res := callBuild(t, broken, map[string]any{}); res["ok"] != false || res["diagnostics"] == nil {
		t.Errorf("broken model: %v", res)
	}
	if _, err := os.Stat(filepath.Join(broken, "dist")); err == nil {
		t.Error("a model with errors wrote output")
	}
}

func anyStrings(v any) []string {
	var out []string
	for _, x := range v.([]any) {
		out = append(out, x.(string))
	}
	return out
}

// diagCodes counts the diagnostics of each code in an MCP result.
func diagCodes(res map[string]any) map[string]int {
	got := map[string]int{}
	list, _ := res["diagnostics"].([]any)
	for _, d := range list {
		got[d.(map[string]any)["code"].(string)]++
	}
	return got
}

// TestMCPBuildWideViewHint is feature 018 story 4 through MCP: build returns
// the wide_view hint in compact form, and validate with severity "warning"
// leaves information out.
func TestMCPBuildWideViewHint(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "first-run")
	res := callBuild(t, root, map[string]any{"formats": []any{"svg"}})
	if n := diagCodes(res)["wide_view"]; n != 1 {
		t.Fatalf("build gave %d wide_view hints, want 1: %v", n, res["diagnostics"])
	}
	for _, d := range res["diagnostics"].([]any) {
		m := d.(map[string]any)
		if m["code"] == "wide_view" && (m["range"] != nil || m["summary"] != nil ||
			!strings.Contains(m["detail"].(string), "system-pricing")) {
			t.Errorf("hint not in compact form, or not about system-pricing: %v", m)
		}
	}
	if n := diagCodes(callBuild(t, root, map[string]any{"formats": []any{"d2"}}))["wide_view"]; n != 0 {
		t.Errorf("a D2-only build gave %d hints", n)
	}
	c := newMCPClient(t, root)
	if n := diagCodes(c.call("validate", map[string]any{}))["docs_root_relative"]; n != 7 {
		t.Errorf("validate gave %d docs_root_relative lines, want 7", n)
	}
	if n := diagCodes(c.call("validate", map[string]any{"severity": "warning"}))["docs_root_relative"]; n != 0 {
		t.Errorf(`validate with severity "warning" shows %d information lines`, n)
	}
}
