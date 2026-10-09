package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

// TestAuthoringPerformance is SC-007 and SC-008 on 1,020 elements. Tokens are
// estimated as whitespace-separated fields of the TOON text.
func TestAuthoringPerformance(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("performance budgets are measured without -short and -race")
	}
	root := writeLargeProject(t, 20, 10, 4)
	svc, enc := usecases.NewAuthoringService(newAuthoringDeps(root)), encoding.NewEncoder()
	timed := func(name string, budget time.Duration, fn func() any) string {
		start := time.Now()
		out := fn()
		if d := time.Since(start); d > budget {
			t.Errorf("%s took %v, budget %v", name, d, budget)
		}
		s, _ := out.(string)
		return s
	}
	describe := tools.NewDescribeTool(svc, enc)
	for level, maxTokens := range map[string]int{"summary": 300, "structure": 2000, "full": 0} {
		out := timed("describe "+level, time.Second, func() any { r, _ := describe.Call(t.Context(), map[string]any{"level": level}); return r })
		if n := len(strings.Fields(out)); maxTokens > 0 && n > maxTokens {
			t.Errorf("describe %s is %d tokens, budget %d", level, n, maxTokens)
		}
	}
	query := tools.NewQueryTool(svc, enc)
	for _, args := range []map[string]any{
		{"kind": "dependents", "address": "container.s10c0", "transitive": true},
		{"kind": "dependencies", "address": "system.s0", "transitive": true},
		{"kind": "path", "address": "container.s0c0", "to": "container.s19c9"},
		{"kind": "orphans"}, {"kind": "coupling"},
	} {
		timed("query "+args["kind"].(string), time.Second, func() any { r, _ := query.Call(t.Context(), args); return r })
	}
	rev := tools.NewValidateTool(svc, enc)
	v, _ := rev.Call(t.Context(), map[string]any{"format": "json"})
	token := v.(string)[strings.Index(v.(string), `"revision":"`)+12:]
	token = token[:strings.Index(token, `"`)]
	apply := tools.NewApplyEditTool(svc, enc)
	out := timed("apply_edit", 2*time.Second, func() any {
		r, _ := apply.Call(t.Context(), map[string]any{"base_revision": token, "format": "json", "edits": []any{
			map[string]any{"op": "update", "target": "element", "address": "container.s5c5", "set": map[string]any{"technology": "Go"}},
		}})
		return r
	})
	if !strings.Contains(out, `"ok":true`) {
		t.Errorf("apply_edit: %s", out)
	}
}
