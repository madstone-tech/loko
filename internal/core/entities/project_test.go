package entities

import (
	"testing"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "MyProject", false},
		{"valid with spaces", "My Project", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, err := NewProject(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProject(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil {
				if proj.Name != tt.input {
					t.Errorf("NewProject(%q).Name = %q", tt.input, proj.Name)
				}
				if proj.Config == nil {
					t.Error("NewProject should set default Config")
				}
				if proj.CreatedAt.IsZero() {
					t.Error("NewProject should set CreatedAt")
				}
			}
		})
	}
}

func TestDefaultProjectConfig(t *testing.T) {
	cfg := DefaultProjectConfig()

	if cfg.SourceDir != "./src" {
		t.Errorf("SourceDir = %q, want ./src", cfg.SourceDir)
	}
	if cfg.OutputDir != "./dist" {
		t.Errorf("OutputDir = %q, want ./dist", cfg.OutputDir)
	}
	if cfg.D2Theme != "neutral-default" {
		t.Errorf("D2Theme = %q, want neutral-default", cfg.D2Theme)
	}
	if !cfg.D2Cache {
		t.Error("D2Cache should be true by default")
	}
	if !cfg.HTMLEnabled {
		t.Error("HTMLEnabled should be true by default")
	}
	if cfg.PDFEnabled {
		t.Error("PDFEnabled should be false by default")
	}
	if cfg.ServePort != 8080 {
		t.Errorf("ServePort = %d, want 8080", cfg.ServePort)
	}
}

func TestProject_Systems(t *testing.T) {
	proj, _ := NewProject("MyProject")

	// Add systems
	sys1, _ := NewSystem("Payment")
	sys2, _ := NewSystem("Order")

	if err := proj.AddSystem(sys1); err != nil {
		t.Errorf("AddSystem failed: %v", err)
	}
	if err := proj.AddSystem(sys2); err != nil {
		t.Errorf("AddSystem failed: %v", err)
	}

	// Add nil
	if err := proj.AddSystem(nil); err == nil {
		t.Error("AddSystem(nil) should fail")
	}

	// Add duplicate
	dup, _ := NewSystem("Payment")
	if err := proj.AddSystem(dup); err == nil {
		t.Error("AddSystem duplicate should fail")
	}

	// Count
	if proj.SystemCount() != 2 {
		t.Errorf("SystemCount() = %d, want 2", proj.SystemCount())
	}

	// Get
	got, err := proj.GetSystem("payment")
	if err != nil {
		t.Errorf("GetSystem failed: %v", err)
	}
	if got.Name != "Payment" {
		t.Errorf("GetSystem returned wrong system")
	}

	// Get non-existent
	_, err = proj.GetSystem("nonexistent")
	if err == nil {
		t.Error("GetSystem(nonexistent) should fail")
	}

	// List
	list := proj.ListSystems()
	if len(list) != 2 {
		t.Errorf("ListSystems() returned %d, want 2", len(list))
	}

	// Remove
	if err := proj.RemoveSystem("payment"); err != nil {
		t.Errorf("RemoveSystem failed: %v", err)
	}
	if proj.SystemCount() != 1 {
		t.Error("SystemCount should be 1 after remove")
	}

	// Remove non-existent
	if err := proj.RemoveSystem("nonexistent"); err == nil {
		t.Error("RemoveSystem(nonexistent) should fail")
	}
}

func TestProject_DeepAccess(t *testing.T) {
	proj, _ := NewProject("MyProject")
	sys, _ := NewSystem("Payment")
	cont, _ := NewContainer("API")
	comp, _ := NewComponent("Handler")

	cont.AddComponent(comp)
	sys.AddContainer(cont)
	proj.AddSystem(sys)

	// GetContainer
	gotCont, err := proj.GetContainer("payment", "api")
	if err != nil {
		t.Errorf("GetContainer failed: %v", err)
	}
	if gotCont.Name != "API" {
		t.Error("GetContainer returned wrong container")
	}

	// GetComponent
	gotComp, err := proj.GetComponent("payment", "api", "handler")
	if err != nil {
		t.Errorf("GetComponent failed: %v", err)
	}
	if gotComp.Name != "Handler" {
		t.Error("GetComponent returned wrong component")
	}

	// Invalid paths
	_, err = proj.GetContainer("nonexistent", "api")
	if err == nil {
		t.Error("GetContainer with invalid system should fail")
	}

	_, err = proj.GetComponent("payment", "nonexistent", "handler")
	if err == nil {
		t.Error("GetComponent with invalid container should fail")
	}
}

func TestProject_Counts(t *testing.T) {
	proj, _ := NewProject("MyProject")
	sys, _ := NewSystem("Payment")
	cont, _ := NewContainer("API")
	comp1, _ := NewComponent("Handler1")
	comp2, _ := NewComponent("Handler2")

	cont.AddComponent(comp1)
	cont.AddComponent(comp2)
	sys.AddContainer(cont)
	proj.AddSystem(sys)

	if proj.ContainerCount() != 1 {
		t.Errorf("ContainerCount() = %d, want 1", proj.ContainerCount())
	}

	if proj.ComponentCount() != 2 {
		t.Errorf("ComponentCount() = %d, want 2", proj.ComponentCount())
	}

	stats := proj.Stats()
	if stats.Systems != 1 || stats.Containers != 1 || stats.Components != 2 {
		t.Errorf("Stats() = %+v, unexpected values", stats)
	}
}

func TestProject_Setters(t *testing.T) {
	proj, _ := NewProject("MyProject")
	initialUpdate := proj.UpdatedAt

	proj.SetDescription("Test description")
	if proj.Description != "Test description" {
		t.Error("SetDescription failed")
	}
	if !proj.UpdatedAt.After(initialUpdate) {
		t.Error("SetDescription should update UpdatedAt")
	}

	proj.SetVersion("1.0.0")
	if proj.Version != "1.0.0" {
		t.Error("SetVersion failed")
	}
}

func TestProject_Validate(t *testing.T) {
	t.Run("valid project", func(t *testing.T) {
		proj, _ := NewProject("MyProject")
		sys, _ := NewSystem("Payment")
		proj.AddSystem(sys)

		if err := proj.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("invalid - bad system", func(t *testing.T) {
		proj, _ := NewProject("MyProject")
		proj.Systems["bad"] = &System{ID: "Bad ID", Name: ""}

		if err := proj.Validate(); err == nil {
			t.Error("Validate() should fail with invalid system")
		}
	})
}
