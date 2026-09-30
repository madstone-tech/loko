package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// visibleResolver lifts an endpoint to its nearest ancestor-or-self in set.
func visibleResolver(ir *arch.IR, set ...arch.Address) resolveFunc {
	in := map[arch.Address]bool{}
	for _, a := range set {
		in[a] = true
	}
	return func(a arch.Address) []string {
		for cur := a; cur != ""; {
			if in[cur] {
				return []string{viewmodel.NodeIDFor(string(cur))}
			}
			e, ok := ir.Element(cur)
			if !ok {
				return nil
			}
			cur = e.Parent
		}
		return nil
	}
}

type edgeSummary struct {
	ID, Label, Tech string
	Rels            []string
	Crossing        bool
}

func summarise(es []viewmodel.Edge) []edgeSummary {
	var out []edgeSummary
	for _, e := range es {
		out = append(out, edgeSummary{e.ID, e.Label, e.Technology, e.Relationships, e.Crossing})
	}
	return out
}

func TestLiftEdges(t *testing.T) {
	t.Parallel()
	ir := sampleIR()
	id := func(a arch.Address) string { return viewmodel.NodeIDFor(string(a)) }

	tests := []struct {
		name        string
		visible     []arch.Address
		want        []edgeSummary
		wantOutside bool
	}{
		{
			// Rule 1 (merge onto a pair), rule 2 (lifted self dropped, authored
			// self kept).
			name:    "top level",
			visible: []arch.Address{aP, aS1, aS2},
			want: []edgeSummary{
				{id(aP) + "--" + id(aS1), "uses", "HTTPS", []string{"person.p.uses.uses"}, false},
				{id(aS1) + "--" + id(aS1) + "--self", "self", "", []string{"component.k1.uses.self"}, false},
				{id(aS1) + "--" + id(aS2), "2 relationships", "gRPC",
					[]string{"component.k1.uses.also", "container.c1.uses.calls"}, false},
			},
		},
		{
			// Rule 3: one end hidden becomes a crossing edge to the single
			// outside node, merged per (inside element, direction).
			name:    "one system's containers",
			visible: []arch.Address{aC1, aC2},
			want: []edgeSummary{
				{id(aC1) + "--" + id(aC1) + "--self", "self", "", []string{"component.k1.uses.self"}, false},
				{id(aC1) + "--" + id(aC2), "x", "", []string{"container.c1.uses.x"}, false},
				{id(aC1) + "--outside", "2 relationships", "gRPC",
					[]string{"component.k1.uses.also", "container.c1.uses.calls"}, true},
				{"outside--" + id(aC1), "uses", "HTTPS", []string{"person.p.uses.uses"}, true},
			},
			wantOutside: true,
		},
		{
			// Rule 4: neither end visible.
			name:    "nothing visible",
			visible: nil,
		},
		{
			// A single visible system: every edge that leaves it crosses.
			name:    "single visible system",
			visible: []arch.Address{aS1},
			want: []edgeSummary{
				{"outside--" + id(aS1), "uses", "HTTPS", []string{"person.p.uses.uses"}, true},
				{id(aS1) + "--outside", "2 relationships", "gRPC",
					[]string{"component.k1.uses.also", "container.c1.uses.calls"}, true},
				{id(aS1) + "--" + id(aS1) + "--self", "self", "", []string{"component.k1.uses.self"}, false},
			},
			wantOutside: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			edges, outside := liftEdges(ir.Relationships, visibleResolver(ir, tt.visible...))
			if got := summarise(edges); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("liftEdges =\n%+v\nwant\n%+v", got, tt.want)
			}
			if outside != tt.wantOutside {
				t.Errorf("outside = %v, want %v", outside, tt.wantOutside)
			}
			for _, e := range edges {
				if e.Style.Dashed != e.Crossing {
					t.Errorf("edge %s: crossing edges and only crossing edges are dashed", e.ID)
				}
			}
		})
	}
}

func TestLiftEdgesMixedTechnology(t *testing.T) {
	t.Parallel()
	ir := irOf(
		[]arch.Element{el(arch.KindSystem, "a", ""), el(arch.KindSystem, "b", "")},
		[]arch.Relationship{
			rl("system.a", "one", "system.b", "one", "HTTP"),
			rl("system.a", "two", "system.b", "two", "gRPC"),
		},
	)
	edges, _ := liftEdges(ir.Relationships, visibleResolver(ir, "system.a", "system.b"))
	if len(edges) != 1 || edges[0].Technology != "" || edges[0].Label != "2 relationships" {
		t.Fatalf("liftEdges = %+v", edges)
	}
}
