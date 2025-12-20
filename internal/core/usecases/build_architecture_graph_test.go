package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestBuildArchitectureGraphBasic tests basic graph building from a project.
func TestBuildArchitectureGraphBasic(t *testing.T) {
	// Create test project
	project, err := entities.NewProject("test-project")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create system
	system, err := entities.NewSystem("E-Commerce")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}

	// Create container
	container, err := entities.NewContainer("API Server")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	if err := system.AddContainer(container); err != nil {
		t.Fatalf("failed to add container to system: %v", err)
	}

	// Create component
	component, err := entities.NewComponent("Authentication")
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}

	if err := container.AddComponent(component); err != nil {
		t.Fatalf("failed to add component to container: %v", err)
	}

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Verify graph structure
	if graph.Size() != 3 {
		t.Errorf("expected 3 nodes, got %d", graph.Size())
	}

	// Verify system node
	systemNode := graph.GetNode(system.ID)
	if systemNode == nil {
		t.Fatal("system node not found")
	}

	if systemNode.Type != "system" || systemNode.Level != 1 {
		t.Error("system node has incorrect type or level")
	}

	// Verify container node
	containerNode := graph.GetNode(container.ID)
	if containerNode == nil {
		t.Fatal("container node not found")
	}

	if containerNode.Type != "container" || containerNode.Level != 2 {
		t.Error("container node has incorrect type or level")
	}

	if containerNode.ParentID != system.ID {
		t.Error("container parent ID incorrect")
	}

	// Verify component node
	componentNode := graph.GetNode(component.ID)
	if componentNode == nil {
		t.Fatal("component node not found")
	}

	if componentNode.Type != "component" || componentNode.Level != 3 {
		t.Error("component node has incorrect type or level")
	}

	if componentNode.ParentID != container.ID {
		t.Error("component parent ID incorrect")
	}
}

// TestBuildArchitectureGraphHierarchy tests hierarchical relationships in the graph.
func TestBuildArchitectureGraphHierarchy(t *testing.T) {
	// Create test project with multiple systems and containers
	project, _ := entities.NewProject("test-project")

	// System 1
	sys1, _ := entities.NewSystem("System 1")
	cont1, _ := entities.NewContainer("Container 1")
	cont2, _ := entities.NewContainer("Container 2")
	sys1.AddContainer(cont1)
	sys1.AddContainer(cont2)

	// System 2
	sys2, _ := entities.NewSystem("System 2")
	cont3, _ := entities.NewContainer("Container 3")
	sys2.AddContainer(cont3)

	// Add components to containers
	comp1, _ := entities.NewComponent("Component 1")
	comp2, _ := entities.NewComponent("Component 2")
	cont1.AddComponent(comp1)
	cont1.AddComponent(comp2)

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{sys1, sys2})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Test GetChildren
	sys1Children := graph.GetChildren(sys1.ID)
	if len(sys1Children) != 2 {
		t.Errorf("expected 2 children for sys1, got %d", len(sys1Children))
	}

	// Test GetDescendants
	sys1Descendants := graph.GetDescendants(sys1.ID)
	if len(sys1Descendants) != 4 {
		t.Errorf("expected 4 descendants for sys1 (2 containers + 2 components), got %d", len(sys1Descendants))
	}

	// Test GetAncestors
	comp1Ancestors := graph.GetAncestors(comp1.ID)
	if len(comp1Ancestors) != 2 {
		t.Errorf("expected 2 ancestors for comp1 (container + system), got %d", len(comp1Ancestors))
	}

	if comp1Ancestors[0].ID != cont1.ID || comp1Ancestors[1].ID != sys1.ID {
		t.Error("ancestor chain incorrect")
	}
}

// TestBuildArchitectureGraphRelationships tests component relationships in the graph.
func TestBuildArchitectureGraphRelationships(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	// Create components with relationships
	comp1, _ := entities.NewComponent("Auth")
	comp2, _ := entities.NewComponent("Database")
	comp3, _ := entities.NewComponent("Cache")

	comp1.AddRelationship(comp2.ID, "queries user data")
	comp1.AddRelationship(comp3.ID, "caches tokens")

	container.AddComponent(comp1)
	container.AddComponent(comp2)
	container.AddComponent(comp3)

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Test GetDependencies
	comp1Deps := graph.GetDependencies(comp1.ID)
	if len(comp1Deps) != 2 {
		t.Errorf("expected 2 dependencies for comp1, got %d", len(comp1Deps))
	}

	// Test GetDependents
	comp2Dependents := graph.GetDependents(comp2.ID)
	if len(comp2Dependents) != 1 {
		t.Errorf("expected 1 dependent for comp2, got %d", len(comp2Dependents))
	}

	if comp2Dependents[0].ID != comp1.ID {
		t.Error("dependent relationship incorrect")
	}
}

