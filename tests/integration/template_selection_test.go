package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// templateDir returns the path to the component templates directory in the repo.
func templateDir(t *testing.T) string {
	t.Helper()
	// Walk up from the test binary to find the repo root.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "templates", "component")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// TestTemplateSelection_DynamoDB verifies that a DynamoDB technology maps to the
// datastore template (T044).
func TestTemplateSelection_DynamoDB(t *testing.T) {
	registry := ason.NewTemplateRegistry(templateDir(t))
	uc := usecases.NewSelectTemplate(registry)

	result, err := uc.Execute(context.Background(), &usecases.SelectTemplateRequest{
		Technology: "DynamoDB",
		EntityType: "component",
	})
	if err != nil {
		t.Fatalf("SelectTemplate failed: %v", err)
	}

	if result.Category != entities.TemplateCategoryDatastore {
		t.Errorf("expected datastore category for DynamoDB, got %q", result.Category)
	}
	if !result.Matched {
		t.Error("expected technology match for DynamoDB")
	}
	if !registry.IsValidTemplate(result.SelectedTemplate) {
		t.Errorf("selected template %q does not exist on disk", result.SelectedTemplate)
	}

	// Verify the template file contains datastore-specific sections
	path := registry.GetTemplatePath(result.SelectedTemplate)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read template file %s: %v", path, err)
	}
	if !strings.Contains(string(content), "Schema") {
		t.Errorf("datastore template should contain 'Schema' section, got:\n%s", content[:200])
	}
}

// TestTemplateSelection_Lambda verifies that a Lambda technology maps to the
// compute template with correct sections (T045).
func TestTemplateSelection_Lambda(t *testing.T) {
	registry := ason.NewTemplateRegistry(templateDir(t))
	uc := usecases.NewSelectTemplate(registry)

	result, err := uc.Execute(context.Background(), &usecases.SelectTemplateRequest{
		Technology: "AWS Lambda",
		EntityType: "component",
	})
	if err != nil {
		t.Fatalf("SelectTemplate failed: %v", err)
	}

	if result.Category != entities.TemplateCategoryCompute {
		t.Errorf("expected compute category for Lambda, got %q", result.Category)
	}
	if !result.Matched {
		t.Error("expected technology match for Lambda")
	}
	if !registry.IsValidTemplate(result.SelectedTemplate) {
		t.Errorf("selected template %q does not exist on disk", result.SelectedTemplate)
	}

	// Verify the template file contains compute-specific sections
	path := registry.GetTemplatePath(result.SelectedTemplate)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read template file %s: %v", path, err)
	}
	if !strings.Contains(string(content), "Trigger") {
		t.Errorf("compute template should contain 'Trigger' section")
	}
	if !strings.Contains(string(content), "Runtime") {
		t.Errorf("compute template should contain 'Runtime' section")
	}
}

// TestTemplateSelection_Override verifies that a --template flag overrides technology
// matching (T046).
func TestTemplateSelection_Override(t *testing.T) {
	registry := ason.NewTemplateRegistry(templateDir(t))
	uc := usecases.NewSelectTemplate(registry)

	// User explicitly requests "datastore" even though technology says "Lambda"
	result, err := uc.Execute(context.Background(), &usecases.SelectTemplateRequest{
		Technology:       "Lambda",
		EntityType:       "component",
		OverrideTemplate: "datastore",
	})
	if err != nil {
		t.Fatalf("SelectTemplate with override failed: %v", err)
	}

	if result.SelectedTemplate != "datastore" {
		t.Errorf("expected override template 'datastore', got %q", result.SelectedTemplate)
	}
	if result.Matched {
		t.Error("override should set Matched=false (explicit, not pattern-matched)")
	}
}

// TestTemplateSelection_ComponentSavedWithTechnology verifies end-to-end:
// a component with DynamoDB technology can be created and saved (T058).
// The template selection logic is separate from the filesystem layer (the
// filesystem layer uses the generic template engine; template-aware saving
// is wired in Phase 5 T055-T057).
func TestTemplateSelection_ComponentSavedWithTechnology(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "test-proj")

	repo := filesystem.NewProjectRepository()
	ctx := context.Background()

	// Scaffold a minimal project
	project, err := entities.NewProject("test-proj")
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	project.Path = projectRoot
	if err := repo.SaveProject(ctx, project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	system, err := entities.NewSystem("payments")
	if err != nil {
		t.Fatalf("NewSystem: %v", err)
	}
	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("SaveSystem: %v", err)
	}

	container, err := entities.NewContainer("api")
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	// Create a DynamoDB component
	component, err := entities.NewComponent("Orders Table")
	if err != nil {
		t.Fatalf("NewComponent: %v", err)
	}
	component.Technology = "DynamoDB"
	component.Description = "Stores order records"

	if err := repo.SaveComponent(ctx, projectRoot, system.ID, container.ID, component); err != nil {
		t.Fatalf("SaveComponent: %v", err)
	}

	// Verify component.md was written
	compDir := filepath.Join(projectRoot, "src", system.ID, container.ID, component.ID)
	compMd := filepath.Join(compDir, "component.md")
	if _, err := os.Stat(compMd); err != nil {
		t.Errorf("component.md not found at %s: %v", compMd, err)
	}

	// Verify technology is preserved in frontmatter
	content, err := os.ReadFile(compMd)
	if err != nil {
		t.Fatalf("failed to read component.md: %v", err)
	}
	if !strings.Contains(string(content), "DynamoDB") {
		t.Errorf("component.md should contain 'DynamoDB', got:\n%s", string(content)[:300])
	}
}
