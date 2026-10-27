package usecases

import (
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// The wide-view thresholds, from measurements of the example views (feature
// 018, research R10): the report's view rendered 3875 × 1441 px, and a view
// this wide cannot be read at page width.
const (
	wideViewMinWidth = 3000
	wideViewMinRatio = 2.5
)

// renderHints are the hints only a laid-out diagram can give, less the codes
// the run and the project ignore. They come after rendering, so the compile
// stage's ignore lists have not seen them.
func renderHints(proj *viewmodel.Projection, artifacts []viewmodel.Artifact, model *arch.SourceModel, ignore []string) arch.Diagnostics {
	views := make([]viewmodel.View, len(proj.Views))
	for i, v := range proj.Views {
		views[i] = v.View
	}
	hints, _ := WideViewHints(views, artifacts, model).Without(slices.Concat(ignore, model.Project.IgnoreWarnings))
	return hints
}

// WideViewHints gives one information diagnostic for each rendered view that
// is at least wideViewMinWidth wide and wideViewMinRatio times wider than
// tall, naming what to try. A view whose author chose a direction or layout,
// or a project that sets a layout, gets none: the author already chose
// (FR-013, FR-015). Sizes come from the SVG artifacts, so a build without
// SVG output gives none.
func WideViewHints(views []viewmodel.View, artifacts []viewmodel.Artifact, model *arch.SourceModel) arch.Diagnostics {
	if model.Project.Layout != "" {
		return nil
	}
	sizes := map[string]viewmodel.Artifact{}
	for _, a := range artifacts {
		if a.Format == viewmodel.FormatSVG {
			sizes[a.Path] = a
		}
	}
	var diags arch.Diagnostics
	for _, v := range views {
		a := sizes[viewmodel.DiagramFile(v.ID, "svg")]
		if a.Width < wideViewMinWidth || a.Height == 0 || float64(a.Width)/float64(a.Height) < wideViewMinRatio {
			continue
		}
		if d, ok := wideViewHint(v, a, model); ok {
			diags = append(diags, d)
		}
	}
	return diags
}

// wideViewHint words the hint: a declared view can set direction or layout
// itself; a generated one has no block to set them on.
func wideViewHint(v viewmodel.View, a viewmodel.Artifact, model *arch.SourceModel) (arch.Diagnostic, bool) {
	d := arch.Diagnostic{Severity: arch.SeverityInfo, Code: arch.CodeWideView, Summary: "View renders wide"}
	size := fmt.Sprintf("renders %d × %d px (ratio %.1f)", a.Width, a.Height, float64(a.Width)/float64(a.Height))
	if v.Kind != viewmodel.KindDeclaredView {
		d.Detail = fmt.Sprintf("View %s %s, too wide to read at page width. Set layout = \"elk\" on the project, "+
			"or declare a view of these elements.", v.ID, size)
		return d, true
	}
	i := slices.IndexFunc(model.Views, func(decl arch.ViewDecl) bool { return string(arch.NewViewAddress(decl.Name)) == v.Subject })
	if i >= 0 && (model.Views[i].Direction != "" || model.Views[i].Layout != "") {
		return d, false
	}
	if i >= 0 {
		d.Range = model.Views[i].Range
	}
	d.Address = arch.Address(v.Subject)
	d.Detail = fmt.Sprintf("%s %s, too wide to read at page width. Try direction = \"right\" or layout = \"elk\" "+
		"on the view.", v.Subject, size)
	return d, true
}
