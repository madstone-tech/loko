package viewmodel

import (
	"strings"
	"testing"
)

func validModel() ViewModel {
	return ViewModel{
		View: View{ID: "system-shop", Kind: KindSystemView},
		Nodes: []Node{
			{ID: "container__api", Role: RoleElement, Parent: "system__shop"},
			{ID: OutsideNodeID, Role: RoleOutside},
			{ID: "system__shop", Role: RoleSubject},
		},
		Edges: []Edge{
			{ID: "container__api--outside", Source: "container__api", Target: OutsideNodeID, Crossing: true},
		},
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mutate  func(*ViewModel)
		wantErr string
	}{
		{"valid", func(*ViewModel) {}, ""},
		{"unknown edge source", func(m *ViewModel) { m.Edges[0].Source = "nope" }, "unknown source"},
		{"unknown edge target", func(m *ViewModel) { m.Edges[0].Target = "nope" }, "unknown target"},
		{"unknown parent", func(m *ViewModel) { m.Nodes[0].Parent = "nope" }, "unknown parent"},
		{"outside without crossing edge", func(m *ViewModel) { m.Edges = nil }, "outside node without"},
		{"crossing edge not touching outside", func(m *ViewModel) {
			m.Nodes = m.Nodes[:1:1]
			m.Nodes = append(m.Nodes, Node{ID: "system__shop", Role: RoleSubject})
			m.Edges[0].Target = "system__shop"
		}, "crossing edge"},
		{"two outside nodes", func(m *ViewModel) {
			m.Nodes = append(m.Nodes, Node{ID: "zz", Role: RoleOutside})
		}, "more than one outside"},
		{"parent cycle", func(m *ViewModel) { m.Nodes[2].Parent = "container__api" }, "cycle"},
		{"nodes unsorted", func(m *ViewModel) { m.Nodes[0], m.Nodes[2] = m.Nodes[2], m.Nodes[0] }, "not sorted"},
		{"duplicate node", func(m *ViewModel) { m.Nodes[1] = m.Nodes[0] }, "duplicate node"},
		{"edges unsorted", func(m *ViewModel) {
			m.Nodes = append([]Node{{ID: "a", Role: RoleElement}}, m.Nodes...)
			m.Edges = append(m.Edges, Edge{ID: "a--container__api", Source: "a", Target: "container__api"})
		}, "not sorted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := validModel()
			tt.mutate(&m)
			err := m.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestEdgeIDFor(t *testing.T) {
	t.Parallel()
	if got := EdgeIDFor("a", "b"); got != "a--b" {
		t.Errorf("EdgeIDFor(a,b) = %q", got)
	}
	if got := EdgeIDFor("a", "a"); got != "a--a--self" {
		t.Errorf("EdgeIDFor(a,a) = %q", got)
	}
}

func TestSortEdges(t *testing.T) {
	t.Parallel()
	es := []Edge{
		{ID: EdgeIDFor("b", "a"), Source: "b", Target: "a"},
		{ID: EdgeIDFor("a", "c"), Source: "a", Target: "c"},
		{ID: EdgeIDFor("a", "b"), Source: "a", Target: "b"},
	}
	SortEdges(es)
	got := es[0].ID + " " + es[1].ID + " " + es[2].ID
	if got != "a--b a--c b--a" {
		t.Errorf("SortEdges order = %s", got)
	}
}
