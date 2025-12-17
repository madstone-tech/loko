package entities

import (
	"testing"
)

func TestNewContainer(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{"valid simple", "API", "api", false},
		{"valid with spaces", "API Server", "api-server", false},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cont, err := NewContainer(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContainer(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil && cont.ID != tt.wantID {
				t.Errorf("NewContainer(%q).ID = %q, want %q", tt.input, cont.ID, tt.wantID)
			}
		})
	}
}

func TestContainer_Components(t *testing.T) {
	cont, _ := NewContainer("API")

	// Add components
	comp1, _ := NewComponent("AuthHandler")
	comp2, _ := NewComponent("UserHandler")

	if err := cont.AddComponent(comp1); err != nil {
		t.Errorf("AddComponent failed: %v", err)
	}
	if err := cont.AddComponent(comp2); err != nil {
		t.Errorf("AddComponent failed: %v", err)
	}

	// Add nil component
	if err := cont.AddComponent(nil); err == nil {
		t.Error("AddComponent(nil) should fail")
	}

	// Add duplicate
	dup, _ := NewComponent("AuthHandler")
	if err := cont.AddComponent(dup); err == nil {
		t.Error("AddComponent duplicate should fail")
	}

	// Count
	if cont.ComponentCount() != 2 {
		t.Errorf("ComponentCount() = %d, want 2", cont.ComponentCount())
	}

	// Get
	got, err := cont.GetComponent("authhandler")
	if err != nil {
		t.Errorf("GetComponent failed: %v", err)
	}
	if got.Name != "AuthHandler" {
		t.Errorf("GetComponent returned wrong component")
	}

	// Get non-existent
	_, err = cont.GetComponent("nonexistent")
	if err == nil {
		t.Error("GetComponent(nonexistent) should fail")
	}

	// List
	list := cont.ListComponents()
	if len(list) != 2 {
		t.Errorf("ListComponents() returned %d, want 2", len(list))
	}

	// Remove
	if err := cont.RemoveComponent("authhandler"); err != nil {
		t.Errorf("RemoveComponent failed: %v", err)
	}
	if cont.ComponentCount() != 1 {
		t.Error("ComponentCount should be 1 after remove")
	}

	// Remove non-existent
	if err := cont.RemoveComponent("nonexistent"); err == nil {
		t.Error("RemoveComponent(nonexistent) should fail")
	}
}

func TestContainer_Validate(t *testing.T) {
	t.Run("valid container", func(t *testing.T) {
		cont, _ := NewContainer("API")
		comp, _ := NewComponent("Handler")
		cont.AddComponent(comp)

		if err := cont.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("invalid - bad component", func(t *testing.T) {
		cont, _ := NewContainer("API")
		cont.Components["bad"] = &Component{ID: "Bad ID", Name: ""}

		if err := cont.Validate(); err == nil {
			t.Error("Validate() should fail with invalid component")
		}
	})
}

func TestContainer_Tags(t *testing.T) {
	cont, _ := NewContainer("API")

	cont.AddTag("backend")
	cont.AddTag("core")
	cont.AddTag("backend") // Duplicate

	if len(cont.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(cont.Tags))
	}

	if !cont.HasTag("backend") {
		t.Error("HasTag(backend) should return true")
	}

	if cont.HasTag("nonexistent") {
		t.Error("HasTag(nonexistent) should return false")
	}
}
