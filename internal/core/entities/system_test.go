package entities

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{"valid simple", "Payment", "payment", false},
		{"valid with spaces", "Payment Service", "payment-service", false},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys, err := NewSystem(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSystem(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil && sys.ID != tt.wantID {
				t.Errorf("NewSystem(%q).ID = %q, want %q", tt.input, sys.ID, tt.wantID)
			}
		})
	}
}

func TestSystem_Containers(t *testing.T) {
	sys, _ := NewSystem("Payment")

	// Add containers
	cont1, _ := NewContainer("API")
	cont2, _ := NewContainer("Database")

	if err := sys.AddContainer(cont1); err != nil {
		t.Errorf("AddContainer failed: %v", err)
	}
	if err := sys.AddContainer(cont2); err != nil {
		t.Errorf("AddContainer failed: %v", err)
	}

	// Check parent ID is set
	if cont1.ParentID != sys.ID {
		t.Errorf("Container.ParentID = %q, want %q", cont1.ParentID, sys.ID)
	}

	// Add nil container
	if err := sys.AddContainer(nil); err == nil {
		t.Error("AddContainer(nil) should fail")
	}

	// Add duplicate
	dup, _ := NewContainer("API")
	if err := sys.AddContainer(dup); err == nil {
		t.Error("AddContainer duplicate should fail")
	}

	// Count
	if sys.ContainerCount() != 2 {
		t.Errorf("ContainerCount() = %d, want 2", sys.ContainerCount())
	}

	// Get
	got, err := sys.GetContainer("api")
	if err != nil {
		t.Errorf("GetContainer failed: %v", err)
	}
	if got.Name != "API" {
		t.Errorf("GetContainer returned wrong container")
	}

	// Get non-existent
	_, err = sys.GetContainer("nonexistent")
	if err == nil {
		t.Error("GetContainer(nonexistent) should fail")
	}

	// List
	list := sys.ListContainers()
	if len(list) != 2 {
		t.Errorf("ListContainers() returned %d, want 2", len(list))
	}

	// Remove
	if err := sys.RemoveContainer("api"); err != nil {
		t.Errorf("RemoveContainer failed: %v", err)
	}
	if sys.ContainerCount() != 1 {
		t.Error("ContainerCount should be 1 after remove")
	}

	// Remove non-existent
	if err := sys.RemoveContainer("nonexistent"); err == nil {
		t.Error("RemoveContainer(nonexistent) should fail")
	}
}

func TestSystem_ComponentCount(t *testing.T) {
	sys, _ := NewSystem("Payment")
	cont, _ := NewContainer("API")
	comp1, _ := NewComponent("Handler1")
	comp2, _ := NewComponent("Handler2")

	cont.AddComponent(comp1)
	cont.AddComponent(comp2)
	sys.AddContainer(cont)

	if sys.ComponentCount() != 2 {
		t.Errorf("ComponentCount() = %d, want 2", sys.ComponentCount())
	}
}

func TestSystem_GetComponent(t *testing.T) {
	sys, _ := NewSystem("Payment")
	cont, _ := NewContainer("API")
	comp, _ := NewComponent("Handler")

	cont.AddComponent(comp)
	sys.AddContainer(cont)

	// Valid path
	got, err := sys.GetComponent("api", "handler")
	if err != nil {
		t.Errorf("GetComponent failed: %v", err)
	}
	if got.Name != "Handler" {
		t.Error("GetComponent returned wrong component")
	}

	// Invalid container
	_, err = sys.GetComponent("nonexistent", "handler")
	if err == nil {
		t.Error("GetComponent with invalid container should fail")
	}

	// Invalid component
	_, err = sys.GetComponent("api", "nonexistent")
	if err == nil {
		t.Error("GetComponent with invalid component should fail")
	}
}

func TestSystem_Validate(t *testing.T) {
	t.Run("valid system", func(t *testing.T) {
		sys, _ := NewSystem("Payment")
		cont, _ := NewContainer("API")
		sys.AddContainer(cont)

		if err := sys.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("invalid - bad container", func(t *testing.T) {
		sys, _ := NewSystem("Payment")
		sys.Containers["bad"] = &Container{ID: "Bad ID", Name: ""}

		if err := sys.Validate(); err == nil {
			t.Error("Validate() should fail with invalid container")
		}
	})
}

func TestSystem_External(t *testing.T) {
	sys, _ := NewSystem("Stripe")
	sys.SetExternal(true)

	if !sys.External {
		t.Error("SetExternal(true) should set External to true")
	}
}

func TestSystem_Tags(t *testing.T) {
	sys, _ := NewSystem("Payment")

	sys.AddTag("core")
	sys.AddTag("payments")
	sys.AddTag("core") // Duplicate

	if len(sys.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(sys.Tags))
	}

	if !sys.HasTag("core") {
		t.Error("HasTag(core) should return true")
	}
}
