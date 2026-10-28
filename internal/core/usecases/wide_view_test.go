package usecases

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func svgOf(id viewmodel.ViewID, w, h int) viewmodel.Artifact {
	return viewmodel.Artifact{Path: viewmodel.DiagramFile(id, "svg"), Format: viewmodel.FormatSVG, Width: w, Height: h}
}

// TestWideViewHints is research R10: a hint for a rendered view at least
// 3,000 px wide and 2.5 times wider than tall, unless the author chose a
// direction or layout.
func TestWideViewHints(t *testing.T) {
	t.Parallel()
	generated := viewmodel.View{ID: "system-pricing", Kind: viewmodel.KindSystemView, Subject: "system.pricing"}
	declared := viewmodel.View{ID: "flow", Kind: viewmodel.KindDeclaredView, Subject: "view.flow"}
	tests := []struct {
		name   string
		view   viewmodel.View
		w, h   int
		model  arch.SourceModel
		detail []string
	}{
		{"wide generated view", generated, 3875, 1441, arch.SourceModel{},
			[]string{"system-pricing", "3875 × 1441 px", "ratio 2.7", `layout = "elk" on the project`, "declare a view"}},
		{"wide declared view", declared, 3875, 1441, arch.SourceModel{Views: []arch.ViewDecl{{Name: "flow", Range: at(3)}}},
			[]string{"view.flow", `direction = "right"`, `layout = "elk" on the view`}},
		{"narrower than 3000 px", generated, 2999, 900, arch.SourceModel{}, nil},
		{"ratio below 2.5", generated, 3200, 1300, arch.SourceModel{}, nil},
		{"direction chosen", declared, 3875, 1441, arch.SourceModel{Views: []arch.ViewDecl{{Name: "flow", Direction: "right"}}}, nil},
		{"view layout chosen", declared, 3875, 1441, arch.SourceModel{Views: []arch.ViewDecl{{Name: "flow", Layout: "dagre"}}}, nil},
		{"project layout chosen", generated, 3875, 1441, arch.SourceModel{Project: arch.ProjectDecl{Layout: "dagre"}}, nil},
		{"no size", generated, 0, 0, arch.SourceModel{}, nil},
		{"sequence style", viewmodel.View{ID: "flow", Kind: viewmodel.KindDeclaredView, Subject: "view.flow",
			Style: viewmodel.StyleSequence}, 5000, 900, arch.SourceModel{Views: []arch.ViewDecl{{Name: "flow"}}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := WideViewHints([]viewmodel.View{tt.view}, []viewmodel.Artifact{svgOf(tt.view.ID, tt.w, tt.h)}, &tt.model)
			if tt.detail == nil {
				if len(d) != 0 {
					t.Fatalf("hints %v, want none", d)
				}
				return
			}
			if len(d) != 1 || d[0].Code != arch.CodeWideView || d[0].Severity != arch.SeverityInfo {
				t.Fatalf("hints %+v, want one wide_view information diagnostic", d)
			}
			for _, want := range tt.detail {
				if !strings.Contains(d[0].Detail, want) {
					t.Errorf("detail %q lacks %q", d[0].Detail, want)
				}
			}
		})
	}
}
