package usecases

import (
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func TestNavTree(t *testing.T) {
	t.Parallel()
	proj, _, err := Project(contextIR(), nil, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	var top []string
	for _, n := range proj.Nav {
		top = append(top, n.Ref.Address)
	}
	if want := []string{"person.p", "person.q", "system.s", "external.ext"}; !slices.Equal(top, want) {
		t.Fatalf("top level = %v, want people, systems, externals", top)
	}
	s := proj.Nav[2]
	if !s.HasChildren || len(s.Children) != 2 || s.Children[0].Ref.Address != "container.api" {
		t.Fatalf("system.s = %+v", s)
	}
	api := s.Children[0]
	var comps []string
	for _, c := range api.Children {
		comps = append(comps, c.Ref.Address)
	}
	if !slices.Equal(comps, []string{"component.auth", "component.cache", "component.router"}) || !api.HasChildren {
		t.Errorf("api children = %v", comps)
	}
	if api.Ref.Technology != "" || api.Ref.Description != "" {
		t.Errorf("tree refs stay small: %+v", api.Ref)
	}
	if db := s.Children[1]; db.HasChildren || db.Children != nil {
		t.Errorf("db has no children: %+v", db)
	}
}

func TestOverview(t *testing.T) {
	t.Parallel()
	ir := contextIR()
	for i := range ir.Elements {
		if ir.Elements[i].Address == "container.db" {
			ir.Elements[i].Shape = "database"
		}
	}
	for i := range ir.Relationships {
		if ir.Relationships[i].LocalName == "reads" {
			ir.Relationships[i].Kind = "async"
		}
	}
	proj, _, err := Project(ir, nil, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	want := viewmodel.Overview{Systems: 1, Externals: 1, Containers: 2, DataStores: 1, Relationships: 5, Async: 1}
	if proj.Overview != want {
		t.Errorf("overview = %+v, want %+v", proj.Overview, want)
	}
	empty, _, err := Project(irOf([]arch.Element{el(arch.KindPerson, "p", "")}, nil), nil, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	if empty.Overview != (viewmodel.Overview{}) {
		t.Errorf("an empty model counts zero: %+v", empty.Overview)
	}
}

func TestViewCountsAndMode(t *testing.T) {
	t.Parallel()
	proj, _, err := Project(contextIR(), nil, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := proj.View("container-api")
	if !ok {
		t.Fatal("no container-api view")
	}
	// auth, cache, router inside api, plus db and person p outside it.
	elements := 0
	for _, n := range v.Nodes {
		if n.Role == viewmodel.RoleElement || n.Role == viewmodel.RoleInstance {
			elements++
		}
	}
	if v.ElementCount != elements || v.ElementCount == 0 {
		t.Errorf("ElementCount = %d, want %d element nodes", v.ElementCount, elements)
	}
	rels := 0
	for _, e := range v.Edges {
		if !e.Crossing {
			rels += len(e.Relationships)
		}
	}
	if v.RelationshipCount != rels || rels == 0 {
		t.Errorf("RelationshipCount = %d, want %d", v.RelationshipCount, rels)
	}
	if v.View.Mode() != "automatic" {
		t.Errorf("container view mode = %q", v.View.Mode())
	}
}
