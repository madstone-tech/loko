package usecases

import (
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// selectDeclared resolves a declared view's selection (research R4):
//
//	(include ∪ its descendants) ∪ (elements carrying any listed tag)
//	  − (exclude ∪ its descendants)
//
// With neither include nor tags the base set is every element. The result is
// sorted by address.
func selectDeclared(ir *arch.IR, v arch.View) []arch.Address {
	all := len(v.Include) == 0 && len(v.Tags) == 0
	tags := map[string]bool{}
	for _, t := range v.Tags {
		tags[t] = true
	}
	var out []arch.Address
	for _, e := range ir.Elements {
		if underAny(ir, e.Address, v.Exclude) {
			continue
		}
		if all || underAny(ir, e.Address, v.Include) || hasAnyTag(e, tags) {
			out = append(out, e.Address)
		}
	}
	return out
}

func underAny(ir *arch.IR, addr arch.Address, roots []arch.Address) bool {
	for _, r := range roots {
		if within(ir, addr, r) {
			return true
		}
	}
	return false
}

func hasAnyTag(e arch.Element, tags map[string]bool) bool {
	for _, t := range e.Tags {
		if tags[t] {
			return true
		}
	}
	return false
}

func declaredSource(ir *arch.IR, subject string) (arch.View, bool) {
	for _, v := range ir.Views {
		if string(v.Address) == subject {
			return v, true
		}
	}
	return arch.View{}, false
}

// projectDeclared nests the selected elements under their nearest selected
// ancestor. An endpoint inside an excluded subtree resolves to nothing, so a
// connection that ended on an excluded element reaches the view boundary
// instead of lifting onto a visible ancestor (US2/AC2).
func projectDeclared(ir *arch.IR, v viewmodel.View) ([]viewmodel.Node, []viewmodel.Edge, bool) {
	src, _ := declaredSource(ir, v.Subject)
	selected := selectDeclared(ir, src)
	visible := make(map[arch.Address]string, len(selected))
	for _, a := range selected {
		visible[a] = viewmodel.NodeIDFor(string(a))
	}
	nodes := make([]viewmodel.Node, 0, len(selected))
	for _, a := range selected {
		e, _ := ir.Element(a)
		nodes = append(nodes, elementNode(e, viewmodel.RoleElement, nearestVisibleAncestor(ir, e, visible)))
	}
	resolve := func(addr arch.Address) []string {
		if underAny(ir, addr, src.Exclude) {
			return nil
		}
		return ancestorResolver(ir, visible)(addr)
	}
	edges, outside := liftEdges(ir.Relationships, resolve)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes, edges, outside
}

func nearestVisibleAncestor(ir *arch.IR, e arch.Element, visible map[arch.Address]string) string {
	for cur := e.Parent; cur != ""; {
		if id, ok := visible[cur]; ok {
			return id
		}
		p, ok := ir.Element(cur)
		if !ok {
			return ""
		}
		cur = p.Parent
	}
	return ""
}
