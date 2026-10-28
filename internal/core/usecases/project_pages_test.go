package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func pagesIR() *arch.IR {
	s := el(arch.KindSystem, "s", "")
	s.Docs = "docs/s.md"
	lone := el(arch.KindSystem, "lone", "")
	lone.Docs = "docs/missing.md"
	c := el(arch.KindContainer, "c", "system.s", "pci")
	c.Description, c.Technology, c.Owner = "the container", "Go", "team"
	return irOf(
		[]arch.Element{
			el(arch.KindPerson, "p", ""),
			s, lone, c,
			el(arch.KindContainer, "b", "system.s"),
			el(arch.KindComponent, "k", "container.c"),
			el(arch.KindExternal, "x", ""),
		},
		[]arch.Relationship{
			rl("container.c", "z", "external.x", "calls", "HTTPS"),
			rl("container.c", "a", "container.b", "reads", ""),
			rl("person.p", "c", "container.c", "uses", ""),
			rl("system.lone", "c", "container.c", "feeds", ""),
		},
	)
}

func pagesFor(t *testing.T, ir *arch.IR, prose ProseSet) map[string]viewmodel.ElementPage {
	t.Helper()
	proj, _, err := Project(ir, prose, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]viewmodel.ElementPage{}
	var order []string
	for _, p := range proj.Pages {
		out[p.Address] = p
		order = append(order, p.Address)
	}
	if len(order) != len(ir.Elements) {
		t.Fatalf("got %d pages for %d elements: every element gets one (FR-025)", len(order), len(ir.Elements))
	}
	for i := 1; i < len(order); i++ {
		if order[i-1] >= order[i] {
			t.Fatalf("pages not sorted by address: %v", order)
		}
	}
	return out
}

func TestProjectPages(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, pagesIR(), ProseSet{
		"docs/s.md":       {Text: "# S\n", Found: true},
		"docs/missing.md": {Found: false},
	})
	c := pages["container.c"]

	if c.Name != "c" || c.Kind != "container" || c.Description != "the container" || c.Technology != "Go" || c.Owner != "team" {
		t.Errorf("header = %+v", c)
	}
	if c.PagePath != "element/container/c.html" {
		t.Errorf("PagePath = %s", c.PagePath)
	}
	if !reflect.DeepEqual(c.Classes, viewmodel.StyleFor(viewmodel.RoleElement, "container", []string{"pci"}).Classes) {
		t.Errorf("Classes = %v", c.Classes)
	}
	if c.Parent == nil || c.Parent.Address != "system.s" || c.Parent.PagePath != "element/system/s.html" {
		t.Errorf("Parent = %+v", c.Parent)
	}
	if len(c.Children) != 1 || c.Children[0].Address != "component.k" {
		t.Errorf("Children = %+v", c.Children)
	}
	// Uses are sorted by the other element's address: container.b before
	// external.x, regardless of local names.
	var uses []string
	for _, r := range c.Uses {
		uses = append(uses, r.Other.Address+":"+r.Description)
	}
	if want := []string{"container.b:reads", "external.x:calls"}; !reflect.DeepEqual(uses, want) {
		t.Errorf("Uses = %v, want %v", uses, want)
	}
	var usedBy []string
	for _, r := range c.UsedBy {
		usedBy = append(usedBy, r.Other.Address)
	}
	if want := []string{"person.p", "system.lone"}; !reflect.DeepEqual(usedBy, want) {
		t.Errorf("UsedBy = %v, want %v", usedBy, want)
	}
	if c.Uses[1].Technology != "HTTPS" || c.Uses[1].Relationship != "container.c.uses.z" {
		t.Errorf("row = %+v", c.Uses[1])
	}

	if s := pages["system.s"]; s.Prose != "# S\n" || s.ProseMissing || s.Parent != nil {
		t.Errorf("system.s = %+v", s)
	}
	if lone := pages["system.lone"]; !lone.ProseMissing || lone.ProsePath != "docs/missing.md" || lone.Prose != "" {
		t.Errorf("a missing prose file is noted, not fatal (FR-026): %+v", lone)
	}
	if p := pages["person.p"]; p.ProseMissing {
		t.Error("an element that declares no docs is not 'missing'")
	}
}

