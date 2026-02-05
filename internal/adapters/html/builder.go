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
	"strings"
	"text/template"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// Builder implements the SiteBuilder interface by generating static HTML documentation.
// It produces a complete website with index, system pages, diagrams, and search functionality.
type Builder struct {
	templates        *template.Template
	cssTokens        map[string]string // Design system tokens for CSS generation
	markdownRenderer *MarkdownRenderer // Renderer for markdown content
}

// NewBuilder creates a new HTML site builder with embedded templates.
func NewBuilder() (*Builder, error) {
	tmpl, err := parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Builder{
		templates:        tmpl,
		cssTokens:        getDefaultCSSTokens(),
		markdownRenderer: NewMarkdownRenderer("", ""),
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

		// Build container pages
		for _, container := range containers {
			if container == nil {
				continue
			}
			components := container.ListComponents()
			if err := b.BuildContainerPage(ctx, system, container, components, outputDir); err != nil {
				return fmt.Errorf("failed to build container page for %s/%s: %w", system.Name, container.Name, err)
			}

			// Build component pages
			for _, component := range components {
				if component == nil {
					continue
				}
				if err := b.BuildComponentPage(ctx, system, container, component, outputDir); err != nil {
					return fmt.Errorf("failed to build component page for %s/%s/%s: %w", system.Name, container.Name, component.Name, err)
				}
			}
		}
	}

	// Build containers overview page
	if err := b.buildContainersOverview(ctx, systems, outputDir); err != nil {
		return fmt.Errorf("failed to build containers overview: %w", err)
	}

	// Build components overview page
	if err := b.buildComponentsOverview(ctx, systems, outputDir); err != nil {
		return fmt.Errorf("failed to build components overview: %w", err)
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

	// Try to read and render markdown content
	markdownContent := ""
	if system.Path != "" {
		markdownPath := filepath.Join(system.Path, "system.md")
		if content, err := os.ReadFile(markdownPath); err == nil {
			// Render markdown to HTML fragment (content only, no HTML wrapper)
			fullHTML := b.markdownRenderer.RenderMarkdownToHTML(string(content))
			// Extract just the content part (between <div class="container"> and </div>)
			markdownContent = b.extractMarkdownContent(fullHTML)
		}
	}

	// Prepare template data
	data := map[string]any{
		"System":          system,
		"Containers":      containers,
		"MarkdownContent": markdownContent,
		"HasMarkdown":     markdownContent != "",
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

// extractMarkdownContent extracts the content body from rendered HTML.
// It removes the HTML wrapper and CSS, returning just the content between <div class="container"> tags.
func (b *Builder) extractMarkdownContent(htmlContent string) string {
	// Find the start of the container div
	startIdx := strings.Index(htmlContent, "<div class=\"container\">")
	if startIdx == -1 {
		return ""
	}
	startIdx += len("<div class=\"container\">")

	// Find the end of the container div
	endIdx := strings.LastIndex(htmlContent, "</div>")
	if endIdx == -1 || endIdx <= startIdx {
		return ""
	}

	return htmlContent[startIdx:endIdx]
}

// BuildContainerPage generates a single container HTML page with embedded diagrams.
func (b *Builder) BuildContainerPage(_ context.Context, system *entities.System, container *entities.Container, components []*entities.Component, outputDir string) error {
	if system == nil {
		return fmt.Errorf("system cannot be nil")
	}
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Try to read and render markdown content
	markdownContent := ""
	if container.Path != "" {
		markdownPath := filepath.Join(container.Path, "container.md")
		if content, err := os.ReadFile(markdownPath); err == nil {
			// Render markdown to HTML fragment (content only, no HTML wrapper)
			fullHTML := b.markdownRenderer.RenderMarkdownToHTML(string(content))
			// Extract just the content part (between <div class="container"> and </div>)
			markdownContent = b.extractMarkdownContent(fullHTML)
		}
	}

	// Prepare template data
	data := map[string]any{
		"System":          system,
		"Container":       container,
		"Components":      components,
		"MarkdownContent": markdownContent,
		"HasMarkdown":     markdownContent != "",
	}

	// Render template
	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "container.html", data); err != nil {
		return fmt.Errorf("failed to render container template: %w", err)
	}

	// Write to file
	containersDir := filepath.Join(outputDir, "containers")
	if err := os.MkdirAll(containersDir, 0755); err != nil {
		return fmt.Errorf("failed to create containers directory: %w", err)
	}

	filePath := filepath.Join(containersDir, system.ID+"_"+container.ID+".html")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write container page %s: %w", filePath, err)
	}

	return nil
}

// buildContainersOverview generates a Level 2 overview page listing all containers.
func (b *Builder) buildContainersOverview(_ context.Context, systems []*entities.System, outputDir string) error {
	// Collect all containers from all systems
	type ContainerInfo struct {
		System    *entities.System
		Container *entities.Container
	}

	var allContainers []ContainerInfo
	for _, system := range systems {
		if system == nil {
			continue
		}
		for _, container := range system.ListContainers() {
			if container == nil {
				continue
			}
			allContainers = append(allContainers, ContainerInfo{
				System:    system,
				Container: container,
			})
		}
	}

	data := map[string]any{
		"Containers": allContainers,
		"Systems":    systems,
	}

	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "containers-overview.html", data); err != nil {
		return fmt.Errorf("failed to render containers overview template: %w", err)
	}

	filePath := filepath.Join(outputDir, "containers.html")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write containers overview page %s: %w", filePath, err)
	}

	return nil
}

