package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMCPMoveRewritesSteps: renaming an element with the move tool rewrites
// every step that names it, in from, to and via (feature 019, US2).
func TestMCPMoveRewritesSteps(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "dynamic-views")
	c := newMCPClient(t, root)
	rev := c.call("validate", map[string]any{})["revision"]
	res := c.call("move", map[string]any{"from": "container.web", "to": "container.storefront", "base_revision": rev})
	if res["ok"] != true {
		t.Fatalf("move: %v", res)
	}
	src, err := os.ReadFile(filepath.Join(root, "main.loko.hcl"))
	if err != nil {
		t.Fatal(err)
	}
	// The moved block that records the rename keeps the old address.
	text, _, _ := strings.Cut(string(src), "\nmoved {")
	for _, want := range []string{"to   = container.storefront", "from = container.storefront", "via  = container.storefront.uses.orders"} {
		if !strings.Contains(text, want) {
			t.Errorf("source lacks %q after the move", want)
		}
	}
	if i := strings.Index(text, "container.web"); i >= 0 {
		t.Errorf("a reference to container.web remains: %q", text[max(0, i-80):min(len(text), i+40)])
	}
	if v := c.call("validate", map[string]any{}); v["errors"].(float64) != 0 || v["warnings"].(float64) != 0 {
		t.Errorf("after the move: %v", v["diagnostics"])
	}
}

// TestMCPAuthorFlow is feature 019 story 4 through MCP only: learn the step
// block, create a flow, insert a step, remove one, build, and read it back.
func TestMCPAuthorFlow(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "dynamic-views")
	c := newMCPClient(t, root)
	step := c.call("schema", map[string]any{"block": "step"})["blocks"].([]any)[0].(map[string]any)
	required := map[string]bool{}
	for _, a := range step["attributes"].([]any) {
		m := a.(map[string]any)
		required[m["name"].(string)] = m["required"] == true
	}
	if len(required) != 5 || !required["from"] || !required["to"] || required["via"] {
		t.Errorf("schema step attributes %v", required)
	}

	apply := func(edit map[string]any) map[string]any {
		t.Helper()
		rev := c.call("validate", map[string]any{})["revision"]
		if p := c.call("apply_edit", map[string]any{"base_revision": rev, "preview": true, "edits": []any{edit}}); p["ok"] != true {
			t.Fatalf("preview %v: %v", edit, p)
		}
		return c.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{edit}})
	}
	addStep := func(label, from, to, before string) {
		t.Helper()
		e := map[string]any{"op": "add", "target": "step", "address": "view.refund.step." + label,
			"set": map[string]any{"from": from, "to": to}}
		if before != "" {
			e["before"] = before
		}
		if r := apply(e); r["ok"] != true {
			t.Fatalf("add step %s: %v", label, r)
		}
	}
	if r := apply(map[string]any{"op": "add", "target": "view", "address": "view.refund",
		"set": map[string]any{"title": "Refund"}}); r["ok"] != true {
		t.Fatalf("add view: %v", r)
	}
	addStep("request", "person.customer", "container.web", "")
	addStep("refund", "container.api", "external.payments", "")
	addStep("store", "container.api", "container.db", "")
	addStep("forward", "container.web", "container.api", "refund")
	if r := apply(map[string]any{"op": "remove", "target": "step", "address": "view.refund.step.store"}); r["ok"] != true {
		t.Fatalf("remove: %v", r)
	}

	rev := c.call("validate", map[string]any{})["revision"]
	bad := c.call("apply_edit", map[string]any{"base_revision": rev, "edits": []any{map[string]any{"op": "add", "target": "step",
		"address": "view.refund.step.ghost", "set": map[string]any{"from": "container.ghost", "to": "container.api"}}}})
	if bad["ok"] == true || bad["refusal"] == nil {
		t.Errorf("a step from an undeclared element was accepted: %v", bad)
	}

	if b := callBuild(t, root, map[string]any{"formats": []any{"d2"}}); b["ok"] != true {
		t.Fatalf("build: %v", b)
	}
	d2, err := os.ReadFile(filepath.Join(root, "dist", "diagrams", "refund.d2"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(d2)
	for _, want := range []string{`"person__customer" -> "system__shop"."container__web": "1`,
		`"system__shop"."container__web" -> "system__shop"."container__api": "2. Places the order`,
		`"system__shop"."container__api" -> "external__payments": "3. Charges the card`} {
		if !strings.Contains(text, want) {
			t.Errorf("refund.d2 lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "container__db") {
		t.Error("the removed step is still drawn")
	}

	desc := c.call("describe", map[string]any{})
	var steps []any
	for _, v := range desc["views"].([]any) {
		if m := v.(map[string]any); m["address"] == "view.refund" {
			steps, _ = m["steps"].([]any)
		}
	}
	if len(steps) != 3 || steps[1].(map[string]any)["name"] != "forward" {
		t.Errorf("describe steps %v", steps)
	}
	flows := c.call("query", map[string]any{"kind": "flows", "address": "container.api"})["flows"].([]any)
	if len(flows) == 0 {
		t.Error("query flows found nothing for container.api")
	}
}
