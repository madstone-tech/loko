package d2

import (
	"bytes"
	"cmp"
	"slices"
	"strconv"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// Emit renders a view model as D2 source. It translates the styling the
// projection decided and decides none itself (FR-008). Output is a pure
// function of the view model: nodes and edges arrive sorted, and nothing here
// iterates a map (FR-021).
func Emit(m vm.ViewModel) []byte {
	src, _ := emitPlanned(m)
	return src
}

// emitPlanned is Emit plus the label plan the SVG pass needs: each card's
// icon and the role of every label line (feature 020, research R2).
func emitPlanned(m vm.ViewModel) ([]byte, LabelPlan) {
	if m.View.Style == vm.StyleSequence {
		return emitSequence(m)
	}
	e := newEmitter()
	b := &e.b
	b.WriteString("# " + vm.NoticeText(m.Sources) + "\n")
	b.WriteString("direction: " + cmp.Or(m.View.Direction, "right") + "\n")
	if m.View.Layout != "" {
		// D2's own setting, so the .d2 file lays out the same under the d2 CLI.
		b.WriteString("vars: {\n  d2-config: {\n    layout-engine: " + m.View.Layout + "\n  }\n}\n")
	}
	writeClasses(b, m.Nodes)
	if m.View.TitleDeclared {
		writeViewTitle(b, m.View.Title)
	}

	children := map[string][]vm.Node{}
	for _, n := range m.Nodes {
		children[n.Parent] = append(children[n.Parent], n)
	}
	for _, n := range children[""] {
		b.WriteString("\n")
		e.writeNode(n, children, nil)
	}

	ids := idIndex(m.Nodes)
	if len(m.Edges) > 0 {
		b.WriteString("\n")
	}
	for _, edge := range m.Edges {
		e.writeEdge(edge, ids)
	}
	return b.Bytes(), e.plan
}

// emitter writes D2 source and records the label plan as it goes.
type emitter struct {
	b    bytes.Buffer
	plan LabelPlan
	// edges counts the edges written per source and target, which is the
	// index d2 gives each in its ID.
	edges map[[2]string]int
}

func newEmitter() *emitter {
	return &emitter{plan: newLabelPlan(), edges: map[[2]string]int{}}
}

// writeClasses declares every class used, so `class:` references resolve.
// The classes carry no styling of their own; they exist so d2 writes them
// onto the SVG for user CSS to target (FR-018).
func writeClasses(b *bytes.Buffer, nodes []vm.Node) {
	seen := map[string]bool{}
	var all []string
	for _, n := range nodes {
		for _, c := range nodeClasses(n) {
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

// nodeClasses are the projection's classes plus, on a card, its declared
// shape: every card is a rectangle, so the shape stays in the source as a
// class (research R12).
func nodeClasses(n vm.Node) []string {
	cs := n.Style.Classes
	if c := shapeClass(n.Style.Shape); c != "" && isCard(n) {
		cs = append(slices.Clone(cs), c)
		sortStrings(cs)
	}
	return cs
}

func shapeClass(s vm.Shape) string {
	switch s {
	case vm.ShapeDatabase, vm.ShapeQueue, vm.ShapeTopic, vm.ShapeFunction, vm.ShapeBucket:
		return "shape-" + string(s)
	}
	return ""
}

// cardRadius rounds every element card (research R13).
const cardRadius = "6"

// writeNode writes a node and its children. scope is the parent's d2 path, so
// the node's absolute ID, the plan's key, is scope plus its own ID.
func (e *emitter) writeNode(n vm.Node, children map[string][]vm.Node, scope []string) {
	b := &e.b
	pad := strings.Repeat("  ", len(scope))
	in := pad + "  "
	text, roles := joinLines(nodeLines(n))
	b.WriteString(pad + quote(n.ID) + ": " + quote(text) + " {\n")
	b.WriteString(in + "shape: " + d2Shape(n.Style.Shape) + "\n")
	if cs := nodeClasses(n); len(cs) > 0 {
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
	if isCard(n) {
		b.WriteString(in + "  border-radius: " + cardRadius + "\n")
	}
	b.WriteString(in + "}\n")
	id := append(slices.Clone(scope), n.ID)
	if isCard(n) {
		e.plan.Nodes[strings.Join(id, ".")] = NodePlan{
			Icon: iconFor(n.Style.Shape, n.Kind), Colour: n.Style.FontColor, Lines: roles, Address: n.Address}
	}
	for _, c := range children[n.ID] {
		e.writeNode(c, children, id)
	}
	b.WriteString(pad + "}\n")
}

func (e *emitter) writeEdge(edge vm.Edge, ids map[string][]string) {
	b := &e.b
	src, dst := ids[edge.Source], ids[edge.Target]
	text, roles := joinLines(edgeLines(edge))
	b.WriteString(quotedPath(src) + " -> " + quotedPath(dst))
	if text != "" {
		b.WriteString(": " + quote(text))
	}
	dash := ""
	switch {
	case edge.Style.Dashed: // crossing the view boundary
		dash = "4"
	case edge.Style.Async: // long dashes, distinct from crossing (016 R4)
		dash = "8"
	}
	b.WriteString(edgeStyle("", dash) + "\n")
	e.recordEdge(src, dst, roles)
}

// edgeStyle is an edge's style block: the neutral edge colours (FR-009) and,
// when set, its dash. pad is the edge line's own indentation.
func edgeStyle(pad, dash string) string {
	s := " {\n" + pad + "  style.stroke: " + quote(vm.EdgePalette.Stroke) + "\n" +
		pad + "  style.font-color: " + quote(vm.EdgePalette.Font) + "\n"
	if dash != "" {
		s += pad + "  style.stroke-dash: " + dash + "\n"
	}
	return s + pad + "}"
}

// recordEdge counts an edge and, when it has lines to restyle, puts its
// roles in the plan under d2's ID for it.
func (e *emitter) recordEdge(src, dst []string, roles []LineRole) {
	pair := [2]string{strings.Join(src, "."), strings.Join(dst, ".")}
	index := e.edges[pair]
	e.edges[pair] = index + 1
	if styledEdge(roles) {
		e.plan.Edges[edgeID(src, dst, index)] = roles
	}
}

// edgeID is d2's absolute ID for an edge, as d2graph.Edge.AbsID builds it:
// the common parent path, then "(<src> -> <dst>)[<index>]" relative to it.
func edgeID(src, dst []string, index int) string {
	common := ""
	for len(src) > 1 && len(dst) > 1 && strings.EqualFold(src[0], dst[0]) {
		common += src[0] + "."
		src, dst = src[1:], dst[1:]
	}
	return common + "(" + strings.Join(src, ".") + " -> " + strings.Join(dst, ".") + ")[" + strconv.Itoa(index) + "]"
}

// idIndex maps each node ID to its D2 path from the root as segments, e.g.
// [system__shop container__api].
func idIndex(nodes []vm.Node) map[string][]string {
	parent := map[string]string{}
	for _, n := range nodes {
		parent[n.ID] = n.Parent
	}
	out := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		var segs []string
		for id := n.ID; id != ""; id = parent[id] {
			segs = append([]string{id}, segs...)
		}
		out[n.ID] = segs
	}
	return out
}

// quotedPath is a D2 path as source text, e.g. "system__shop"."container__api".
func quotedPath(segs []string) string {
	q := make([]string, len(segs))
	for i, s := range segs {
		q[i] = quote(s)
	}
	return strings.Join(q, ".")
}

// d2Shape draws every card and boundary as a rectangle; the outside node is
// an oval. An element's declared shape is its icon and class instead (020).
func d2Shape(s vm.Shape) string {
	if s == vm.ShapeOval {
		return "oval"
	}
	return "rectangle"
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

// writeViewTitle draws a declared view title as a heading at the top of the
// canvas, so an SVG embedded on its own says what it shows (feature 018, R7).
// A root-level text object placed with near, which both dagre and ELK honour.
func writeViewTitle(b *bytes.Buffer, title string) {
	b.WriteString("\nloko-view-title: " + quote(wrapTitle(title)) + " {\n")
	b.WriteString("  shape: text\n  near: top-center\n")
	b.WriteString("  style: {\n    bold: true\n    font-size: 28\n  }\n}\n")
}
