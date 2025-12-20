package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestComponentPersistence verifies that components are persisted to disk correctly.
func TestComponentPersistence(t *testing.T) {
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
	system, err := entities.NewSystem("API")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	
	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}
	
	// Create container
	container, err := entities.NewContainer("Server")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	container.SetTechnology("Go")
	
	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}
	
	// Create component
	component, err := entities.NewComponent("Auth Handler")
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}
	component.SetDescription("Handles JWT authentication")
	component.SetTechnology("Go middleware")
	component.AddTag("security")
	component.AddTag("auth")
	
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, component); err != nil {
		t.Fatalf("failed to save component: %v", err)
	}
	
	// Verify component directory exists
	componentDir := filepath.Join(projectRoot, "src", system.ID, container.ID, component.ID)
	if _, err := os.Stat(componentDir); err != nil {
		t.Fatalf("component directory not created: %v", err)
	}
	
	// Verify component.md exists
	componentMdPath := filepath.Join(componentDir, "component.md")
	if _, err := os.Stat(componentMdPath); err != nil {
		t.Fatalf("component.md not created: %v", err)
	}
	
	// Verify D2 template exists
	d2Path := filepath.Join(componentDir, "auth-handler.d2")
	if _, err := os.Stat(d2Path); err != nil {
		t.Fatalf("component D2 template not created: %v", err)
	}
	
	// Read and verify D2 template content
	content, err := os.ReadFile(d2Path)
	if err != nil {
		t.Fatalf("failed to read D2 template: %v", err)
	}
	
	contentStr := string(content)
	if !contains(contentStr, "Auth Handler") {
		t.Errorf("D2 template missing component name")
	}
	if !contains(contentStr, "C4 Level 3") {
		t.Errorf("D2 template missing C4 Level 3 marker")
	}
	
	// Load component back
	loadedComponent, err := repo.LoadComponent(ctx, projectRoot, system.ID, container.ID, component.ID)
	if err != nil {
		t.Fatalf("failed to load component: %v", err)
	}
	
	if loadedComponent.Name != "Auth Handler" {
		t.Errorf("expected name 'Auth Handler', got %q", loadedComponent.Name)
	}
	
	if loadedComponent.Technology != "Go middleware" {
		t.Errorf("expected technology 'Go middleware', got %q", loadedComponent.Technology)
	}
	
	if loadedComponent.Description != "Handles JWT authentication" {
		t.Errorf("expected description 'Handles JWT authentication', got %q", loadedComponent.Description)
	}
	
	if len(loadedComponent.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(loadedComponent.Tags))
	}
}

