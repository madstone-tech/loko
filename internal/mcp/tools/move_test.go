package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

func TestMoveTool(t *testing.T) {
	t.Parallel()
	root := fixture(t, "two-systems")
	svc, enc := newService(root), encoding.NewEncoder()
	move, validate := NewMoveTool(svc, enc), NewValidateTool(svc, enc)

	inUse := callJSON(t, move, map[string]any{"from": "container.api", "to": "container.web", "base_revision": readRevision(t, validate)})
	if r, _ := inUse["refusal"].(map[string]any); inUse["ok"] != false || r["reason"] != "address_in_use" {
		t.Errorf("onto an existing address: %v", inUse)
	}
	got := callJSON(t, move, map[string]any{"from": "container.api", "to": "container.orders_api", "base_revision": readRevision(t, validate)})
	if got["ok"] != true {
		t.Fatalf("move: %v", got["refusal"])
	}
	main, _ := os.ReadFile(filepath.Join(root, "main.loko.hcl"))
	if !strings.Contains(string(main), "moved {\n  from = container.api\n  to   = container.orders_api\n}\n") {
		t.Error("no moved block recorded")
	}
	if v := callJSON(t, validate, map[string]any{}); v["ok"] != true {
		t.Errorf("moved project does not compile: %v", v["diagnostics"])
	}
	deps := callJSON(t, NewQueryTool(svc, enc), map[string]any{"kind": "dependents", "address": "container.orders_db"})
	if names(deps["elements"], "address") != "component.repo container.orders_api" {
		t.Errorf("references follow the rename: %v", deps["elements"])
	}
}

// TestMoveKindChangeBatch: a container becoming a component needs a new
// parent, so it is a rename then an update in one apply_edit batch.
func TestMoveKindChangeBatch(t *testing.T) {
	t.Parallel()
	root := fixture(t, "two-systems")
	svc, enc := newService(root), encoding.NewEncoder()
	validate := NewValidateTool(svc, enc)
	res := callJSON(t, NewApplyEditTool(svc, enc), map[string]any{"base_revision": readRevision(t, validate), "edits": []any{
		edit("rename", "element", "container.fraud", "to", "component.fraud"),
		edit("update", "element", "component.fraud", "set", map[string]any{"container": "container.ledger"}),
	}})
	if res["ok"] != true {
		t.Fatalf("kind change batch: %v %v", res["refusal"], res["diagnostics"])
	}
	if v := callJSON(t, validate, map[string]any{}); v["ok"] != true {
		t.Errorf("does not compile: %v", v["diagnostics"])
	}
}
