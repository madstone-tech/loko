package usecases

import (
	"context"
	"os"
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

// mockD2Parser is a test double for the D2Parser port.
// It maps d2 source content to a fixed list of relationships so tests can
// control which relationships are "found" in each file.
type mockD2Parser struct {
	// byContent maps d2 source string → relationships to return.
	// If the content isn't in the map, returns the default slice.
	byContent map[string][]entities.D2Relationship
	// default relationships returned when content isn't in byContent
	defaultRels []entities.D2Relationship
	err         error
}

func (m *mockD2Parser) ParseRelationships(_ context.Context, content string) ([]entities.D2Relationship, error) {
	if m.err != nil {
		return nil, m.err
	}
	if rels, ok := m.byContent[content]; ok {
		return rels, nil
	}
	return m.defaultRels, nil
}

// writeD2File writes content to a .d2 file inside dir and returns the content written.
func writeD2File(t *testing.T, dir, filename, content string) string {
	t.Helper()
	if err := os.WriteFile(dir+"/"+filename, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write D2 file: %v", err)
	}
	return content
}

// TestBuildArchitectureGraph_UnionMerge_D2Only verifies that D2-sourced relationships
// are added to the graph when a D2Parser is injected.
func TestBuildArchitectureGraph_UnionMerge_D2Only(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	// Give each component a real temp dir so parseComponentD2 can find .d2 files
	apiDir := t.TempDir()
	dbDir := t.TempDir()
	cacheDir := t.TempDir()

	apiD2Content := writeD2File(t, apiDir, "api.d2", "api -> database: Queries\napi -> cache: Caches tokens")

	api, _ := entities.NewComponent("API")
	api.Path = apiDir
	db, _ := entities.NewComponent("Database")
	db.Path = dbDir // no .d2 file — nothing to parse
	cache, _ := entities.NewComponent("Cache")
	cache.Path = cacheDir

	container.AddComponent(api)
	container.AddComponent(db)
	container.AddComponent(cache)

	dbRel, _ := entities.NewD2Relationship("api", "database", "Queries")
	cacheRel, _ := entities.NewD2Relationship("api", "cache", "Caches tokens")
	mock := &mockD2Parser{
		byContent: map[string][]entities.D2Relationship{
			apiD2Content: {*dbRel, *cacheRel},
		},
	}

	uc := NewBuildArchitectureGraphWithD2(mock)
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	apiID := entities.QualifiedNodeID("component", system.ID, container.ID, api.ID)
	deps := graph.GetDependencies(apiID)
	if len(deps) != 2 {
		t.Errorf("expected 2 D2-sourced deps for api, got %d", len(deps))
	}
}

// TestBuildArchitectureGraph_UnionMerge_BothSameTypeDeduplicated verifies that
// a relationship present in both frontmatter and D2 is added only once.
func TestBuildArchitectureGraph_UnionMerge_BothSameTypeDeduplicated(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	svcA, _ := entities.NewComponent("Service A")
	svcADir := t.TempDir()
	svcAContent := writeD2File(t, svcADir, "service-a.d2", "service-a -> service-b: Sends requests")
	svcA.Path = svcADir
	svcA.AddRelationship("service-b", "Sends requests") // frontmatter
	svcB, _ := entities.NewComponent("Service B")
	svcB.Path = t.TempDir()

	container.AddComponent(svcA)
	container.AddComponent(svcB)

	// D2 duplicates the same relationship
	dupRel, _ := entities.NewD2Relationship("service-a", "service-b", "Sends requests")
	mock := &mockD2Parser{
		byContent: map[string][]entities.D2Relationship{
			svcAContent: {*dupRel},
		},
	}

	uc := NewBuildArchitectureGraphWithD2(mock)
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	svcAID := entities.QualifiedNodeID("component", system.ID, container.ID, svcA.ID)
	deps := graph.GetDependencies(svcAID)
	if len(deps) != 1 {
		t.Errorf("expected 1 deduplicated edge, got %d", len(deps))
	}
}

// TestBuildArchitectureGraph_UnionMerge_BothDifferentTargetsKeepBoth verifies that
// distinct relationships from frontmatter and D2 are both preserved.
func TestBuildArchitectureGraph_UnionMerge_BothDifferentTargetsKeepBoth(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	api, _ := entities.NewComponent("API")
	apiDir := t.TempDir()
	apiContent := writeD2File(t, apiDir, "api.d2", "api -> replica-db: Reads for reports")
	api.Path = apiDir
	api.AddRelationship("primary-db", "Reads user data") // frontmatter
	primaryDB, _ := entities.NewComponent("Primary DB")
	primaryDB.Path = t.TempDir()
	replicaDB, _ := entities.NewComponent("Replica DB")
	replicaDB.Path = t.TempDir()

	container.AddComponent(api)
	container.AddComponent(primaryDB)
	container.AddComponent(replicaDB)

	// D2 adds a different target
	d2Rel, _ := entities.NewD2Relationship("api", "replica-db", "Reads for reports")
	mock := &mockD2Parser{
		byContent: map[string][]entities.D2Relationship{
			apiContent: {*d2Rel},
		},
	}

	uc := NewBuildArchitectureGraphWithD2(mock)
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	apiID := entities.QualifiedNodeID("component", system.ID, container.ID, api.ID)
	deps := graph.GetDependencies(apiID)
	if len(deps) != 2 {
		t.Errorf("expected 2 edges (different targets kept), got %d", len(deps))
	}
}

// TestBuildArchitectureGraph_UnionMerge_DeduplicationKey verifies the deduplication
// key is per (source, target): same pair → 1 edge; different pairs → 2 edges.
func TestBuildArchitectureGraph_UnionMerge_DeduplicationKey(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("System")
	container, _ := entities.NewContainer("Container")
	system.AddContainer(container)

	svc, _ := entities.NewComponent("Service")
	svcDir := t.TempDir()
	svcContent := writeD2File(t, svcDir, "service.d2", "service -> db-a: Uses db-a\nservice -> db-b: Uses db-b")
	svc.Path = svcDir
	svc.AddRelationship("db-a", "Uses db-a") // frontmatter
	dbA, _ := entities.NewComponent("DB A")
	dbA.Path = t.TempDir()
	dbB, _ := entities.NewComponent("DB B")
	dbB.Path = t.TempDir()

	container.AddComponent(svc)
	container.AddComponent(dbA)
	container.AddComponent(dbB)

	// D2 duplicates db-a and adds db-b
	dupRel, _ := entities.NewD2Relationship("service", "db-a", "Uses db-a")
	newRel, _ := entities.NewD2Relationship("service", "db-b", "Uses db-b")
	mock := &mockD2Parser{
		byContent: map[string][]entities.D2Relationship{
			svcContent: {*dupRel, *newRel},
		},
	}

	uc := NewBuildArchitectureGraphWithD2(mock)
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	svcID := entities.QualifiedNodeID("component", system.ID, container.ID, svc.ID)
	deps := graph.GetDependencies(svcID)
	// db-a deduped (1), db-b new (1) = 2 total
	if len(deps) != 2 {
		t.Errorf("expected 2 edges (db-a deduped + db-b new), got %d", len(deps))
	}
}

// TestBuildArchitectureGraph_WithRelationshipRepository verifies that stored
// relationships from RelationshipRepository are added as graph edges.
func TestBuildArchitectureGraph_WithRelationshipRepository(t *testing.T) {
	// Build a project with two containers.
	project, _ := entities.NewProject("test-project")
	project.Path = "/tmp/test-project"

	system, _ := entities.NewSystem("My System")
	cont1, _ := entities.NewContainer("API")
	cont2, _ := entities.NewContainer("Worker")
	_ = system.AddContainer(cont1)
	_ = system.AddContainer(cont2)

	// Pre-seed the mock repository with one relationship.
	relRepo := newMockRelationshipRepository()
	relID := entities.GenerateRelationshipID(
		system.ID+"/"+cont1.ID,
		system.ID+"/"+cont2.ID,
		"dispatches",
	)
	relRepo.seed("/tmp/test-project", system.ID, []entities.Relationship{
		{
			ID:     relID,
			Source: system.ID + "/" + cont1.ID,
			Target: system.ID + "/" + cont2.ID,
			Label:  "dispatches",
			Type:   "async",
		},
	})

	uc := NewBuildArchitectureGraphWithRelRepo(relRepo)
	graph, err := uc.Execute(context.Background(), project, []*entities.System{system})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Graph should have 3 nodes: system + 2 containers.
	if graph.Size() != 3 {
		t.Errorf("expected 3 nodes, got %d", graph.Size())
	}

	// The stored relationship should be an edge in the graph.
	if graph.EdgeCount() == 0 {
		t.Error("expected at least 1 edge from stored relationship, got 0")
	}
}
