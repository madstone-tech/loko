package d2

import (
	"bytes"
	"strconv"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// sequenceFlow is the root object holding a sequence diagram.
const sequenceFlow = "loko-flow"

// emitSequence draws a dynamic view as a D2 sequence diagram (feature 019,
// R7): one lifeline per element in the order it first takes part, then one
// arrow per step in step order, which is what D2 lays out top to bottom.
// Boundaries are not drawn, and direction and layout do not apply: the
// sequence layout is D2's own.
func emitSequence(m vm.ViewModel) []byte {
	var b bytes.Buffer
	b.WriteString("# " + vm.NoticeText(m.Sources) + "\n")
	actors := lifelines(m)
	writeClasses(&b, actors)
	if m.View.TitleDeclared {
		writeViewTitle(&b, m.View.Title)
	}
	b.WriteString("\n" + sequenceFlow + ": \"\" {\n  shape: sequence_diagram\n")
	for _, n := range actors {
		writeNode(&b, n, nil, 1)
	}
	for _, s := range m.Steps {
		e := vm.Edge{Label: sequenceLabel(s), Technology: s.Technology}
		b.WriteString("  " + quote(s.FromNode) + " -> " + quote(s.ToNode) + ": " + quote(baseEdgeLabel(e)))
		if s.Reply {
			b.WriteString(" {\n    style.stroke-dash: 4\n  }")
		}
		b.WriteString("\n")
	}
	b.WriteString("}\n")
	return b.Bytes()
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
