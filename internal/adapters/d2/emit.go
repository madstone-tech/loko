package d2

import (
	"bytes"
	"cmp"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// Emit renders a view model as D2 source. It translates the styling the
// projection decided and decides none itself (FR-008). Output is a pure
// function of the view model: nodes and edges arrive sorted, and nothing here
// iterates a map (FR-021).
func Emit(m vm.ViewModel) []byte {
	var b bytes.Buffer
	b.WriteString("# " + vm.NoticeText(m.Sources) + "\n")
	b.WriteString("direction: " + cmp.Or(m.View.Direction, "right") + "\n")
	if m.View.Layout != "" {
		// D2's own setting, so the .d2 file lays out the same under the d2 CLI.
		b.WriteString("vars: {\n  d2-config: {\n    layout-engine: " + m.View.Layout + "\n  }\n}\n")
	}
	writeClasses(&b, m.Nodes)

	children := map[string][]vm.Node{}
	for _, n := range m.Nodes {
		children[n.Parent] = append(children[n.Parent], n)
	}
	for _, n := range children[""] {
		b.WriteString("\n")
		writeNode(&b, n, children, 0)
	}

	paths := pathIndex(m.Nodes)
	if len(m.Edges) > 0 {
		b.WriteString("\n")
	}
	for _, e := range m.Edges {
		writeEdge(&b, e, paths)
	}
	return b.Bytes()
}

// writeClasses declares every class used, so `class:` references resolve.
// The classes carry no styling of their own; they exist so d2 writes them
// onto the SVG for user CSS to target (FR-018).
func writeClasses(b *bytes.Buffer, nodes []vm.Node) {
	seen := map[string]bool{}
	var all []string
	for _, n := range nodes {
		for _, c := range n.Style.Classes {
			if !seen[c] {
				seen[c] = true
				all = append(all, c)
			}
		}
	}
	if len(all) == 0 {
		return
	}
	sortStrings(all)
	b.WriteString("classes: {\n")
	for _, c := range all {
		b.WriteString("  " + c + ": {}\n")
	}
	b.WriteString("}\n")
}

func writeNode(b *bytes.Buffer, n vm.Node, children map[string][]vm.Node, depth int) {
	pad := strings.Repeat("  ", depth)
	in := pad + "  "
	b.WriteString(pad + quote(n.ID) + ": " + quote(nodeLabel(n)) + " {\n")
	b.WriteString(in + "shape: " + d2Shape(n.Style.Shape) + "\n")
	if cs := n.Style.Classes; len(cs) > 0 {
		b.WriteString(in + "class: [" + strings.Join(cs, "; ") + "]\n")
	}
	b.WriteString(in + "style: {\n")
	fill := n.Style.Fill
	if fill == "" {
		fill = "transparent"
	} else {
		fill = quote(fill)
	}
	b.WriteString(in + "  fill: " + fill + "\n")
	if n.Style.Stroke != "" {
		b.WriteString(in + "  stroke: " + quote(n.Style.Stroke) + "\n")
	}
	if n.Style.FontColor != "" {
		b.WriteString(in + "  font-color: " + quote(n.Style.FontColor) + "\n")
	}
	if n.Style.Dashed {
		b.WriteString(in + "  stroke-dash: 4\n")
	}
	b.WriteString(in + "}\n")
	for _, c := range children[n.ID] {
		writeNode(b, c, children, depth+1)
	}
	b.WriteString(pad + "}\n")
}

func writeEdge(b *bytes.Buffer, e vm.Edge, paths map[string]string) {
	b.WriteString(paths[e.Source] + " -> " + paths[e.Target])
	if l := edgeLabel(e); l != "" {
		b.WriteString(": " + quote(l))
	}
	switch {
	case e.Style.Dashed: // crossing the view boundary
		b.WriteString(" {\n  style.stroke-dash: 4\n}")
	case e.Style.Async: // long dashes, distinct from crossing (016 R4)
		b.WriteString(" {\n  style.stroke-dash: 8\n}")
	}
	b.WriteString("\n")
}

// pathIndex maps each node ID to its quoted D2 path from the root, e.g.
// "system__shop"."container__api".
func pathIndex(nodes []vm.Node) map[string]string {
	parent := map[string]string{}
	for _, n := range nodes {
		parent[n.ID] = n.Parent
	}
	out := make(map[string]string, len(nodes))
	for _, n := range nodes {
		var segs []string
		for id := n.ID; id != ""; id = parent[id] {
			segs = append([]string{quote(id)}, segs...)
		}
		out[n.ID] = strings.Join(segs, ".")
	}
	return out
}

var kindTitles = map[string]string{
	vm.KindPerson: "Person", vm.KindSystem: "System", vm.KindContainer: "Container",
	vm.KindComponent: "Component", vm.KindExternal: "External",
}

// nodeLabel follows the C4 convention: name, then [Kind: Technology], then
// the description.
func nodeLabel(n vm.Node) string {
	if n.Role == vm.RoleOutside || n.Role == vm.RoleGroup || (n.Kind == "" && n.Role == vm.RoleSubject) {
		return n.Label
	}
	tag := kindTitles[n.Kind]
	if n.Technology != "" {
		if tag != "" {
			tag += ": "
		}
		tag += n.Technology
	}
	label := n.Label
	if n.Title != "" {
		// Titled: the title first, then the name beside the C4 tag. Plain
		// text, not a markdown label, so it renders in any SVG viewer (016 R2).
		if tag == "" {
			tag = "Element"
		}
		label = wrapTitle(n.Title) + "\n[" + tag + "] · " + n.Label
	} else if tag != "" {
		label += "\n[" + tag + "]"
	}
	if n.Description != "" && n.Role != vm.RoleSubject {
		label += "\n\n" + n.Description
	}
	return label
}

func edgeLabel(e vm.Edge) string {
	label := baseEdgeLabel(e)
	if len(e.Tags) == 0 {
		return label
	}
	tags := "#" + strings.Join(e.Tags, " #")
	if label == "" {
		return tags
	}
	return label + "\n" + tags
}

func baseEdgeLabel(e vm.Edge) string {
	switch {
	case e.Label != "" && e.Technology != "":
		return e.Label + "\n[" + e.Technology + "]"
	case e.Technology != "":
		return "[" + e.Technology + "]"
	default:
		return e.Label
	}
}

func d2Shape(s vm.Shape) string {
	switch s {
	case vm.ShapePerson:
		return "c4-person"
	case vm.ShapeOval:
		return "oval"
	case vm.ShapeDatabase:
		return "cylinder"
	case vm.ShapeQueue:
		return "queue"
	case vm.ShapeTopic:
		return "hexagon"
	case vm.ShapeFunction:
		return "step"
	case vm.ShapeBucket:
		return "stored_data"
	default:
		return "rectangle"
	}
}

// quote renders s as a D2 double-quoted string. Backslash, quote and '$' (D2
// substitution) are escaped; newlines become \n.
func quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "$", `\$`)
	return `"` + r.Replace(s) + `"`
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// titleWidth is the longest line a title is drawn with: D2 does not wrap plain
// labels, so a long title would otherwise stretch its box across the diagram.
const titleWidth = 40

// wrapTitle breaks a title at word boundaries into lines of at most
// titleWidth characters, never truncating it.
func wrapTitle(title string) string {
	var lines []string
	line := ""
	for _, w := range strings.Fields(title) {
		switch {
		case line == "":
			line = w
		case len([]rune(line))+1+len([]rune(w)) <= titleWidth:
			line += " " + w
		default:
			lines = append(lines, line)
			line = w
		}
	}
	return strings.Join(append(lines, line), "\n")
}
