package template

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestRegistry_GetTemplateName(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name         string
		category     entities.TemplateCategory
		entityType   string
		expectedName string
	}{
		{
			name:         "Compute component",
			category:     entities.TemplateCategoryCompute,
			entityType:   "component",
			expectedName: "compute-component",
		},
		{
			name:         "Datastore container",
			category:     entities.TemplateCategoryDatastore,
			entityType:   "container",
			expectedName: "datastore-container",
		},
		{
			name:         "Messaging system",
			category:     entities.TemplateCategoryMessaging,
			entityType:   "system",
			expectedName: "standard-system", // Falls back to generic
		},
		{
			name:         "Unknown category",
			category:     "unknown",
			entityType:   "component",
			expectedName: "standard-component", // Falls back to generic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateName := registry.GetTemplateName(tt.category, tt.entityType)
			if templateName != tt.expectedName {
				t.Errorf("GetTemplateName(%q, %q) = %q, want %q", tt.category, tt.entityType, templateName, tt.expectedName)
			}
		})
	}
}

func TestRegistry_IsValidTemplate(t *testing.T) {
	registry := NewRegistry()

	// Since our implementation always returns true for non-empty names
	tests := []struct {
		name     string
		template string
		expected bool
	}{
		{"Valid template", "some-template", true},
		{"Empty template", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := registry.IsValidTemplate(tt.template)
			if valid != tt.expected {
				t.Errorf("IsValidTemplate(%q) = %v, want %v", tt.template, valid, tt.expected)
			}
		})
	}
}
