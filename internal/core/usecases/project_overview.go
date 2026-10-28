package usecases

import (
	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// navTree is the site's Architecture tree (feature 021, data-model NavNode):
// people, then systems with their containers and components, then externals,
// each in address order. Refs carry no technology or description, since the
// tree is on every page.
func navTree(ir *arch.IR) []viewmodel.NavNode {
	var out []viewmodel.NavNode
	for _, kind := range []arch.ElementKind{arch.KindPerson, arch.KindSystem, arch.KindExternal} {
		for _, e := range ir.ElementsByKind(kind) {
			if e.Parent == "" {
				out = append(out, navNode(ir, e))
			}
		}
	}
	return out
}

func navNode(ir *arch.IR, e arch.Element) viewmodel.NavNode {
	n := viewmodel.NavNode{Ref: linkTo(e)}
	for _, c := range ir.Children(e.Address) {
		n.Children = append(n.Children, navNode(ir, c))
	}
	n.HasChildren = len(n.Children) > 0
	return n
}

// overview counts what the index page's stat cards show. The diagnostics
// counts are set by BuildArtifacts, which alone sees them (research R12).
func overview(ir *arch.IR) viewmodel.Overview {
	var o viewmodel.Overview
	for _, e := range ir.Elements {
		switch e.Kind {
		case arch.KindSystem:
			o.Systems++
		case arch.KindExternal:
			o.Externals++
		case arch.KindContainer:
			o.Containers++
			if e.Shape == arch.ShapeDatabase || e.Shape == arch.ShapeQueue || e.Shape == arch.ShapeTopic {
				o.DataStores++
			}
		}
	}
	for _, r := range ir.Relationships {
		o.Relationships++
		if r.Kind == "async" {
			o.Async++
		}
	}
	return o
}

// viewCounts are the element and relationship counts on a view's page.
func viewCounts(nodes []viewmodel.Node, edges []viewmodel.Edge) (elements, relationships int) {
	for _, n := range nodes {
		if n.Role == viewmodel.RoleElement || n.Role == viewmodel.RoleInstance {
			elements++
		}
	}
	for _, e := range edges {
		if !e.Crossing {
			relationships += len(e.Relationships)
		}
	}
	return elements, relationships
}
