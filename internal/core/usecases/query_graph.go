package usecases

import (
	"cmp"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// queryGraph answers graph questions with descendant inclusion (research R5):
// an element stands for itself and everything it contains, and results are
// reported at the address of the element on the other side of each
// relationship. It is the same lifting rule views use.
type queryGraph struct {
	parent map[arch.Address]arch.Address
	// rels are sorted by address, self-relationships dropped: they make an
	// element neither depend on itself nor be coupled to itself.
	rels []arch.Relationship
}

func newQueryGraph(ir *arch.IR) *queryGraph {
	g := &queryGraph{parent: make(map[arch.Address]arch.Address, len(ir.Elements))}
	for _, e := range ir.Elements {
		g.parent[e.Address] = e.Parent
	}
	for _, r := range ir.Relationships {
		if r.Source != r.Target {
			g.rels = append(g.rels, r)
		}
	}
	slices.SortFunc(g.rels, func(a, b arch.Relationship) int { return cmp.Compare(a.Address, b.Address) })
	return g
}

// within reports whether a is root or one of root's descendants.
func (g *queryGraph) within(a, root arch.Address) bool {
	for ; a != ""; a = g.parent[a] {
		if a == root {
			return true
		}
	}
	return false
}

// direct returns the elements one relationship away from start, outward
// (dependencies) or inward (dependents), sorted. Relationships wholly inside
// start's subtree are not counted.
func (g *queryGraph) direct(start arch.Address, dependents bool) []arch.Address {
	var out []arch.Address
	for _, r := range g.rels {
		near, far := r.Source, r.Target
		if dependents {
			near, far = far, near
		}
		if g.within(near, start) && !g.within(far, start) && !slices.Contains(out, far) {
			out = append(out, far)
		}
	}
	slices.Sort(out)
	return out
}

// closure is the breadth-first closure of direct. Each element is reported
// once, at its shortest distance; cycles terminate on the visited set.
func (g *queryGraph) closure(start arch.Address, dependents, transitive bool) map[arch.Address]int {
	dist := map[arch.Address]int{start: 0}
	frontier := []arch.Address{start}
	for d := 1; len(frontier) > 0; d++ {
		var next []arch.Address
		for _, n := range frontier {
			for _, m := range g.direct(n, dependents) {
				if _, seen := dist[m]; !seen {
					dist[m] = d
					next = append(next, m)
				}
			}
		}
		if !transitive {
			break
		}
		slices.Sort(next)
		frontier = next
	}
	delete(dist, start)
	return dist
}

// path is the shortest directed chain of relationships from a (or inside a)
// to b (or inside b). Relationships are tried in address order, so ties are
// broken deterministically.
func (g *queryGraph) path(a, b arch.Address) ([]PathStep, bool) {
	if g.within(a, b) {
		return []PathStep{}, true
	}
	type hop struct {
		prev arch.Address
		rel  arch.Relationship
	}
	via := map[arch.Address]hop{a: {}}
	for queue := []arch.Address{a}; len(queue) > 0; queue = queue[1:] {
		n := queue[0]
		for _, r := range g.rels {
			if _, seen := via[r.Target]; seen || !g.within(r.Source, n) {
				continue
			}
			via[r.Target] = hop{prev: n, rel: r}
			if g.within(r.Target, b) {
				var steps []PathStep
				for at := r.Target; at != a; at = via[at].prev {
					h := via[at].rel
					steps = append(steps, PathStep{From: string(h.Source), To: string(h.Target), Relationship: string(h.Address)})
				}
				slices.Reverse(steps)
				return steps, true
			}
			queue = append(queue, r.Target)
		}
	}
	return []PathStep{}, false
}

// connected marks every relationship endpoint and, upward, every ancestor of
// one: the connectivity rule of the orphan_element warning.
func (g *queryGraph) connected(ir *arch.IR) map[arch.Address]bool {
	out := map[arch.Address]bool{}
	for _, r := range ir.Relationships {
		for _, end := range []arch.Address{r.Source, r.Target} {
			for a := end; a != "" && !out[a]; a = g.parent[a] {
				out[a] = true
			}
		}
	}
	return out
}
