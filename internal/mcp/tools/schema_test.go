package tools

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

func TestSchemaTool(t *testing.T) {
	t.Parallel()
	tool := NewSchemaTool(encoding.NewEncoder())
	if !strings.Contains(tool.Description(), "Read-only") {
		t.Errorf("description %q does not say the tool is read-only", tool.Description())
	}
	all := callJSON(t, tool, map[string]any{})
	if all["ok"] != true || all["addresses"] == nil || all["docs"] == nil {
		t.Fatalf("full answer: %v", all)
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
}

// TestSchemaTokenBudget is research R2: the whole language in at most 3,000
// tokens of TOON, one block in at most 500, counted as TestAuthoringPerformance
// counts them.
func TestSchemaTokenBudget(t *testing.T) {
	t.Parallel()
	tool := NewSchemaTool(encoding.NewEncoder())
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
