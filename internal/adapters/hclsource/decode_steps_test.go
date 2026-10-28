package hclsource

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestDecodeSteps: a view's steps reach the source model in declaration
// order, with every attribute and range (feature 019).
func TestDecodeSteps(t *testing.T) {
	t.Parallel()
	m, diags, err := New().Load(t.Context(), "testdata/golden/dynamic_views/input")
	if err != nil || len(diags) != 0 {
		t.Fatal(err, diags)
	}
	v := m.Views[0]
	if v.Name != "checkout" || v.Style != "sequence" || v.StyleRange.StartLine != 19 || len(v.Steps) != 3 {
		t.Fatalf("view: %+v", v)
	}
	var names []string
	for _, s := range v.Steps {
		names = append(names, s.Name)
	}
	if names[0] != "order" || names[1] != "ack" || names[2] != "again" {
		t.Errorf("steps out of declaration order: %v", names)
	}
	order := v.Steps[0]
	if order.From.Raw != "container.web" || order.To.Raw != "container.api" || order.Via.Raw != "container.web.uses.orders" ||
		order.Description != "Places the order" || order.Reply || order.Range.StartLine != 21 || order.From.Range.StartLine != 22 {
		t.Errorf("order: %+v", order)
	}
	if ack := v.Steps[1]; !ack.Reply || !ack.Via.IsZero() {
		t.Errorf("ack: %+v", ack)
	}
	sel := m.Views[1]
	for name, r := range map[string]arch.SourceRange{"include": sel.IncludeRange, "exclude": sel.ExcludeRange, "tags": sel.TagsRange} {
		if r.StartLine == 0 {
			t.Errorf("%s range not recorded", name)
		}
	}
}
