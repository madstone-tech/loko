package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestGraphPerformanceLargeProject tests performance on a large project.
func TestGraphPerformanceLargeProject(t *testing.T) {
	// Create test project with 100 components
	project, err := entities.NewProject("large-project")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create system with 10 containers, each with 10 components
	system, err := entities.NewSystem("Backend")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}

	for i := 0; i < 10; i++ {
		container, err := entities.NewContainer(fmt.Sprintf("Container%d", i))
		if err != nil {
			t.Fatalf("failed to create container: %v", err)
		}

		for j := 0; j < 10; j++ {
			component, err := entities.NewComponent(fmt.Sprintf("Comp%d_%d", i, j))
			if err != nil {
				t.Fatalf("failed to create component: %v", err)
			}

			// Add relationships to create a dependency graph
			// Each component depends on the previous component in the container
			if j > 0 {
				prevCompID := fmt.Sprintf("comp%d_%d", i, j-1)
				component.AddRelationship(prevCompID, "sequential dependency")
			}

			// Also add cross-container dependencies
			if i > 0 {
				crossCompID := fmt.Sprintf("comp%d_%d", i-1, j)
				component.AddRelationship(crossCompID, "cross-container dependency")
			}

			container.AddComponent(component)
		}

		system.AddContainer(container)
	}

	// Build the architecture graph
	uc := usecases.NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Test GetIncomingEdges performance - should be <1ms per call
	t.Run("GetIncomingEdges performance", func(t *testing.T) {
		targetNode := entities.QualifiedNodeID("component", system.ID, "container5", "comp5_5")

		start := time.Now()
		for i := 0; i < 100; i++ {
			_ = graph.GetIncomingEdges(targetNode)
		}
		duration := time.Since(start)

		avgTime := duration / 100
		if avgTime > time.Millisecond {
			t.Errorf("GetIncomingEdges took %v per call, expected <1ms", avgTime)
		}
		t.Logf("GetIncomingEdges average time: %v", avgTime)
	})

	// Test GetChildren performance - should be <1ms per call
	t.Run("GetChildren performance", func(t *testing.T) {
		containerNode := entities.QualifiedNodeID("container", system.ID, "container5", "")

		start := time.Now()
		for i := 0; i < 100; i++ {
			_ = graph.GetChildren(containerNode)
		}
		duration := time.Since(start)

		avgTime := duration / 100
		if avgTime > time.Millisecond {
			t.Errorf("GetChildren took %v per call, expected <1ms", avgTime)
		}
		t.Logf("GetChildren average time: %v", avgTime)
	})

	// Test AnalyzeDependencies performance - should be <2s for 100 components
	t.Run("AnalyzeDependencies performance", func(t *testing.T) {
		validateUC := usecases.NewValidateArchitecture()

		start := time.Now()
		report := validateUC.Execute(graph, []*entities.System{system})
		duration := time.Since(start)

		if duration > 2*time.Second {
			t.Errorf("AnalyzeDependencies took %v, expected <2s", duration)
		}
		t.Logf("AnalyzeDependencies time: %v", duration)
		t.Logf("Validation report: %d issues found", len(report.Issues))
	})
}

// TestGraphPerformanceDeepHierarchy tests performance on deep hierarchy.
func TestGraphPerformanceDeepHierarchy(t *testing.T) {
	// Create project with 5-level deep hierarchy
	project, err := entities.NewProject("deep-hierarchy")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create system with multiple containers
	system, err := entities.NewSystem("System")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}

	// Create 5 containers, each with 20 components
	for i := 0; i < 5; i++ {
		container, err := entities.NewContainer(fmt.Sprintf("Container%d", i))
		if err != nil {
			t.Fatalf("failed to create container: %v", err)
		}

		for j := 0; j < 20; j++ {
			component, err := entities.NewComponent(fmt.Sprintf("Comp%d_%d", i, j))
			if err != nil {
				t.Fatalf("failed to create component: %v", err)
			}

			container.AddComponent(component)
		}

		system.AddContainer(container)
	}

	// Build graph
	uc := usecases.NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Test GetDescendants on root - should be <100ms
	t.Run("GetDescendants performance", func(t *testing.T) {
		systemID := entities.QualifiedNodeID("system", system.ID, "", "")

		start := time.Now()
		descendants := graph.GetDescendants(systemID)
		duration := time.Since(start)

		// Should get all containers (5) + all components (5*20 = 100) = 105 descendants
		expectedCount := 105
		if len(descendants) != expectedCount {
			t.Errorf("expected %d descendants, got %d", expectedCount, len(descendants))
		}

		if duration > 100*time.Millisecond {
			t.Errorf("GetDescendants took %v, expected <100ms", duration)
		}
		t.Logf("GetDescendants time: %v for %d descendants", duration, len(descendants))
	})
}
