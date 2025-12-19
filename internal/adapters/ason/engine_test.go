package ason

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestRenderTemplate tests basic template rendering with variable substitution.
func TestRenderTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test template file
	templateDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	templateContent := `# {{SystemName}}

{{Description}}

## Technology

{{Technology}}
`

	templatePath := filepath.Join(templateDir, "system.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	engine := NewTemplateEngine()
	engine.AddSearchPath(templateDir)

	tests := []struct {
		name      string
		template  string
		variables map[string]string
		wantErr   bool
		validate  func(t *testing.T, content string)
	}{
		{
			name:     "simple substitution",
			template: "system.md",
			variables: map[string]string{
				"SystemName":  "Payment Service",
				"Description": "Handles payment processing",
				"Technology":  "Go + PostgreSQL",
			},
			wantErr: false,
			validate: func(t *testing.T, content string) {
				if !contains(content, "Payment Service") {
					t.Errorf("expected 'Payment Service' in output")
				}
				if !contains(content, "Handles payment processing") {
					t.Errorf("expected description in output")
				}
				if !contains(content, "Go + PostgreSQL") {
					t.Errorf("expected technology in output")
				}
			},
		},
		{
			name:      "missing template",
			template:  "nonexistent.md",
			variables: map[string]string{},
			wantErr:   true,
		},
		{
			name:     "empty variables",
			template: "system.md",
			variables: map[string]string{
				"SystemName":  "",
				"Description": "",
				"Technology":  "",
			},
			wantErr: false,
			validate: func(t *testing.T, content string) {
				if !contains(content, "# \n") {
					t.Errorf("expected empty substitution")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := engine.RenderTemplate(context.Background(), tt.template, tt.variables)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, content)
			}
		})
	}
}

// TestListTemplates tests template discovery.
func TestListTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create template files
	templateDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	templates := []string{"system.md", "container.md", "component.md"}
	for _, tmpl := range templates {
		path := filepath.Join(templateDir, tmpl)
		if err := os.WriteFile(path, []byte("template"), 0644); err != nil {
			t.Fatalf("failed to write template %s: %v", tmpl, err)
		}
	}

	engine := NewTemplateEngine()
	engine.AddSearchPath(templateDir)

	list, err := engine.ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	if len(list) != 3 {
		t.Errorf("expected 3 templates, got %d", len(list))
	}

	for _, expected := range templates {
		found := false
		for _, actual := range list {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected template %q not found", expected)
		}
	}
}

// TestAddSearchPath tests adding search paths.
func TestAddSearchPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two template directories
	dir1 := filepath.Join(tmpDir, "templates1")
	dir2 := filepath.Join(tmpDir, "templates2")

	for _, dir := range []string{dir1, dir2} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}

	// Add template to dir1
	if err := os.WriteFile(filepath.Join(dir1, "system.md"), []byte("template"), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	// Add template to dir2
	if err := os.WriteFile(filepath.Join(dir2, "container.md"), []byte("template"), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	engine := NewTemplateEngine()
	engine.AddSearchPath(dir1)
	engine.AddSearchPath(dir2)

	list, err := engine.ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 templates from both dirs, got %d", len(list))
	}
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
