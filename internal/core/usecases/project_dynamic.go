package usecases

import (
	"cmp"
	"fmt"
	"strconv"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// projectDynamic draws a dynamic view (feature 019, R6): each element taking
// part once, a container or component inside its system's boundary, and one
// numbered edge per step. Edges are not merged as liftEdges merges them: two
// steps between the same elements are two edges. The steps, in order, are
// returned as rows for the page and the sequence diagram.
func projectDynamic(ir *arch.IR, v arch.View) ([]viewmodel.Node, []viewmodel.Edge, []viewmodel.StepRow) {
	nodes, ids := dynamicNodes(ir, participants(v.Steps))
	edges := make([]viewmodel.Edge, 0, len(v.Steps))
	rows := make([]viewmodel.StepRow, 0, len(v.Steps))
	for _, s := range v.Steps {
		rel, _ := ir.Relationship(s.Relationship)
		text := cmp.Or(s.Description, rel.Description)
		e := viewmodel.Edge{
			ID:            fmt.Sprintf("step__%04d__%s", s.Number, s.Name),
			Source:        ids[s.From],
			Target:        ids[s.To],
			Label:         stepLabel(s.Number, text),
			Technology:    rel.Technology,
			Relationships: []string{},
			Style:         viewmodel.EdgeStyle{Dashed: s.Reply, Async: !s.Reply && rel.Kind == arch.RelAsync},
		}
		if s.Relationship != "" {
			e.Relationships = []string{string(s.Relationship)}
		}
		edges = append(edges, e)
		rows = append(rows, viewmodel.StepRow{
			Number: s.Number, Name: s.Name, From: stepLink(ir, s.From), To: stepLink(ir, s.To),
			FromNode: ids[s.From], ToNode: ids[s.To], Text: text, Technology: rel.Technology, Reply: s.Reply,
		})
	}
	viewmodel.SortEdges(edges)
	return nodes, edges, rows
}

// participants lists the steps' elements once each, in first-appearance order.
func participants(steps []arch.Step) []arch.Address {
	seen := map[arch.Address]bool{}
	var out []arch.Address
	for _, s := range steps {
		for _, a := range []arch.Address{s.From, s.To} {
			if a != "" && !seen[a] {
				seen[a] = true
				out = append(out, a)
			}
		}
	}
	return out
}

// dynamicNodes places each participant under its nearest participating
// ancestor, else under its system, drawn as a boundary when the system takes
// no step (clarification Q5). No other ancestor is drawn.
func dynamicNodes(ir *arch.IR, parts []arch.Address) ([]viewmodel.Node, map[arch.Address]string) {
	ids := map[arch.Address]string{}
	for _, a := range parts {
		ids[a] = viewmodel.NodeIDFor(string(a))
	}
	var nodes []viewmodel.Node
	boundaries := map[arch.Address]bool{}
	for _, a := range parts {
		e, _ := ir.Element(a)
		parent := nearestVisibleAncestor(ir, e, ids)
		if parent == "" && (e.Kind == arch.KindContainer || e.Kind == arch.KindComponent) {
			sys := systemOf(ir, e)
			if !boundaries[sys.Address] {
				boundaries[sys.Address] = true
				nodes = append(nodes, elementNode(sys, viewmodel.RoleSubject, ""))
			}
			parent = viewmodel.NodeIDFor(string(sys.Address))
		}
		nodes = append(nodes, elementNode(e, viewmodel.RoleElement, parent))
	}
	return nodes, ids
}

// systemOf returns the system a container or component belongs to.
func systemOf(ir *arch.IR, e arch.Element) arch.Element {
	for e.Kind != arch.KindSystem && e.Parent != "" {
		e, _ = ir.Element(e.Parent)
	}
	return e
}

// stepLink is the page link of a step's element.
func stepLink(ir *arch.IR, addr arch.Address) viewmodel.LinkRef {
	e, _ := ir.Element(addr)
	return linkTo(e)
}

// stepLabel is "<n>. <text>", or the number alone when the step has no text.
func stepLabel(n int, text string) string {
	if text == "" {
		return strconv.Itoa(n)
	}
	return strconv.Itoa(n) + ". " + text
}
