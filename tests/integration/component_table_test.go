package integration_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestGenerateComponentTable_Integration tests the GenerateComponentTable function
// in an integration context with real filesystem operations.
func TestGenerateComponentTable_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-proj")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Scaffold a minimal project
	project, err := entities.NewProject("test-proj")
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	project.Path = projectRoot
	if err := repo.SaveProject(ctx, project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	system, err := entities.NewSystem("payments")
	if err != nil {
		t.Fatalf("NewSystem: %v", err)
	}
	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("SaveSystem: %v", err)
	}

	container, err := entities.NewContainer("api")
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	container.SetTechnology("Go + Fiber")
	container.SetDescription("REST API for payment processing")
	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	// Create components
	authComp, err := entities.NewComponent("Authentication Service")
	if err != nil {
		t.Fatalf("NewComponent: %v", err)
	}
	authComp.Technology = "Go package"
	authComp.Description = "Handles user authentication"
	if err := container.AddComponent(authComp); err != nil {
		t.Fatalf("AddComponent: %v", err)
	}

	paymentComp, err := entities.NewComponent("Payment Processor")
	if err != nil {
		t.Fatalf("NewComponent: %v", err)
	}
	paymentComp.Technology = "Stripe API"
	paymentComp.Description = "Processes payment transactions"
	if err := container.AddComponent(paymentComp); err != nil {
		t.Fatalf("AddComponent: %v", err)
	}

	// Generate component table
	result := usecases.GenerateComponentTable(container)

	// Verify the table is correctly formatted
	expectedHeader := "| Name | Technology | Description |"
	expectedSeparator := "|------|------------|-------------|"
	expectedRow1 := "| Authentication Service | Go package | Handles user authentication |"
	expectedRow2 := "| Payment Processor | Stripe API | Processes payment transactions |"

	if !containsSubstring(result, expectedHeader) {
		t.Errorf("Missing header in result:\n%s", result)
	}
	if !containsSubstring(result, expectedSeparator) {
		t.Errorf("Missing separator in result:\n%s", result)
	}
	if !containsSubstring(result, expectedRow1) {
		t.Errorf("Missing first row in result:\n%s", result)
	}
	if !containsSubstring(result, expectedRow2) {
		t.Errorf("Missing second row in result:\n%s", result)
	}

	// Verify ordering (should be alphabetical by name)
	lines := splitLines(result)
	authLineIndex := findLineIndex(lines, "Authentication Service")
	paymentLineIndex := findLineIndex(lines, "Payment Processor")

	if authLineIndex >= paymentLineIndex {
		t.Errorf("Components should be sorted alphabetically. Auth line index: %d, Payment line index: %d", authLineIndex, paymentLineIndex)
	}
}

// TestGenerateContainerTable_Integration tests the GenerateContainerTable function
// in an integration context.
func TestGenerateContainerTable_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-proj")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Scaffold a minimal project
	project, err := entities.NewProject("test-proj")
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	project.Path = projectRoot
	if err := repo.SaveProject(ctx, project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	system, err := entities.NewSystem("ecommerce-platform")
	if err != nil {
		t.Fatalf("NewSystem: %v", err)
	}
	system.SetDescription("E-commerce platform system")
	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("SaveSystem: %v", err)
	}

	// Create containers
	apiContainer, err := entities.NewContainer("API Server")
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	apiContainer.SetTechnology("Go + Fiber")
	apiContainer.SetDescription("REST API for frontend clients")
	if err := system.AddContainer(apiContainer); err != nil {
		t.Fatalf("AddContainer: %v", err)
	}

	dbContainer, err := entities.NewContainer("Database")
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	dbContainer.SetTechnology("PostgreSQL 15")
	dbContainer.SetDescription("Primary data store")
	if err := system.AddContainer(dbContainer); err != nil {
		t.Fatalf("AddContainer: %v", err)
	}

	cacheContainer, err := entities.NewContainer("Cache")
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	cacheContainer.SetTechnology("Redis 7")
	cacheContainer.SetDescription("In-memory cache layer")
	if err := system.AddContainer(cacheContainer); err != nil {
		t.Fatalf("AddContainer: %v", err)
	}

	// Generate container table
	result := usecases.GenerateContainerTable(system)

	// Verify the table is correctly formatted
	expectedHeader := "| Name | Technology | Description |"
	expectedSeparator := "|------|------------|-------------|"
	expectedRow1 := "| API Server | Go + Fiber | REST API for frontend clients |"
	expectedRow2 := "| Cache | Redis 7 | In-memory cache layer |"
	expectedRow3 := "| Database | PostgreSQL 15 | Primary data store |"

	if !containsSubstring(result, expectedHeader) {
		t.Errorf("Missing header in result:\n%s", result)
	}
	if !containsSubstring(result, expectedSeparator) {
		t.Errorf("Missing separator in result:\n%s", result)
	}
	if !containsSubstring(result, expectedRow1) {
		t.Errorf("Missing first row in result:\n%s", result)
	}
	if !containsSubstring(result, expectedRow2) {
		t.Errorf("Missing second row in result:\n%s", result)
	}
	if !containsSubstring(result, expectedRow3) {
		t.Errorf("Missing third row in result:\n%s", result)
	}

	// Verify ordering (should be alphabetical by name)
	lines := splitLines(result)
	apiLineIndex := findLineIndex(lines, "API Server")
	cacheLineIndex := findLineIndex(lines, "Cache")
	dbLineIndex := findLineIndex(lines, "Database")

	if apiLineIndex >= cacheLineIndex || cacheLineIndex >= dbLineIndex {
		t.Errorf("Containers should be sorted alphabetically. API: %d, Cache: %d, DB: %d",
			apiLineIndex, cacheLineIndex, dbLineIndex)
	}
}

// Helper functions for testing
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func findLineIndex(lines []string, substr string) int {
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i
		}
	}
	return -1
}
