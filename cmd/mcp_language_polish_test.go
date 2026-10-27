package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/hclsource"
)

// schemaAttrs lists the attribute names the schema tool gives for one block.
func schemaAttrs(t *testing.T, c *mcpClient, block string) []string {
	t.Helper()
	res := c.call("schema", map[string]any{"block": block})
	var names []string
	for _, a := range res["blocks"].([]any)[0].(map[string]any)["attributes"].([]any) {
		names = append(names, a.(map[string]any)["name"].(string))
	}
	return names
}

// applyAndCheck previews then applies one update, and checks the file holds
// (or no longer holds) the given text.
func applyAndCheck(t *testing.T, c *mcpClient, root string, edit map[string]any, text string, present bool) {
	t.Helper()
	rev := c.call("validate", map[string]any{})["revision"]
	prev := c.call("apply_edit", map[string]any{"base_revision": rev, "preview": true, "edits": []any{edit}})
	if prev["ok"] != true || len(prev["files"].([]any)) != 1 {
		t.Fatalf("preview %v: %v", edit, prev)
	}
	if done := c.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{edit}}); done["ok"] != true {
		t.Fatalf("apply %v: %v", edit, done)
	}
	src, err := os.ReadFile(filepath.Join(root, "main.loko.hcl"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), text) != present {
		t.Errorf("after %v, %q present = %v, want %v", edit, text, !present, present)
	}
}

// TestMCPNodeAndInstanceAttributes is feature 018 story 1 through MCP: the
// schema tool lists the new attributes, and apply_edit sets and removes each
// on a node and an instance, with preview.
func TestMCPNodeAndInstanceAttributes(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "deployment-details")
	c := newMCPClient(t, root)
	if got := schemaAttrs(t, c, "node"); !slices.Equal(got, []string{"description", "tags", "technology", "title"}) {
		t.Errorf("schema node attributes = %v", got)
	}
	if got := schemaAttrs(t, c, "instance"); !slices.Contains(got, "technology") {
		t.Errorf("schema instance attributes = %v", got)
	}
	node := "deployment.prod.node.aws.eu_west_1"
	for _, a := range []struct {
		name  string
		value any
		text  string
	}{
		{"title", "Region node", `"Region node"`},
		{"technology", "AWS region", `"AWS region"`},
		{"description", "Ireland", `"Ireland"`},
		{"tags", []any{"zz-region"}, `"zz-region"`},
	} {
		applyAndCheck(t, c, root, map[string]any{"op": "update", "target": "group", "address": node,
			"set": map[string]any{a.name: a.value}}, a.text, true)
		applyAndCheck(t, c, root, map[string]any{"op": "update", "target": "group", "address": node,
			"clear": []any{a.name}}, a.text, false)
	}
	inst := "deployment.prod.instance.orders"
	applyAndCheck(t, c, root, map[string]any{"op": "update", "target": "instance", "address": inst,
		"set": map[string]any{"technology": "DynamoDB on-demand"}}, "DynamoDB on-demand", true)
	applyAndCheck(t, c, root, map[string]any{"op": "update", "target": "instance", "address": inst,
		"clear": []any{"technology"}}, "DynamoDB on-demand", false)
	rev := c.call("validate", map[string]any{})["revision"]
	res := c.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{
		map[string]any{"op": "update", "target": "group", "address": node, "set": map[string]any{"title": ""}}}})
	if res["ok"] == true {
		t.Errorf("an empty node title was accepted: %v", res)
	}
}

// TestMCPViewTitleAndDescription is FR-017 for views: apply_edit sets and
// removes title and description, with preview, and refuses an empty title.
func TestMCPViewTitleAndDescription(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "deployment-details")
	c := newMCPClient(t, root)
	if got := schemaAttrs(t, c, "view"); !slices.Contains(got, "title") || !slices.Contains(got, "description") {
		t.Errorf("schema view attributes = %v", got)
	}
	view := "view.request_path"
	for _, a := range []struct{ name, value string }{{"title", "Order storage"}, {"description", "Receipts left out"}} {
		applyAndCheck(t, c, root, map[string]any{"op": "update", "target": "view", "address": view,
			"set": map[string]any{a.name: a.value}}, `"`+a.value+`"`, true)
		applyAndCheck(t, c, root, map[string]any{"op": "update", "target": "view", "address": view,
			"clear": []any{a.name}}, `"`+a.value+`"`, false)
	}
	rev := c.call("validate", map[string]any{})["revision"]
	res := c.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{
		map[string]any{"op": "update", "target": "view", "address": view, "set": map[string]any{"title": ""}}}})
	if res["ok"] == true {
		t.Errorf("an empty view title was accepted: %v", res)
	}
}

// TestLanguagePolishFixtureIsCanonical: fmt leaves the feature 018 fixture
// unchanged, so the new attributes format like every other.
func TestLanguagePolishFixtureIsCanonical(t *testing.T) {
	t.Parallel()
	files, diags, err := hclsource.NewFormatter().Canonical(t.Context(), filepath.Join(fixtures, "deployment-details"))
	if err != nil || diags.HasErrors() || len(files) != 0 {
		t.Errorf("deployment-details is not canonical: %v %v %+v", err, diags, files)
	}
}
