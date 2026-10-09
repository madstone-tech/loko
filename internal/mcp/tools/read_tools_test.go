package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// fixture copies a test project into a temp dir so a test may change it.
func fixture(t *testing.T, name string) string {
	t.Helper()
	dst := t.TempDir()
	if err := os.CopyFS(dst, os.DirFS(filepath.Join("..", "..", "..", "testdata", "projects", name))); err != nil {
		t.Fatal(err)
	}
	return dst
}

// newService wires real adapters the way cmd/wiring.go does. Tests may
// construct adapters; production code under internal/** may not.
func newService(root string) *usecases.AuthoringService {
	return usecases.NewAuthoringService(usecases.AuthoringDeps{
		Root: root, Source: hclsource.New(), Editor: hclsource.NewEditor(),
	})
}

type toolCaller interface {
	Name() string
	Call(ctx context.Context, args map[string]any) (any, error)
}

// callJSON runs a tool with JSON output and decodes the result.
func callJSON(t *testing.T, tl toolCaller, args map[string]any) map[string]any {
	t.Helper()
	args["format"] = "json"
	out, err := tl.Call(t.Context(), args)
	if err != nil {
		t.Fatalf("%s: %v", tl.Name(), err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out.(string)), &m); err != nil {
		t.Fatalf("%s: result is not JSON: %v\n%s", tl.Name(), err, out)
	}
	return m
}

func names(list any, key string) string {
	var out []string
	for _, v := range list.([]any) {
		if key == "" {
			out = append(out, v.(string))
			continue
		}
		out = append(out, v.(map[string]any)[key].(string))
	}
	return strings.Join(out, " ")
}

func TestDescribeTool(t *testing.T) {
	t.Parallel()
	svc, enc := newService(fixture(t, "two-systems")), encoding.NewEncoder()
	tl := NewDescribeTool(svc, enc)

	toon, err := tl.Call(t.Context(), map[string]any{})
	if err != nil || strings.HasPrefix(strings.TrimSpace(toon.(string)), "{") {
		t.Fatalf("describe defaults to TOON (FR-008): %v\n%s", err, toon)
	}
	got := callJSON(t, tl, map[string]any{"level": "summary"})
	if got["ok"] != true || got["project"].(map[string]any)["name"] != "two-systems" || got["revision"] == "" {
		t.Errorf("summary: %v", got)
	}
	if n := names(got["elements"], "address"); n != "external.bank person.customer system.archive system.payments system.shop" {
		t.Errorf("summary elements = %q", n)
	}
}

func TestQueryTool(t *testing.T) {
	t.Parallel()
	tl := NewQueryTool(newService(fixture(t, "two-systems")), encoding.NewEncoder())
	tests := []struct {
		args map[string]any
		key  string
		want string
	}{
		{map[string]any{"kind": "dependents", "address": "container.orders_db"}, "elements", "component.repo container.api"},
		{map[string]any{"kind": "dependencies", "address": "container.gateway"}, "elements", "container.ledger external.bank"},
		{map[string]any{"kind": "path", "address": "person.customer", "to": "external.bank"}, "path", "person.customer container.web container.api container.gateway"},
	}
	for _, tt := range tests {
		got := callJSON(t, tl, tt.args)
		if got["ok"] != true || names(got[tt.key], map[string]string{"elements": "address", "path": "from"}[tt.key]) != tt.want {
			t.Errorf("%v = %v, want %s", tt.args, got[tt.key], tt.want)
		}
	}
	missing := callJSON(t, tl, map[string]any{"kind": "dependents", "address": "container.orders"})
	e := missing["error"].(map[string]any)
	if missing["ok"] != false || e["reason"] != "not_found" || !strings.Contains(names(e["suggestions"], ""), "container.orders_db") {
		t.Errorf("not_found: %v", missing)
	}
	if _, err := tl.Call(t.Context(), map[string]any{"kind": 3.0}); err == nil {
		t.Error("a wrongly typed argument is a protocol error")
	}
}

func TestValidateToolReportsPositions(t *testing.T) {
	t.Parallel()
	root := fixture(t, "two-systems")
	main := filepath.Join(root, "main.loko.hcl")
	src, _ := os.ReadFile(main)
	broken := strings.Replace(string(src), "target      = container.api", "target      = container.apii", 1)
	if err := os.WriteFile(main, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	svc, enc := newService(root), encoding.NewEncoder()
	got := callJSON(t, NewValidateTool(svc, enc), map[string]any{})
	diags := got["diagnostics"].([]any)
	if got["ok"] != false || len(diags) == 0 {
		t.Fatalf("validate on a broken project: %v", got)
	}
	r := diags[0].(map[string]any)["range"].(map[string]any)
	if r["file"] != "main.loko.hcl" || r["startLine"].(float64) == 0 {
		t.Errorf("diagnostic range: %v", r)
	}
	desc := callJSON(t, NewDescribeTool(svc, enc), map[string]any{})
	if desc["ok"] != false || desc["diagnostics"] == nil || desc["elements"] != nil {
		t.Errorf("describe on a broken project returns diagnostics and no data (FR-007): %v", desc)
	}
}
