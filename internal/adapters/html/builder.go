// Package html provides an HTML site builder adapter that generates static documentation.
// It implements the SiteBuilder interface by producing semantic HTML5 pages with
// embedded CSS/JS, responsive design, and navigation features.
package html

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// Builder implements the SiteBuilder interface by generating static HTML documentation.
// It produces a complete website with index, system pages, diagrams, and search functionality.
type Builder struct {
	templates *template.Template
}

// NewBuilder creates a new HTML site builder with embedded templates.
func NewBuilder() (*Builder, error) {
	tmpl, err := parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Builder{
		templates: tmpl,
	}, nil
}

// BuildSite generates HTML documentation from a project.
// Creates an output directory with index.html, system pages, diagrams, and static assets.
func (b *Builder) BuildSite(ctx context.Context, project *entities.Project, systems []*entities.System, outputDir string) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}
	if outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Create output directory structure
	if err := b.createDirectories(outputDir); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write static assets (CSS and JS)
	if err := b.writeAssets(outputDir); err != nil {
		return fmt.Errorf("failed to write assets: %w", err)
	}

	// Build index page
	if err := b.buildIndexPage(ctx, project, systems, outputDir); err != nil {
		return fmt.Errorf("failed to build index page: %w", err)
	}

	// Build system pages
	for _, system := range systems {
		if system == nil {
			continue
		}
		containers := system.ListContainers()
		if err := b.BuildSystemPage(ctx, system, containers, outputDir); err != nil {
			return fmt.Errorf("failed to build system page for %s: %w", system.Name, err)
		}
	}

	// Build search index
	if err := b.buildSearchIndex(systems, outputDir); err != nil {
		return fmt.Errorf("failed to build search index: %w", err)
	}

	return nil
}

// BuildSystemPage generates a single system HTML page with embedded diagrams.
func (b *Builder) BuildSystemPage(_ context.Context, system *entities.System, containers []*entities.Container, outputDir string) error {
	if system == nil {
		return fmt.Errorf("system cannot be nil")
	}
	if outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Prepare template data
	data := map[string]interface{}{
		"System":     system,
		"Containers": containers,
	}

	// Render template
	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "system.html", data); err != nil {
		return fmt.Errorf("failed to render system template: %w", err)
	}

	// Write to file
	systemsDir := filepath.Join(outputDir, "systems")
	if err := os.MkdirAll(systemsDir, 0755); err != nil {
		return fmt.Errorf("failed to create systems directory: %w", err)
	}

	filePath := filepath.Join(systemsDir, system.ID+".html")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write system page %s: %w", filePath, err)
	}

	return nil
}

// buildIndexPage generates the project index page.
func (b *Builder) buildIndexPage(_ context.Context, project *entities.Project, systems []*entities.System, outputDir string) error {
	data := map[string]interface{}{
		"Project": project,
		"Systems": systems,
	}

	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "index.html", data); err != nil {
		return fmt.Errorf("failed to render index template: %w", err)
	}

	filePath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write index page: %w", err)
	}

	return nil
}

// buildSearchIndex generates a JSON search index for client-side search.
func (b *Builder) buildSearchIndex(systems []*entities.System, outputDir string) error {
	type SearchResult struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	type SearchIndex struct {
		Results []SearchResult `json:"results"`
	}

	var results []SearchResult

	// Add systems to search index
	for _, system := range systems {
		if system == nil {
			continue
		}
		results = append(results, SearchResult{
			Title:       system.Name,
			URL:         fmt.Sprintf("systems/%s.html", system.ID),
			Description: system.Description,
			Type:        "system",
		})

		// Add containers to search index
		for _, container := range system.ListContainers() {
			if container == nil {
				continue
			}
			results = append(results, SearchResult{
				Title:       container.Name,
				URL:         fmt.Sprintf("systems/%s.html#%s", system.ID, container.ID),
				Description: container.Description,
				Type:        "container",
			})
		}
	}

	index := SearchIndex{Results: results}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal search index: %w", err)
	}

	filePath := filepath.Join(outputDir, "search.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write search index: %w", err)
	}

	return nil
}

// createDirectories creates the output directory structure.
func (b *Builder) createDirectories(outputDir string) error {
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "systems"),
		filepath.Join(outputDir, "diagrams"),
		filepath.Join(outputDir, "styles"),
		filepath.Join(outputDir, "js"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// writeAssets writes CSS and JavaScript files to the output directory.
func (b *Builder) writeAssets(outputDir string) error {
	// Write CSS
	cssPath := filepath.Join(outputDir, "styles", "style.css")
	if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("failed to write CSS: %w", err)
	}

	// Write JavaScript
	jsPath := filepath.Join(outputDir, "js", "main.js")
	if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
		return fmt.Errorf("failed to write JavaScript: %w", err)
	}

	return nil
}

// parseTemplates parses all embedded HTML templates.
func parseTemplates() (*template.Template, error) {
	tmpl := template.New("base")

	// Parse all templates
	for name, content := range templateMap {
		_, err := tmpl.New(name).Parse(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return tmpl, nil
}
