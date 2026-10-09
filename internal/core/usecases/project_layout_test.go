package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestViewLayouts: the project's layout is every view's default, a view's own
// layout wins, and dagre projects as "" so unchanged projects stay identical.
func TestViewLayouts(t *testing.T) {
	t.Parallel()
	views := func(project string, decl ...arch.View) map[string]string {
		ir := declaredIR(decl...)
		ir.Project.Layout = project
		got, _ := ResolveViews(ir, Provenance{})
		out := map[string]string{}
		for _, v := range got {
			out[string(v.ID)] = v.Layout
		}
		return out
	}
	shop := []arch.Address{"system.shop"}
	tests := []struct {
		name, project string
		decl          []arch.View
		want          map[string]string
	}{
		{"unset", "", []arch.View{{Address: "view.v", Name: "v", Include: shop}},
			map[string]string{"landscape": "", "system-shop": "", "v": ""}},
		{"project elk", "elk", []arch.View{{Address: "view.v", Name: "v", Include: shop}},
			map[string]string{"landscape": "elk", "system-shop": "elk", "v": "elk"}},
		{"view overrides project", "elk", []arch.View{{Address: "view.v", Name: "v", Include: shop, Layout: "dagre"}},
			map[string]string{"landscape": "elk", "v": ""}},
		{"view elk alone", "", []arch.View{{Address: "view.v", Name: "v", Include: shop, Layout: "elk"}},
			map[string]string{"landscape": "", "v": "elk"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := views(tt.project, tt.decl...)
			for id, want := range tt.want {
				if got[id] != want {
					t.Errorf("%s: layout %q, want %q (all: %v)", id, got[id], want, got)
				}
			}
		})
	}
}

func TestValidateLayout(t *testing.T) {
	t.Parallel()
	m := renderModel()
	m.Views[0].Layout, m.Views[0].LayoutRange = "neato", at(10)
	m.Project.Layout, m.Project.LayoutRange = "elk", at(1)
	d := ValidateRenderAttributes(m)
	if len(d) != 1 || d[0].Code != arch.CodeInvalidAttributeValue || d[0].Range.StartLine != 10 {
		t.Fatalf("view layout: %v", d)
	}
	m.Views[0].Layout, m.Project.Layout = "elk", "spring"
	d = ValidateRenderAttributes(m)
	if len(d) != 1 || d[0].Range.StartLine != 1 {
		t.Fatalf("project layout: %v", d)
	}
}
