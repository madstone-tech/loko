package usecases

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// QueryRequest asks one graph question (FR-003..FR-005).
type QueryRequest struct {
	Kind       string // dependents | dependencies | path | orphans | coupling
	Address    string
	To         string
	Transitive bool
	Limit      int // coupling rows; 0 means 20
}

// QueryResult answers it. Every list is sorted (FR-009).
type QueryResult struct {
	OK       bool             `json:"ok" toon:"ok"`
	Kind     string           `json:"kind" toon:"kind"`
	Elements []ElementHit     `json:"elements,omitempty" toon:"elements,omitempty"`
	Path     []PathStep       `json:"path,omitempty" toon:"path,omitempty"`
	Found    *bool            `json:"found,omitempty" toon:"found,omitempty"`
	Coupling []CouplingRow    `json:"coupling,omitempty" toon:"coupling,omitempty"`
	Error    *ReadError       `json:"error,omitempty" toon:"error,omitempty"`
	Diags    arch.Diagnostics `json:"diagnostics,omitempty" toon:"diagnostics,omitempty"`
	Revision string           `json:"revision,omitempty" toon:"revision,omitempty"`
}

// ElementHit is one element in a result. Distance is 1 for a direct
// relationship and grows along a transitive closure; 0 for orphans.
type ElementHit struct {
	Address  string `json:"address" toon:"address"`
	Kind     string `json:"kind" toon:"kind"`
	Name     string `json:"name" toon:"name"`
	Distance int    `json:"distance" toon:"distance"`
}

// PathStep is one relationship along a path.
type PathStep struct {
	From         string `json:"from" toon:"from"`
	To           string `json:"to" toon:"to"`
	Relationship string `json:"relationship" toon:"relationship"`
}

// CouplingRow is one element's distinct fan-in and fan-out.
type CouplingRow struct {
	Address string `json:"address" toon:"address"`
	FanIn   int    `json:"fanIn" toon:"fanIn"`
	FanOut  int    `json:"fanOut" toon:"fanOut"`
}

// ReadError is a read that could not be answered as asked.
type ReadError struct {
	Reason      string   `json:"reason" toon:"reason"`
	Detail      string   `json:"detail" toon:"detail"`
	Suggestions []string `json:"suggestions,omitempty" toon:"suggestions,omitempty"`
}

// Query compiles the project and answers req. A project that does not compile
// returns its diagnostics instead of an answer (FR-007).
func Query(ctx context.Context, deps AuthoringDeps, req QueryRequest) (*QueryResult, error) {
	ir, res, err := compileIR(ctx, deps)
	if err != nil {
		return nil, err
	}
	out := &QueryResult{Kind: req.Kind, Diags: res.Diags.SortedForOutput()}
	if ir != nil {
		out = QueryIR(ir, req)
	}
	if out.Revision, err = currentRevision(ctx, deps); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryIR answers req on a compiled IR. It is pure, and shared by the MCP tool
// and `loko query` through Query.
func QueryIR(ir *arch.IR, req QueryRequest) *QueryResult {
	if e := checkQuery(ir, req); e != nil {
		return &QueryResult{Kind: req.Kind, Error: e}
	}
	g := newQueryGraph(ir)
	out := &QueryResult{OK: true, Kind: req.Kind}
	switch req.Kind {
	case "dependents", "dependencies":
		out.Elements = elementHits(ir, g.closure(arch.Address(req.Address), req.Kind == "dependents", req.Transitive))
	case "path":
		found := false
		out.Path, found = g.path(arch.Address(req.Address), arch.Address(req.To))
		out.Found = &found
	case "orphans":
		conn, orphans := g.connected(ir), map[arch.Address]int{}
		for _, e := range ir.Elements {
			if !conn[e.Address] {
				orphans[e.Address] = 0
			}
		}
		out.Elements = elementHits(ir, orphans)
	case "coupling":
		out.Coupling = coupling(ir, g, cmp.Or(req.Limit, 20))
	}
	return out
}

func checkQuery(ir *arch.IR, req QueryRequest) *ReadError {
	var need []string
	switch req.Kind {
	case "dependents", "dependencies":
		need = []string{req.Address}
	case "path":
		need = []string{req.Address, req.To}
	case "orphans", "coupling":
	default:
		return &ReadError{Reason: "invalid_query", Detail: fmt.Sprintf(
			"%q is not one of dependents, dependencies, path, orphans, coupling", req.Kind)}
	}
	for _, a := range need {
		if a == "" {
			return &ReadError{Reason: "invalid_query", Detail: req.Kind + " needs an address (and `to` for a path)"}
		}
		if _, ok := ir.Element(arch.Address(a)); !ok {
			return &ReadError{Reason: "not_found", Detail: fmt.Sprintf("no element %s", a), Suggestions: suggestAddresses(ir, a)}
		}
	}
	return nil
}

func elementHits(ir *arch.IR, dist map[arch.Address]int) []ElementHit {
	out := make([]ElementHit, 0, len(dist))
	for a, d := range dist {
		e, _ := ir.Element(a)
		out = append(out, ElementHit{Address: string(a), Kind: string(e.Kind), Name: e.Name, Distance: d})
	}
	slices.SortFunc(out, func(x, y ElementHit) int {
		return cmp.Or(cmp.Compare(x.Distance, y.Distance), cmp.Compare(x.Address, y.Address))
	})
	return out
}

func coupling(ir *arch.IR, g *queryGraph, limit int) []CouplingRow {
	var rows []CouplingRow
	for _, e := range ir.Elements {
		in, out := len(g.direct(e.Address, true)), len(g.direct(e.Address, false))
		if in+out > 0 {
			rows = append(rows, CouplingRow{Address: string(e.Address), FanIn: in, FanOut: out})
		}
	}
	slices.SortFunc(rows, func(a, b CouplingRow) int {
		return cmp.Or(cmp.Compare(b.FanIn+b.FanOut, a.FanIn+a.FanOut), cmp.Compare(a.Address, b.Address))
	})
	return rows[:min(limit, len(rows))]
}

// suggestAddresses returns up to three declared addresses closest to a typo,
// nearest first (FR-030), with the compiler's budget: within a third of the
// string.
func suggestAddresses(ir *arch.IR, typo string) []string {
	type cand struct {
		addr string
		dist int
	}
	var cands []cand
	for _, e := range ir.Elements {
		if d := editDistance(typo, string(e.Address)); d <= max(len(typo)/3, 1) {
			cands = append(cands, cand{string(e.Address), d})
		}
	}
	slices.SortFunc(cands, func(a, b cand) int { return cmp.Or(cmp.Compare(a.dist, b.dist), cmp.Compare(a.addr, b.addr)) })
	// Only near misses: anything more than one edit worse than the best is noise.
	var out []string
	for _, c := range cands[:min(3, len(cands))] {
		if c.dist <= cands[0].dist+1 {
			out = append(out, c.addr)
		}
	}
	return out
}
