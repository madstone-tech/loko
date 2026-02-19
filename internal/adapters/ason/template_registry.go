package ason

import (
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TemplateRegistry maps technology categories to template files on disk.
// It implements the usecases.TemplateRegistry interface.
type TemplateRegistry struct {
	// templateDir is the directory containing template .md files
	templateDir string
}

// NewTemplateRegistry creates a TemplateRegistry backed by the given directory.
// Template files are expected as "<name>.md" inside templateDir.
func NewTemplateRegistry(templateDir string) *TemplateRegistry {
	return &TemplateRegistry{templateDir: templateDir}
}

// categoryToName maps TemplateCategory values to their file-system name.
var categoryToName = map[entities.TemplateCategory]string{
	entities.TemplateCategoryCompute:   "compute",
	entities.TemplateCategoryDatastore: "datastore",
	entities.TemplateCategoryMessaging: "messaging",
	entities.TemplateCategoryAPI:       "api",
	entities.TemplateCategoryEvent:     "event",
	entities.TemplateCategoryStorage:   "storage",
	entities.TemplateCategoryGeneric:   "generic",
}

// GetTemplateName maps a category to a template name.
// Unknown categories fall back to "generic".
// entityType is accepted for future extension (e.g. "container" templates)
// but currently only "component"-level templates are supported.
func (r *TemplateRegistry) GetTemplateName(category entities.TemplateCategory, _ string) string {
	if name, ok := categoryToName[category]; ok {
		return name
	}
	return "generic"
}

// GetTemplatePath returns the absolute file path for a named template.
// The path is always computed; callers should use IsValidTemplate to check existence.
func (r *TemplateRegistry) GetTemplatePath(name string) string {
	return filepath.Join(r.templateDir, name+".md")
}

// IsValidTemplate returns true if a template file named "<name>.md" exists
// in the registry's template directory.
func (r *TemplateRegistry) IsValidTemplate(name string) bool {
	path := filepath.Join(r.templateDir, name+".md")
	_, err := os.Stat(path)
	return err == nil
}
