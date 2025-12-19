package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// MockProjectRepository is a concrete mock for testing.
type MockProjectRepository struct {
	SaveProjectFunc func(ctx context.Context, project *entities.Project) error
	LoadProjectFunc func(ctx context.Context, projectRoot string) (*entities.Project, error)
	SaveSystemFunc  func(ctx context.Context, projectRoot string, system *entities.System) error
	ListSystemsFunc func(ctx context.Context, projectRoot string) ([]*entities.System, error)
	LoadSystemFunc  func(ctx context.Context, projectRoot, systemName string) (*entities.System, error)
}

func (m *MockProjectRepository) LoadProject(ctx context.Context, projectRoot string) (*entities.Project, error) {
	if m.LoadProjectFunc != nil {
		return m.LoadProjectFunc(ctx, projectRoot)
	}
	return nil, nil
}

func (m *MockProjectRepository) SaveProject(ctx context.Context, project *entities.Project) error {
	if m.SaveProjectFunc != nil {
		return m.SaveProjectFunc(ctx, project)
	}
	return nil
}

func (m *MockProjectRepository) SaveSystem(ctx context.Context, projectRoot string, system *entities.System) error {
	if m.SaveSystemFunc != nil {
		return m.SaveSystemFunc(ctx, projectRoot, system)
	}
	return nil
}

func (m *MockProjectRepository) ListSystems(ctx context.Context, projectRoot string) ([]*entities.System, error) {
	if m.ListSystemsFunc != nil {
		return m.ListSystemsFunc(ctx, projectRoot)
	}
	return nil, nil
}

func (m *MockProjectRepository) LoadSystem(ctx context.Context, projectRoot, systemName string) (*entities.System, error) {
	if m.LoadSystemFunc != nil {
		return m.LoadSystemFunc(ctx, projectRoot, systemName)
	}
	return nil, nil
}

func (m *MockProjectRepository) LoadContainer(ctx context.Context, projectRoot, systemName, containerName string) (*entities.Container, error) {
	return nil, nil
}

func (m *MockProjectRepository) SaveContainer(ctx context.Context, projectRoot, systemName string, container *entities.Container) error {
	return nil
}

// TestCreateSystemUseCase tests the CreateSystem use case.
func TestCreateSystemUseCase(t *testing.T) {
	tests := []struct {
		name      string
		systemReq *CreateSystemRequest
		wantErr   bool
		errMsg    string
		validate  func(t *testing.T, sys *entities.System)
	}{
		{
			name: "valid system creation",
			systemReq: &CreateSystemRequest{
				Name:        "Payment Service",
				Description: "Handles payment processing",
			},
			wantErr: false,
			validate: func(t *testing.T, sys *entities.System) {
				if sys.Name != "Payment Service" {
					t.Errorf("expected name 'Payment Service', got %q", sys.Name)
				}
				if sys.ID != "payment-service" {
					t.Errorf("expected ID 'payment-service', got %q", sys.ID)
				}
				if sys.Description != "Handles payment processing" {
					t.Errorf("expected description, got %q", sys.Description)
				}
			},
		},
		{
			name: "empty name",
			systemReq: &CreateSystemRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "invalid characters in name",
			systemReq: &CreateSystemRequest{
				Name: "System@#$",
			},
			wantErr: true,
		},
		{
			name: "system with tags",
			systemReq: &CreateSystemRequest{
				Name: "Auth Service",
				Tags: []string{"security", "critical"},
			},
			wantErr: false,
			validate: func(t *testing.T, sys *entities.System) {
				if len(sys.Tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(sys.Tags))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProjectRepository{}
			uc := NewCreateSystem(mockRepo)

			sys, err := uc.Execute(context.Background(), tt.systemReq)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !tt.wantErr && sys == nil {
				t.Errorf("Execute() returned nil system")
				return
			}

			if tt.validate != nil {
				tt.validate(t, sys)
			}
		})
	}
}

// TestCreateSystemValidation tests validation of system creation.
func TestCreateSystemValidation(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	uc := NewCreateSystem(mockRepo)

	sys, err := uc.Execute(context.Background(), &CreateSystemRequest{
		Name: "Valid System",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Validate the created system
	if err := sys.Validate(); err != nil {
		t.Errorf("System validation failed: %v", err)
	}
}
