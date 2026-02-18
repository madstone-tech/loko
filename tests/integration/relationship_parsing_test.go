package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	d2adapter "github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestFrontmatterRelationshipParsing verifies that relationships defined in component
// frontmatter are correctly parsed and added to the architecture graph as edges.
// This is the T022 integration test for US1.1.
func TestFrontmatterRelationshipParsing(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Create project
	project, err := entities.NewProject("test-project")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	project.Path = projectRoot

	if err := repo.SaveProject(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Create system
	system, err := entities.NewSystem("E-Commerce")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}

	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}

	// Create container
	container, err := entities.NewContainer("Backend Services")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}

	// Create components with relationships
	orderService, err := entities.NewComponent("Order Service")
	if err != nil {
		t.Fatalf("failed to create order service: %v", err)
	}
	orderService.SetDescription("Manages customer orders")
	orderService.SetTechnology("Node.js")
	orderService.AddRelationship("inventory-service", "Checks product availability")
	orderService.AddRelationship("payment-service", "Processes payments")

	inventoryService, err := entities.NewComponent("Inventory Service")
	if err != nil {
		t.Fatalf("failed to create inventory service: %v", err)
	}
	inventoryService.SetDescription("Manages product inventory")
	inventoryService.SetTechnology("Go")
	inventoryService.AddRelationship("inventory-db", "Reads/writes inventory data")

	paymentService, err := entities.NewComponent("Payment Service")
	if err != nil {
		t.Fatalf("failed to create payment service: %v", err)
	}
	paymentService.SetDescription("Payment processing")
	paymentService.SetTechnology("Python")

	inventoryDB, err := entities.NewComponent("Inventory DB")
	if err != nil {
		t.Fatalf("failed to create inventory db: %v", err)
	}
	inventoryDB.SetDescription("PostgreSQL database for inventory")
	inventoryDB.SetTechnology("PostgreSQL")

	// Save components
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, orderService); err != nil {
		t.Fatalf("failed to save order service: %v", err)
	}
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, inventoryService); err != nil {
		t.Fatalf("failed to save inventory service: %v", err)
	}
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, paymentService); err != nil {
		t.Fatalf("failed to save payment service: %v", err)
	}
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, inventoryDB); err != nil {
		t.Fatalf("failed to save inventory db: %v", err)
	}

	// Load systems (which loads components with relationships)
	systems, err := repo.ListSystems(ctx, projectRoot)
	if err != nil {
		t.Fatalf("failed to list systems: %v", err)
	}

	if len(systems) != 1 {
		t.Fatalf("expected 1 system, got %d", len(systems))
	}

	// Build architecture graph
	buildGraph := usecases.NewBuildArchitectureGraph()
	graph, err := buildGraph.Execute(ctx, project, systems)
	if err != nil {
		t.Fatalf("failed to build architecture graph: %v", err)
	}

	// Verify graph has nodes for all components
	if graph.Size() < 6 {
		t.Errorf("expected at least 6 nodes (1 system + 1 container + 4 components), got %d", graph.Size())
	}

	// Verify relationship edges exist
	orderQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, orderService.ID)
	inventoryQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, inventoryService.ID)
	paymentQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, paymentService.ID)
	dbQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, inventoryDB.ID)

	// Test 1: Order Service → Inventory Service relationship
	orderDeps := graph.GetDependencies(orderQualifiedID)
	if len(orderDeps) != 2 {
		t.Errorf("Order Service should have 2 dependencies, got %d", len(orderDeps))
	}

	foundInventoryDep := false
	foundPaymentDep := false
	for _, dep := range orderDeps {
		if dep.ID == inventoryQualifiedID {
			foundInventoryDep = true
		}
		if dep.ID == paymentQualifiedID {
			foundPaymentDep = true
		}
	}

	if !foundInventoryDep {
		t.Error("Order Service → Inventory Service relationship not found in graph")
	}
	if !foundPaymentDep {
		t.Error("Order Service → Payment Service relationship not found in graph")
	}

	// Test 2: Inventory Service → Inventory DB relationship
	inventoryDeps := graph.GetDependencies(inventoryQualifiedID)
	if len(inventoryDeps) != 1 {
		t.Errorf("Inventory Service should have 1 dependency, got %d", len(inventoryDeps))
	}

	if len(inventoryDeps) > 0 && inventoryDeps[0].ID != dbQualifiedID {
		t.Errorf("Inventory Service should depend on %q, got %q", dbQualifiedID, inventoryDeps[0].ID)
	}

	// Test 3: Verify edge descriptions are preserved
	orderOutgoingEdges := graph.GetOutgoingEdges(orderQualifiedID)
	if len(orderOutgoingEdges) != 2 {
		t.Errorf("Order Service should have 2 outgoing edges, got %d", len(orderOutgoingEdges))
	}

	for _, edge := range orderOutgoingEdges {
		if edge.Target == inventoryQualifiedID {
			if edge.Description != "Checks product availability" {
				t.Errorf("Inventory edge description = %q, want %q",
					edge.Description, "Checks product availability")
			}
		}
		if edge.Target == paymentQualifiedID {
			if edge.Description != "Processes payments" {
				t.Errorf("Payment edge description = %q, want %q",
					edge.Description, "Processes payments")
			}
		}
	}

	// Test 4: Count total relationship edges
	totalRelationshipEdges := 0
	for _, edges := range graph.Edges {
		for _, edge := range edges {
			if edge.Type == "depends-on" {
				totalRelationshipEdges++
			}
		}
	}

	expectedEdges := 3 // order→inventory, order→payment, inventory→db
	if totalRelationshipEdges != expectedEdges {
		t.Errorf("expected %d relationship edges, got %d", expectedEdges, totalRelationshipEdges)
	}
}

