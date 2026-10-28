package d2

import (
	"strconv"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// sequenceFlow is the root object holding a sequence diagram.
const sequenceFlow = "loko-flow"

// emitSequence draws a dynamic view as a D2 sequence diagram (feature 019,
// R7): one lifeline per element in the order it first takes part, then one
// arrow per step in step order, which is what D2 lays out top to bottom.
// Boundaries are not drawn, and direction and layout do not apply: the
// sequence layout is D2's own. Participants are cards like in any other view
// and step labels have edge line roles (feature 020, R9).
func emitSequence(m vm.ViewModel) ([]byte, LabelPlan) {
	e := newEmitter()
	b := &e.b
	b.WriteString("# " + vm.NoticeText(m.Sources) + "\n")
	actors := lifelines(m)
	writeClasses(b, actors)
	if m.View.TitleDeclared {
		writeViewTitle(b, m.View.Title)
	}
	b.WriteString("\n" + sequenceFlow + ": \"\" {\n  shape: sequence_diagram\n")
	scope := []string{sequenceFlow}
	for _, n := range actors {
		e.writeNode(n, nil, scope)
	}
	for _, s := range m.Steps {
		text, roles := joinLines(edgeTextLines(sequenceLabel(s), s.Technology))
		b.WriteString("  " + quote(s.FromNode) + " -> " + quote(s.ToNode) + ": " + quote(text))
		dash := ""
		if s.Reply {
			dash = "4"
		}
		b.WriteString(edgeStyle("  ", dash) + "\n")
		e.recordEdge([]string{sequenceFlow, s.FromNode}, []string{sequenceFlow, s.ToNode}, roles)
	}
	b.WriteString("}\n")
	return b.Bytes(), e.plan
}

// lifelines are the steps' elements, once each, in first-appearance order.
func lifelines(m vm.ViewModel) []vm.Node {
	nodes := map[string]vm.Node{}
	for _, n := range m.Nodes {
		nodes[n.ID] = n
	}
	seen := map[string]bool{}
	var out []vm.Node
	for _, s := range m.Steps {
		for _, id := range []string{s.FromNode, s.ToNode} {
			if !seen[id] {
				seen[id] = true
				out = append(out, nodes[id])
			}
		}
	}
	return out
}

// sequenceLabel numbers a step as the edges style does.
func sequenceLabel(s vm.StepRow) string {
	if s.Text == "" {
		return strconv.Itoa(s.Number)
	}
	return strconv.Itoa(s.Number) + ". " + s.Text
}
