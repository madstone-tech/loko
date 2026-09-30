package usecases

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// SC-006: 1,000 elements across 200 files compiles in under 2 seconds, and
// 5,000 elements in under 10.
//
// The bounds are generous on purpose. The point is not to chase a number but
// to notice a regression in complexity class — an accidental O(n²) in
// resolution or in the containment walk would blow past them long before it
// became visible on a hand-written fixture.
//
// The model is built in memory rather than on disk so the measurement is of
// the pipeline, not of the file system. End-to-end timing including discovery
// and parsing is exercised separately via tools/genfixture.
func TestCompilePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}
	t.Parallel()

	tests := []struct {
		name     string
		elements int
		budget   time.Duration
	}{
		{"1k elements", 1000, 2 * time.Second},
		{"5k elements", 5000, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := stubSource{model: syntheticModel(tt.elements)}
			start := time.Now()
			res, err := CompileArchitecture(context.Background(), src,
				CompileRequest{BuildVersion: "1.0.0"})
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			if res.HasErrors() {
				t.Fatalf("synthetic model produced errors: %v", codes(res.Diags)[:min(5, len(res.Diags))])
			}
			if elapsed > tt.budget {
				t.Errorf("compiled %d elements in %v, budget %v", tt.elements, elapsed, tt.budget)
			}
			t.Logf("%d elements in %v (%.0f%% of budget)",
				tt.elements, elapsed.Round(time.Millisecond),
				100*float64(elapsed)/float64(tt.budget))
		})
	}
}

// TestBuildIRPerformance guards the ordering work specifically: every slice is
// sorted at construction, so a careless comparator would show up here.
func TestBuildIRPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}
	t.Parallel()

	model := syntheticModel(5000)
	res, _ := ResolveModel(model)

	start := time.Now()
	ir := BuildIR(model, res)
	elapsed := time.Since(start)

	if got := len(ir.Elements); got != 5000 {
		t.Fatalf("built %d elements, want 5000", got)
	}
	if elapsed > 2*time.Second {
		t.Errorf("BuildIR took %v for 5000 elements, budget 2s", elapsed)
	}
	t.Logf("BuildIR: 5000 elements in %v", elapsed.Round(time.Millisecond))
}

// syntheticModel builds a SourceModel with the requested element count, in the
// same shape tools/genfixture produces: roughly one system per twelve
// elements, half containers, the rest components, with each container using
// the next so edges are dense enough to be meaningful.
func syntheticModel(elements int) *arch.SourceModel {
	systems := max(elements/12, 1)
	containers := (elements - systems) / 2
	components := elements - systems - containers

	decls := make([]arch.ElementDecl, 0, elements)
	rng := func(i int) arch.SourceRange {
		return arch.SourceRange{File: fmt.Sprintf("part_%d.loko.hcl", i%200), StartLine: i, StartColumn: 1}
	}

	for i := range systems {
		decls = append(decls, arch.ElementDecl{
			Kind: arch.KindSystem, Name: fmt.Sprintf("sys_%d", i), Range: rng(i),
			Docs: "docs.md",
		})
	}
	for i := range containers {
		decls = append(decls, arch.ElementDecl{
			Kind: arch.KindContainer, Name: fmt.Sprintf("con_%d", i), Range: rng(i),
			Docs:   "docs.md",
			Parent: arch.Reference{Raw: fmt.Sprintf("system.sys_%d", i%systems), Range: rng(i)},
			Relations: []arch.RelationDecl{{
				LocalName: "next", Range: rng(i),
				Target: arch.Reference{Raw: fmt.Sprintf("container.con_%d", (i+1)%containers), Range: rng(i)},
			}},
		})
	}
	for i := range components {
		decls = append(decls, arch.ElementDecl{
			Kind: arch.KindComponent, Name: fmt.Sprintf("cmp_%d", i), Range: rng(i),
			Docs:   "docs.md",
			Parent: arch.Reference{Raw: fmt.Sprintf("container.con_%d", i%containers), Range: rng(i)},
		})
	}

	return &arch.SourceModel{
		Project:  arch.ProjectDecl{Name: "synthetic", Declared: true},
		Elements: decls,
	}
}

// Kept so `go test -bench .` can track the trend rather than only the ceiling.
func BenchmarkCompile1k(b *testing.B) {
	model := syntheticModel(1000)
	src := stubSource{model: model}
	for b.Loop() {
		if _, err := CompileArchitecture(context.Background(), src,
			CompileRequest{BuildVersion: "1.0.0"}); err != nil {
			b.Fatal(err)
		}
	}
}
