package usecases

import (
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// ProjectView projects one view to its view model (FR-007): which elements
// appear, how they nest, which connections join them, and how each is styled.
// It is a pure function of its inputs (FR-010).
func ProjectView(ir *arch.IR, v viewmodel.View, prov Provenance) viewmodel.ViewModel {
	var nodes []viewmodel.Node
	var edges []viewmodel.Edge
	var outside bool
	switch v.Kind {
	case viewmodel.KindLandscapeView:
		nodes, edges, outside = projectLandscape(ir)
	case viewmodel.KindSystemView, viewmodel.KindContainerView:
		nodes, edges, outside = projectSubject(ir, arch.Address(v.Subject), v.Kind)
	case viewmodel.KindDeploymentView:
		nodes, edges, outside = projectDeployment(ir, arch.Address(v.Subject))
	case viewmodel.KindDeclaredView:
		nodes, edges, outside = projectDeclared(ir, v)
	}
	return finishView(ir, v, prov, nodes, edges, outside)
}

func finishView(ir *arch.IR, v viewmodel.View, prov Provenance,
	nodes []viewmodel.Node, edges []viewmodel.Edge, outside bool) viewmodel.ViewModel {

	if outside {
		nodes = append(nodes, outsideNode())
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	var depicted []arch.Address
	for _, n := range nodes {
		if n.Address != "" {
			depicted = append(depicted, arch.Address(n.Address))
		}
	}
	for _, e := range edges {
		for _, r := range e.Relationships {
			depicted = append(depicted, arch.Address(r))
		}
	}
	if v.Subject != "" {
		depicted = append(depicted, arch.Address(v.Subject))
	}
	return viewmodel.ViewModel{
		View:        v,
		Nodes:       nodes,
		Edges:       edges,
		Sources:     sourcesFor(ir, prov, depicted),
		DiagramPath: viewmodel.DiagramFile(v.ID, "svg"),
		PagePath:    viewmodel.ViewPage(v.ID),
	}
}

// projectLandscape shows every top-level element; relationships lift to
// their top-level ancestors, so nothing is outside.
func projectLandscape(ir *arch.IR) ([]viewmodel.Node, []viewmodel.Edge, bool) {
	visible := map[arch.Address]string{}
	var nodes []viewmodel.Node
	for _, e := range ir.Elements {
		if e.Parent == "" {
			n := elementNode(e, viewmodel.RoleElement, "")
			visible[e.Address] = n.ID
			nodes = append(nodes, n)
		}
	}
	edges, outside := liftEdges(ir.Relationships, ancestorResolver(ir, visible))
	return nodes, edges, outside
}

// projectSubject draws a system or container as a boundary with its children
// inside and every neighbour outside it (research R4). A neighbour is lifted
// to a sibling of the subject when it shares the subject's parent (a
// container view's sibling containers), and to its top-level ancestor
// otherwise.
func projectSubject(ir *arch.IR, subject arch.Address, kind viewmodel.ViewKind) ([]viewmodel.Node, []viewmodel.Edge, bool) {
	subj, _ := ir.Element(subject)
	visible := map[arch.Address]string{}
	sn := elementNode(subj, viewmodel.RoleSubject, "")
	visible[subject] = sn.ID
	nodes := []viewmodel.Node{sn}
	for _, c := range ir.Children(subject) {
		n := elementNode(c, viewmodel.RoleElement, sn.ID)
		visible[c.Address] = n.ID
		nodes = append(nodes, n)
	}

	scope := arch.Address("")
	if kind == viewmodel.KindContainerView {
		scope = subj.Parent
	}
	var touching []arch.Relationship
	for _, r := range ir.Relationships {
		inS, inT := within(ir, r.Source, subject), within(ir, r.Target, subject)
		if !inS && !inT {
			continue
		}
		touching = append(touching, r)
		for _, end := range []arch.Address{r.Source, r.Target} {
			if within(ir, end, subject) {
				continue
			}
			nb := neighbourOf(ir, end, scope)
			if _, ok := visible[nb]; !ok {
				e, _ := ir.Element(nb)
				n := elementNode(e, viewmodel.RoleElement, "")
				visible[nb] = n.ID
				nodes = append(nodes, n)
			}
		}
	}
	edges, outside := liftEdges(touching, ancestorResolver(ir, visible))
	return nodes, edges, outside
}

// within reports whether addr is root or one of its descendants.
func within(ir *arch.IR, addr, root arch.Address) bool {
	for cur := addr; cur != ""; {
		if cur == root {
			return true
		}
		e, ok := ir.Element(cur)
		if !ok {
			return false
		}
		cur = e.Parent
	}
	return false
}

// neighbourOf returns the ancestor-or-self of addr that is a child of scope,
// or its top-level ancestor when it is not under scope.
func neighbourOf(ir *arch.IR, addr, scope arch.Address) arch.Address {
	cur := addr
	for {
		e, ok := ir.Element(cur)
		if !ok || e.Parent == "" || (scope != "" && e.Parent == scope) {
			return cur
		}
		cur = e.Parent
	}
}

// ancestorResolver resolves an endpoint to its nearest visible
// ancestor-or-self.
func ancestorResolver(ir *arch.IR, visible map[arch.Address]string) resolveFunc {
	return func(addr arch.Address) []string {
		for cur := addr; cur != ""; {
			if id, ok := visible[cur]; ok {
				return []string{id}
			}
			e, ok := ir.Element(cur)
			if !ok {
				return nil
			}
			cur = e.Parent
		}
		return nil
	}
}

func elementNode(e arch.Element, role viewmodel.NodeRole, parent string) viewmodel.Node {
	return viewmodel.Node{
		ID:          viewmodel.NodeIDFor(string(e.Address)),
		Address:     string(e.Address),
		Role:        role,
		Kind:        string(e.Kind),
		Label:       e.Name,
		Technology:  e.Technology,
		Description: e.Description,
		Parent:      parent,
		Style:       viewmodel.StyleFor(role, string(e.Kind), e.Tags),
		Link:        viewmodel.ElementPath(string(e.Address), "html"),
	}
}

// sourcesFor returns the files that declared addrs. Provenance is preferred;
// an address it does not know falls back to the range the IR recorded, and a
// relationship with no range falls back to its source element's file.
func sourcesFor(ir *arch.IR, prov Provenance, addrs []arch.Address) []string {
	seen := map[string]bool{}
	var out []string
	for _, a := range addrs {
		f := fileOf(ir, prov, a)
		if f != "" && !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	sort.Strings(out)
	return out
}

func fileOf(ir *arch.IR, prov Provenance, a arch.Address) string {
	if r, ok := prov.RangeOf(a); ok && r.File != "" {
		return r.File
	}
	if e, ok := ir.Element(a); ok {
		return e.Range.File
	}
	if r, ok := ir.Relationship(a); ok {
		if r.Range.File != "" {
			return r.Range.File
		}
		return fileOf(ir, prov, r.Source)
	}
	if inst, ok := ir.Instance(a); ok {
		return fileOf(ir, prov, inst.Of)
	}
	return ""
}
