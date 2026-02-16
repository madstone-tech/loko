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

	// Verify system node using qualified ID
	systemQualifiedID := entities.QualifiedNodeID("system", system.ID, "", "")
	systemNode := graph.GetNode(systemQualifiedID)
	if systemNode == nil {
		t.Fatal("system node not found")
	}

	if systemNode.Type != "system" || systemNode.Level != 1 {
		t.Error("system node has incorrect type or level")
	}

	// Verify container node using qualified ID
	containerQualifiedID := entities.QualifiedNodeID("container", system.ID, container.ID, "")
	containerNode := graph.GetNode(containerQualifiedID)
	if containerNode == nil {
		t.Fatal("container node not found")
	}

	if containerNode.Type != "container" || containerNode.Level != 2 {
		t.Error("container node has incorrect type or level")
	}

	if containerNode.ParentID != systemQualifiedID {
		t.Error("container parent ID incorrect")
	}

	// Verify component node using qualified ID
	componentQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, component.ID)
	componentNode := graph.GetNode(componentQualifiedID)
	if componentNode == nil {
		t.Fatal("component node not found")
	}

	if componentNode.Type != "component" || componentNode.Level != 3 {
		t.Error("component node has incorrect type or level")
	}

	if componentNode.ParentID != containerQualifiedID {
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

	// Test GetChildren using qualified IDs
	sys1QualifiedID := entities.QualifiedNodeID("system", sys1.ID, "", "")
	sys1Children := graph.GetChildren(sys1QualifiedID)
	if len(sys1Children) != 2 {
		t.Errorf("expected 2 children for sys1, got %d", len(sys1Children))
	}

	// Test GetDescendants
	sys1Descendants := graph.GetDescendants(sys1QualifiedID)
	if len(sys1Descendants) != 4 {
		t.Errorf("expected 4 descendants for sys1 (2 containers + 2 components), got %d", len(sys1Descendants))
	}

	// Test GetAncestors using qualified IDs
	comp1QualifiedID := entities.QualifiedNodeID("component", sys1.ID, cont1.ID, comp1.ID)
	comp1Ancestors := graph.GetAncestors(comp1QualifiedID)
	if len(comp1Ancestors) != 2 {
		t.Errorf("expected 2 ancestors for comp1 (container + system), got %d", len(comp1Ancestors))
	}

	cont1QualifiedID := entities.QualifiedNodeID("container", sys1.ID, cont1.ID, "")
	if comp1Ancestors[0].ID != cont1QualifiedID || comp1Ancestors[1].ID != sys1QualifiedID {
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

	// Test GetDependencies using qualified IDs
	comp1QualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, comp1.ID)
	comp1Deps := graph.GetDependencies(comp1QualifiedID)
	if len(comp1Deps) != 2 {
		t.Errorf("expected 2 dependencies for comp1, got %d", len(comp1Deps))
	}

	// Test GetDependents using qualified IDs
	comp2QualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, comp2.ID)
	comp2Dependents := graph.GetDependents(comp2QualifiedID)
	if len(comp2Dependents) != 1 {
		t.Errorf("expected 1 dependent for comp2, got %d", len(comp2Dependents))
	}

	if comp2Dependents[0].ID != comp1QualifiedID {
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

	// Extract system 1 subgraph using qualified ID
	sys1QualifiedID := entities.QualifiedNodeID("system", sys1.ID, "", "")
	subgraph, err := uc.GetSystemGraph(graph, sys1QualifiedID)
	if err != nil {
		t.Fatalf("failed to get system graph: %v", err)
	}

	// Verify subgraph contains only system 1 hierarchy
	if subgraph.Size() != 3 {
		t.Errorf("expected 3 nodes in subgraph, got %d", subgraph.Size())
	}

	// Verify system 2 components are not in subgraph using qualified IDs
	comp2QualifiedID := entities.QualifiedNodeID("component", sys2.ID, cont2.ID, comp2.ID)
	if subgraph.GetNode(comp2QualifiedID) != nil {
		t.Error("system 2 component should not be in system 1 subgraph")
	}

	// Verify system 1 components are in subgraph using qualified IDs
	comp1QualifiedID := entities.QualifiedNodeID("component", sys1.ID, cont1.ID, comp1.ID)
	if subgraph.GetNode(comp1QualifiedID) == nil {
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
	if report.ComponentsCount != 4 {
		t.Errorf("expected 4 components, got %v", report.ComponentsCount)
	}

	if len(report.IsolatedComponents) != 1 {
		t.Errorf("expected 1 isolated component, got %d: %v", len(report.IsolatedComponents), report.IsolatedComponents)
	}

	// Auth component has 2 dependencies, which is not > 2, so shouldn't be marked highly coupled
	// DB component has 1 dependent, which is not > 2, so shouldn't be marked central
	// The thresholds in AnalyzeDependencies are > 2 for both

	if len(report.HighlyCoupledComponents) != 0 {
		t.Errorf("expected 0 highly coupled components (threshold > 2), got %d: %v", len(report.HighlyCoupledComponents), report.HighlyCoupledComponents)
	}

	if len(report.CentralComponents) != 0 {
		t.Errorf("expected 0 central components (threshold > 2), got %d: %v", len(report.CentralComponents), report.CentralComponents)
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

// TestGraphBuildingWithDuplicateComponentNames tests that components with the same name in different systems don't collide.
func TestGraphBuildingWithDuplicateComponentNames(t *testing.T) {
	project, _ := entities.NewProject("test-project")

	// System 1 with "auth" component
	sys1, _ := entities.NewSystem("Backend")
	cont1, _ := entities.NewContainer("API")
	comp1, _ := entities.NewComponent("Auth")
	cont1.AddComponent(comp1)
	sys1.AddContainer(cont1)

	// System 2 with "auth" component (same short name)
	sys2, _ := entities.NewSystem("Admin")
	cont2, _ := entities.NewContainer("UI")
	comp2, _ := entities.NewComponent("Auth")
	cont2.AddComponent(comp2)
	sys2.AddContainer(cont2)

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{sys1, sys2})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Verify both components exist in graph (with qualified IDs)
	// After implementation, these should be:
	// - backend/api/auth
	// - admin/ui/auth

	// Count all nodes - should be 6 (2 systems, 2 containers, 2 components)
	if graph.Size() != 6 {
		t.Errorf("expected 6 nodes (both auth components should be added), got %d", graph.Size())
	}

	// Verify both auth components are retrievable
	// This test will fail with current implementation (collision bug)
	// After fix, use qualified IDs
	backendAuth := graph.GetNode(entities.QualifiedNodeID("component", sys1.ID, cont1.ID, comp1.ID))
	if backendAuth == nil {
		t.Error("backend auth component not found - node ID collision occurred")
	}

	adminAuth := graph.GetNode(entities.QualifiedNodeID("component", sys2.ID, cont2.ID, comp2.ID))
	if adminAuth == nil {
		t.Error("admin auth component not found - node ID collision occurred")
	}
}

// TestGraphBuildingWithDuplicateContainerNames tests that containers with the same name in different systems don't collide.
func TestGraphBuildingWithDuplicateContainerNames(t *testing.T) {
	project, _ := entities.NewProject("test-project")

	// System 1 with "api" container
	sys1, _ := entities.NewSystem("Backend")
	cont1, _ := entities.NewContainer("API")
	sys1.AddContainer(cont1)

	// System 2 with "api" container (same short name)
	sys2, _ := entities.NewSystem("Frontend")
	cont2, _ := entities.NewContainer("API")
	sys2.AddContainer(cont2)

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{sys1, sys2})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// Verify both containers exist (with qualified IDs)
	// Should be: backend/api and frontend/api
	if graph.Size() != 4 {
		t.Errorf("expected 4 nodes (2 systems + 2 containers), got %d", graph.Size())
	}

	backendAPI := graph.GetNode(entities.QualifiedNodeID("container", sys1.ID, cont1.ID, ""))
	if backendAPI == nil {
		t.Error("backend/api container not found - node ID collision occurred")
	}

	frontendAPI := graph.GetNode(entities.QualifiedNodeID("container", sys2.ID, cont2.ID, ""))
	if frontendAPI == nil {
		t.Error("frontend/api container not found - node ID collision occurred")
	}
}

// TestRelationshipResolutionUsingShortIDs tests that component relationships using short IDs are resolved correctly.
func TestRelationshipResolutionUsingShortIDs(t *testing.T) {
	project, _ := entities.NewProject("test-project")

	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")
	system.AddContainer(container)

	// Create components
	auth, _ := entities.NewComponent("Auth")
	db, _ := entities.NewComponent("Database")

	// Add relationship using short ID (current behavior)
	auth.AddRelationship(db.ID, "queries user data")

	container.AddComponent(auth)
	container.AddComponent(db)

	// Build graph
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// After implementation, verify relationship was resolved from short ID to qualified ID
	authQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, auth.ID)
	dbQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, db.ID)

	// Verify auth component has dependency on db component
	authDeps := graph.GetDependencies(authQualifiedID)
	if len(authDeps) != 1 {
		t.Errorf("expected 1 dependency for auth, got %d", len(authDeps))
	}

	if len(authDeps) > 0 && authDeps[0].ID != dbQualifiedID {
		t.Errorf("auth dependency should point to %q, got %q", dbQualifiedID, authDeps[0].ID)
	}
}

