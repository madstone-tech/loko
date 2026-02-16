package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// mockTemplateRegistry is a mock implementation of TemplateRegistry for testing.
type mockTemplateRegistry struct {
	templates map[string]bool
}

func (m *mockTemplateRegistry) GetTemplateName(category entities.TemplateCategory, entityType string) string {
	// Simple mapping for testing
	templateName := string(category) + "-" + entityType

	// For generic category, use standard
	if category == entities.TemplateCategoryGeneric {
		templateName = "standard-" + entityType
	}

	// If we have a specific template map, use it
	if m.templates != nil {
		// Try to find a valid template name in our map
		for tmpl := range m.templates {
			if tmpl == templateName {
				return templateName
			}
		}
		// If not found, return a known valid template
		for tmpl := range m.templates {
			if m.templates[tmpl] {
				return tmpl
			}
		}
	}
	return templateName
}

func (m *mockTemplateRegistry) IsValidTemplate(name string) bool {
	if m.templates == nil {
		// Default behavior: all non-empty names are valid
		return name != ""
	}
	valid, exists := m.templates[name]
	if !exists {
		// If not in our map, check if it follows our naming pattern
		// For testing purposes, we'll consider it invalid if it's not in our map
		return false
	}
	return valid
}

func TestSelectTemplate_Execute(t *testing.T) {
	registry := &mockTemplateRegistry{
		templates: map[string]bool{
			"compute-component":   true,
			"datastore-component": true,
			"standard-component":  true,
			"custom-template":     true, // Add this for the override test
		},
	}

	usecase := NewSelectTemplate(registry)

	tests := []struct {
		name          string
		request       *SelectTemplateRequest
		expectError   bool
		expectedName  string
		expectedMatch bool
	}{
		{
			name: "Override template",
			request: &SelectTemplateRequest{
				Technology:       "AWS Lambda",
				EntityType:       "component",
				OverrideTemplate: "custom-template",
			},
			expectError:   false,
			expectedName:  "custom-template",
			expectedMatch: false,
		},
		{
			name: "Match compute technology",
			request: &SelectTemplateRequest{
				Technology: "AWS Lambda function",
				EntityType: "component",
			},
			expectError:   false,
			expectedName:  "compute-component",
			expectedMatch: true,
		},
		{
			name: "Match datastore technology",
			request: &SelectTemplateRequest{
				Technology: "DynamoDB table",
				EntityType: "component",
			},
			expectError:   false,
			expectedName:  "datastore-component",
			expectedMatch: true,
		},
		{
			name: "No technology match - fallback to generic",
			request: &SelectTemplateRequest{
				Technology: "unknown technology",
				EntityType: "component",
			},
			expectError:   false,
			expectedName:  "standard-component",
			expectedMatch: false,
		},
		{
			name: "Empty technology - use default",
			request: &SelectTemplateRequest{
				Technology: "",
				EntityType: "component",
			},
			expectError:   false,
			expectedName:  "standard-component",
			expectedMatch: false,
		},
		{
			name: "Invalid override template",
			request: &SelectTemplateRequest{
				Technology:       "AWS Lambda",
				EntityType:       "component",
				OverrideTemplate: "nonexistent-template",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := usecase.Execute(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.SelectedTemplate != tt.expectedName {
				t.Errorf("SelectedTemplate = %q, want %q", result.SelectedTemplate, tt.expectedName)
			}

			if result.Matched != tt.expectedMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.expectedMatch)
			}
		})
	}
}

func TestSelectTemplate_WithInvalidTemplate(t *testing.T) {
	registry := &mockTemplateRegistry{
		templates: map[string]bool{
			"compute-component":  false, // Mark as invalid
			"standard-component": true,
		},
	}

	usecase := NewSelectTemplate(registry)

	request := &SelectTemplateRequest{
		Technology: "AWS Lambda function",
		EntityType: "component",
	}

	result, err := usecase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should fall back to generic template
	if result.SelectedTemplate != "standard-component" {
		t.Errorf("Expected fallback to standard template, got %q", result.SelectedTemplate)
	}
}
