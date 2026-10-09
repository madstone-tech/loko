package usecases

import (
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
//  3. exactly one end visible → a crossing edge to the single outside node,
//     merged per (inside node, direction) (FR-009);
//  4. neither visible → not in this view.
//
// It returns the edges sorted, and whether an outside node is needed.
func liftEdges(rels []arch.Relationship, resolve resolveFunc) ([]viewmodel.Edge, bool) {
	acc := map[edgeKey]*edgeAcc{}
	var order []edgeKey
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
		srcs, tgts := resolve(r.Source), resolve(r.Target)
		if r.Kind == arch.RelTrigger {
			// Declared on the invoked element, targeting its trigger; drawn
			// from the trigger. Only rendering sees the swap (016 R4).
			srcs, tgts = tgts, srcs
		}
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
			for _, s := range srcs {
				put(edgeKey{s, viewmodel.OutsideNodeID}, r, true)
			}
		case len(tgts) > 0:
			for _, t := range tgts {
				put(edgeKey{viewmodel.OutsideNodeID, t}, r, true)
			}
		}
	}

	edges := make([]viewmodel.Edge, 0, len(order))
	outside := false
	for _, k := range order {
		a := acc[k]
		outside = outside || a.cross
		edges = append(edges, mergeEdge(k, a))
	}
	viewmodel.SortEdges(edges)
	return edges, outside
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
