package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/mcp"
)

type mcpClient struct {
	t     *testing.T
	tools map[string]mcp.Tool
}

func newMCPClient(t *testing.T, root string) *mcpClient {
	c := &mcpClient{t: t, tools: map[string]mcp.Tool{}}
	for _, tool := range newMCPTools(root) {
		c.tools[tool.Name()] = tool
	}
	return c
}

func (c *mcpClient) call(name string, args map[string]any) map[string]any {
	c.t.Helper()
	args["format"] = "json"
	out, err := c.tools[name].Call(c.t.Context(), args)
	if err != nil {
		c.t.Fatalf("%s: %v", name, err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out.(string)), &m); err != nil {
		c.t.Fatalf("%s: %v\n%s", name, err, out)
	}
	return m
}

// TestMCPFromEmptyDirectory is SC-005: from an empty directory to a rendered
// site, through MCP tools only, starting from describe's bootstrap request.
func TestMCPFromEmptyDirectory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	c := newMCPClient(t, root)

	desc := c.call("describe", map[string]any{})
	boot, ok := desc["bootstrap"].(map[string]any)
	if desc["ok"] != false || !ok {
		t.Fatalf("describe in an empty directory: %v", desc)
	}
	req := boot["request"].(map[string]any)
	edit := req["edits"].([]any)[0].(map[string]any)
	edit["set"] = map[string]any{"name": "shop", "description": "Online shop"}

	req["preview"] = true
	prev := c.call("apply_edit", req)
	if prev["ok"] != true || len(prev["files"].([]any)) != 1 {
		t.Fatalf("preview: %v", prev)
	}
	if _, err := os.Stat(filepath.Join(root, "architecture.loko.hcl")); err == nil {
		t.Fatal("the preview wrote the file")
	}
	delete(req, "preview")
	done := c.call("apply_edit", req)
	if done["ok"] != true {
		t.Fatalf("create project: %v", done)
	}

	again := c.call("apply_edit", map[string]any{"base_revision": done["revision"], "edits": req["edits"]})
	if again["refusal"].(map[string]any)["reason"] != "project_exists" {
		t.Errorf("second project: %v", again)
	}

	model := c.call("apply_edit", map[string]any{"base_revision": done["revision"], "edits": []any{
		map[string]any{"op": "add", "target": "element", "address": "system.shop"},
		map[string]any{"op": "add", "target": "element", "address": "person.customer"},
		map[string]any{"op": "add", "target": "relationship", "address": "person.customer.uses.browse",
			"set": map[string]any{"target": "system.shop"}},
	}})
	if model["ok"] != true {
		t.Fatalf("add elements: %v", model)
	}
	if v := c.call("validate", map[string]any{}); v["errors"].(float64) != 0 {
		t.Errorf("validate: %v", v)
	}
	if b := c.call("build", map[string]any{"formats": []any{"html"}}); b["ok"] != true || b["written"].(float64) == 0 {
		t.Errorf("build: %v", b)
	}
	if _, err := os.Stat(filepath.Join(root, "dist", "index.html")); err != nil {
		t.Errorf("no site: %v", err)
	}
}

// TestMCPElementBeforeProject is FR-023.
func TestMCPElementBeforeProject(t *testing.T) {
	t.Parallel()
	c := newMCPClient(t, t.TempDir())
	rev := c.call("validate", map[string]any{})["revision"]
	res := c.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{
		map[string]any{"op": "add", "target": "element", "address": "person.user"}}})
	if res["refusal"].(map[string]any)["reason"] != "no_project" {
		t.Errorf("element before project: %v", res)
	}
}

// TestMCPConcurrentProjectCreation: two clients on one directory; the second
// creation is refused as stale, as any concurrent edit is.
func TestMCPConcurrentProjectCreation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a, b := newMCPClient(t, root), newMCPClient(t, root)
	rev := a.call("describe", map[string]any{})["revision"]
	b.call("describe", map[string]any{})
	edit := map[string]any{"op": "add", "target": "project", "address": "project", "file": "architecture.loko.hcl",
		"set": map[string]any{"name": "shop"}}
	if r := a.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{edit}}); r["ok"] != true {
		t.Fatalf("first: %v", r)
	}
	r := b.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{edit}})
	if reason := r["refusal"].(map[string]any)["reason"]; reason != "stale_revision" && reason != "project_exists" {
		t.Errorf("second: %v", r)
	}
}
