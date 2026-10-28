package html

import (
	"cmp"
	"html"
	"html/template"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// treeHTML renders a page's pruned Architecture tree (spec FR-001, FR-002).
// It is built here rather than by a recursive template: the tree is on every
// page, and template calls per node made large sites slow to rebuild in
// `loko serve` (feature 021). A theme still replaces it by overriding nav.
func treeHTML(root string, nodes []vm.NavNode) template.HTML {
	var b strings.Builder
	b.WriteString(`<nav class="tree" aria-label="Architecture"><ul>`)
	for _, n := range nodes {
		writeTreeNode(&b, root, n)
	}
	b.WriteString(`</ul></nav>`)
	return template.HTML(b.String()) //nolint:gosec // every value is escaped below
}

func writeTreeNode(b *strings.Builder, root string, n vm.NavNode) {
	name := html.EscapeString(cmp.Or(n.Ref.Title, n.Ref.Name))
	b.WriteString(`<li class="tree-item`)
	if n.Current {
		b.WriteString(` active`)
	}
	b.WriteString(`"><div class="tree-row">`)
	switch {
	case n.HasChildren && n.Children != nil:
		b.WriteString(`<button type="button" class="twisty" aria-expanded="true" aria-label="Expand or collapse ` + name + `">▾</button>`)
	case n.HasChildren:
		b.WriteString(`<button type="button" class="twisty" aria-expanded="false" aria-label="Expand or collapse ` + name + `">▸</button>`)
	default:
		b.WriteString(`<span class="twisty-spacer" aria-hidden="true"></span>`)
	}
	kind := html.EscapeString(n.Ref.Kind)
	b.WriteString(`<i class="icon kind-icon-` + kind + `" data-icon="` + vm.IconFor("", n.Ref.Kind) + `" aria-hidden="true"></i>`)
	b.WriteString(`<a href="` + html.EscapeString(root+n.Ref.PagePath) + `" title="` + name + `"`)
	if n.Current {
		b.WriteString(` aria-current="page"`)
	}
	b.WriteString(`>` + name + `</a></div>`)
	if len(n.Children) > 0 {
		b.WriteString(`<ul>`)
		for _, c := range n.Children {
			writeTreeNode(b, root, c)
		}
		b.WriteString(`</ul>`)
	}
	b.WriteString(`</li>`)
}