// TestBuildArchitectureGraph_UnionMerge_FrontmatterOnly verifies that when only
// frontmatter relationships exist (no D2), they are added to the graph.
// This is part of T029 union merge logic tests.
func TestBuildArchitectureGraph_UnionMerge_FrontmatterOnly(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	// Component with frontmatter relationships only (no D2 diagram)
	serviceA, _ := entities.NewComponent("Service A")
	serviceA.AddRelationship("service-b", "Calls via HTTP")

	serviceB, _ := entities.NewComponent("Service B")

	container.AddComponent(serviceA)
	container.AddComponent(serviceB)

	// Build graph (no D2Parser provided yet)
	uc := NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify frontmatter relationship exists in graph
	serviceAID := entities.QualifiedNodeID("component", system.ID, container.ID, serviceA.ID)
	serviceBID := entities.QualifiedNodeID("component", system.ID, container.ID, serviceB.ID)

	deps := graph.GetDependencies(serviceAID)
	if len(deps) != 1 {
		t.Errorf("expected 1 dependency from frontmatter, got %d", len(deps))
	}

	if len(deps) > 0 && deps[0].ID != serviceBID {
		t.Errorf("dependency target = %q, want %q", deps[0].ID, serviceBID)
	}

	// Verify edge has correct description
	edges := graph.GetOutgoingEdges(serviceAID)
	if len(edges) > 0 && edges[0].Description != "Calls via HTTP" {
		t.Errorf("edge description = %q, want %q", edges[0].Description, "Calls via HTTP")
	}
}

