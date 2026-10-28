package usecases

import (
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// fillContext sets an element page's place in the model (feature 021,
// data-model.md) and returns the elements it depicts beyond the element and
// its relationships, for the page's generated-file notice (research R9).
func fillContext(p *viewmodel.ElementPage, ir *arch.IR, e arch.Element, views []viewmodel.ViewModel) []arch.Address {
	var depicted []arch.Address
	for parent, ok := ir.Element(e.Parent); ok; parent, ok = ir.Element(parent.Parent) {
		p.Ancestors = append([]viewmodel.LinkRef{detailLink(parent)}, p.Ancestors...)
		depicted = append(depicted, parent.Address)
	}
	siblings := siblingsOf(ir, e)
	p.Siblings, p.Prev, p.Next = siblingRows(ir, e, siblings)
	p.Neighbours = neighbourRows(ir, e, siblings)
	for _, s := range siblings {
		depicted = append(depicted, s.Address)
	}
	for _, n := range p.Neighbours {
		depicted = append(depicted, arch.Address(n.Other.Address))
	}
	p.ContextDiagram = contextDiagramFor(ir, e, views)
	p.Source = viewmodel.SourceRef{File: e.Range.File, Line: e.Range.StartLine}
	if parent, ok := ir.Element(e.Parent); ok && e.Docs != "" && parent.Docs == e.Docs {
		p.ProseInherited = true
	}
	return depicted
}

// siblingsOf is the parent's children, or for a top-level element the other
// top-level elements of its kind, in address order and including e. Nil when
// e is alone.
func siblingsOf(ir *arch.IR, e arch.Element) []arch.Element {
	var out []arch.Element
	if e.Parent != "" {
		out = ir.Children(e.Parent)
	} else {
		for _, o := range ir.ElementsByKind(e.Kind) {
			if o.Parent == "" {
				out = append(out, o)
			}
		}
	}
	if len(out) < 2 {
		return nil
	}
	return out
}

func siblingRows(ir *arch.IR, e arch.Element, siblings []arch.Element) (rows []viewmodel.SiblingRow, prev, next *viewmodel.LinkRef) {
	calls, calledBy := map[arch.Address]bool{}, map[arch.Address]bool{}
	for _, r := range ir.OutgoingFrom(e.Address) {
		calls[r.Target] = true
	}
	for _, r := range ir.IncomingTo(e.Address) {
		calledBy[r.Source] = true
	}
	for i, s := range siblings {
		current := s.Address == e.Address
		rows = append(rows, viewmodel.SiblingRow{Ref: detailLink(s), Current: current,
			Calls: !current && calls[s.Address], CalledBy: !current && calledBy[s.Address]})
		if !current {
			continue
		}
		if i > 0 {
			ref := linkTo(siblings[i-1])
			prev = &ref
		}
		if i+1 < len(siblings) {
			ref := linkTo(siblings[i+1])
			next = &ref
		}
	}
	return rows, prev, next
}

// neighbourRows rolls e's relationships up to one row per other element, in
// address order (spec FR-013). Relationships from e to itself are left out.
func neighbourRows(ir *arch.IR, e arch.Element, siblings []arch.Element) []viewmodel.NeighbourRow {
	rels := map[arch.Address][]arch.Relationship{}
	out, in := map[arch.Address]bool{}, map[arch.Address]bool{}
	for _, r := range ir.OutgoingFrom(e.Address) {
		rels[r.Target], out[r.Target] = append(rels[r.Target], r), true
	}
	for _, r := range ir.IncomingTo(e.Address) {
		rels[r.Source], in[r.Source] = append(rels[r.Source], r), true
	}
	delete(rels, e.Address)
	others := make([]arch.Address, 0, len(rels))
	for a := range rels {
		others = append(others, a)
	}
	slices.Sort(others)
	rows := make([]viewmodel.NeighbourRow, 0, len(others))
	for _, a := range others {
		o, _ := ir.Element(a)
		row := rollUp(rels[a])
		row.Other = detailLink(o)
		row.Direction = direction(out[a], in[a])
		row.Sibling = slices.ContainsFunc(siblings, func(s arch.Element) bool { return s.Address == a })
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil
	}
	return rows
}

// rollUp merges the relationships between two elements into one row.
func rollUp(rels []arch.Relationship) viewmodel.NeighbourRow {
	slices.SortFunc(rels, func(a, b arch.Relationship) int { return strings.Compare(string(a.Address), string(b.Address)) })
	row := viewmodel.NeighbourRow{Relationships: len(rels)}
	var techs, tags []string
	for _, r := range rels {
		if row.Description == "" {
			row.Description = r.Description
		}
		if r.Technology != "" {
			techs = append(techs, r.Technology)
		}
		tags = append(tags, r.Tags...)
	}
	slices.Sort(techs)
	slices.Sort(tags)
	row.Technology = strings.Join(slices.Compact(techs), ", ")
	if tags = slices.Compact(tags); len(tags) > 0 {
		row.Tags = tags
	}
	return row
}

func direction(out, in bool) string {
	switch {
	case out && in:
		return viewmodel.DirectionBoth
	case in:
		return viewmodel.DirectionIn
	default:
		return viewmodel.DirectionOut
	}
}

// contextDiagramFor picks the view that shows e among its siblings (research
// R8): its parent's own view, the landscape for a top-level element, else the
// first view that draws e.
func contextDiagramFor(ir *arch.IR, e arch.Element, views []viewmodel.ViewModel) viewmodel.ViewID {
	want := viewmodel.ViewID("landscape")
	if parent, ok := ir.Element(e.Parent); ok {
		want = viewmodel.ViewID(string(parent.Kind) + "-" + parent.Name)
	}
	for _, v := range views {
		if v.View.ID == want {
			return want
		}
	}
	for _, v := range views {
		for _, n := range v.Nodes {
			if n.Address == string(e.Address) && n.Role == viewmodel.RoleElement {
				return v.View.ID
			}
		}
	}
	return ""
}

// detailLink is linkTo with the element's technology and description, for
// the lists that show them (feature 021).
func detailLink(e arch.Element) viewmodel.LinkRef {
	ref := linkTo(e)
	ref.Technology, ref.Description = e.Technology, e.Description
	return ref
}

func viewByID(views []viewmodel.ViewModel, id viewmodel.ViewID) (viewmodel.ViewModel, bool) {
	for _, v := range views {
		if id != "" && v.View.ID == id {
			return v, true
		}
	}
	return viewmodel.ViewModel{}, false
}

// mergeSorted returns the sorted, de-duplicated union of a and b.
func mergeSorted(a, b []string) []string {
	out := slices.Concat(a, b)
	slices.Sort(out)
	return slices.Compact(out)
}
