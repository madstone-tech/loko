package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestScaffoldEntityExecuteSystem tests scaffolding a system.
func TestScaffoldEntityExecuteSystem(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}
	mockRepo.SaveSystemFunc = func(ctx context.Context, projectRoot string, system *entities.System) error {
		return nil
	}

	uc := NewScaffoldEntity(mockRepo)

	req := &ScaffoldEntityRequest{
		ProjectRoot: "/test/project",
		EntityType:  "system",
		Name:        "Payment Service",
		Description: "Handles payment processing",
		Tags:        []string{"finance", "critical"},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	if result.EntityID != "payment-service" {
		t.Errorf("expected entity ID 'payment-service', got %q", result.EntityID)
	}

	if len(result.FilesCreated) == 0 {
		t.Error("expected files to be created")
	}
}
