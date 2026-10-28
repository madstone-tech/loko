package usecases

import (
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// contextIR is feature 021's page-context fixture:
//
//	person p, person q ──▶ container api
//	system s ⊃ container api ⊃ component auth, cache, router; container db
//	router ──▶ auth (twice), auth ──▶ router, router ──▶ db; external ext
func contextIR() *arch.IR {
	api := el(arch.KindContainer, "api", "system.s")
	api.Technology, api.Description = "Go", "the API"
	auth := el(arch.KindComponent, "auth", "container.api")
	auth.Technology = "Go"
	r1 := rl("component.router", "r1", "component.auth", "", "gRPC")
	r2 := rl("component.router", "r2", "component.auth", "", "HTTP")
	r2.Tags = []string{"hot"}
	back := rl("component.auth", "back", "component.router", "callbacks", "gRPC")
	back.Tags = []string{"cold"}
	return irOf(
		[]arch.Element{
			el(arch.KindPerson, "p", ""),
			el(arch.KindPerson, "q", ""),
			el(arch.KindSystem, "s", ""),
			el(arch.KindExternal, "ext", ""),
			api,
			el(arch.KindContainer, "db", "system.s"),
			auth,
			el(arch.KindComponent, "cache", "container.api"),
			el(arch.KindComponent, "router", "container.api"),
		},
		[]arch.Relationship{
			r1, r2, back,
			rl("component.router", "reads", "container.db", "reads", "SQL"),
			rl("person.p", "uses", "container.api", "uses", "HTTPS"),
		},
	)
}

func addresses(refs []viewmodel.LinkRef) []string {
	var out []string
	for _, r := range refs {
		out = append(out, r.Address)
	}
	return out
}

func TestPageSiblings(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, contextIR(), nil)
	router := pages["component.router"]
	var got []string
	for _, s := range router.Siblings {
		got = append(got, s.Ref.Address)
	}
	if want := []string{"component.auth", "component.cache", "component.router"}; !slices.Equal(got, want) {
		t.Fatalf("router siblings = %v, want %v", got, want)
	}
	auth, cache, self := router.Siblings[0], router.Siblings[1], router.Siblings[2]
	if !auth.Calls || !auth.CalledBy || auth.Current {
		t.Errorf("auth = %+v, want calls and called by", auth)
	}
	if cache.Calls || cache.CalledBy || cache.Current {
		t.Errorf("cache = %+v, want unrelated", cache)
	}
	if !self.Current || self.Calls || self.CalledBy {
		t.Errorf("router itself = %+v, want only Current", self)
	}
	if auth.Ref.Technology != "Go" {
		t.Errorf("sibling refs carry technology: %+v", auth.Ref)
	}

	p := pages["person.p"].Siblings
	if len(p) != 2 || p[0].Ref.Address != "person.p" || !p[0].Current || p[1].Ref.Address != "person.q" {
		t.Errorf("person p siblings = %+v, want the other top-level people", p)
	}
	if s := pages["system.s"].Siblings; s != nil {
		t.Errorf("system s has no other system, so no siblings: %+v", s)
	}
}

func TestPageNeighbours(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, contextIR(), nil)
	n := pages["component.router"].Neighbours
	if len(n) != 2 {
		t.Fatalf("router neighbours = %+v, want auth and db", n)
	}
	auth, db := n[0], n[1]
	if auth.Other.Address != "component.auth" || auth.Direction != viewmodel.DirectionBoth ||
		auth.Relationships != 3 || !auth.Sibling {
		t.Errorf("auth row = %+v", auth)
	}
	if auth.Technology != "HTTP, gRPC" || auth.Description != "callbacks" {
		t.Errorf("auth technology %q description %q; want sorted unique join and the first description by relationship address",
			auth.Technology, auth.Description)
	}
	if !slices.Equal(auth.Tags, []string{"cold", "hot"}) {
		t.Errorf("auth tags = %v, want the sorted union", auth.Tags)
	}
	if db.Other.Address != "container.db" || db.Direction != viewmodel.DirectionOut || db.Sibling || db.Relationships != 1 {
		t.Errorf("db row = %+v", db)
	}
	if in := pages["container.api"].Neighbours; len(in) != 1 || in[0].Direction != viewmodel.DirectionIn {
		t.Errorf("api neighbours = %+v, want person p incoming", in)
	}
	for addr, p := range pages {
		for _, row := range p.Neighbours {
			if row.Other.Address == addr {
				t.Errorf("%s lists itself as a neighbour", addr)
			}
		}
	}
}

func TestPageAncestorsAndPager(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, contextIR(), nil)
	if got := addresses(pages["component.router"].Ancestors); !slices.Equal(got, []string{"system.s", "container.api"}) {
		t.Errorf("router ancestors = %v", got)
	}
	if a := pages["system.s"].Ancestors; a != nil {
		t.Errorf("a top-level element has no ancestors: %v", a)
	}
	auth, cache, router := pages["component.auth"], pages["component.cache"], pages["component.router"]
	if auth.Prev != nil || auth.Next == nil || auth.Next.Address != "component.cache" {
		t.Errorf("auth prev/next = %v / %v", auth.Prev, auth.Next)
	}
	if cache.Prev == nil || cache.Next == nil || cache.Prev.Address != "component.auth" || cache.Next.Address != "component.router" {
		t.Errorf("cache prev/next = %v / %v", cache.Prev, cache.Next)
	}
	if router.Next != nil || router.Prev == nil || router.Prev.Address != "component.cache" {
		t.Errorf("router prev/next = %v / %v; the pager does not wrap", router.Prev, router.Next)
	}
	if s := pages["system.s"]; s.Prev != nil || s.Next != nil {
		t.Errorf("an element without siblings has no pager: %v / %v", s.Prev, s.Next)
	}
}

func TestPageSourcesWiden(t *testing.T) {
	t.Parallel()
	pages := pagesFor(t, contextIR(), nil)
	got := pages["component.cache"].SiteSources
	for _, want := range []string{"component.loko.hcl", "container.loko.hcl", "system.loko.hcl"} {
		if !slices.Contains(got, want) {
			t.Errorf("cache site sources %v lack %s (parent, ancestors and siblings it depicts)", got, want)
		}
	}
	if s := pages["component.cache"].Sources; !slices.Equal(s, []string{"component.loko.hcl"}) {
		t.Errorf("Sources = %v; it feeds the markdown notice and must not widen (FR-043)", s)
	}
}
