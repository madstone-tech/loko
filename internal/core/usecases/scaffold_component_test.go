package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestScaffoldEntityExecuteComponent tests scaffolding a component.
func TestScaffoldEntityExecuteComponent(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("Payment Service")
	container, _ := entities.NewContainer("API Server")
	system.AddContainer(container)

	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}
	mockRepo.LoadSystemFunc = func(ctx context.Context, projectRoot, systemName string) (*entities.System, error) {
		return system, nil
	}
	// Note: The existing MockProjectRepository doesn't have SaveComponentFunc field,
	// but the SaveComponent method exists and returns nil by default

	uc := NewScaffoldEntity(mockRepo)

	req := &ScaffoldEntityRequest{
		ProjectRoot: "/test/project",
		EntityType:  "component",
		ParentPath:  []string{"Payment Service", "API Server"},
		Name:        "Auth Handler",
		Description: "Handles authentication",
		Technology:  "Go",
		Tags:        []string{"security", "core"},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	if result.EntityID != "auth-handler" {
		t.Errorf("expected entity ID 'auth-handler', got %q", result.EntityID)
	}

	if len(result.FilesCreated) == 0 {
		t.Error("expected files to be created")
	}
}
