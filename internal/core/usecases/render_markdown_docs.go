package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// RenderMarkdownDocs is a use case that renders markdown files to HTML documentation.
// It reads system.md, container.md, and component.md files and converts them to
// standalone HTML pages with visualization alongside D2 diagrams.
type RenderMarkdownDocs struct {
	markdownRenderer MarkdownRenderer
	progressReporter ProgressReporter
}

// NewRenderMarkdownDocs creates a new RenderMarkdownDocs use case.
func NewRenderMarkdownDocs(
	markdownRenderer MarkdownRenderer,
	progressReporter ProgressReporter,
) *RenderMarkdownDocs {
	return &RenderMarkdownDocs{
		markdownRenderer: markdownRenderer,
		progressReporter: progressReporter,
	}
}

// Execute renders all markdown files in a project to HTML.
// It iterates through systems, containers, and components,
// reads their associated markdown files, and renders them as HTML with embedded diagrams.
func (uc *RenderMarkdownDocs) Execute(
	ctx context.Context,
	project *entities.Project,
	systems []*entities.System,
	outputDir string,
) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}
	if len(systems) == 0 {
		uc.progressReporter.ReportInfo("No systems found to render")
		return nil
	}

	uc.progressReporter.ReportInfo("Starting markdown rendering...")

	// Count total items for progress reporting
	totalCount := 0
	for _, sys := range systems {
		totalCount++ // system itself
		totalCount += len(sys.Containers)
		for _, container := range sys.Containers {
			totalCount += len(container.Components)
		}
	}

	count := 0
	for _, sys := range systems {
		count++
		uc.progressReporter.ReportProgress(
			fmt.Sprintf("Rendering system markdown: %s", sys.Name),
			count,
			totalCount,
			fmt.Sprintf("Rendering %s markdown", sys.Name),
		)

		// Render system markdown
		if err := uc.renderSystemMarkdown(ctx, sys, outputDir); err != nil {
			uc.progressReporter.ReportError(fmt.Errorf("failed to render markdown for system %s: %w", sys.Name, err))
			return fmt.Errorf("failed to render markdown for system %s: %w", sys.Name, err)
		}

		// Render container markdowns
		for _, container := range sys.Containers {
			count++
			uc.progressReporter.ReportProgress(
				fmt.Sprintf("Rendering container markdown: %s/%s", sys.Name, container.Name),
				count,
				totalCount,
				fmt.Sprintf("Rendering %s markdown", container.Name),
			)

			if err := uc.renderContainerMarkdown(ctx, sys, container, outputDir); err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to render markdown for container %s/%s: %w", sys.Name, container.Name, err))
				return fmt.Errorf("failed to render markdown for container %s/%s: %w", sys.Name, container.Name, err)
			}

			// Render component markdowns
			for _, component := range container.Components {
				count++
				uc.progressReporter.ReportProgress(
					fmt.Sprintf("Rendering component markdown: %s/%s/%s", sys.Name, container.Name, component.Name),
					count,
					totalCount,
					fmt.Sprintf("Rendering %s markdown", component.Name),
				)

				if err := uc.renderComponentMarkdown(ctx, sys, container, component, outputDir); err != nil {
					uc.progressReporter.ReportError(fmt.Errorf("failed to render markdown for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err))
					return fmt.Errorf("failed to render markdown for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
				}
			}
		}
	}

	uc.progressReporter.ReportSuccess(fmt.Sprintf("Markdown rendering completed in %s", outputDir))
	return nil
}

// renderSystemMarkdown renders a system's markdown to HTML.
func (uc *RenderMarkdownDocs) renderSystemMarkdown(_ context.Context, system *entities.System, outputDir string) error {
	markdownPath := filepath.Join(system.Path, "system.md")
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		// If markdown file doesn't exist, skip rendering
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Replace container table placeholder
	contentStr := string(content)
	if strings.Contains(contentStr, "{{container_table}}") {
		containerTable := GenerateContainerTable(system)
		contentStr = strings.ReplaceAll(contentStr, "{{container_table}}", containerTable)
	}

	htmlContent := uc.markdownRenderer.RenderMarkdownToHTML(contentStr)

	// Create output directory
	htmlDir := filepath.Join(outputDir, "markdown", "systems")
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file
	htmlPath := filepath.Join(htmlDir, system.ID+".html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// renderContainerMarkdown renders a container's markdown to HTML.
func (uc *RenderMarkdownDocs) renderContainerMarkdown(_ context.Context, system *entities.System, container *entities.Container, outputDir string) error {
	markdownPath := filepath.Join(container.Path, "container.md")
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		// If markdown file doesn't exist, skip rendering
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Replace component table placeholder
	contentStr := string(content)
	if strings.Contains(contentStr, "{{component_table}}") {
		componentTable := GenerateComponentTable(container)
		contentStr = strings.ReplaceAll(contentStr, "{{component_table}}", componentTable)
	}

	htmlContent := uc.markdownRenderer.RenderMarkdownToHTML(contentStr)

	// Create output directory
	htmlDir := filepath.Join(outputDir, "markdown", "containers")
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file with naming: system_container.html
	htmlPath := filepath.Join(htmlDir, system.ID+"_"+container.ID+".html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// renderComponentMarkdown renders a component's markdown to HTML.
func (uc *RenderMarkdownDocs) renderComponentMarkdown(_ context.Context, system *entities.System, container *entities.Container, component *entities.Component, outputDir string) error {
	markdownPath := filepath.Join(component.Path, "component.md")
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		// If markdown file doesn't exist, skip rendering
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	htmlContent := uc.markdownRenderer.RenderMarkdownToHTML(string(content))

	// Create output directory
	htmlDir := filepath.Join(outputDir, "markdown", "components")
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file with naming: system_container_component.html
	htmlPath := filepath.Join(htmlDir, system.ID+"_"+container.ID+"_"+component.ID+".html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}
