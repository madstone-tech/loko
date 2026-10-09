package usecases

import (
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
