package benchmarks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// buildD2BenchProject creates a temporary loko project with n components,
// each having a .d2 file with a single relationship arrow.
func buildD2BenchProject(tb testing.TB, n int) (*entities.Project, []*entities.System) {
	tb.Helper()
	root := tb.TempDir()

	project := &entities.Project{
		Name: "bench-project",
		Path: root,
	}

	// One system with one container holding all n components.
	system := &entities.System{
		ID:   "bench-system",
		Name: "Bench System",
	}
	container := &entities.Container{
		ID:   "bench-container",
		Name: "Bench Container",
	}

	container.Components = make(map[string]*entities.Component, n)
	for i := range n {
		compID := fmt.Sprintf("component-%03d", i)
		compPath := filepath.Join(root, "src", "bench-system", "bench-container", compID)
		if err := os.MkdirAll(compPath, 0o755); err != nil {
			tb.Fatalf("mkdir %s: %v", compPath, err)
		}

		// Write a .d2 file: this component depends on the next one (circular ok for bench)
		nextID := fmt.Sprintf("component-%03d", (i+1)%n)
		d2Content := fmt.Sprintf("%s -> %s: calls\n", compID, nextID)
		d2File := filepath.Join(compPath, "architecture.d2")
		if err := os.WriteFile(d2File, []byte(d2Content), 0o644); err != nil {
			tb.Fatalf("write d2 %s: %v", d2File, err)
		}

		container.Components[compID] = &entities.Component{
			ID:            compID,
			Name:          compID,
			Path:          compPath,
			Relationships: map[string]string{},
		}
	}

	system.Containers = map[string]*entities.Container{container.ID: container}

	return project, []*entities.System{system}
}

// BenchmarkD2Parsing_100Components measures throughput of parsing 100 components with
// D2 files using the worker pool (T037). Each call to d2lib.Compile runs a full DAG
// layout engine, so realistic throughput on modern hardware is 400-600 ms for 100
// files with 10 concurrent workers.
func BenchmarkD2Parsing_100Components(b *testing.B) {
	project, systems := buildD2BenchProject(b, 100)
	parser := d2.NewD2Parser()
	uc := usecases.NewBuildArchitectureGraphWithD2(parser)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := uc.Execute(ctx, project, systems)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

// TestD2Parsing_100Components_Under30s is a non-benchmark wall-clock test that
// asserts parsing 100 components completes within a CI-safe time budget.
//
// Note: the real d2lib.Compile runs a full DAG layout engine per file (~4-5 ms each
// on modern developer hardware, up to ~150-200ms on GitHub Actions shared runners).
// 100 files × 10 concurrent workers → realistically 400-600ms locally, but up to
// ~20s on constrained CI runners. The 30s budget provides a generous safety margin
// for all CI environments while still catching unbounded hangs or regressions.
func TestD2Parsing_100Components_Under30s(t *testing.T) {
	project, systems := buildD2BenchProject(t, 100)
	parser := d2.NewD2Parser()
	uc := usecases.NewBuildArchitectureGraphWithD2(parser)
	ctx := context.Background()

	start := time.Now()
	_, err := uc.Execute(ctx, project, systems)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if elapsed > 30*time.Second {
		t.Errorf("D2 parsing 100 components took %v, want <30s", elapsed)
	}
	t.Logf("D2 parsing 100 components: %v", elapsed)
}
