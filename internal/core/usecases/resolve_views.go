package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// ResolveViews derives the views available without configuration and adds
// the declared ones (research R3).
//
// Derived: a landscape when anything exists, a view per system that has
// containers, per container that has components, and per environment that
// has instances. A view that would contain nothing is never produced and is
// not an error (FR-002). View IDs are functions of what the view depicts, so
// their files keep the same names across runs (FR-006).
func ResolveViews(ir *arch.IR, prov Provenance) ([]viewmodel.View, arch.Diagnostics) {
	views := derivedViews(ir)
	views, diags := addDeclaredViews(ir, prov, views)
	viewmodel.SortViews(views)
	return views, diags
}

func derivedViews(ir *arch.IR) []viewmodel.View {
	if len(ir.Elements) == 0 {
		return nil
	}
	views := []viewmodel.View{{ID: "landscape", Kind: viewmodel.KindLandscapeView, Title: ir.Project.Name}}
	for _, e := range ir.Elements {
		kind, childKind := viewKindFor(e.Kind)
		if kind == "" || !hasChildOfKind(ir, e.Address, childKind) {
			continue
		}
		views = append(views, viewmodel.View{
			ID:      viewmodel.ViewID(string(e.Kind) + "-" + e.Name),
			Kind:    kind,
			Title:   e.Name,
			Subject: string(e.Address),
		})
	}
	for _, env := range ir.Environments {
		if len(env.Instances) == 0 {
			continue
		}
		views = append(views, viewmodel.View{
			ID:      viewmodel.ViewID("deployment-" + env.Name),
			Kind:    viewmodel.KindDeploymentView,
			Title:   env.Name,
			Subject: string(env.Address),
		})
	}
	return views
}

// viewKindFor names the view an element of this kind gets, and the child kind
// it must contain to get one.
func viewKindFor(k arch.ElementKind) (viewmodel.ViewKind, arch.ElementKind) {
	switch k {
	case arch.KindSystem:
		return viewmodel.KindSystemView, arch.KindContainer
	case arch.KindContainer:
		return viewmodel.KindContainerView, arch.KindComponent
	default:
		return "", ""
	}
}

func hasChildOfKind(ir *arch.IR, parent arch.Address, kind arch.ElementKind) bool {
	for _, c := range ir.Children(parent) {
		if c.Kind == kind {
			return true
		}
	}
	return false
}

// addDeclaredViews adds each declared view. One that selects nothing is
// reported and not produced (FR-005); one whose ID matches a derived view
// replaces it, and the shadowing is reported so the author is not left
// wondering where the other picture went (FR-004).
func addDeclaredViews(ir *arch.IR, prov Provenance, views []viewmodel.View) ([]viewmodel.View, arch.Diagnostics) {
	var diags arch.Diagnostics
	for _, d := range ir.Views {
		rng, _ := prov.RangeOf(d.Address)
		if len(selectDeclared(ir, d)) == 0 {
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning, Code: arch.CodeViewEmpty, Address: d.Address, Range: rng,
				Summary: "View selects nothing",
				Detail:  fmt.Sprintf("%s resolves to no elements, so no diagram is produced for it.", d.Address),
			})
			continue
		}
		id := viewmodel.ViewID(d.Name)
		for i, v := range views {
			if v.ID == id && v.Kind != viewmodel.KindDeclaredView {
				views = append(views[:i], views[i+1:]...)
				diags = append(diags, arch.Diagnostic{
					Severity: arch.SeverityWarning, Code: arch.CodeViewShadowed, Address: d.Address, Range: rng,
					Summary: "View replaces an automatic view",
					Detail: fmt.Sprintf("%s has the same name as the automatic %s view, which is not produced. "+
						"Rename the view to keep both.", d.Address, v.Kind),
				})
				break
			}
		}
		views = append(views, viewmodel.View{
			ID: id, Kind: viewmodel.KindDeclaredView, Title: d.Name, Subject: string(d.Address),
			Selection: &viewmodel.Selection{
				Include: addressStrings(d.Include), Exclude: addressStrings(d.Exclude), Tags: d.Tags,
			},
		})
	}
	return views, diags
}

func addressStrings(as []arch.Address) []string {
	if len(as) == 0 {
		return nil
	}
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = string(a)
	}
	return out
}
