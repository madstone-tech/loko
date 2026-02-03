// Package ason provides template rendering using the ason library.
// Ason is a simple template engine with variable substitution support.
package ason

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateEngine implements the TemplateEngine port using simple variable substitution.
// It supports template discovery from multiple search paths.
type TemplateEngine struct {
	searchPaths []string
}

// NewTemplateEngine creates a new template engine.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		searchPaths: []string{},
	}
}

// AddSearchPath adds a directory to the template search path.
// Paths are searched in the order they were added.
func (te *TemplateEngine) AddSearchPath(path string) {
	if path != "" {
		te.searchPaths = append(te.searchPaths, path)
	}
}

// RenderTemplate loads a template by name and applies variable substitution.
// Variables are substituted using {{VariableName}} syntax.
// Returns the rendered content or error if template not found.
func (te *TemplateEngine) RenderTemplate(ctx context.Context, templateName string, variables map[string]string) (string, error) {
	if templateName == "" {
		return "", fmt.Errorf("template name cannot be empty")
	}

	// Find the template file
	templatePath, err := te.findTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}

	// Read the template content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Apply variable substitution
	rendered := te.substitute(string(content), variables)

	return rendered, nil
}

// ListTemplates returns available template names from discovery paths.
// Returns a list of template filenames found in all search paths.
func (te *TemplateEngine) ListTemplates(ctx context.Context) ([]string, error) {
	templates := make(map[string]bool) // Use map to deduplicate

	for _, searchPath := range te.searchPaths {
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			// Skip paths that don't exist or can't be read
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				templates[entry.Name()] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(templates))
	for name := range templates {
		result = append(result, name)
	}

	return result, nil
}

// findTemplate searches for a template file in all search paths.
// Returns the full path to the template or an error if not found.
func (te *TemplateEngine) findTemplate(name string) (string, error) {
	for _, searchPath := range te.searchPaths {
		fullPath := filepath.Join(searchPath, name)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("template %q not found in any search path", name)
}

// substitute replaces {{VariableName}} with variable values.
// Variables not found in the map are replaced with empty strings.
func (te *TemplateEngine) substitute(content string, variables map[string]string) string {
	result := content

	// Simple variable substitution: {{VarName}} -> value
	for varName, varValue := range variables {
		placeholder := "{{" + varName + "}}"
		result = strings.ReplaceAll(result, placeholder, varValue)
	}

	return result
}
