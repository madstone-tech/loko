package entities

import (
	"testing"
)

func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		inputType TemplateType
		wantID    string
		wantErr   bool
	}{
		{"valid system", "c4-system", TemplateTypeSystem, "c4-system", false},
		{"valid container", "c4-container", TemplateTypeContainer, "c4-container", false},
		{"valid with spaces", "My Template", TemplateTypeProject, "my-template", false},
		{"empty name", "", TemplateTypeSystem, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewTemplate(tt.inputName, tt.inputType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplate(%q, %q) error = %v, wantErr %v", tt.inputName, tt.inputType, err, tt.wantErr)
				return
			}
			if err == nil {
				if tmpl.ID != tt.wantID {
					t.Errorf("NewTemplate(%q).ID = %q, want %q", tt.inputName, tmpl.ID, tt.wantID)
				}
				if tmpl.Type != tt.inputType {
					t.Errorf("NewTemplate Type = %q, want %q", tmpl.Type, tt.inputType)
				}
			}
		})
	}
}

func TestTemplate_Validate(t *testing.T) {
	t.Run("valid template", func(t *testing.T) {
		tmpl, _ := NewTemplate("c4-system", TemplateTypeSystem)
		tmpl.AddFile(TemplateFile{Source: "system.md.tmpl", Target: "{{.name}}/system.md"})

		if err := tmpl.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("invalid - no type", func(t *testing.T) {
		tmpl := &Template{ID: "test", Name: "Test", Files: []TemplateFile{{Source: "a", Target: "b"}}}
		if err := tmpl.Validate(); err == nil {
			t.Error("Validate() should fail for empty type")
		}
	})

	t.Run("invalid - no files", func(t *testing.T) {
		tmpl, _ := NewTemplate("empty", TemplateTypeSystem)
		if err := tmpl.Validate(); err == nil {
			t.Error("Validate() should fail for no files")
		}
	})

	t.Run("invalid - empty variable name", func(t *testing.T) {
		tmpl, _ := NewTemplate("test", TemplateTypeSystem)
		tmpl.AddFile(TemplateFile{Source: "a", Target: "b"})
		tmpl.AddVariable(TemplateVariable{Name: "", Required: true})

		if err := tmpl.Validate(); err == nil {
			t.Error("Validate() should fail for variable with empty name")
		}
	})
}

func TestTemplate_Variables(t *testing.T) {
	tmpl, _ := NewTemplate("c4-system", TemplateTypeSystem)

	tmpl.AddVariable(TemplateVariable{
		Name:        "name",
		Type:        "string",
		Description: "System name",
		Required:    true,
		Prompt:      "Enter system name:",
	})
	tmpl.AddVariable(TemplateVariable{
		Name:        "description",
		Type:        "string",
		Description: "System description",
		Required:    false,
		Default:     "A system",
	})

	// GetVariable
	v, found := tmpl.GetVariable("name")
	if !found {
		t.Error("GetVariable(name) should return true")
	}
	if v.Name != "name" {
		t.Error("GetVariable returned wrong variable")
	}

	_, found = tmpl.GetVariable("nonexistent")
	if found {
		t.Error("GetVariable(nonexistent) should return false")
	}

	// RequiredVariables
	required := tmpl.RequiredVariables()
	if len(required) != 1 {
		t.Errorf("RequiredVariables() = %d, want 1", len(required))
	}
	if required[0].Name != "name" {
		t.Error("RequiredVariables should return 'name'")
	}

	// DefaultValues
	defaults := tmpl.DefaultValues()
	if len(defaults) != 1 {
		t.Errorf("DefaultValues() = %d, want 1", len(defaults))
	}
	if defaults["description"] != "A system" {
		t.Errorf("DefaultValues[description] = %q, want 'A system'", defaults["description"])
	}
}

func TestTemplate_Files(t *testing.T) {
	tmpl, _ := NewTemplate("c4-system", TemplateTypeSystem)

	tmpl.AddFile(TemplateFile{
		Source: "system.md.tmpl",
		Target: "{{.name}}/system.md",
	})
	tmpl.AddFile(TemplateFile{
		Source:    "diagram.d2.tmpl",
		Target:    "{{.name}}/system.d2",
		Condition: "{{.includeDiagram}}",
	})

	if len(tmpl.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(tmpl.Files))
	}

	if tmpl.Files[0].Source != "system.md.tmpl" {
		t.Error("First file source incorrect")
	}
	if tmpl.Files[1].Condition != "{{.includeDiagram}}" {
		t.Error("Second file condition incorrect")
	}
}

func TestTemplateTypes(t *testing.T) {
	// Ensure constants are defined correctly
	if TemplateTypeSystem != "system" {
		t.Error("TemplateTypeSystem should be 'system'")
	}
	if TemplateTypeContainer != "container" {
		t.Error("TemplateTypeContainer should be 'container'")
	}
	if TemplateTypeComponent != "component" {
		t.Error("TemplateTypeComponent should be 'component'")
	}
	if TemplateTypeProject != "project" {
		t.Error("TemplateTypeProject should be 'project'")
	}
}
