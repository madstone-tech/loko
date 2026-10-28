package html

import (
	"cmp"
	"fmt"
	"html/template"
	"slices"
	"strconv"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// pageData is what every template receives: only view-model values, so a
// template cannot reach the file system or the IR (contracts/theme.md).
// Fields after Nav were added by feature 021 (contracts/theme-blocks.md).
type pageData struct {
	Kind       string // "index", "view" or "element"
	Title      string
	Subtitle   string // the address name, under a display title (feature 016)
	Root       string // relative href from this page to the site root
	BodyClass  string
	Project    vm.ProjectHeader
	Projection *vm.Projection
	Element    *vm.ElementPage
	View       *vm.ViewModel
	Diagram    *vm.ViewModel
	// Nav is the rendered nav block for this page.
	Nav template.HTML

	// Tree is the Architecture tree pruned for this page (research R7).
	Tree []vm.NavNode
	// Crumbs lead from the project to this page's parent.
	Crumbs []vm.LinkRef
	// Label sits above the title, e.g. "COMPONENT · IN GATEWAY".
	Label string
	// Meta is the metadata line: technology, full id, file:line, counts.
	Meta []string
	// Lead is the description under the title.
	Lead string
	// Icon and IconKind draw the title's icon tile; empty for no tile.
	Icon, IconKind string
	// Context is the view an element page shows in context; Inline is that
	// view's SVG (element pages) or the view's own (view pages), ready to
	// embed. Inline is empty when the SVG was not rendered in this build.
	Context *vm.ViewModel
	Inline  template.HTML
	// HasD2 is set when this build writes the D2 sources a view page links.
	HasD2 bool
}

// site is what every page job shares.
type site struct {
	in       *vm.Projection
	views    map[vm.ViewID]*vm.ViewModel
	pages    map[string]*vm.ElementPage
	upstream map[string][]byte
	hasD2    bool
}

func newSite(in *vm.Projection, upstream map[string][]byte, formats []vm.Format) *site {
	s := &site{in: in, views: map[vm.ViewID]*vm.ViewModel{}, pages: map[string]*vm.ElementPage{},
		upstream: upstream, hasD2: slices.Contains(formats, vm.FormatD2)}
	for i := range in.Views {
		s.views[in.Views[i].View.ID] = &in.Views[i]
	}
	for i := range in.Pages {
		s.pages[in.Pages[i].Address] = &in.Pages[i]
	}
	return s
}

func (s *site) indexData() pageData {
	return pageData{Kind: "index", Title: s.in.Project.Name, BodyClass: "page-index",
		Diagram: s.views["landscape"], Tree: vm.NavFor(s.in.Nav, ""), Label: "OVERVIEW",
		Lead: s.in.Project.Description, HasD2: s.hasD2}
}

func (s *site) viewData(v *vm.ViewModel) pageData {
	d := pageData{Kind: "view", Title: v.View.Title, BodyClass: "page-view view-" + string(v.View.Kind), View: v,
		Tree: vm.NavFor(s.in.Nav, ""), Crumbs: s.crumbsFor(v.View.Subject, true),
		Label: strings.ToUpper(string(v.View.Kind)) + " VIEW · " + strings.ToUpper(v.View.Mode()),
		Meta:  []string{plural(v.ElementCount, "element"), plural(v.RelationshipCount, "relationship")},
		Lead:  v.View.Description, HasD2: s.hasD2}
	d.Inline = inlineSVG(s.upstream[v.DiagramPath], "view-caption", "in-this-view", "")
	return d
}

func (s *site) elementData(p *vm.ElementPage) pageData {
	d := pageData{Kind: "element", Title: cmp.Or(p.Title, p.Name), Subtitle: subtitle(p),
		BodyClass: strings.Join(append([]string{"page-element"}, p.Classes...), " "),
		Element:   p, Diagram: s.views[p.Diagram], Tree: vm.NavFor(s.in.Nav, p.Address),
		Crumbs: s.crumbsFor(p.Address, false), Label: elementLabel(p), Meta: elementMeta(p),
		Lead: p.Description, Icon: vm.IconFor(vm.Shape(p.Shape), p.Kind), IconKind: p.Kind, HasD2: s.hasD2}
	if v := s.views[p.ContextDiagram]; v != nil {
		d.Context = v
		d.Inline = inlineSVG(s.upstream[v.DiagramPath], "context-caption", "", p.Address)
	}
	return d
}

// crumbsFor is the breadcrumb up to address: the project, then its
// ancestors, then the element itself when self is set (for a view of it).
func (s *site) crumbsFor(address string, self bool) []vm.LinkRef {
	crumbs := []vm.LinkRef{{Name: s.in.Project.Name, PagePath: vm.IndexPage("html")}}
	p := s.pages[address]
	if p == nil {
		return crumbs
	}
	crumbs = append(crumbs, p.Ancestors...)
	if self {
		crumbs = append(crumbs, vm.LinkRef{Address: p.Address, Name: p.Name, Title: p.Title, Kind: p.Kind, PagePath: p.PagePath})
	}
	return crumbs
}

// elementLabel is the small label above an element's title.
func elementLabel(p *vm.ElementPage) string {
	label := strings.ToUpper(p.Kind)
	if p.Parent != nil {
		label += " · IN " + strings.ToUpper(cmp.Or(p.Parent.Title, p.Parent.Name))
	}
	return label
}

func elementMeta(p *vm.ElementPage) []string {
	var meta []string
	if p.Technology != "" {
		meta = append(meta, p.Technology)
	}
	meta = append(meta, p.Address)
	if p.Source.File != "" {
		src := p.Source.File
		if p.Source.Line > 0 {
			src += ":" + strconv.Itoa(p.Source.Line)
		}
		meta = append(meta, src)
	}
	return meta
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// systemRow is one row of the index's Systems list.
type systemRow struct {
	Ref        vm.LinkRef
	Containers int
	External   bool
}

// systemRows lists the systems and externals for the index, in page order.
func systemRows(in *vm.Projection) []systemRow {
	var rows []systemRow
	for _, p := range in.Pages {
		if p.Kind != vm.KindSystem && p.Kind != vm.KindExternal {
			continue
		}
		rows = append(rows, systemRow{
			Ref: vm.LinkRef{Address: p.Address, Name: p.Name, Title: p.Title, Kind: p.Kind, PagePath: p.PagePath,
				Technology: p.Technology, Description: p.Description},
			Containers: len(p.Children), External: p.Kind == vm.KindExternal,
		})
	}
	return rows
}
