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
