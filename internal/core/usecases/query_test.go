package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// queryIR is the query fixture:
//
//	system s1 { container c1 { component k1 }, container c2 }
//	system s2 { container db }
//	system s3 { container c4, container a, container b, container c }
//	system s4 { container c5 }            (no relationships: orphans)
//	person p, external x                  (x is an orphan)
//
//	p → c1, k1 → db (into db from inside c1), c4 → db, c1 → c2, k1 → k1 (self)
//	a → b → c → a (a cycle)
func queryIR() *arch.IR {
	s1, s2, s3, s4 := sys("s1"), sys("s2"), sys("s3"), sys("s4")
	c1 := ctr("c1")
	return irOf(
		[]arch.Element{
			el(arch.KindPerson, "p", ""), el(arch.KindExternal, "x", ""),
			el(arch.KindSystem, "s1", ""), el(arch.KindSystem, "s2", ""),
			el(arch.KindSystem, "s3", ""), el(arch.KindSystem, "s4", ""),
			el(arch.KindContainer, "c1", s1), el(arch.KindContainer, "c2", s1),
			el(arch.KindComponent, "k1", c1), el(arch.KindContainer, "db", s2),
			el(arch.KindContainer, "c4", s3), el(arch.KindContainer, "a", s3),
			el(arch.KindContainer, "b", s3), el(arch.KindContainer, "c", s3),
			el(arch.KindContainer, "c5", s4),
		},
		[]arch.Relationship{
			rl(aP, "uses", c1, "", ""),
			rl(comp("k1"), "reads", ctr("db"), "", ""),
			rl(ctr("c4"), "writes", ctr("db"), "", ""),
			rl(c1, "x", ctr("c2"), "", ""),
			rl(comp("k1"), "self", comp("k1"), "", ""),
			rl(ctr("a"), "next", ctr("b"), "", ""),
			rl(ctr("b"), "next", ctr("c"), "", ""),
			rl(ctr("c"), "next", ctr("a"), "", ""),
		},
	)
}

func sys(n string) arch.Address  { return arch.NewElementAddress(arch.KindSystem, n) }
func ctr(n string) arch.Address  { return arch.NewElementAddress(arch.KindContainer, n) }
func comp(n string) arch.Address { return arch.NewElementAddress(arch.KindComponent, n) }

func hits(r *QueryResult) string {
	var out []string
	for _, h := range r.Elements {
		out = append(out, fmt.Sprintf("%s@%d", h.Address, h.Distance))
	}
	return strings.Join(out, " ")
}

func TestQueryDirect(t *testing.T) {
	t.Parallel()
	ir := queryIR()
	tests := []struct {
		req  QueryRequest
		want string
	}{
		// A component inside another container counts, reported at its own address.
		{QueryRequest{Kind: "dependents", Address: "container.db"}, "component.k1@1 container.c4@1"},
		// Descendant inclusion on the target side: s2 stands for db.
		{QueryRequest{Kind: "dependents", Address: "system.s2"}, "component.k1@1 container.c4@1"},
		// Relationships inside the subtree (k1 → k1) are not dependencies.
		{QueryRequest{Kind: "dependencies", Address: "container.c1"}, "container.c2@1 container.db@1"},
		{QueryRequest{Kind: "dependencies", Address: "person.p"}, "container.c1@1"},
		{QueryRequest{Kind: "dependents", Address: "person.p"}, ""},
	}
	for _, tt := range tests {
		got := QueryIR(ir, tt.req)
		if !got.OK || hits(got) != tt.want {
			t.Errorf("%s(%s) = %q (ok=%v), want %q", tt.req.Kind, tt.req.Address, hits(got), got.OK, tt.want)
		}
	}
}

func TestQueryTransitiveTerminatesOnCycles(t *testing.T) {
	t.Parallel()
	got := QueryIR(queryIR(), QueryRequest{Kind: "dependencies", Address: "container.a", Transitive: true})
	if want := "container.b@1 container.c@2"; hits(got) != want {
		t.Errorf("transitive dependencies(a) = %q, want %q", hits(got), want)
	}
	got = QueryIR(queryIR(), QueryRequest{Kind: "dependents", Address: "container.c2", Transitive: true})
	if want := "container.c1@1 person.p@2"; hits(got) != want {
		t.Errorf("transitive dependents(c2) = %q, want %q", hits(got), want)
	}
}

