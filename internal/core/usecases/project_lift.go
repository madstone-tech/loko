package usecases

import (
	"cmp"
	"fmt"
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// resolveFunc maps a relationship endpoint to the nodes that represent it in
// the current view — its nearest visible ancestor-or-self. Logical views
// return at most one node; a deployment view returns one per instance of that
// element in the environment. No nodes means the endpoint is not visible.
type resolveFunc func(arch.Address) []string

// outsideFunc maps an endpoint that has no node in the view to the element
// its outside node stands for. The empty address is the single shared
// outside node of 1.5.
type outsideFunc func(arch.Address) arch.Address

// collapsedOutside sends every outside endpoint to the single shared node.
func collapsedOutside(arch.Address) arch.Address { return "" }

// outsideID is the node ID of the outside node that stands for addr.
func outsideID(addr arch.Address) string {
	if addr == "" {
		return viewmodel.OutsideNodeID
	}
	return viewmodel.OutsideNodeIDFor(string(addr))
}

type edgeKey struct{ source, target string }

type edgeAcc struct {
	rels  []arch.Relationship
	cross bool
}

// liftEdges applies the four projection rules of research R4 to every
// relationship, given how endpoints resolve in this view:
//
//  1. both ends visible and distinct → an edge; relationships lifting onto the
//     same pair merge into one;
//  2. both ends lift to the same node → kept as a self-loop only when the
//     authored relationship was itself a self-connection;
//  3. exactly one end visible → a crossing edge to the outside node that
//     outside picks for the other end, merged per (inside node, outside
//     node, direction) (FR-009, ADR-0022);
//  4. neither visible → not in this view.
//
// It returns the edges sorted, and the outside elements the crossing edges
// reach, in first-use order ("" for the single shared node).
func liftEdges(rels []arch.Relationship, resolve resolveFunc, outside outsideFunc) ([]viewmodel.Edge, []arch.Address) {
	acc := map[edgeKey]*edgeAcc{}
	var order []edgeKey
	var used []arch.Address
	seen := map[arch.Address]bool{}
	crossTo := func(end arch.Address) string {
		out := outside(end)
		if !seen[out] {
			seen[out], used = true, append(used, out)
		}
		return outsideID(out)
	}
	put := func(k edgeKey, r arch.Relationship, cross bool) {
		a, ok := acc[k]
		if !ok {
			a = &edgeAcc{cross: cross}
			acc[k] = a
			order = append(order, k)
		}
		a.rels = append(a.rels, r)
	}
	for _, r := range rels {
		src, tgt := r.Source, r.Target
		if r.Kind == arch.RelTrigger {
			// Declared on the invoked element, targeting its trigger; drawn
			// from the trigger. Only rendering sees the swap (016 R4).
			src, tgt = tgt, src
		}
		srcs, tgts := resolve(src), resolve(tgt)
		switch {
		case len(srcs) > 0 && len(tgts) > 0:
			for _, s := range srcs {
				for _, t := range tgts {
					if s != t || r.Source == r.Target {
						put(edgeKey{s, t}, r, false)
					}
				}
			}
		case len(srcs) > 0:
			o := crossTo(tgt)
			for _, s := range srcs {
				put(edgeKey{s, o}, r, true)
			}
		case len(tgts) > 0:
			o := crossTo(src)
			for _, t := range tgts {
				put(edgeKey{o, t}, r, true)
			}
		}
	}

	edges := make([]viewmodel.Edge, 0, len(order))
	for _, k := range order {
		edges = append(edges, mergeEdge(k, acc[k]))
	}
	viewmodel.SortEdges(edges)
	return edges, used
}

func mergeEdge(k edgeKey, a *edgeAcc) viewmodel.Edge {
	sort.Slice(a.rels, func(i, j int) bool { return a.rels[i].Address < a.rels[j].Address })
	addrs := make([]string, len(a.rels))
	tech := a.rels[0].Technology
	async := !a.cross
	var tags []string
	for i, r := range a.rels {
		addrs[i] = string(r.Address)
		if r.Technology != tech {
			tech = ""
		}
		async = async && r.Kind == arch.RelAsync
		tags = append(tags, r.Tags...)
	}
	label := a.rels[0].Description
	if len(a.rels) > 1 {
		label = fmt.Sprintf("%d relationships", len(a.rels))
	}
	return viewmodel.Edge{
		ID:            viewmodel.EdgeIDFor(k.source, k.target),
		Source:        k.source,
		Target:        k.target,
		Label:         label,
		Technology:    tech,
		Relationships: addrs,
		Crossing:      a.cross,
		Tags:          sortedUnique(tags),
		Style:         viewmodel.EdgeStyle{Dashed: a.cross, Async: async},
	}
}

// outsideNode is the single synthetic end of a view's boundary connections.
func outsideNode() viewmodel.Node {
	return viewmodel.Node{
		ID:    viewmodel.OutsideNodeID,
		Role:  viewmodel.RoleOutside,
		Label: "outside this view",
		Style: viewmodel.StyleFor(viewmodel.RoleOutside, "", nil),
	}
}

// outsideNodes builds the outside nodes the crossing edges reach: the single
// shared node for "", and one node per element otherwise (ADR-0022). An
// element's outside node keeps the outside role and style and has no
// address, so the view does not claim to depict it.
func outsideNodes(ir *arch.IR, used []arch.Address) []viewmodel.Node {
	nodes := make([]viewmodel.Node, 0, len(used))
	for _, a := range used {
		if a == "" {
			nodes = append(nodes, outsideNode())
			continue
		}
		e, _ := ir.Element(a)
		nodes = append(nodes, viewmodel.Node{
			ID:    outsideID(a),
			Role:  viewmodel.RoleOutside,
			Kind:  string(e.Kind),
			Label: cmp.Or(e.Title, e.Name, string(a)),
			Style: viewmodel.StyleFor(viewmodel.RoleOutside, "", nil),
		})
	}
	return nodes
}