// TestBuildArchitectureGraph_UnionMerge_D2Only verifies that D2-only relationships
// (when frontmatter has no relationships) are added to the graph.
// Note: This test will be fully functional after T034-T035 (D2Parser integration).
func TestBuildArchitectureGraph_UnionMerge_D2Only(t *testing.T) {
	t.Skip("Skipping until T034-T035: D2Parser not yet wired to BuildArchitectureGraph")

	// TODO (T034-T035): Implement this test once D2Parser is integrated
	// Expected behavior:
	// - Component has D2 file with arrows: api -> db, api -> cache
	// - Component frontmatter has empty relationships map
	// - Result: Graph has 2 edges from D2 parsing
}

// TestBuildArchitectureGraph_UnionMerge_BothSameTypeDeduplicated verifies that
// when the same relationship exists in both frontmatter AND D2, it is deduplicated
// (only appears once in the graph).
func TestBuildArchitectureGraph_UnionMerge_BothSameTypeDeduplicated(t *testing.T) {
	t.Skip("Skipping until T034-T036: Union merge and deduplication not yet implemented")

	// TODO (T034-T036): Implement this test once union merge is implemented
	// Expected behavior:
	// - Frontmatter: service-a -> service-b: "Sends requests"
	// - D2 file: service-a -> service-b: "Sends requests"
	// - Result: Graph has 1 edge (deduplicated by source+target+type tuple)
}

// TestBuildArchitectureGraph_UnionMerge_BothDifferentTypesKeepBoth verifies that
// when frontmatter and D2 define different relationship types between the same
// source/target pair, both are kept (different edge types).
func TestBuildArchitectureGraph_UnionMerge_BothDifferentTypesKeepBoth(t *testing.T) {
	t.Skip("Skipping until T034-T036: Union merge with different types not yet implemented")

	// TODO (T034-T036): Implement this test once union merge supports multiple edge types
	// Expected behavior:
	// - Frontmatter: service-a -> service-b: "depends-on"
	// - D2 file: service-a -> service-b: "uses" (if D2 supports type annotations)
	// - Result: Graph has 2 edges with different types
	//
	// Note: Current implementation uses single type "depends-on" for all relationships.
	// This test documents future enhancement if we add relationship type support.
}

// TestBuildArchitectureGraph_UnionMerge_DeduplicationKey verifies that the
// deduplication key correctly uses (source, target, type) tuple to identify duplicates.
func TestBuildArchitectureGraph_UnionMerge_DeduplicationKey(t *testing.T) {
	t.Skip("Skipping until T036: Deduplication logic not yet implemented")

	// TODO (T036): Implement this test once deduplication is added
	// Test cases:
	// 1. Same source, same target, same type → deduplicate
	// 2. Same source, different target → keep both
	// 3. Different source, same target → keep both
	// 4. Same source+target, different type → keep both (if types supported)
}