func TestQueryPath(t *testing.T) {
	t.Parallel()
	got := QueryIR(queryIR(), QueryRequest{Kind: "path", Address: "person.p", To: "container.db"})
	var steps []string
	for _, s := range got.Path {
		steps = append(steps, fmt.Sprintf("%s>%s via %s", s.From, s.To, s.Relationship))
	}
	want := "person.p>container.c1 via person.p.uses.uses | component.k1>container.db via component.k1.uses.reads"
	if got.Found == nil || !*got.Found || strings.Join(steps, " | ") != want {
		t.Errorf("path = %q found=%v, want %q", strings.Join(steps, " | "), got.Found, want)
	}
	none := QueryIR(queryIR(), QueryRequest{Kind: "path", Address: "container.db", To: "person.p"})
	if !none.OK || none.Found == nil || *none.Found || len(none.Path) != 0 {
		t.Errorf("no path: %+v", none)
	}
}

func TestQueryOrphans(t *testing.T) {
	t.Parallel()
	got := QueryIR(queryIR(), QueryRequest{Kind: "orphans"})
	if want := "container.c5@0 external.x@0 system.s4@0"; hits(got) != want {
		t.Errorf("orphans = %q, want %q", hits(got), want)
	}
}

func TestQueryCoupling(t *testing.T) {
	t.Parallel()
	got := QueryIR(queryIR(), QueryRequest{Kind: "coupling", Limit: 7})
	var rows []string
	for _, r := range got.Coupling {
		rows = append(rows, fmt.Sprintf("%s:%d/%d", r.Address, r.FanIn, r.FanOut))
	}
	// Ranked by fan-in + fan-out, then address. c1 (with k1 inside) is used by
	// p and uses c2 and db. The a → b → c cycle gives each 1/1. Elements with
	// no coupling at all are not listed.
	want := "container.c1:1/2 container.a:1/1 container.b:1/1 container.c:1/1 container.db:2/0 system.s1:1/1 system.s2:2/0"
	if strings.Join(rows, " ") != want {
		t.Errorf("coupling = %q, want %q", strings.Join(rows, " "), want)
	}
}

func TestQueryNotFoundSuggests(t *testing.T) {
	t.Parallel()
	got := QueryIR(queryIR(), QueryRequest{Kind: "dependents", Address: "container.dbb"})
	if got.OK || got.Error == nil || got.Error.Reason != "not_found" {
		t.Fatalf("unknown address: %+v", got)
	}
	if len(got.Error.Suggestions) == 0 || got.Error.Suggestions[0] != "container.db" || len(got.Error.Suggestions) > 3 {
		t.Errorf("suggestions = %v", got.Error.Suggestions)
	}
	bad := QueryIR(queryIR(), QueryRequest{Kind: "fan-in", Address: "container.db"})
	if bad.OK || bad.Error.Reason != "invalid_query" {
		t.Errorf("unknown kind: %+v", bad.Error)
	}
	missing := QueryIR(queryIR(), QueryRequest{Kind: "path", Address: "person.p"})
	if missing.OK || missing.Error.Reason != "invalid_query" {
		t.Errorf("path without to: %+v", missing.Error)
	}
}

func TestQueryDeterministic(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"dependents", "dependencies", "orphans", "coupling"} {
		req := QueryRequest{Kind: k, Address: "container.db", Transitive: true}
		a, b := QueryIR(queryIR(), req), QueryIR(queryIR(), req)
		if fmt.Sprint(a) != fmt.Sprint(b) {
			t.Errorf("%s is not deterministic", k)
		}
	}
}

// TestKindsDoNotChangeQueries is FR-004 / SC-004: relationship kinds and tags
// affect drawing only.
func TestKindsDoNotChangeQueries(t *testing.T) {
	t.Parallel()
	plain := queryIR()
	kinded := queryIR()
	for i := range kinded.Relationships {
		kinded.Relationships[i].Kind = []string{arch.RelTrigger, arch.RelAsync, ""}[i%3]
		kinded.Relationships[i].Tags = []string{"t"}
	}
	reqs := []QueryRequest{
		{Kind: "dependents", Address: "container.db", Transitive: true},
		{Kind: "dependencies", Address: "container.c1", Transitive: true},
		{Kind: "path", Address: "person.p", To: "container.db"},
		{Kind: "orphans"}, {Kind: "coupling"},
	}
	for _, req := range reqs {
		a, _ := json.Marshal(QueryIR(plain, req))
		b, _ := json.Marshal(QueryIR(kinded, req))
		if string(a) != string(b) {
			t.Errorf("%s changed with kinds:\n%s\n%s", req.Kind, a, b)
		}
	}
}