// BuildComponentPage generates a single component HTML page.
func (b *Builder) BuildComponentPage(_ context.Context, system *entities.System, container *entities.Container, component *entities.Component, outputDir string) error {
	if system == nil {
		return fmt.Errorf("system cannot be nil")
	}
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}
	if outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Try to read and render markdown content
	markdownContent := ""
	if component.Path != "" {
		markdownPath := filepath.Join(component.Path, "component.md")
		if content, err := os.ReadFile(markdownPath); err == nil {
			// Render markdown to HTML fragment (content only, no HTML wrapper)
			fullHTML := b.markdownRenderer.RenderMarkdownToHTML(string(content))
			// Extract just the content part (between <div class="container"> and </div>)
			markdownContent = b.extractMarkdownContent(fullHTML)
		}
	}

	// Prepare template data
	data := map[string]any{
		"System":          system,
		"Container":       container,
		"Component":       component,
		"MarkdownContent": markdownContent,
		"HasMarkdown":     markdownContent != "",
	}

	// Render template
	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "component.html", data); err != nil {
		return fmt.Errorf("failed to render component template: %w", err)
	}

	// Write to file
	componentsDir := filepath.Join(outputDir, "components")
	if err := os.MkdirAll(componentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create components directory: %w", err)
	}

	filePath := filepath.Join(componentsDir, component.ID+".html")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write component page %s: %w", filePath, err)
	}

	return nil
}

// buildComponentsOverview generates a Level 3 overview page listing all components.
func (b *Builder) buildComponentsOverview(_ context.Context, systems []*entities.System, outputDir string) error {
	// Collect all components from all containers
	type ComponentInfo struct {
		System    *entities.System
		Container *entities.Container
		Component *entities.Component
	}

	var allComponents []ComponentInfo
	for _, system := range systems {
		if system == nil {
			continue
		}
		for _, container := range system.ListContainers() {
			if container == nil {
				continue
			}
			for _, component := range container.ListComponents() {
				if component == nil {
					continue
				}
				allComponents = append(allComponents, ComponentInfo{
					System:    system,
					Container: container,
					Component: component,
				})
			}
		}
	}

	data := map[string]any{
		"Components": allComponents,
		"Systems":    systems,
	}

	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "components-overview.html", data); err != nil {
		return fmt.Errorf("failed to render components overview template: %w", err)
	}

	filePath := filepath.Join(outputDir, "components.html")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write components overview page %s: %w", filePath, err)
	}

	return nil
}

// buildIndexPage generates the project index page.
func (b *Builder) buildIndexPage(_ context.Context, project *entities.Project, systems []*entities.System, outputDir string) error {
	data := map[string]any{
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
		filepath.Join(outputDir, "containers"),
		filepath.Join(outputDir, "components"),
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

// getDefaultCSSTokens returns the default design system tokens for CSS generation.
func getDefaultCSSTokens() map[string]string {
	return map[string]string{
		// Primary colors
		"color-primary":       "#2563eb",
		"color-primary-dark":  "#1e40af",
		"color-primary-light": "#dbeafe",

		// Semantic colors
		"color-success": "#10b981",
		"color-warning": "#f59e0b",
		"color-error":   "#ef4444",

		// Neutral colors
		"color-text":       "#1f2937",
		"color-text-light": "#6b7280",
		"color-bg":         "#ffffff",
		"color-bg-alt":     "#f9fafb",
		"color-border":     "#e5e7eb",

		// Spacing
		"spacing-xs":  "0.25rem",
		"spacing-sm":  "0.5rem",
		"spacing-md":  "1rem",
		"spacing-lg":  "1.5rem",
		"spacing-xl":  "2rem",
		"spacing-2xl": "3rem",

		// Typography
		"font-family":   "-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
		"font-mono":     "\"Menlo\", \"Monaco\", \"Courier New\", monospace",
		"border-radius": "0.375rem",
		"shadow-sm":     "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
		"shadow-md":     "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
		"shadow-lg":     "0 10px 15px -3px rgba(0, 0, 0, 0.1)",
	}
}
