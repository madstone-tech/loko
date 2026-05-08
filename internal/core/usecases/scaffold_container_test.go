package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestScaffoldEntityExecuteContainer tests scaffolding a container.
func TestScaffoldEntityExecuteContainer(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("Payment Service")

	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}
	mockRepo.LoadSystemFunc = func(ctx context.Context, projectRoot, systemName string) (*entities.System, error) {
		return system, nil
	}
	// Note: The existing MockProjectRepository doesn't have SaveContainerFunc field,
	// but the SaveContainer method exists and returns nil by default

	uc := NewScaffoldEntity(mockRepo)

	req := &ScaffoldEntityRequest{
		ProjectRoot: "/test/project",
		EntityType:  "container",
		ParentPath:  []string{"Payment Service"},
		Name:        "API Server",
		Description: "REST API endpoints",
		Technology:  "Go + gRPC",
		Tags:        []string{"api", "backend"},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	if result.EntityID != "api-server" {
		t.Errorf("expected entity ID 'api-server', got %q", result.EntityID)
	}

	if len(result.FilesCreated) == 0 {
		t.Error("expected files to be created")
	}
}