// listingIR is queryIR with a declared view that shadows the generated
// system-s1 view and a declared view of its own, a title on c1, and an async
// relationship with a description.
func listingIR() *arch.IR {
	q := queryIR()
	for i := range q.Elements {
		if q.Elements[i].Address == ctr("c1") {
			q.Elements[i].Title = "Checkout"
		}
	}
	for i := range q.Relationships {
		if q.Relationships[i].Address == arch.NewRelationshipAddress(ctr("c4"), "writes") {
			q.Relationships[i].Kind, q.Relationships[i].Description = arch.RelAsync, "Stores rows"
		}
	}
	views := []arch.View{
		{Address: "view.flow", Name: "flow", Include: []arch.Address{ctr("c1")}},
		{Address: "view.system-s1", Name: "system-s1", Title: "Shop", Include: []arch.Address{ctr("c1"), ctr("c2")}},
	}
	return arch.NewIR(q.Project, q.Elements, q.Relationships, nil, views, nil)
}

func TestQueryListViews(t *testing.T) {
	t.Parallel()
	res := QueryIR(listingIR(), QueryRequest{Kind: "views"})
	var got []string
	for _, v := range res.Views {
		got = append(got, v.Name+":"+v.Kind+":"+v.Title)
	}
	want := "container-c1:container:c1 flow:declared:flow landscape:landscape:test system-s1:declared:Shop " +
		"system-s2:system:s2 system-s3:system:s3 system-s4:system:s4"
	if !res.OK || strings.Join(got, " ") != want {
		t.Errorf("views:\n%s\nwant:\n%s", strings.Join(got, " "), want)
	}
}

func TestQueryListElements(t *testing.T) {
	t.Parallel()
	all := QueryIR(listingIR(), QueryRequest{Kind: "elements"})
	if !all.OK || len(all.Elements) != 15 || all.Elements[0].Address != "component.k1" || all.Elements[14].Address != "system.s4" {
		t.Fatalf("elements: %+v", all.Elements)
	}
	for _, e := range all.Elements {
		if (e.Address == "container.c1") != (e.Title == "Checkout") || e.Distance != 0 {
			t.Errorf("element %+v: only c1 has a title, and listings have no distance", e)
		}
	}
	ctrs := QueryIR(listingIR(), QueryRequest{Kind: "elements", ElementKind: "container"})
	if !ctrs.OK || len(ctrs.Elements) != 8 || ctrs.Elements[0].Address != "container.a" {
		t.Errorf("containers: %+v", ctrs.Elements)
	}
	for _, e := range ctrs.Elements {
		if e.Kind != "container" {
			t.Errorf("filter let through %+v", e)
		}
	}
	bad := QueryIR(listingIR(), QueryRequest{Kind: "elements", ElementKind: "node"})
	if bad.OK || bad.Error == nil || bad.Error.Reason != "invalid_query" ||
		!strings.Contains(bad.Error.Detail, "person, system, container, component, external") {
		t.Errorf("unknown element kind: %+v", bad.Error)
	}
}

func TestQueryListRelationships(t *testing.T) {
	t.Parallel()
	res := QueryIR(listingIR(), QueryRequest{Kind: "relationships"})
	if !res.OK || len(res.Relationships) != 8 {
		t.Fatalf("relationships: %+v", res.Relationships)
	}
	first, writes := res.Relationships[0], res.Relationships[6]
	if first.Address != "component.k1.uses.reads" || first.Source != "component.k1" || first.Target != "container.db" ||
		first.Kind != "sync" || first.Description != "" {
		t.Errorf("first row: %+v", first)
	}
	if writes.Address != "container.c4.uses.writes" || writes.Kind != "async" || writes.Description != "Stores rows" {
		t.Errorf("c4 writes: %+v", writes)
	}
	for i := 1; i < len(res.Relationships); i++ {
		if res.Relationships[i-1].Address >= res.Relationships[i].Address {
			t.Errorf("not sorted by address at %d", i)
		}
	}
}

// TestQueryListingLeavesOtherKinds: the new title field and listings do not
// change the encoding of the 1.5 query kinds, and listings need no address.
func TestQueryListingLeavesOtherKinds(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"views", "elements", "relationships"} {
		if r := QueryIR(queryIR(), QueryRequest{Kind: k}); !r.OK {
			t.Errorf("%s: %+v", k, r.Error)
		}
	}
	b, _ := json.Marshal(QueryIR(listingIR(), QueryRequest{Kind: "dependents", Address: "container.c2"}))
	if strings.Contains(string(b), "title") || strings.Contains(string(b), "views") || strings.Contains(string(b), "relationships") {
		t.Errorf("an existing kind gained a field: %s", b)
	}
	unknown := QueryIR(queryIR(), QueryRequest{Kind: "nodes"})
	if unknown.Error == nil || !strings.Contains(unknown.Error.Detail, "views, elements, relationships") {
		t.Errorf("unknown kind lists the new kinds: %+v", unknown.Error)
	}
}
