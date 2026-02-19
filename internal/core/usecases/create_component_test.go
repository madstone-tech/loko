package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// mockComponentTemplateRegistry is a test double for TemplateRegistry.
type mockComponentTemplateRegistry struct {
	getTemplateNameFunc func(category entities.TemplateCategory, entityType string) string
	isValidTemplateFunc func(name string) bool
}

func (m *mockComponentTemplateRegistry) GetTemplateName(category entities.TemplateCategory, entityType string) string {
	if m.getTemplateNameFunc != nil {
		return m.getTemplateNameFunc(category, entityType)
	}
	return "default-template"
}

func (m *mockComponentTemplateRegistry) IsValidTemplate(name string) bool {
	if m.isValidTemplateFunc != nil {
		return m.isValidTemplateFunc(name)
	}
	return true
}

// TestNewCreateComponent tests creating a CreateComponent use case without templates.
func TestNewCreateComponent(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	uc := NewCreateComponent(mockRepo)

	if uc == nil {
		t.Error("NewCreateComponent() returned nil")
	}

	if uc.repo != mockRepo {
		t.Error("NewCreateComponent() did not set repo correctly")
	}

	if uc.templateRegistry != nil {
		t.Error("NewCreateComponent() should not set templateRegistry")
	}
}

// TestNewCreateComponentWithTemplates tests creating a CreateComponent use case with templates.
func TestNewCreateComponentWithTemplates(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	mockRegistry := &mockTemplateRegistry{}

	uc := NewCreateComponentWithTemplates(mockRepo, mockRegistry)

	if uc == nil {
		t.Error("NewCreateComponentWithTemplates() returned nil")
	}

	if uc.repo != mockRepo {
		t.Error("NewCreateComponentWithTemplates() did not set repo correctly")
	}

	if uc.templateRegistry != mockRegistry {
		t.Error("NewCreateComponentWithTemplates() did not set templateRegistry correctly")
	}
}

// TestCreateComponentExecute tests the Execute method of CreateComponent.
func TestCreateComponentExecute(t *testing.T) {
	tests := []struct {
		name     string
		request  *CreateComponentRequest
		wantErr  bool
		validate func(t *testing.T, result *CreateComponentResult)
	}{
		{
			name: "valid component creation without templates",
			request: &CreateComponentRequest{
				Name:        "Auth Handler",
				Description: "Handles authentication",
				Technology:  "Go",
				Tags:        []string{"security", "core"},
			},
			wantErr: false,
			validate: func(t *testing.T, result *CreateComponentResult) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				if result.Component == nil {
					t.Fatal("component should not be nil")
				}
				if result.Component.Name != "Auth Handler" {
					t.Errorf("expected name 'Auth Handler', got %q", result.Component.Name)
				}
				if result.Component.Description != "Handles authentication" {
					t.Errorf("expected description 'Handles authentication', got %q", result.Component.Description)
				}
				if result.Component.Technology != "Go" {
					t.Errorf("expected technology 'Go', got %q", result.Component.Technology)
				}
				if len(result.Component.Tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(result.Component.Tags))
				}
				if result.SelectedTemplate != "" {
					t.Error("selected template should be empty when no registry provided")
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
			request: &CreateComponentRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "invalid name characters",
			request: &CreateComponentRequest{
				Name: "Component@Name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProjectRepository{}
			uc := NewCreateComponent(mockRepo)

			result, err := uc.Execute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestCreateComponentWithTemplateSelection tests component creation with template selection.
func TestCreateComponentWithTemplateSelection(t *testing.T) {
	mockRepo := &MockProjectRepository{}

	// Create a mock registry that will return a specific template for Go technology
	mockRegistry := &mockComponentTemplateRegistry{
		getTemplateNameFunc: func(category entities.TemplateCategory, entityType string) string {
			if category == entities.TemplateCategoryCompute && entityType == "component" {
				return "go-component-template"
			}
			return "default-template"
		},
		isValidTemplateFunc: func(name string) bool {
			return name == "go-component-template" || name == "default-template"
		},
	}

	uc := NewCreateComponentWithTemplates(mockRepo, mockRegistry)

	req := &CreateComponentRequest{
		Name:        "API Service",
		Description: "REST API service",
		Technology:  "Go",
		Tags:        []string{"api", "microservice"},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Since we're not directly testing the SelectTemplate use case here,
	// we just verify that template selection happened (template info is not empty)
	if result.SelectedTemplate == "" {
		t.Error("expected non-empty selected template")
	}

	// We can't easily test the exact category without reimplementing the selector logic
	// but we can verify that the result struct has the fields populated
}

// TestCreateComponentTemplateSelectionFailure tests graceful degradation when template selection fails.
func TestCreateComponentTemplateSelectionFailure(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	mockRegistry := &mockComponentTemplateRegistry{
		getTemplateNameFunc: func(category entities.TemplateCategory, entityType string) string {
			return "default-template"
		},
		isValidTemplateFunc: func(name string) bool {
			// Simulate template not found
			return false
		},
	}

	uc := NewCreateComponentWithTemplates(mockRepo, mockRegistry)

	req := &CreateComponentRequest{
		Name:        "Web Service",
		Description: "Web frontend service",
		Technology:  "React",
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should still create component even if template selection fails
	if result.Component == nil {
		t.Error("component should still be created even if template selection fails")
	}

	// Template selection failure should result in empty template info
	if result.SelectedTemplate != "" {
		t.Errorf("expected empty selected template on failure, got %q", result.SelectedTemplate)
	}
}
