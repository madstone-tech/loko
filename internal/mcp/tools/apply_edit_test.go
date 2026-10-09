package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

// readRevision returns the revision from a validate call.
func readRevision(t *testing.T, svcTools toolCaller) string {
	t.Helper()
	return callJSON(t, svcTools, map[string]any{})["revision"].(string)
}

func edit(op, target, address string, extra ...any) map[string]any {
	e := map[string]any{"op": op, "target": target, "address": address}
	for i := 0; i < len(extra); i += 2 {
		e[extra[i].(string)] = extra[i+1]
	}
	return e
}

func TestApplyEditTool(t *testing.T) {
	t.Parallel()
	root := fixture(t, "two-systems")
	svc, enc := newService(root), encoding.NewEncoder()
	apply, validate := NewApplyEditTool(svc, enc), NewValidateTool(svc, enc)
	rev := readRevision(t, validate)
	mainBefore, _ := os.ReadFile(filepath.Join(root, "main.loko.hcl"))

	preview := callJSON(t, apply, map[string]any{"base_revision": rev, "preview": true, "edits": []any{
		edit("add", "element", "container.cache", "set", map[string]any{"system": "system.shop", "technology": "Redis"}),
	}})
	files := preview["files"].([]any)
	if preview["ok"] != true || preview["preview"] != true || len(files) != 1 ||
		!strings.Contains(files[0].(map[string]any)["diff"].(string), "+container \"cache\" {") {
		t.Fatalf("preview: %v", preview)
	}
	if now, _ := os.ReadFile(filepath.Join(root, "main.loko.hcl")); string(now) != string(mainBefore) {
		t.Fatal("preview wrote to disk")
	}

	refusals := map[string]map[string]any{
		"invalid_edit":        edit("add", "element", "system.x", "set", map[string]any{"region": "eu"}),
		"not_found":           edit("remove", "relationship", "container.api.uses.nope"),
		"address_in_use":      edit("add", "element", "system.shop"),
		"dangling_references": edit("remove", "element", "container.ledger"),
		"compile_errors":      edit("update", "relationship", "container.api.uses.orders", "set", map[string]any{"target": "container.nowhere"}),
	}
	for reason, e := range refusals {
		got := callJSON(t, apply, map[string]any{"base_revision": rev, "edits": []any{e}})
		r, _ := got["refusal"].(map[string]any)
		if got["ok"] != false || r["reason"] != reason {
			t.Errorf("%s: %v", reason, got)
		}
	}
	stale := callJSON(t, apply, map[string]any{"base_revision": "r1-0123456789abcdef", "edits": []any{edit("add", "element", "system.x")}})
	if stale["refusal"].(map[string]any)["reason"] != "stale_revision" {
		t.Errorf("unknown revision: %v", stale)
	}

	done := callJSON(t, apply, map[string]any{"base_revision": rev, "edits": []any{
		edit("add", "element", "container.cache", "set", map[string]any{"system": "system.shop"}),
		edit("add", "relationship", "container.api.uses.cache", "set", map[string]any{"target": "container.cache"}),
	}})
	if done["ok"] != true || done["revision"] == rev {
		t.Fatalf("apply: %v", done)
	}
	if now, _ := os.ReadFile(filepath.Join(root, "main.loko.hcl")); !strings.Contains(string(now), "target = container.cache") {
		t.Error("the edit was not written")
	}
	if v := callJSON(t, validate, map[string]any{}); v["ok"] != true {
		t.Errorf("the written project does not compile: %v", v["diagnostics"])
	}

	view := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("add", "view", "view.checkout", "set", map[string]any{"include": []any{"container.api", "external.bank"}, "tags": []any{"pci"}}),
	}})
	if view["ok"] != true {
		t.Fatalf("add view: %v", view["refusal"])
	}
	titled := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("update", "element", "container.api", "set", map[string]any{"title": "Orders API"}),
	}})
	if titled["ok"] != true {
		t.Fatalf("set title: %v", titled["refusal"])
	}
	empty := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("update", "element", "container.api", "set", map[string]any{"title": ""}),
	}})
	if r, _ := empty["refusal"].(map[string]any); r["reason"] != "invalid_edit" {
		t.Errorf("an empty title is refused: %v", empty)
	}
	shaped := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("update", "element", "container.orders_db", "set", map[string]any{"shape": "database"}),
	}})
	if shaped["ok"] != true {
		t.Fatalf("set shape: %v", shaped["refusal"])
	}
	onSystem := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("update", "element", "system.shop", "set", map[string]any{"shape": "database"}),
	}})
	if r, _ := onSystem["refusal"].(map[string]any); r["reason"] != "invalid_edit" || !strings.Contains(fmt.Sprint(r["detail"]), "set.shape") {
		t.Errorf("shape on a system is refused naming set.shape: %v", onSystem)
	}
	kinded := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("update", "relationship", "container.api.uses.orders", "set", map[string]any{"kind": "async", "tags": []any{"write"}}),
	}})
	if kinded["ok"] != true {
		t.Fatalf("set kind and tags: %v", kinded["refusal"])
	}
	badKind := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("update", "relationship", "container.api.uses.orders", "set", map[string]any{"kind": "event"}),
	}})
	if r, _ := badKind["refusal"].(map[string]any); r["reason"] != "invalid_edit" || !strings.Contains(fmt.Sprint(r["detail"]), "sync, async, trigger") {
		t.Errorf("an unknown kind is refused listing the allowed ones: %v", badKind)
	}
	desc := callJSON(t, NewDescribeTool(svc, enc), map[string]any{"level": "full"})
	if !strings.Contains(fmt.Sprint(desc["elements"]), "Orders API") {
		t.Errorf("describe shows the title: %v", desc["elements"])
	}
	if !strings.Contains(fmt.Sprint(desc["views"]), "view.checkout") || !strings.Contains(fmt.Sprint(desc["views"]), "external.bank") {
		t.Errorf("describe lists the new view: %v", desc["views"])
	}
}
