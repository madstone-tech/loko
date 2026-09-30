package usecases

import (
	"reflect"
	"sort"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func nodeIndex(vm viewmodel.ViewModel) map[string]viewmodel.Node {
	m := map[string]viewmodel.Node{}
	for _, n := range vm.Nodes {
		m[n.ID] = n
	}
	return m
}

func nodeIDs(vm viewmodel.ViewModel) []string {
	var out []string
	for _, n := range vm.Nodes {
		out = append(out, n.ID)
	}
	return out
}

func edgeIDs(vm viewmodel.ViewModel) []string {
	var out []string
	for _, e := range vm.Edges {
		out = append(out, e.ID)
	}
	return out
}

func nid(a arch.Address) string { return viewmodel.NodeIDFor(string(a)) }

func projectOne(t *testing.T, ir *arch.IR, id viewmodel.ViewID) viewmodel.ViewModel {
	t.Helper()
	views, _ := ResolveViews(ir, Provenance{})
	for _, v := range views {
		if v.ID == id {
			vm := ProjectView(ir, v, Provenance{})
			if err := vm.Validate(); err != nil {
				t.Fatalf("Validate(%s): %v", id, err)
			}
			return vm
		}
	}
	t.Fatalf("view %s not resolved", id)
	return viewmodel.ViewModel{}
}

func TestProjectLandscape(t *testing.T) {
	t.Parallel()
	vm := projectOne(t, sampleIR(), "landscape")
	if got, want := nodeIDs(vm), []string{nid(aP), nid(aS1), nid(aS2)}; !reflect.DeepEqual(got, want) {
		t.Errorf("nodes = %v, want %v", got, want)
	}
	for _, e := range vm.Edges {
		if e.Crossing {
			t.Errorf("landscape has a crossing edge %s: nothing is outside the landscape", e.ID)
		}
	}
	n := nodeIndex(vm)[nid(aS2)]
	if !reflect.DeepEqual(n.Style, viewmodel.StyleFor(viewmodel.RoleElement, "system", []string{"pci"})) {
		t.Errorf("style = %+v", n.Style)
	}
	if n.Link != viewmodel.ElementPath("system.s2", "html") || n.Parent != "" {
		t.Errorf("node = %+v", n)
	}
	if vm.DiagramPath != "diagrams/landscape.svg" || vm.PagePath != "view/landscape.html" {
		t.Errorf("paths = %s %s", vm.DiagramPath, vm.PagePath)
	}
}

func TestProjectSystemView(t *testing.T) {
	t.Parallel()
	vm := projectOne(t, sampleIR(), "system-s1")
	idx := nodeIndex(vm)

	// Subject boundary, its containers inside, neighbours lifted to their
	// top-level ancestor and drawn outside the boundary.
	want := []string{nid(aC1), nid(aC2), nid(aP), nid(aS1), nid(aS2)}
	if got := nodeIDs(vm); !reflect.DeepEqual(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}
	if idx[nid(aS1)].Role != viewmodel.RoleSubject || idx[nid(aS1)].Style.Shape != viewmodel.ShapeBoundary {
		t.Errorf("subject = %+v", idx[nid(aS1)])
	}
	for _, c := range []arch.Address{aC1, aC2} {
		if idx[nid(c)].Parent != nid(aS1) {
			t.Errorf("%s parent = %q, want the subject", c, idx[nid(c)].Parent)
		}
	}
	for _, n := range []arch.Address{aP, aS2} {
		if idx[nid(n)].Parent != "" || idx[nid(n)].Role != viewmodel.RoleElement {
			t.Errorf("neighbour %s = %+v", n, idx[nid(n)])
		}
	}
	wantEdges := []string{
		nid(aC1) + "--" + nid(aC1) + "--self",
		nid(aC1) + "--" + nid(aC2),
		nid(aC1) + "--" + nid(aS2),
		nid(aP) + "--" + nid(aC1),
	}
	if got := edgeIDs(vm); !reflect.DeepEqual(got, wantEdges) {
		t.Errorf("edges = %v, want %v", got, wantEdges)
	}
	if !reflect.DeepEqual(vm.Sources, []string{"component.loko.hcl", "container.loko.hcl", "person.loko.hcl", "system.loko.hcl"}) {
		// Sources come from provenance; with none supplied they fall back to
		// the ranges recorded on IR elements.
		t.Errorf("sources = %v", vm.Sources)
	}
}

func TestProjectContainerView(t *testing.T) {
	t.Parallel()
	ir := irOf(
		[]arch.Element{
			el(arch.KindSystem, "s", ""),
			el(arch.KindSystem, "other", ""),
			el(arch.KindContainer, "c", "system.s"),
			el(arch.KindContainer, "sib", "system.s"),
			el(arch.KindContainer, "far", "system.other"),
			el(arch.KindComponent, "k", "container.c"),
			el(arch.KindComponent, "k2", "container.c"),
			el(arch.KindComponent, "sibk", "container.sib"),
		},
		[]arch.Relationship{
			rl("component.k", "a", "component.k2", "a", ""),
			rl("component.k", "b", "component.sibk", "b", ""),  // same system → sibling container
			rl("component.k2", "c", "container.far", "c", ""),  // elsewhere → top-level
			rl("container.sib", "d", "container.far", "d", ""), // not touching c: absent
		},
	)
	vm := projectOne(t, ir, "container-c")
	want := []string{"component__k", "component__k2", "container__c", "container__sib", "system__other"}
	if got := nodeIDs(vm); !reflect.DeepEqual(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}
	idx := nodeIndex(vm)
	if idx["component__k"].Parent != "container__c" || idx["container__c"].Role != viewmodel.RoleSubject {
		t.Errorf("nesting wrong: %+v / %+v", idx["component__k"], idx["container__c"])
	}
	wantEdges := []string{"component__k--component__k2", "component__k--container__sib", "component__k2--system__other"}
	if got := edgeIDs(vm); !reflect.DeepEqual(got, wantEdges) {
		t.Errorf("edges = %v, want %v", got, wantEdges)
	}
}

func TestProjectViewIsPure(t *testing.T) {
	t.Parallel()
	ir := sampleIR()
	views, _ := ResolveViews(ir, Provenance{})
	for _, v := range views {
		a, b := ProjectView(ir, v, Provenance{}), ProjectView(ir, v, Provenance{})
		if !reflect.DeepEqual(a, b) {
			t.Errorf("ProjectView(%s) is not deterministic (FR-010)", v.ID)
		}
		if !sort.SliceIsSorted(a.Nodes, func(i, j int) bool { return a.Nodes[i].ID < a.Nodes[j].ID }) {
			t.Errorf("%s nodes not sorted", v.ID)
		}
	}
}
