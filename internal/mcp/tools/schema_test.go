package tools

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

func TestSchemaTool(t *testing.T) {
	t.Parallel()
	tool := NewSchemaTool(encoding.NewEncoder(), "1.3.0")
	if !strings.Contains(tool.Description(), "Read-only") {
		t.Errorf("description %q does not say the tool is read-only", tool.Description())
	}
	all := callJSON(t, tool, map[string]any{})
	if all["ok"] != true || all["addresses"] == nil || all["docs"] == nil {
		t.Fatalf("full answer: %v", all)
	}
	if all["server_version"] != "1.3.0" {
		t.Errorf("server_version = %v, want 1.3.0", all["server_version"])
	}
	if !strings.Contains(names(all["blocks"], "name"), "binding") {
		t.Errorf("blocks: %s", names(all["blocks"], "name"))
	}
	one := callJSON(t, tool, map[string]any{"block": "instance"})
	if got := names(one["blocks"], "name"); got != "instance" || one["addresses"] != nil {
		t.Errorf("one-block answer: blocks %q, addresses %v", got, one["addresses"])
	}
	bad := callJSON(t, tool, map[string]any{"block": "sequence"})
	if bad["ok"] != false || bad["error"].(map[string]any)["reason"] != "unknown_block" {
		t.Errorf("unknown block: %v", bad)
	}
	// The view block names the reserved view names so a client can choose a
	// non-colliding name without building first.
	view := callJSON(t, tool, map[string]any{"block": "view"})
	reserved := names(view["blocks"].([]any)[0].(map[string]any)["reserved"], "")
	if !strings.Contains(reserved, "landscape") {
		t.Errorf("view reserved names = %q", reserved)
	}
}

// TestSchemaTokenBudget is research R2: the whole language in at most 3,000
// tokens of TOON, one block in at most 500, counted as TestAuthoringPerformance
// counts them.
func TestSchemaTokenBudget(t *testing.T) {
	t.Parallel()
	tool := NewSchemaTool(encoding.NewEncoder(), "1.3.0")
	tokens := func(args map[string]any) int {
		out, err := tool.Call(t.Context(), args)
		if err != nil {
			t.Fatal(err)
		}
		return len(strings.Fields(out.(string)))
	}
	if n := tokens(map[string]any{}); n > 3000 {
		t.Errorf("full schema is %d tokens, budget 3000", n)
	}
	for _, b := range []string{"container", "view", "instance", "uses"} {
		if n := tokens(map[string]any{"block": b}); n > 500 {
			t.Errorf("schema %s is %d tokens, budget 500", b, n)
		}
	}
}
