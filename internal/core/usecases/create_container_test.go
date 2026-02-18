package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestNewCreateContainer tests creating a CreateContainer use case.
func TestNewCreateContainer(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	uc := NewCreateContainer(mockRepo)

	if uc == nil {
		t.Error("NewCreateContainer() returned nil")
	}

	if uc.repo != mockRepo {
		t.Error("NewCreateContainer() did not set repo correctly")
	}
}

// TestCreateContainerExecute tests the Execute method of CreateContainer.
func TestCreateContainerExecute(t *testing.T) {
	tests := []struct {
		name     string
		request  *CreateContainerRequest
		wantErr  bool
		validate func(t *testing.T, container *entities.Container)
	}{
		{
			name: "valid container creation",
			request: &CreateContainerRequest{
				Name:        "API Server",
				Description: "Handles REST API requests",
				Technology:  "Go + gRPC",
				Tags:        []string{"api", "backend"},
			},
			wantErr: false,
			validate: func(t *testing.T, container *entities.Container) {
				if container == nil {
					t.Fatal("container should not be nil")
				}
				if container.Name != "API Server" {
					t.Errorf("expected name 'API Server', got %q", container.Name)
				}
				if container.Description != "Handles REST API requests" {
					t.Errorf("expected description 'Handles REST API requests', got %q", container.Description)
				}
				if container.Technology != "Go + gRPC" {
					t.Errorf("expected technology 'Go + gRPC', got %q", container.Technology)
				}
				if len(container.Tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(container.Tags))
				}
			},
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
		{
			name: "empty name",
			request: &CreateContainerRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "invalid name characters",
			request: &CreateContainerRequest{
				Name: "Container@Name",
			},
			wantErr: true,
		},
		{
			name: "container with no optional fields",
			request: &CreateContainerRequest{
				Name: "Minimal Container",
			},
			wantErr: false,
			validate: func(t *testing.T, container *entities.Container) {
				if container.Name != "Minimal Container" {
					t.Errorf("expected name 'Minimal Container', got %q", container.Name)
				}
				if container.Description != "" {
					t.Errorf("expected empty description, got %q", container.Description)
				}
				if container.Technology != "" {
					t.Errorf("expected empty technology, got %q", container.Technology)
				}
				if len(container.Tags) != 0 {
					t.Errorf("expected 0 tags, got %d", len(container.Tags))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProjectRepository{}
			uc := NewCreateContainer(mockRepo)

			container, err := uc.Execute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.validate != nil {
				tt.validate(t, container)
			}

			// Validate the created container
			if err := container.Validate(); err != nil {
				t.Errorf("Container validation failed: %v", err)
			}
		})
	}
}
