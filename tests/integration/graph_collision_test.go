package integration

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestMultiSystemCollisionScenario tests the complete scenario of multiple systems
// with components that have the same name, ensuring no silent failures occur.
func TestMultiSystemCollisionScenario(t *testing.T) {
	// Create test project
	project, err := entities.NewProject("multi-system-test")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// System 1: Backend with API container and Auth component
	backend, err := entities.NewSystem("Backend")
	if err != nil {
		t.Fatalf("failed to create backend system: %v", err)
	}

	backendAPI, err := entities.NewContainer("API")
	if err != nil {
		t.Fatalf("failed to create backend API container: %v", err)
	}

	backendAuth, err := entities.NewComponent("Auth")
	if err != nil {
		t.Fatalf("failed to create backend auth component: %v", err)
	}

	backendDB, err := entities.NewComponent("Database")
	if err != nil {
		t.Fatalf("failed to create backend database component: %v", err)
	}

	// Backend auth depends on backend database
	backendAuth.AddRelationship(backendDB.ID, "stores user credentials")

	backendAPI.AddComponent(backendAuth)
	backendAPI.AddComponent(backendDB)
	backend.AddContainer(backendAPI)

	// System 2: Admin with UI container and Auth component (same name as backend auth)
	admin, err := entities.NewSystem("Admin")
	if err != nil {
		t.Fatalf("failed to create admin system: %v", err)
	}

	adminUI, err := entities.NewContainer("UI")
	if err != nil {
		t.Fatalf("failed to create admin UI container: %v", err)
	}

	adminAuth, err := entities.NewComponent("Auth")
	if err != nil {
		t.Fatalf("failed to create admin auth component: %v", err)
	}

	// Admin auth depends on backend auth (cross-system relationship)
	adminAuth.AddRelationship(backendAuth.ID, "delegates to backend authentication")

	adminUI.AddComponent(adminAuth)
	admin.AddContainer(adminUI)

	// Build the architecture graph
	uc := usecases.NewBuildArchitectureGraph()
	graph, err := uc.Execute(context.Background(), project, []*entities.System{backend, admin})
	if err != nil {
		t.Fatalf("failed to build architecture graph: %v", err)
	}

	// Verify: Both components should exist as distinct nodes
	t.Run("both auth components exist", func(t *testing.T) {
		expectedNodeCount := 6 // 2 systems + 2 containers + 2 auth components (+ 1 DB component = 7 total)
		// Wait, let me recount: Backend system, Admin system, Backend API container, Admin UI container,
		// Backend Auth component, Backend DB component, Admin Auth component = 7 nodes
		expectedNodeCount = 7

		if graph.Size() != expectedNodeCount {
			t.Errorf("expected %d nodes in graph, got %d (both auth components should exist)", expectedNodeCount, graph.Size())
		}

		// Verify backend auth component exists with qualified ID
		backendAuthQualifiedID := entities.QualifiedNodeID("component", backend.ID, backendAPI.ID, backendAuth.ID)
		backendAuthNode := graph.GetNode(backendAuthQualifiedID)
		if backendAuthNode == nil {
			t.Errorf("backend auth component not found at ID %q - silent collision occurred", backendAuthQualifiedID)
		}

		// Verify admin auth component exists with qualified ID
		adminAuthQualifiedID := entities.QualifiedNodeID("component", admin.ID, adminUI.ID, adminAuth.ID)
		adminAuthNode := graph.GetNode(adminAuthQualifiedID)
		if adminAuthNode == nil {
			t.Errorf("admin auth component not found at ID %q - silent collision occurred", adminAuthQualifiedID)
		}
	})

	// Verify: Relationship edges exist between qualified IDs
	t.Run("relationships are correctly wired", func(t *testing.T) {
		backendAuthQualifiedID := entities.QualifiedNodeID("component", backend.ID, backendAPI.ID, backendAuth.ID)
		backendDBQualifiedID := entities.QualifiedNodeID("component", backend.ID, backendAPI.ID, backendDB.ID)

		// Backend auth should have dependency on backend database
		backendAuthDeps := graph.GetDependencies(backendAuthQualifiedID)
		if len(backendAuthDeps) != 1 {
			t.Errorf("backend auth should have 1 dependency, got %d", len(backendAuthDeps))
		}

		if len(backendAuthDeps) > 0 && backendAuthDeps[0].ID != backendDBQualifiedID {
			t.Errorf("backend auth should depend on %q, got %q", backendDBQualifiedID, backendAuthDeps[0].ID)
		}

		// Admin auth should have dependency on backend auth (cross-system)
		adminAuthQualifiedID := entities.QualifiedNodeID("component", admin.ID, adminUI.ID, adminAuth.ID)
		adminAuthDeps := graph.GetDependencies(adminAuthQualifiedID)
		if len(adminAuthDeps) != 1 {
			t.Errorf("admin auth should have 1 dependency, got %d", len(adminAuthDeps))
		}

		if len(adminAuthDeps) > 0 && adminAuthDeps[0].ID != backendAuthQualifiedID {
			t.Errorf("admin auth should depend on %q, got %q", backendAuthQualifiedID, adminAuthDeps[0].ID)
		}
	})

	// Verify: GetNode works with both qualified and short IDs
	t.Run("short ID resolution works", func(t *testing.T) {
		// ResolveID should work for unambiguous short IDs
		dbQualifiedID, ok := graph.ResolveID("database")
		if !ok {
			t.Error("ResolveID failed for unambiguous short ID 'database'")
		}

		expectedDBID := entities.QualifiedNodeID("component", backend.ID, backendAPI.ID, backendDB.ID)
		if dbQualifiedID != expectedDBID {
			t.Errorf("ResolveID('database') = %q, want %q", dbQualifiedID, expectedDBID)
		}

		// ResolveID should handle ambiguous short ID "auth" (appears in both systems)
		// Implementation may return one or handle ambiguity differently
		_, ok = graph.ResolveID("auth")
		// We just document that it can be resolved to one of them
		// The exact behavior depends on implementation choice
		if ok {
			// If it resolves, it should be one of the two auth components
			t.Log("ResolveID for ambiguous 'auth' resolved (implementation-specific behavior)")
		}
	})

	// Verify: Validation passes without errors
	t.Run("graph validation passes", func(t *testing.T) {
		if err := graph.Validate(); err != nil {
			t.Errorf("graph validation failed: %v", err)
		}
	})

	// Verify: No silent failures - check edge count
	t.Run("all edges created", func(t *testing.T) {
		expectedEdgeCount := 2 // backend auth -> backend db, admin auth -> backend auth
		if graph.EdgeCount() != expectedEdgeCount {
			t.Errorf("expected %d edges, got %d", expectedEdgeCount, graph.EdgeCount())
		}
	})
}
