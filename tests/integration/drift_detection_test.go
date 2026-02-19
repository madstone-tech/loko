package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestDriftDetection_NoDrift tests drift detection with consistent components.
func TestDriftDetection_NoDrift(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	// Create project directory structure
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

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
	system, err := entities.NewSystem("Backend")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	system.Path = filepath.Join(projectRoot, "backend")

	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}

	// Create container
	container, err := entities.NewContainer("API")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	container.Path = filepath.Join(projectRoot, "backend", "api")

	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}

	// Create components with consistent relationships
	auth, err := entities.NewComponent("Auth")
	if err != nil {
		t.Fatalf("failed to create auth component: %v", err)
	}
	auth.SetDescription("Authentication service")
	auth.Path = filepath.Join(projectRoot, "backend", "api", "auth")

	db, err := entities.NewComponent("Database")
	if err != nil {
		t.Fatalf("failed to create db component: %v", err)
	}
	db.SetDescription("Database service")
	db.Path = filepath.Join(projectRoot, "backend", "api", "database")

	// Add relationship between existing components
	auth.AddRelationship(db.ID, "queries user data")

	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, auth); err != nil {
		t.Fatalf("failed to save auth component: %v", err)
	}

	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, db); err != nil {
		t.Fatalf("failed to save db component: %v", err)
	}

	// Run drift detection
	uc := usecases.NewDetectDrift(repo)
	req := &usecases.DetectDriftRequest{
		ProjectRoot: projectRoot,
	}

	result, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify no drift issues
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d: %+v", len(result.Issues), result.Issues)
	}

	if result.HasErrors {
		t.Error("expected HasErrors to be false")
	}

	if result.ComponentsChecked == 0 {
		t.Error("expected ComponentsChecked to be > 0")
	}
}

// TestDriftDetection_OrphanedRelationship tests detection of orphaned relationships.
func TestDriftDetection_OrphanedRelationship(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	// Create project directory structure
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

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
	system, err := entities.NewSystem("Backend")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	system.Path = filepath.Join(projectRoot, "backend")

	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}

	// Create container
	container, err := entities.NewContainer("API")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	container.Path = filepath.Join(projectRoot, "backend", "api")

	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}

	// Create component with relationship to non-existent component
	auth, err := entities.NewComponent("Auth")
	if err != nil {
		t.Fatalf("failed to create auth component: %v", err)
	}
	auth.SetDescription("Authentication service")
	auth.Path = filepath.Join(projectRoot, "backend", "api", "auth")

	// Add relationship to deleted component
	auth.AddRelationship("deleted-comp", "calls")

	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, auth); err != nil {
		t.Fatalf("failed to save auth component: %v", err)
	}

	// Run drift detection
	uc := usecases.NewDetectDrift(repo)
	req := &usecases.DetectDriftRequest{
		ProjectRoot: projectRoot,
	}

	result, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check that we have an orphaned relationship issue
	foundOrphaned := false
	for _, issue := range result.Issues {
		if issue.Type == entities.DriftOrphanedRelationship {
			foundOrphaned = true
			break
		}
	}

	if !foundOrphaned {
		t.Error("expected DriftOrphanedRelationship issue, but none found")
	}

	if !result.HasErrors {
		t.Error("expected HasErrors to be true")
	}

	if result.ComponentsChecked != 1 {
		t.Errorf("expected ComponentsChecked to be 1, got %d", result.ComponentsChecked)
	}
}

// TestDriftDetection_HasErrorsFlag tests the HasErrors flag behavior.
func TestDriftDetection_HasErrorsFlag(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")

	// Create project directory structure
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

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
	system, err := entities.NewSystem("Backend")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	system.Path = filepath.Join(projectRoot, "backend")

	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}

	// Create container
	container, err := entities.NewContainer("API")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	container.Path = filepath.Join(projectRoot, "backend", "api")

	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}

	// Create component with relationship to non-existent component
	auth, err := entities.NewComponent("Auth")
	if err != nil {
		t.Fatalf("failed to create auth component: %v", err)
	}
	auth.SetDescription("Authentication service")
	auth.Path = filepath.Join(projectRoot, "backend", "api", "auth")

	// Add relationship to deleted component
	auth.AddRelationship("deleted-comp", "calls")

	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, auth); err != nil {
		t.Fatalf("failed to save auth component: %v", err)
	}

	// Run drift detection
	uc := usecases.NewDetectDrift(repo)
	req := &usecases.DetectDriftRequest{
		ProjectRoot: projectRoot,
	}

	result, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify HasErrors flag is true
	if !result.HasErrors {
		t.Error("expected HasErrors to be true when orphaned relationships exist")
	}

	// Verify HasWarnings flag is false (no warnings in this test)
	if result.HasWarnings {
		t.Error("expected HasWarnings to be false")
	}
}
