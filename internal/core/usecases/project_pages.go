package usecases

import (
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// ProjectPages builds one page per logical element (FR-025): its prose, the
// diagram it embeds, what it uses and what uses it, and what it contains. An
// element whose prose file is missing still gets a page, with the absence
// recorded (FR-026).
func ProjectPages(ir *arch.IR, prose ProseSet, prov Provenance, views []viewmodel.ViewModel) []viewmodel.ElementPage {
	produced := make(map[viewmodel.ViewID]bool, len(views))
	for _, v := range views {
		produced[v.View.ID] = true
	}
	pages := make([]viewmodel.ElementPage, 0, len(ir.Elements))
	for _, e := range ir.Elements {
		pages = append(pages, elementPage(ir, e, prose, prov, produced))
	}
	return pages
}

func elementPage(ir *arch.IR, e arch.Element, prose ProseSet, prov Provenance,
	produced map[viewmodel.ViewID]bool) viewmodel.ElementPage {

	p := viewmodel.ElementPage{
		Address:     string(e.Address),
		Kind:        string(e.Kind),
		Name:        e.Name,
		Description: e.Description,
		Technology:  e.Technology,
		Owner:       e.Owner,
		Tags:        e.Tags,
		Classes:     viewmodel.StyleFor(viewmodel.RoleElement, string(e.Kind), e.Tags).Classes,
		Diagram:     diagramFor(ir, e, produced),
		PagePath:    viewmodel.ElementPath(string(e.Address), "html"),
	}
	if e.Docs != "" {
		doc := prose[e.Docs]
		p.ProsePath = e.Docs
		p.Prose = doc.Text
		p.ProseMissing = !doc.Found
	}
	if parent, ok := ir.Element(e.Parent); ok {
		ref := linkTo(parent)
		p.Parent = &ref
	}
	for _, c := range ir.Children(e.Address) {
		p.Children = append(p.Children, linkTo(c))
	}
	out, in := ir.OutgoingFrom(e.Address), ir.IncomingTo(e.Address)
	p.Uses = relationRows(ir, out, func(r arch.Relationship) arch.Address { return r.Target })
	p.UsedBy = relationRows(ir, in, func(r arch.Relationship) arch.Address { return r.Source })

	depicted := []arch.Address{e.Address}
	for _, r := range append(append([]arch.Relationship(nil), out...), in...) {
		depicted = append(depicted, r.Address)
	}
	p.Sources = sourcesFor(ir, prov, depicted)
	return p
}

// diagramFor picks the view a page embeds: the element's own view when one
// was produced, else its parent's, else the landscape.
func diagramFor(ir *arch.IR, e arch.Element, produced map[viewmodel.ViewID]bool) viewmodel.ViewID {
	own := viewmodel.ViewID(string(e.Kind) + "-" + e.Name)
	if produced[own] {
		return own
	}
	if parent, ok := ir.Element(e.Parent); ok {
		if pv := viewmodel.ViewID(string(parent.Kind) + "-" + parent.Name); produced[pv] {
			return pv
		}
	}
	if produced["landscape"] {
		return "landscape"
	}
	return ""
}

func linkTo(e arch.Element) viewmodel.LinkRef {
	return viewmodel.LinkRef{
		Address:  string(e.Address),
		Name:     e.Name,
		Kind:     string(e.Kind),
		PagePath: viewmodel.ElementPath(string(e.Address), "html"),
	}
}

// relationRows builds a uses or used-by table, sorted by the other element's
// address and then by the relationship's. Rows carry no description of the
// other element, so editing it leaves this page unchanged (FR-022).
func relationRows(ir *arch.IR, rels []arch.Relationship, other func(arch.Relationship) arch.Address) []viewmodel.RelationRow {
	rows := make([]viewmodel.RelationRow, 0, len(rels))
	for _, r := range rels {
		o, _ := ir.Element(other(r))
		rows = append(rows, viewmodel.RelationRow{
			Relationship: string(r.Address),
			Other:        linkTo(o),
			Description:  r.Description,
			Technology:   r.Technology,
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Other.Address != rows[j].Other.Address {
			return rows[i].Other.Address < rows[j].Other.Address
		}
		return rows[i].Relationship < rows[j].Relationship
	})
	if len(rows) == 0 {
		return nil
	}
	return rows
}
