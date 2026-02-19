package entities

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{"valid simple", "AuthHandler", "authhandler", false},
		{"valid with spaces", "Auth Handler", "auth-handler", false},
		{"empty", "", "", true},
		{"invalid chars", "Auth@Handler", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := NewComponent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewComponent(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil {
				if comp.ID != tt.wantID {
					t.Errorf("NewComponent(%q).ID = %q, want %q", tt.input, comp.ID, tt.wantID)
				}
				if comp.Name != tt.input {
					t.Errorf("NewComponent(%q).Name = %q, want %q", tt.input, comp.Name, tt.input)
				}
			}
		})
	}
}

func TestComponent_Validate(t *testing.T) {
	t.Run("valid component", func(t *testing.T) {
		comp, _ := NewComponent("AuthHandler")
		comp.Description = "Handles authentication"
		comp.Technology = "Go"

		if err := comp.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("invalid component - empty name", func(t *testing.T) {
		comp := &Component{ID: "auth", Name: ""}
		if err := comp.Validate(); err == nil {
			t.Error("Validate() expected error for empty name")
		}
	})

	t.Run("invalid component - bad id", func(t *testing.T) {
		comp := &Component{ID: "Auth Handler", Name: "Auth Handler"}
		if err := comp.Validate(); err == nil {
			t.Error("Validate() expected error for invalid id")
		}
	})
}

func TestComponent_Tags(t *testing.T) {
	comp, _ := NewComponent("AuthHandler")

	// Add tags
	comp.AddTag("security")
	comp.AddTag("core")
	comp.AddTag("security") // Duplicate

	if len(comp.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(comp.Tags))
	}

	if !comp.HasTag("security") {
		t.Error("HasTag(security) should return true")
	}

	if comp.HasTag("nonexistent") {
		t.Error("HasTag(nonexistent) should return false")
	}
}

func TestComponent_SettersGetters(t *testing.T) {
	comp, _ := NewComponent("Handler")

	comp.SetDescription("Test description")
	if comp.Description != "Test description" {
		t.Error("SetDescription failed")
	}

	comp.SetTechnology("Go + gRPC")
	if comp.Technology != "Go + gRPC" {
		t.Error("SetTechnology failed")
	}
}

// TestComponent_RelationshipMethods tests relationship-related methods.
func TestComponent_RelationshipMethods(t *testing.T) {
	comp, _ := NewComponent("TestComponent")

	// Initially no relationships
	if comp.RelationshipCount() != 0 {
		t.Errorf("Expected 0 relationships, got %d", comp.RelationshipCount())
	}

	// Add relationships
	comp.AddRelationship("target1", "calls target1")
	comp.AddRelationship("target2", "uses target2")

	if comp.RelationshipCount() != 2 {
		t.Errorf("Expected 2 relationships, got %d", comp.RelationshipCount())
	}

	// Check if relationship exists
	desc, exists := comp.GetRelationship("target1")
	if !exists {
		t.Error("Expected relationship to exist")
	}
	if desc != "calls target1" {
		t.Errorf("Expected description 'calls target1', got %q", desc)
	}

	// Check non-existent relationship
	_, exists = comp.GetRelationship("nonexistent")
	if exists {
		t.Error("Expected relationship to not exist")
	}

	// List relationships
	rels := comp.ListRelationships()
	if len(rels) != 2 {
		t.Errorf("Expected 2 relationships in list, got %d", len(rels))
	}

	// Remove relationship
	comp.RemoveRelationship("target1")
	if comp.RelationshipCount() != 1 {
		t.Errorf("Expected 1 relationship after removal, got %d", comp.RelationshipCount())
	}

	// Try to remove non-existent relationship (should not panic)
	comp.RemoveRelationship("nonexistent")
	if comp.RelationshipCount() != 1 {
		t.Errorf("Expected 1 relationship after removing non-existent, got %d", comp.RelationshipCount())
	}

	// Add duplicate relationship (should update)
	comp.AddRelationship("target2", "updated description")
	desc, _ = comp.GetRelationship("target2")
	if desc != "updated description" {
		t.Errorf("Expected updated description, got %q", desc)
	}
}

// TestComponent_CodeAnnotationMethods tests code annotation-related methods.
func TestComponent_CodeAnnotationMethods(t *testing.T) {
	comp, _ := NewComponent("TestComponent")

	// Initially no code annotations
	if comp.CodeAnnotationCount() != 0 {
		t.Errorf("Expected 0 code annotations, got %d", comp.CodeAnnotationCount())
	}

	// Add code annotations
	comp.AddCodeAnnotation("internal/handler", "Main handler logic")
	comp.AddCodeAnnotation("internal/model", "Data models")

	if comp.CodeAnnotationCount() != 2 {
		t.Errorf("Expected 2 code annotations, got %d", comp.CodeAnnotationCount())
	}

	// Check if annotation exists
	desc, exists := comp.GetCodeAnnotation("internal/handler")
	if !exists {
		t.Error("Expected code annotation to exist")
	}
	if desc != "Main handler logic" {
		t.Errorf("Expected description 'Main handler logic', got %q", desc)
	}

	// Check non-existent annotation
	_, exists = comp.GetCodeAnnotation("nonexistent")
	if exists {
		t.Error("Expected code annotation to not exist")
	}

	// List code annotations
	annotations := comp.ListCodeAnnotations()
	if len(annotations) != 2 {
		t.Errorf("Expected 2 annotations in list, got %d", len(annotations))
	}

	// Remove annotation
	comp.RemoveCodeAnnotation("internal/handler")
	if comp.CodeAnnotationCount() != 1 {
		t.Errorf("Expected 1 annotation after removal, got %d", comp.CodeAnnotationCount())
	}

	// Try to remove non-existent annotation (should not panic)
	comp.RemoveCodeAnnotation("nonexistent")
	if comp.CodeAnnotationCount() != 1 {
		t.Errorf("Expected 1 annotation after removing non-existent, got %d", comp.CodeAnnotationCount())
	}

	// Add duplicate annotation (should update)
	comp.AddCodeAnnotation("internal/model", "Updated data models")
	desc, _ = comp.GetCodeAnnotation("internal/model")
	if desc != "Updated data models" {
		t.Errorf("Expected updated description, got %q", desc)
	}
}

// TestComponent_DependencyMethods tests dependency-related methods.
func TestComponent_DependencyMethods(t *testing.T) {
	comp, _ := NewComponent("TestComponent")

	// Initially no dependencies
	if comp.DependencyCount() != 0 {
		t.Errorf("Expected 0 dependencies, got %d", comp.DependencyCount())
	}

	// Add dependencies
	comp.AddDependency("github.com/gorilla/mux")
	comp.AddDependency("github.com/lib/pq")
	comp.AddDependency("github.com/golang-jwt/jwt")

	if comp.DependencyCount() != 3 {
		t.Errorf("Expected 3 dependencies, got %d", comp.DependencyCount())
	}

	// Check if dependency exists
	if !comp.HasDependency("github.com/gorilla/mux") {
		t.Error("Expected dependency to exist")
	}

	// Check non-existent dependency
	if comp.HasDependency("nonexistent") {
		t.Error("Expected dependency to not exist")
	}

	// List dependencies
	deps := comp.ListDependencies()
	if len(deps) != 3 {
		t.Errorf("Expected 3 dependencies in list, got %d", len(deps))
	}

	// Remove dependency
	comp.RemoveDependency("github.com/gorilla/mux")
	if comp.DependencyCount() != 2 {
		t.Errorf("Expected 2 dependencies after removal, got %d", comp.DependencyCount())
	}

	// Try to remove non-existent dependency (should not panic)
	comp.RemoveDependency("nonexistent")
	if comp.DependencyCount() != 2 {
		t.Errorf("Expected 2 dependencies after removing non-existent, got %d", comp.DependencyCount())
	}

	// Add duplicate dependency (should not add twice)
	comp.AddDependency("github.com/lib/pq")
	if comp.DependencyCount() != 2 {
		t.Errorf("Expected 2 dependencies after adding duplicate, got %d", comp.DependencyCount())
	}

	// Test empty dependency (should not be added)
	initialCount := comp.DependencyCount()
	comp.AddDependency("")
	if comp.DependencyCount() != initialCount {
		t.Error("Expected no change when adding empty dependency")
	}
}

// TestComponent_EntityTypeMethods tests entity type-related methods.
func TestComponent_EntityTypeMethods(t *testing.T) {
	comp, _ := NewComponent("TestComponent")

	if comp.GetID() != "testcomponent" {
		t.Errorf("Expected ID 'testcomponent', got %q", comp.GetID())
	}

	if comp.GetName() != "TestComponent" {
		t.Errorf("Expected name 'TestComponent', got %q", comp.GetName())
	}

	if comp.GetEntityType() != "component" {
		t.Errorf("Expected entity type 'component', got %q", comp.GetEntityType())
	}
}