// TestProjectPagesDiagram pins which view each page embeds: its own when one
// exists, else its parent's, else the landscape — never one not produced.
func TestProjectPagesDiagram(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, pagesIR(), nil)
	for addr, want := range map[string]viewmodel.ViewID{
		"system.s":    "system-s",
		"container.c": "container-c",
		"container.b": "system-s",    // no components: its parent's view
		"component.k": "container-c", // its parent's view
		"system.lone": "landscape",   // no containers, no parent
		"person.p":    "landscape",
		"external.x":  "landscape",
	} {
		if got := pages[addr].Diagram; got != want {
			t.Errorf("%s embeds %q, want %q", addr, got, want)
		}
	}
}

// TestContextDiagram is feature 021 research R8: In context shows the
// element among its siblings, in its parent's view; a top-level element in the
// landscape. Diagram, the page's own view, is unchanged.
func TestContextDiagram(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, contextIR(), nil)
	for addr, want := range map[string]viewmodel.ViewID{
		"component.router": "container-api",
		"container.api":    "system-s",
		"system.s":         "landscape",
		"person.p":         "landscape",
	} {
		if got := pages[addr].ContextDiagram; got != want {
			t.Errorf("%s ContextDiagram = %q, want %q", addr, got, want)
		}
	}
	if got := pages["container.api"].Diagram; got != "container-api" {
		t.Errorf("container.api Diagram = %q, want its own view", got)
	}
}

// TestProseInherited is spec FR-009: a component whose docs is its parent's
// file links to the parent instead of repeating it.
func TestProseInherited(t *testing.T) {
	t.Parallel()
	ir := contextIR()
	for i := range ir.Elements {
		switch ir.Elements[i].Address {
		case "container.api", "component.router":
			ir.Elements[i].Docs = "docs/api.md"
		case "component.auth":
			ir.Elements[i].Docs = "docs/auth.md"
		}
	}
	pages := pagesFor(t, ir, ProseSet{
		"docs/api.md":  {Text: "# API", Found: true},
		"docs/auth.md": {Text: "# Auth", Found: true},
	})
	router := pages["component.router"]
	// Prose stays in the projection, so the markdown output is unchanged
	// (FR-043); the HTML page shows a link to the parent instead.
	if !router.ProseInherited || router.Prose != "# API" || router.Parent == nil {
		t.Errorf("router: inherited %v prose %q parent %v; want inherited, prose kept, parent set",
			router.ProseInherited, router.Prose, router.Parent)
	}
	if auth := pages["component.auth"]; auth.ProseInherited || auth.Prose != "# Auth" {
		t.Errorf("auth has its own prose: inherited %v prose %q", auth.ProseInherited, auth.Prose)
	}
	if api := pages["container.api"]; api.ProseInherited || api.Prose != "# API" {
		t.Errorf("api owns the file: inherited %v prose %q", api.ProseInherited, api.Prose)
	}
}

func TestElementSource(t *testing.T) {
	t.Parallel()
	ir := contextIR()
	for i := range ir.Elements {
		if ir.Elements[i].Address == "component.cache" {
			ir.Elements[i].Range = arch.SourceRange{File: "svc/api.loko.hcl", StartLine: 42}
		}
	}
	if got := pagesFor(t, ir, nil)["component.cache"].Source; got != (viewmodel.SourceRef{File: "svc/api.loko.hcl", Line: 42}) {
		t.Errorf("source = %+v", got)
	}
}

func TestChildrenCarryDetail(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, contextIR(), nil)
	s := pages["system.s"]
	if len(s.Children) != 2 || s.Children[0].Technology != "Go" || s.Children[0].Description != "the API" {
		t.Errorf("system children = %+v, want technology and description", s.Children)
	}
	if p := pages["component.auth"].Parent; p == nil || p.Technology != "Go" || p.Description != "the API" {
		t.Errorf("auth parent = %+v, want technology and description", p)
	}
	for _, row := range pages["container.db"].UsedBy {
		if row.Other.Technology != "" || row.Other.Description != "" {
			t.Errorf("used-by rows stay without the other element's detail (markdown unchanged): %+v", row.Other)
		}
	}
}