// TestFrontmatterRelationshipParsing_NoRelationships verifies that components
// without relationships don't create spurious edges.
func TestFrontmatterRelationshipParsing_NoRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Create minimal project structure
	project, _ := entities.NewProject("test-project")
	project.Path = projectRoot
	repo.SaveProject(ctx, project)

	system, _ := entities.NewSystem("Simple System")
	repo.SaveSystem(ctx, projectRoot, system)

	container, _ := entities.NewContainer("Simple Container")
	repo.SaveContainer(ctx, projectRoot, system.ID, container)

	// Create isolated component with NO relationships
	isolatedComponent, _ := entities.NewComponent("Isolated Service")
	isolatedComponent.SetDescription("Standalone service")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, isolatedComponent)

	// Load and build graph
	systems, _ := repo.ListSystems(ctx, projectRoot)
	buildGraph := usecases.NewBuildArchitectureGraph()
	graph, _ := buildGraph.Execute(ctx, project, systems)

	// Verify no relationship edges exist
	componentQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, isolatedComponent.ID)
	deps := graph.GetDependencies(componentQualifiedID)

	if len(deps) != 0 {
		t.Errorf("isolated component should have 0 dependencies, got %d", len(deps))
	}

	dependents := graph.GetDependents(componentQualifiedID)
	if len(dependents) != 0 {
		t.Errorf("isolated component should have 0 dependents, got %d", len(dependents))
	}
}

// TestFrontmatterRelationshipParsing_InvalidTarget verifies that relationships
// to non-existent components are gracefully skipped without breaking the graph.
func TestFrontmatterRelationshipParsing_InvalidTarget(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Create project structure
	project, _ := entities.NewProject("test-project")
	project.Path = projectRoot
	repo.SaveProject(ctx, project)

	system, _ := entities.NewSystem("Test System")
	repo.SaveSystem(ctx, projectRoot, system)

	container, _ := entities.NewContainer("Test Container")
	repo.SaveContainer(ctx, projectRoot, system.ID, container)

	// Create component with valid AND invalid relationships
	serviceA, _ := entities.NewComponent("Service A")
	serviceA.AddRelationship("service-b", "Valid relationship")
	serviceA.AddRelationship("non-existent-service", "Invalid relationship - target doesn't exist")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, serviceA)

	serviceB, _ := entities.NewComponent("Service B")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, serviceB)

	// Load and build graph
	systems, _ := repo.ListSystems(ctx, projectRoot)
	buildGraph := usecases.NewBuildArchitectureGraph()
	graph, err := buildGraph.Execute(ctx, project, systems)

	if err != nil {
		t.Fatalf("graph building should not error on invalid relationships, got: %v", err)
	}

	// Verify only the valid relationship created an edge
	serviceAQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, serviceA.ID)
	serviceBQualifiedID := entities.QualifiedNodeID("component", system.ID, container.ID, serviceB.ID)

	deps := graph.GetDependencies(serviceAQualifiedID)
	if len(deps) != 1 {
		t.Errorf("Service A should have 1 valid dependency (invalid one skipped), got %d", len(deps))
	}

	if len(deps) > 0 && deps[0].ID != serviceBQualifiedID {
		t.Errorf("Service A dependency should be Service B, got %q", deps[0].ID)
	}
}

