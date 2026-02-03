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

// TestInitProjectWorkflow tests the full init → new system workflow.
func TestInitProjectWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "myproject")

	// Step 1: Create project directory
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Step 2: Initialize project with loko.toml
	project, err := entities.NewProject("myproject")
	if err != nil {
		t.Fatalf("NewProject() error = %v", err)
	}

	project.Path = projectRoot
	project.Description = "Test project"
	project.Version = "0.1.0"

	// Step 3: Save project (creates loko.toml)
	repo := filesystem.NewProjectRepository()
	if err := repo.SaveProject(context.Background(), project); err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}

	// Verify loko.toml was created
	configPath := filepath.Join(projectRoot, "loko.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("loko.toml not created: %v", err)
	}

	// Step 4: Create a system
	system, err := entities.NewSystem("Payment Service")
	if err != nil {
		t.Fatalf("NewSystem() error = %v", err)
	}

	system.Description = "Handles payment processing"
	system.Path = filepath.Join(projectRoot, "src", system.ID)

	// Step 5: Add system to project
	if err := project.AddSystem(system); err != nil {
		t.Fatalf("AddSystem() error = %v", err)
	}

	// Step 6: Save system
	if err := repo.SaveSystem(context.Background(), projectRoot, system); err != nil {
		t.Fatalf("SaveSystem() error = %v", err)
	}

	// Verify system.md was created
	systemMdPath := filepath.Join(projectRoot, "src", system.ID, "system.md")
	if _, err := os.Stat(systemMdPath); err != nil {
		t.Errorf("system.md not created: %v", err)
	}

	// Step 7: Create a container
	container, err := entities.NewContainer("API")
	if err != nil {
		t.Fatalf("NewContainer() error = %v", err)
	}

	container.Description = "REST API"
	container.Technology = "Go + Fiber"
	container.Path = filepath.Join(projectRoot, "src", system.ID, container.ID)

	// Step 8: Add container to system
	if err := system.AddContainer(container); err != nil {
		t.Fatalf("AddContainer() error = %v", err)
	}

	// Step 9: Save container
	if err := repo.SaveContainer(context.Background(), projectRoot, system.ID, container); err != nil {
		t.Fatalf("SaveContainer() error = %v", err)
	}

	// Verify container.md was created
	containerMdPath := filepath.Join(projectRoot, "src", system.ID, container.ID, "container.md")
	if _, err := os.Stat(containerMdPath); err != nil {
		t.Errorf("container.md not created: %v", err)
	}

	// Step 10: Verify directory structure
	expectedDirs := []string{
		filepath.Join(projectRoot, "src"),
		filepath.Join(projectRoot, "src", system.ID),
		filepath.Join(projectRoot, "src", system.ID, container.ID),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("expected directory not created: %s", dir)
		}
	}

	// Step 11: Load project back and verify
	loadedProject, err := repo.LoadProject(context.Background(), projectRoot)
	if err != nil {
		t.Fatalf("LoadProject() error = %v", err)
	}

	if loadedProject.Name != "myproject" {
		t.Errorf("expected project name 'myproject', got %q", loadedProject.Name)
	}

	if loadedProject.SystemCount() != 1 {
		t.Errorf("expected 1 system, got %d", loadedProject.SystemCount())
	}

	loadedSystem, err := loadedProject.GetSystem(system.ID)
	if err != nil {
		t.Fatalf("GetSystem() error = %v", err)
	}

	if loadedSystem.Name != "Payment Service" {
		t.Errorf("expected system name 'Payment Service', got %q", loadedSystem.Name)
	}

	if loadedSystem.ContainerCount() != 1 {
		t.Errorf("expected 1 container, got %d", loadedSystem.ContainerCount())
	}
}

// TestCreateSystemUseCase tests the CreateSystem use case in integration.
func TestCreateSystemUseCase(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "myproject")

	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	repo := filesystem.NewProjectRepository()

	// Create use case
	uc := usecases.NewCreateSystem(repo)

	// Execute
	system, err := uc.Execute(context.Background(), &usecases.CreateSystemRequest{
		Name:        "Auth Service",
		Description: "Handles authentication",
		Tags:        []string{"security"},
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if system.Name != "Auth Service" {
		t.Errorf("expected name 'Auth Service', got %q", system.Name)
	}

	if system.ID != "auth-service" {
		t.Errorf("expected ID 'auth-service', got %q", system.ID)
	}

	if len(system.Tags) != 1 || system.Tags[0] != "security" {
		t.Errorf("expected tags [security], got %v", system.Tags)
	}
}

// TestMultipleSystemsWorkflow tests creating multiple systems.
func TestMultipleSystemsWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "myproject")

	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	project, err := entities.NewProject("myproject")
	if err != nil {
		t.Fatalf("NewProject() error = %v", err)
	}

	project.Path = projectRoot

	repo := filesystem.NewProjectRepository()

	// Create multiple systems
	systemNames := []string{"Payment Service", "Auth Service", "Notification Service"}
	for _, name := range systemNames {
		sys, err := entities.NewSystem(name)
		if err != nil {
			t.Fatalf("NewSystem(%q) error = %v", name, err)
		}

		sys.Path = filepath.Join(projectRoot, "src", sys.ID)

		if err := project.AddSystem(sys); err != nil {
			t.Fatalf("AddSystem(%q) error = %v", name, err)
		}

		if err := repo.SaveSystem(context.Background(), projectRoot, sys); err != nil {
			t.Fatalf("SaveSystem(%q) error = %v", name, err)
		}
	}

	if project.SystemCount() != 3 {
		t.Errorf("expected 3 systems, got %d", project.SystemCount())
	}

	// Verify all systems were saved
	for _, name := range systemNames {
		sys, err := entities.NewSystem(name)
		if err != nil {
			t.Fatalf("NewSystem(%q) error = %v", name, err)
		}

		systemMdPath := filepath.Join(projectRoot, "src", sys.ID, "system.md")
		if _, err := os.Stat(systemMdPath); err != nil {
			t.Errorf("system.md not created for %q: %v", name, err)
		}
	}
}
