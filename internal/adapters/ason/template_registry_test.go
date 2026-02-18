package ason_test

import (
	"os"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestTemplateRegistry_GetTemplateName verifies the TemplateRegistry maps all 7
// categories to correct template names (T043).
func TestTemplateRegistry_GetTemplateName(t *testing.T) {
	tests := []struct {
		name       string
		category   entities.TemplateCategory
		entityType string
		want       string
	}{
		{
			name:       "compute template",
			category:   entities.TemplateCategoryCompute,
			entityType: "component",
			want:       "compute",
		},
		{
			name:       "datastore template",
			category:   entities.TemplateCategoryDatastore,
			entityType: "component",
			want:       "datastore",
		},
		{
			name:       "messaging template",
			category:   entities.TemplateCategoryMessaging,
			entityType: "component",
			want:       "messaging",
		},
		{
			name:       "api template",
			category:   entities.TemplateCategoryAPI,
			entityType: "component",
			want:       "api",
		},
		{
			name:       "event template",
			category:   entities.TemplateCategoryEvent,
			entityType: "component",
			want:       "event",
		},
		{
			name:       "storage template",
			category:   entities.TemplateCategoryStorage,
			entityType: "component",
			want:       "storage",
		},
		{
			name:       "generic template",
			category:   entities.TemplateCategoryGeneric,
			entityType: "component",
			want:       "generic",
		},
		{
			name:       "unknown category falls back to generic",
			category:   entities.TemplateCategory("unknown"),
			entityType: "component",
			want:       "generic",
		},
	}

	registry := ason.NewTemplateRegistry("/some/template/dir")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.GetTemplateName(tt.category, tt.entityType)
			if got != tt.want {
				t.Errorf("GetTemplateName(%q, %q) = %q, want %q", tt.category, tt.entityType, got, tt.want)
			}
		})
	}
}

// TestTemplateRegistry_IsValidTemplate verifies that template validation
// works correctly with a real directory (T043).
func TestTemplateRegistry_IsValidTemplate_KnownTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake template file
	templateFile := tmpDir + "/compute.md"
	if err := os.WriteFile(templateFile, []byte("# Compute Template"), 0o644); err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	registry := ason.NewTemplateRegistry(tmpDir)

	if !registry.IsValidTemplate("compute") {
		t.Error("expected 'compute' to be valid when compute.md exists")
	}
	if registry.IsValidTemplate("nonexistent") {
		t.Error("expected 'nonexistent' to be invalid when no file exists")
	}
}

// TestTemplateRegistry_GetTemplatePath verifies absolute path resolution (T043).
func TestTemplateRegistry_GetTemplatePath(t *testing.T) {
	registry := ason.NewTemplateRegistry("/templates/component")

	tests := []struct {
		category   entities.TemplateCategory
		entityType string
		wantSuffix string
	}{
		{entities.TemplateCategoryCompute, "component", "compute.md"},
		{entities.TemplateCategoryDatastore, "component", "datastore.md"},
		{entities.TemplateCategoryGeneric, "component", "generic.md"},
	}

	for _, tt := range tests {
		name := registry.GetTemplateName(tt.category, tt.entityType)
		path := registry.GetTemplatePath(name)
		if path == "" {
			t.Errorf("GetTemplatePath(%q) returned empty string", name)
		}
		// Path should end with the expected filename
		if len(path) < len(tt.wantSuffix) || path[len(path)-len(tt.wantSuffix):] != tt.wantSuffix {
			t.Errorf("GetTemplatePath(%q) = %q, want suffix %q", name, path, tt.wantSuffix)
		}
	}
}
