package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

// TestMCPFindRelationships_BasicFunctionality verifies that the find_relationships
// MCP tool returns populated results when components have frontmatter relationships.
// This is T025-T026 verification.
func TestMCPFindRelationships_BasicFunctionality(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Create test project with relationships
	setupTestProjectWithRelationships(t, ctx, repo, projectRoot)

	// Create MCP tool
	tool := tools.NewFindRelationshipsTool(repo)

	// Test 1: Find all relationships (wildcard pattern)
	t.Run("find all relationships", func(t *testing.T) {
		args := map[string]any{
			"project_root":   projectRoot,
			"source_pattern": "*", // Wildcard to match all sources
		}

		result, err := tool.Call(ctx, args)
		if err != nil {
			t.Fatalf("find_relationships failed: %v", err)
		}

		response, ok := result.(*entities.FindRelationshipsResponse)
		if !ok {
			t.Fatalf("expected FindRelationshipsResponse, got %T", result)
		}

		if len(response.Relationships) == 0 {
			t.Error("expected non-empty relationships array, got empty")
		}

		if response.TotalMatched == 0 {
			t.Error("expected TotalMatched > 0, got 0")
		}

		// Verify we have the 3 relationships we created
		// order→inventory, order→payment, inventory→db
		if response.TotalMatched < 3 {
			t.Errorf("expected at least 3 relationships, got %d", response.TotalMatched)
		}
	})

	// Test 2: Find relationships with source filter
	t.Run("filter by source pattern", func(t *testing.T) {
		args := map[string]any{
			"project_root":   projectRoot,
			"source_pattern": "*order-service*",
		}

		result, err := tool.Call(ctx, args)
		if err != nil {
			t.Fatalf("find_relationships failed: %v", err)
		}

		response, ok := result.(*entities.FindRelationshipsResponse)
		if !ok {
			t.Fatalf("expected FindRelationshipsResponse, got %T", result)
		}

		// Order service has 2 relationships
		if len(response.Relationships) != 2 {
			t.Errorf("expected 2 relationships from order-service, got %d", len(response.Relationships))
		}

		// Verify all returned relationships have order-service as source
		for _, rel := range response.Relationships {
			if rel.SourceID != "e-commerce/backend-services/order-service" {
				t.Errorf("expected source to be e-commerce/backend-services/order-service, got %q", rel.SourceID)
			}
		}
	})

	// Test 3: Find relationships with target filter
	t.Run("filter by target pattern", func(t *testing.T) {
		args := map[string]any{
			"project_root":   projectRoot,
			"target_pattern": "*inventory*",
		}

		result, err := tool.Call(ctx, args)
		if err != nil {
			t.Fatalf("find_relationships failed: %v", err)
		}

		response, ok := result.(*entities.FindRelationshipsResponse)
		if !ok {
			t.Fatalf("expected FindRelationshipsResponse, got %T", result)
		}

		// Should find: order→inventory-service, inventory-service→inventory-db
		if len(response.Relationships) < 1 {
			t.Errorf("expected at least 1 relationship to inventory components, got %d", len(response.Relationships))
		}
	})

	// Test 4: Verify relationship type filter
	t.Run("filter by relationship type", func(t *testing.T) {
		args := map[string]any{
			"project_root":      projectRoot,
			"source_pattern":    "*", // Need at least one pattern
			"relationship_type": "depends-on",
		}

		result, err := tool.Call(ctx, args)
		if err != nil {
			t.Fatalf("find_relationships failed: %v", err)
		}

		response, ok := result.(*entities.FindRelationshipsResponse)
		if !ok {
			t.Fatalf("expected FindRelationshipsResponse, got %T", result)
		}

		if len(response.Relationships) == 0 {
			t.Error("expected depends-on relationships, got none")
		}

		// Verify all relationships have correct type
		for _, rel := range response.Relationships {
			if rel.Type != "depends-on" {
				t.Errorf("expected type 'depends-on', got %q", rel.Type)
			}
		}
	})

	// Test 5: Verify relationship descriptions are included
	t.Run("relationship descriptions included", func(t *testing.T) {
		args := map[string]any{
			"project_root":   projectRoot,
			"source_pattern": "*order-service*",
		}

		result, err := tool.Call(ctx, args)
		if err != nil {
			t.Fatalf("find_relationships failed: %v", err)
		}

		response, ok := result.(*entities.FindRelationshipsResponse)
		if !ok {
			t.Fatalf("expected FindRelationshipsResponse, got %T", result)
		}

		// Verify descriptions are populated
		foundDescriptions := 0
		for _, rel := range response.Relationships {
			if rel.Description != "" {
				foundDescriptions++
			}
		}

		if foundDescriptions == 0 {
			t.Error("expected at least some relationships to have descriptions")
		}
	})
}

// TestMCPQueryDependencies_WithRelationships verifies that query_dependencies
// (via AnalyzeDependencies) shows graph connections from relationships.
// This is T027 verification.
func TestMCPQueryDependencies_WithRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Create test project with relationships
	setupTestProjectWithRelationships(t, ctx, repo, projectRoot)

	// Build graph and analyze dependencies
	project, err := repo.LoadProject(ctx, projectRoot)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	systems, err := repo.ListSystems(ctx, projectRoot)
	if err != nil {
		t.Fatalf("failed to list systems: %v", err)
	}

	buildGraph := usecases.NewBuildArchitectureGraph()
	graph, err := buildGraph.Execute(ctx, project, systems)
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	report := buildGraph.AnalyzeDependencies(graph)

	// Verify dependency analysis includes our relationships
	if report.TotalEdges == 0 {
		t.Error("expected TotalEdges > 0 from relationships, got 0")
	}

	// We created 4 components, none should be isolated (all have relationships)
	// except payment-service which is only a target
	if len(report.IsolatedComponents) > 2 {
		t.Errorf("expected at most 2 isolated components, got %d: %v",
			len(report.IsolatedComponents), report.IsolatedComponents)
	}

	// Verify component counts
	if report.ComponentsCount != 4 {
		t.Errorf("expected 4 components, got %d", report.ComponentsCount)
	}

	if report.TotalNodes < 6 {
		t.Errorf("expected at least 6 nodes (system+container+4 components), got %d", report.TotalNodes)
	}
}

// setupTestProjectWithRelationships creates a test project with components
// that have frontmatter relationships for MCP tool testing.
func setupTestProjectWithRelationships(t *testing.T, ctx context.Context, repo *filesystem.ProjectRepository, projectRoot string) {
	t.Helper()

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

	// Create components with relationships (same as T022 test)
	orderService, _ := entities.NewComponent("Order Service")
	orderService.SetDescription("Manages customer orders")
	orderService.AddRelationship("inventory-service", "Checks product availability")
	orderService.AddRelationship("payment-service", "Processes payments")

	inventoryService, _ := entities.NewComponent("Inventory Service")
	inventoryService.SetDescription("Manages product inventory")
	inventoryService.AddRelationship("inventory-db", "Reads/writes inventory data")

	paymentService, _ := entities.NewComponent("Payment Service")
	paymentService.SetDescription("Payment processing")

	inventoryDB, _ := entities.NewComponent("Inventory DB")
	inventoryDB.SetDescription("PostgreSQL database")

	// Save all components
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, orderService)
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, inventoryService)
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, paymentService)
	repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, inventoryDB)
}