// TestD2RelationshipParsing_FrontmatterAndD2 verifies that relationships from both
// frontmatter and D2 files are merged into the architecture graph (union merge).
// This is T030 integration test for US1.2.
func TestD2RelationshipParsing_FrontmatterAndD2(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Build minimal project structure
	project, _ := entities.NewProject("test-project")
	project.Path = projectRoot
	repo.SaveProject(ctx, project)

	system, _ := entities.NewSystem("System")
	repo.SaveSystem(ctx, projectRoot, system)

	container, _ := entities.NewContainer("Container")
	repo.SaveContainer(ctx, projectRoot, system.ID, container)

	// api-service has frontmatter relationship: api-service -> database
	apiService, _ := entities.NewComponent("API Service")
	apiService.AddRelationship("cache-service", "Caches session data") // frontmatter
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, apiService)

	// Manually write a .d2 file alongside component.md with an extra relationship
	cacheService, _ := entities.NewComponent("Cache Service")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, cacheService)

	database, _ := entities.NewComponent("Database")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, database)

	// Write a D2 file into api-service's component directory
	apiComponentDir := filepath.Join(projectRoot, "src", system.ID, container.ID, apiService.ID)
	d2Content := "api-service -> database: Reads data\napi-service -> cache-service: Caches session data"
	if err := os.WriteFile(filepath.Join(apiComponentDir, "api-service.d2"), []byte(d2Content), 0o644); err != nil {
		t.Fatalf("failed to write D2 file: %v", err)
	}

	// Set Path so D2 parsing activates
	apiService.Path = apiComponentDir

	// Build graph with real D2 parser
	d2Parser := d2adapter.NewD2Parser()
	buildGraph := usecases.NewBuildArchitectureGraphWithD2(d2Parser)
	systems, _ := repo.ListSystems(ctx, projectRoot)

	// Attach path to loaded component
	for _, sys := range systems {
		for _, cont := range sys.Containers {
			for _, comp := range cont.Components {
				if comp.ID == apiService.ID {
					comp.Path = apiComponentDir
				}
			}
		}
	}

	graph, err := buildGraph.Execute(ctx, project, systems)
	if err != nil {
		t.Fatalf("failed to build architecture graph: %v", err)
	}

	// api-service should have 2 outgoing edges:
	// - cache-service (frontmatter, deduplicated from D2)
	// - database (D2 only)
	apiID := entities.QualifiedNodeID("component", system.ID, container.ID, apiService.ID)
	deps := graph.GetDependencies(apiID)

	if len(deps) != 2 {
		t.Errorf("expected 2 edges (frontmatter + D2 union, deduplicated), got %d", len(deps))
	}
}

// TestD2RelationshipParsing_ErrorHandling verifies that D2 parse errors are
// handled gracefully: the component is skipped and graph building continues.
// This is T031 integration test for US1.2.
func TestD2RelationshipParsing_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	project, _ := entities.NewProject("test-project")
	project.Path = projectRoot
	repo.SaveProject(ctx, project)

	system, _ := entities.NewSystem("System")
	repo.SaveSystem(ctx, projectRoot, system)

	container, _ := entities.NewContainer("Container")
	repo.SaveContainer(ctx, projectRoot, system.ID, container)

	// Component A: valid frontmatter relationship
	compA, _ := entities.NewComponent("Component A")
	compA.AddRelationship("component-c", "Calls C")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, compA)

	// Component B: gets a malformed D2 file
	compB, _ := entities.NewComponent("Component B")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, compB)
	compBDir := filepath.Join(projectRoot, "src", system.ID, container.ID, compB.ID)
	os.WriteFile(filepath.Join(compBDir, "bad.d2"), []byte("{ invalid d2 syntax :::"), 0o644)

	// Component C: no D2 file, just a target
	compC, _ := entities.NewComponent("Component C")
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, compC)

	systems, _ := repo.ListSystems(ctx, projectRoot)

	// Attach D2 paths
	for _, sys := range systems {
		for _, cont := range sys.Containers {
			for _, comp := range cont.Components {
				comp.Path = filepath.Join(projectRoot, "src", sys.ID, cont.ID, comp.ID)
			}
		}
	}

	d2Parser := d2adapter.NewD2Parser()
	buildGraph := usecases.NewBuildArchitectureGraphWithD2(d2Parser)
	graph, err := buildGraph.Execute(ctx, project, systems)

	// Graph build must succeed even though component B's D2 file is malformed
	if err != nil {
		t.Fatalf("graph building should succeed despite malformed D2 file, got: %v", err)
	}

	// Component A → Component C edge should still exist (frontmatter)
	compAID := entities.QualifiedNodeID("component", system.ID, container.ID, compA.ID)
	compCID := entities.QualifiedNodeID("component", system.ID, container.ID, compC.ID)

	deps := graph.GetDependencies(compAID)
	if len(deps) != 1 {
		t.Errorf("component A should have 1 frontmatter dependency, got %d", len(deps))
	}
	if len(deps) > 0 && deps[0].ID != compCID {
		t.Errorf("component A dependency = %q, want %q", deps[0].ID, compCID)
	}

	// Component B should have 0 edges (malformed D2 skipped gracefully)
	compBID := entities.QualifiedNodeID("component", system.ID, container.ID, compB.ID)
	bDeps := graph.GetDependencies(compBID)
	if len(bDeps) != 0 {
		t.Errorf("component B should have 0 edges (D2 parse failed), got %d", len(bDeps))
	}
}
