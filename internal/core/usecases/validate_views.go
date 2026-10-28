package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// DeclaredViewDiagnostics reports a declared view that shadows an automatic
// one, or one that selects nothing (FR-004, FR-005).
//
// It runs as part of compilation, so `validate` and every other consumer of
// the compiler surface the finding before a build — where an agent authoring
// a view previously learned about a collision only from the rendered output.
// The automatic-view set is derivedViews over the built IR, exactly the set
// ResolveViews replaces from, so the two can never disagree.
//
// ResolveViews keeps the replacement behaviour (a shadowing view still
// replaces the derived one, an empty one is still not produced) but no longer
// reports: the diagnostic is emitted once, here, and flows to validate, build
// and serve through CompileArchitecture's result.
func DeclaredViewDiagnostics(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	ir := BuildIR(model, res)
	prov := BuildProvenance(model)

	kindOf := make(map[viewmodel.ViewID]string, len(ir.Views))
	for _, v := range derivedViews(ir) {
		kindOf[v.ID] = string(v.Kind)
	}

	var diags arch.Diagnostics
	for _, v := range ir.Views {
		rng, _ := prov.RangeOf(v.Address)

		// A view with steps is a dynamic flow; it takes no selection
		// (feature 019) and is never "empty" for this rule.
		if len(v.Steps) == 0 && len(selectDeclared(ir, v)) == 0 {
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning, Code: arch.CodeViewEmpty, Address: v.Address, Range: rng,
				Summary: "View selects nothing",
				Detail:  fmt.Sprintf("%s resolves to no elements, so no diagram is produced for it.", v.Address),
			})
			continue
		}
		if kind := kindOf[viewmodel.ViewID(v.Name)]; kind != "" {
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning, Code: arch.CodeViewShadowed, Address: v.Address, Range: rng,
				Summary: "View replaces an automatic view",
				Detail: fmt.Sprintf("%s has the same name as the automatic %s view, which is not produced. "+
					"Rename the view to keep both.", v.Address, kind),
			})
		}
	}
	return diags
}
