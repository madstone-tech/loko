package usecases

import (
	"cmp"

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
//
// The shadowing and empty-view diagnostics live in DeclaredViewDiagnostics,
// which runs at compile; ResolveViews applies the same rules silently.
func ResolveViews(ir *arch.IR, prov Provenance) []viewmodel.View {
	views := derivedViews(ir)
	views = addDeclaredViews(ir, prov, views)
	viewmodel.SortViews(views)
	return views
}

func derivedViews(ir *arch.IR) []viewmodel.View {
	if len(ir.Elements) == 0 {
		return nil
	}
	views := []viewmodel.View{{ID: "landscape", Kind: viewmodel.KindLandscapeView, Title: ir.Project.Name, Direction: arch.DirRight}}
	for _, e := range ir.Elements {
		kind, childKind := viewKindFor(e.Kind)
		if kind == "" || !hasChildOfKind(ir, e.Address, childKind) {
			continue
		}
		views = append(views, viewmodel.View{
			ID:        viewmodel.ViewID(string(e.Kind) + "-" + e.Name),
			Kind:      kind,
			Title:     e.Name,
			Subject:   string(e.Address),
			Direction: arch.DirDown,
		})
	}
	for _, env := range ir.Environments {
		if len(env.Instances) == 0 {
			continue
		}
		views = append(views, viewmodel.View{
			ID:        viewmodel.ViewID("deployment-" + env.Name),
			Kind:      viewmodel.KindDeploymentView,
			Title:     env.Name,
			Subject:   string(env.Address),
			Direction: arch.DirRight,
		})
	}
	for i := range views {
		views[i].Layout = layoutFor(ir.Project.Layout)
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

// addDeclaredViews adds each declared view. One that selects nothing is not
// produced (FR-005); one whose ID matches a derived view replaces it (FR-004).
// Neither is reported here: DeclaredViewDiagnostics emits those diagnostics
// at compile, so validate and build see them from one place.
func addDeclaredViews(ir *arch.IR, prov Provenance, views []viewmodel.View) []viewmodel.View {
	for _, d := range ir.Views {
		if len(d.Steps) == 0 && len(selectDeclared(ir, d)) == 0 {
			continue
		}
		id := viewmodel.ViewID(d.Name)
		for i, v := range views {
			if v.ID == id && v.Kind != viewmodel.KindDeclaredView {
				views = append(views[:i], views[i+1:]...)
				break
			}
		}
		views = append(views, viewmodel.View{
			ID: id, Kind: viewmodel.KindDeclaredView, Subject: string(d.Address),
			Title: cmp.Or(d.Title, d.Name), TitleDeclared: d.Title != "", Description: d.Description,
			Style:     sequenceOnly(d.Style),
			Direction: cmp.Or(d.Direction, arch.DirDown),
			Layout:    layoutFor(cmp.Or(d.Layout, ir.Project.Layout)),
			Selection: &viewmodel.Selection{
				Include: addressStrings(d.Include), Exclude: addressStrings(d.Exclude), Tags: d.Tags,
			},
		})
	}
	return views
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

// layoutFor normalises an authored layout engine for the view model, where ""
// stands for dagre so that projects which never chose an engine project as
// before.
func layoutFor(l string) string {
	if l == arch.LayoutDagre {
		return ""
	}
	return l
}

// sequenceOnly keeps a view's style only when it changes the drawing: the
// numbered-edge style is the default, so it is left empty and views without a
// style project exactly as before (feature 019).
func sequenceOnly(style string) string {
	if style == arch.StyleSequence {
		return viewmodel.StyleSequence
	}
	return ""
}