// TestComponentLoadingFromContainer verifies that components are loaded when loading a container.
func TestComponentLoadingFromContainer(t *testing.T) {
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
	system, err := entities.NewSystem("API")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	
	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}
	
	// Create container
	container, err := entities.NewContainer("Server")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	
	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}
	
	// Create and save multiple components
	components := []string{"Auth Handler", "Logger", "Database Client"}
	for _, compName := range components {
		component, err := entities.NewComponent(compName)
		if err != nil {
			t.Fatalf("failed to create component: %v", err)
		}
		
		if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, component); err != nil {
			t.Fatalf("failed to save component: %v", err)
		}
	}
	
	// Load container
	loadedContainer, err := repo.LoadContainer(ctx, projectRoot, system.ID, container.ID)
	if err != nil {
		t.Fatalf("failed to load container: %v", err)
	}
	
	// Verify all components were loaded
	if loadedContainer.ComponentCount() != 3 {
		t.Errorf("expected 3 components, got %d", loadedContainer.ComponentCount())
	}
	
	loadedComps := loadedContainer.ListComponents()
	if len(loadedComps) != 3 {
		t.Errorf("expected 3 components in list, got %d", len(loadedComps))
	}
	
	// Verify component names
	names := make(map[string]bool)
	for _, comp := range loadedComps {
		names[comp.Name] = true
	}
	
	for _, expectedName := range components {
		if !names[expectedName] {
			t.Errorf("expected component %q not found", expectedName)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestComponentRelationships verifies that component relationships are persisted correctly.
func TestComponentRelationships(t *testing.T) {
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
	system, err := entities.NewSystem("API")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	
	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("failed to save system: %v", err)
	}
	
	// Create container
	container, err := entities.NewContainer("Server")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	
	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("failed to save container: %v", err)
	}
	
	// Create components
	authHandler, err := entities.NewComponent("Auth Handler")
	if err != nil {
		t.Fatalf("failed to create auth handler: %v", err)
	}
	authHandler.SetDescription("Handles JWT authentication")
	
	userService, err := entities.NewComponent("User Service")
	if err != nil {
		t.Fatalf("failed to create user service: %v", err)
	}
	userService.SetDescription("Manages user data")
	
	// Add relationship: Auth Handler depends on User Service
	authHandler.AddRelationship("user-service", "validates users via")
	
	// Save components
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, authHandler); err != nil {
		t.Fatalf("failed to save auth handler: %v", err)
	}
	
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, userService); err != nil {
		t.Fatalf("failed to save user service: %v", err)
	}
	
	// Load auth handler back
	loadedAuthHandler, err := repo.LoadComponent(ctx, projectRoot, system.ID, container.ID, authHandler.ID)
	if err != nil {
		t.Fatalf("failed to load auth handler: %v", err)
	}
	
	// Verify relationship was persisted
	if loadedAuthHandler.RelationshipCount() != 1 {
		t.Errorf("expected 1 relationship, got %d", loadedAuthHandler.RelationshipCount())
	}
	
	// Verify relationship content
	desc, exists := loadedAuthHandler.GetRelationship("user-service")
	if !exists {
		t.Errorf("expected relationship to 'user-service' not found")
	}
	
	if desc != "validates users via" {
		t.Errorf("expected relationship description 'validates users via', got %q", desc)
	}
}

// TestComponentAnnotationsAndDependencies verifies component code annotations and dependencies are persisted.
func TestComponentAnnotationsAndDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-project")
	
	repo := filesystem.NewProjectRepository()
	ctx := context.Background()
	
	// Setup
	project, _ := entities.NewProject("test-project")
	project.Path = projectRoot
	repo.SaveProject(ctx, project)
	
	system, _ := entities.NewSystem("API")
	repo.SaveSystem(ctx, projectRoot, system)
	
	container, _ := entities.NewContainer("Server")
	repo.SaveContainer(ctx, projectRoot, system.ID, container)
	
	// Create component with annotations and dependencies
	component, _ := entities.NewComponent("Auth Handler")
	component.SetDescription("Handles JWT authentication")
	
	// Add code annotations
	component.AddCodeAnnotation("internal/auth", "JWT token validation")
	component.AddCodeAnnotation("internal/middleware", "HTTP middleware for auth checks")
	
	// Add dependencies
	component.AddDependency("github.com/golang-jwt/jwt")
	component.AddDependency("github.com/gorilla/mux")
	
	// Save component
	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, component); err != nil {
		t.Fatalf("failed to save component: %v", err)
	}
	
	// Load component back
	loadedComponent, err := repo.LoadComponent(ctx, projectRoot, system.ID, container.ID, component.ID)
	if err != nil {
		t.Fatalf("failed to load component: %v", err)
	}
	
	// Verify annotations were persisted
	if loadedComponent.CodeAnnotationCount() != 2 {
		t.Errorf("expected 2 annotations, got %d", loadedComponent.CodeAnnotationCount())
	}
	
	if desc, exists := loadedComponent.GetCodeAnnotation("internal/auth"); !exists || desc != "JWT token validation" {
		t.Errorf("expected annotation for 'internal/auth', got %q", desc)
	}
	
	// Verify dependencies were persisted
	if loadedComponent.DependencyCount() != 2 {
		t.Errorf("expected 2 dependencies, got %d", loadedComponent.DependencyCount())
	}
	
	if !loadedComponent.HasDependency("github.com/golang-jwt/jwt") {
		t.Errorf("expected dependency 'github.com/golang-jwt/jwt' not found")
	}
	
	if !loadedComponent.HasDependency("github.com/gorilla/mux") {
		t.Errorf("expected dependency 'github.com/gorilla/mux' not found")
	}
}
