package markdown

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Backend is the "md" backend: an index, one document per view, and one per
// element, with prose inlined verbatim. Diagrams are embedded by path.
type Backend struct{}

// New returns the "md" backend.
func New() *Backend { return &Backend{} }

// Format implements usecases.Backend.
func (*Backend) Format() vm.Format { return vm.FormatMD }

// Render implements usecases.Backend.
func (*Backend) Render(_ context.Context, in *vm.Projection, _ usecases.RenderOptions) ([]vm.Artifact, error) {
	out := []vm.Artifact{doc(vm.IndexPage("md"), "", in.Sources, index(in))}
	for _, v := range in.Views {
		out = append(out, doc(vm.MarkdownViewPage(v.View.ID), ownerOfView(v), v.Sources, view(v)))
	}
	for _, p := range in.Pages {
		out = append(out, doc(vm.ElementPath(p.Address, "md"), p.Address, p.Sources, element(in, p)))
	}
	return out, nil
}

func doc(path, owner string, sources []string, body string) vm.Artifact {
	return vm.Artifact{
		Path:   path,
		Format: vm.FormatMD,
		Bytes:  []byte("<!-- " + vm.NoticeText(sources) + " -->\n" + body),
		Owner:  owner,
	}
}

func ownerOfView(v vm.ViewModel) string {
	if v.View.Subject != "" {
		return v.View.Subject
	}
	return "view " + string(v.View.ID)
}

func index(in *vm.Projection) string {
	const self = "md/index.md"
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", in.Project.Name)
	if in.Project.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", in.Project.Description)
	}
	if land, ok := in.View("landscape"); ok {
		fmt.Fprintf(&b, "![%s](%s)\n\n", land.View.ID, vm.Rel(self, land.DiagramPath))
	}
	b.WriteString("## Views\n\n| View | Kind |\n|---|---|\n")
	for _, v := range in.Views {
		fmt.Fprintf(&b, "| [%s](%s) | %s |\n", v.View.ID, vm.Rel(self, vm.MarkdownViewPage(v.View.ID)), v.View.Kind)
	}
	// No descriptions here: editing one element's description must not
	// change the index (FR-022, contracts/output-layout.md).
	b.WriteString("\n## Elements\n\n| Element | Kind |\n|---|---|\n")
	for _, p := range in.Pages {
		fmt.Fprintf(&b, "| [%s](%s) | %s |\n", cell(p.Name), vm.Rel(self, vm.ElementPath(p.Address, "md")), p.Kind)
	}
	return b.String()
}

func view(v vm.ViewModel) string {
	self := vm.MarkdownViewPage(v.View.ID)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n_%s view_\n\n", v.View.Title, v.View.Kind)
	fmt.Fprintf(&b, "![%s](%s)\n\n", v.View.ID, vm.Rel(self, v.DiagramPath))
	b.WriteString("| Element | Kind | Technology |\n|---|---|---|\n")
	labels := map[string]string{}
	for _, n := range v.Nodes {
		labels[n.ID] = n.Label
		if n.Role == vm.RoleOutside {
			continue
		}
		name := cell(n.Label)
		if n.Link != "" {
			name = fmt.Sprintf("[%s](%s)", name, vm.Rel(self, vm.MarkdownPage(n.Link)))
		}
		kind := n.Kind
		if kind == "" {
			kind = string(n.Role)
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", name, kind, cell(n.Technology))
	}
	var leaves []string
	for _, e := range v.Edges {
		if !e.Crossing {
			continue
		}
		if e.Target == vm.OutsideNodeID {
			leaves = append(leaves, fmt.Sprintf("- **%s** → outside: %s", labels[e.Source], e.Label))
		} else {
			leaves = append(leaves, fmt.Sprintf("- outside → **%s**: %s", labels[e.Target], e.Label))
		}
	}
	if len(leaves) > 0 {
		b.WriteString("\n## Leaves this view\n\n" + strings.Join(leaves, "\n") + "\n")
	}
	return b.String()
}

func element(in *vm.Projection, p vm.ElementPage) string {
	self := vm.ElementPath(p.Address, "md")
	link := func(r vm.LinkRef) string {
		return fmt.Sprintf("[%s](%s)", cell(cmp.Or(r.Title, r.Name)), vm.Rel(self, vm.ElementPath(r.Address, "md")))
	}
	var b strings.Builder
	if p.Title != "" {
		fmt.Fprintf(&b, "# %s\n\n`%s`\n\n%s\n\n", p.Title, p.Name, facts(p))
	} else {
		fmt.Fprintf(&b, "# %s\n\n%s\n\n", p.Name, facts(p))
	}
	if p.Parent != nil {
		fmt.Fprintf(&b, "Part of %s\n\n", link(*p.Parent))
	}
	if p.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", p.Description)
	}
	if d, ok := in.View(p.Diagram); ok {
		fmt.Fprintf(&b, "![%s](%s)\n\n", d.View.ID, vm.Rel(self, d.DiagramPath))
	}
	switch {
	case p.ProseMissing:
		fmt.Fprintf(&b, "> Prose file `%s` not found.\n\n", p.ProsePath)
	case p.Prose != "":
		b.WriteString("## Prose\n\n" + strings.TrimRight(p.Prose, "\n") + "\n\n")
	}
	writeRows(&b, "Uses", p.Uses, link)
	writeRows(&b, "Used by", p.UsedBy, link)
	if len(p.Children) > 0 {
		b.WriteString("## Contains\n\n")
		for _, c := range p.Children {
			fmt.Fprintf(&b, "- %s\n", link(c))
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func facts(p vm.ElementPage) string {
	parts := []string{"**" + p.Kind + "**"}
	if p.Shape != "" {
		parts = append(parts, p.Shape)
	}
	if p.Technology != "" {
		parts = append(parts, p.Technology)
	}
	if p.Owner != "" {
		parts = append(parts, "owned by "+p.Owner)
	}
	line := strings.Join(parts, " · ")
	if len(p.Tags) > 0 {
		line += "\n\nTags: `" + strings.Join(p.Tags, "`, `") + "`"
	}
	return line
}

func writeRows(b *strings.Builder, title string, rows []vm.RelationRow, link func(vm.LinkRef) string) {
	if len(rows) == 0 {
		return
	}
	tagged := slices.ContainsFunc(rows, func(r vm.RelationRow) bool { return len(r.Tags) > 0 })
	if tagged {
		fmt.Fprintf(b, "## %s\n\n| Element | Description | Technology | Tags |\n|---|---|---|---|\n", title)
	} else {
		fmt.Fprintf(b, "## %s\n\n| Element | Description | Technology |\n|---|---|---|\n", title)
	}
	for _, r := range rows {
		fmt.Fprintf(b, "| %s | %s | %s |", link(r.Other), cell(r.Description), cell(r.Technology))
		if tagged {
			fmt.Fprintf(b, " %s |", cell(hashTags(r.Tags)))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// cell makes text safe inside a markdown table cell.
func cell(s string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ").Replace(s)
}

func hashTags(tags []string) string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = "#" + t
	}
	return strings.Join(out, " ")
}