// TestGetSystemGraph tests extracting a subgraph for a specific system.
func TestGetSystemGraph(t *testing.T) {
	project, _ := entities.NewProject("test-project")

	// Create two systems
	sys1, _ := entities.NewSystem("System 1")
	cont1, _ := entities.NewContainer("Container 1")
	comp1, _ := entities.NewComponent("Component 1")
	cont1.AddComponent(comp1)
	sys1.AddContainer(cont1)

	sys2, _ := entities.NewSystem("System 2")
	cont2, _ := entities.NewContainer("Container 2")
	comp2, _ := entities.NewComponent("Component 2")
	cont2.AddComponent(comp2)
	sys2.AddContainer(cont2)

	// Build full graph
	uc := NewBuildArchitectureGraph()
	graph, _ := uc.Execute(context.Background(), project, []*entities.System{sys1, sys2})

	// Extract system 1 subgraph
	subgraph, err := uc.GetSystemGraph(graph, sys1.ID)
	if err != nil {
		t.Fatalf("failed to get system graph: %v", err)
	}

	// Verify subgraph contains only system 1 hierarchy
	if subgraph.Size() != 3 {
		t.Errorf("expected 3 nodes in subgraph, got %d", subgraph.Size())
	}

	// Verify system 2 components are not in subgraph
	if subgraph.GetNode(comp2.ID) != nil {
		t.Error("system 2 component should not be in system 1 subgraph")
	}

	// Verify system 1 components are in subgraph
	if subgraph.GetNode(comp1.ID) == nil {
		t.Error("system 1 component should be in subgraph")
	}
}

// TestAnalyzeDependencies tests dependency analysis.
func TestAnalyzeDependencies(t *testing.T) {
	project, _ := entities.NewProject("test-project")

	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	// Create components
	auth, _ := entities.NewComponent("Auth")
	db, _ := entities.NewComponent("Database")
	cache, _ := entities.NewComponent("Cache")
	isolComp, _ := entities.NewComponent("Isolated")

	auth.AddRelationship(db.ID, "queries data")
	auth.AddRelationship(cache.ID, "stores tokens")

	container.AddComponent(auth)
	container.AddComponent(db)
	container.AddComponent(cache)
	container.AddComponent(isolComp)

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, _ := uc.Execute(context.Background(), project, []*entities.System{system})

	// Analyze dependencies
	report := uc.AnalyzeDependencies(graph)

	// Verify report
	if report["components_count"] != 4 {
		t.Errorf("expected 4 components, got %v", report["components_count"])
	}

	isolatedList := report["isolated_components"].([]string)
	if len(isolatedList) != 1 {
		t.Errorf("expected 1 isolated component, got %d: %v", len(isolatedList), isolatedList)
	}

	// Auth component has 2 dependencies, which is not > 2, so shouldn't be marked highly coupled
	// DB component has 1 dependent, which is not > 2, so shouldn't be marked central
	// The thresholds in AnalyzeDependencies are > 2 for both

	highlyCoupled := report["highly_coupled_components"].(map[string]int)
	if len(highlyCoupled) != 0 {
		t.Errorf("expected 0 highly coupled components (threshold > 2), got %d: %v", len(highlyCoupled), highlyCoupled)
	}

	centralComps := report["central_components"].(map[string]int)
	if len(centralComps) != 0 {
		t.Errorf("expected 0 central components (threshold > 2), got %d: %v", len(centralComps), centralComps)
	}
}

// TestBuildArchitectureGraphValidation tests graph validation.
func TestBuildArchitectureGraphValidation(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	component, _ := entities.NewComponent("Component")

	container.AddComponent(component)
	system.AddContainer(container)

	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Verify graph is valid
	if err := graph.Validate(); err != nil {
		t.Fatalf("graph validation failed: %v", err)
	}
}
