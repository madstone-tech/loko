package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// OutputFormat represents a documentation output format.
type OutputFormat string

const (
	// FormatHTML generates HTML documentation.
	FormatHTML OutputFormat = "html"
	// FormatMarkdown generates a single README.md file.
	FormatMarkdown OutputFormat = "markdown"
	// FormatPDF generates PDF documentation from HTML.
	FormatPDF OutputFormat = "pdf"
)

// BuildDocsOptions configures what output formats to generate.
type BuildDocsOptions struct {
	// Formats specifies which output formats to generate.
	// If empty, defaults to HTML only.
	Formats []OutputFormat
}

// DefaultBuildDocsOptions returns the default build options (HTML only).
func DefaultBuildDocsOptions() BuildDocsOptions {
	return BuildDocsOptions{
		Formats: []OutputFormat{FormatHTML},
	}
}

// BuildDocs orchestrates the process of rendering diagrams and building documentation.
//
// This use case:
// 1. Iterates through all systems, containers, and components
// 2. Renders D2 diagrams to SVG using the DiagramRenderer (C4 levels 1-3)
// 3. Calls the SiteBuilder to generate HTML documentation
// 4. Optionally generates Markdown and PDF outputs
// 5. Reports progress via ProgressReporter
type BuildDocs struct {
	diagramRenderer  DiagramRenderer
	siteBuilder      SiteBuilder
	markdownBuilder  MarkdownBuilder
	pdfRenderer      PDFRenderer
	progressReporter ProgressReporter
}

// NewBuildDocs creates a new BuildDocs use case with the given adapters.
func NewBuildDocs(
	diagramRenderer DiagramRenderer,
	siteBuilder SiteBuilder,
	progressReporter ProgressReporter,
) *BuildDocs {
	return &BuildDocs{
		diagramRenderer:  diagramRenderer,
		siteBuilder:      siteBuilder,
		progressReporter: progressReporter,
	}
}

// WithMarkdownBuilder sets the markdown builder for markdown output.
func (uc *BuildDocs) WithMarkdownBuilder(mb MarkdownBuilder) *BuildDocs {
	uc.markdownBuilder = mb
	return uc
}

// WithPDFRenderer sets the PDF renderer for PDF output.
func (uc *BuildDocs) WithPDFRenderer(pr PDFRenderer) *BuildDocs {
	uc.pdfRenderer = pr
	return uc
}

// Execute performs a complete documentation build.
//
// It:
// 1. Renders all diagrams in the project (systems, containers, and components)
// 2. Calls BuildSite to generate HTML documentation
// 3. Reports progress and errors
// 4. Returns error if any rendering fails
func (uc *BuildDocs) Execute(
	ctx context.Context,
	project *entities.Project,
	systems []*entities.System,
	outputDir string,
) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}
	if len(systems) == 0 {
		uc.progressReporter.ReportInfo("No systems found to build")
		return nil
	}

	uc.progressReporter.ReportInfo("Starting documentation build...")

	// Count total diagrams for progress reporting
	totalDiagrams := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			totalDiagrams++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				totalDiagrams++
			}
			for _, component := range container.Components {
				if component.Diagram != nil {
					totalDiagrams++
				}
			}
		}
	}

	// Render all diagrams and save to files
	diagramCount := 0
	for _, sys := range systems {
		// Render system diagram
		if sys.Diagram != nil {
			diagramCount++
			uc.progressReporter.ReportProgress(
				fmt.Sprintf("Rendering system diagram: %s", sys.Name),
				diagramCount,
				totalDiagrams,
				fmt.Sprintf("Rendering %s diagram", sys.Name),
			)

			// Create unique filename for diagram
			diagramFileName := fmt.Sprintf("%s.svg", sys.ID)
			diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

			svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, sys.Diagram.Source)
			if err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to render diagram for system %s: %w", sys.Name, err))
				return fmt.Errorf("failed to render diagram for system %s: %w", sys.Name, err)
			}

			// Save SVG to file
			if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
				return fmt.Errorf("failed to create diagrams directory: %w", err)
			}
			if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
				return fmt.Errorf("failed to save diagram for system %s: %w", sys.Name, err)
			}

			// Store diagram path in entity for use by HTML builder
			sys.DiagramPath = filepath.Join("diagrams", diagramFileName)
		}

		// Render container diagrams
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				diagramCount++
				uc.progressReporter.ReportProgress(
					fmt.Sprintf("Rendering container diagram: %s/%s", sys.Name, container.Name),
					diagramCount,
					totalDiagrams,
					fmt.Sprintf("Rendering %s diagram in %s", container.Name, sys.Name),
				)

				// Create unique filename for diagram
				diagramFileName := fmt.Sprintf("%s_%s.svg", sys.ID, container.ID)
				diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

				svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, container.Diagram.Source)
				if err != nil {
					uc.progressReporter.ReportError(fmt.Errorf("failed to render diagram for container %s/%s: %w", sys.Name, container.Name, err))
					return fmt.Errorf("failed to render diagram for container %s/%s: %w", sys.Name, container.Name, err)
				}

				// Save SVG to file
				if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
					return fmt.Errorf("failed to create diagrams directory: %w", err)
				}
				if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
					return fmt.Errorf("failed to save diagram for container %s/%s: %w", sys.Name, container.Name, err)
				}

				// Store diagram path in entity for use by HTML builder
				container.DiagramPath = filepath.Join("diagrams", diagramFileName)
			}

			// Render component diagrams
			enhancer := NewEnhanceComponentDiagram()
			for _, component := range container.Components {
				if component.Diagram != nil {
					diagramCount++
					uc.progressReporter.ReportProgress(
						fmt.Sprintf("Rendering component diagram: %s/%s/%s", sys.Name, container.Name, component.Name),
						diagramCount,
						totalDiagrams,
						fmt.Sprintf("Rendering %s diagram in %s", component.Name, container.Name),
					)

					// Enhance component diagram with relationships and metadata
					enhancedD2Source, err := enhancer.Execute(component, container, sys)
					if err != nil {
						uc.progressReporter.ReportError(fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err))
						return fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					// Create unique filename for diagram
					diagramFileName := fmt.Sprintf("%s_%s_%s.svg", sys.ID, container.ID, component.ID)
					diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

					svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, enhancedD2Source)
					if err != nil {
						uc.progressReporter.ReportError(fmt.Errorf("failed to render diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err))
						return fmt.Errorf("failed to render diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					// Save SVG to file
					if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
						return fmt.Errorf("failed to create diagrams directory: %w", err)
					}
					if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
						return fmt.Errorf("failed to save diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					// Store diagram path in entity for use by HTML builder
					component.DiagramPath = filepath.Join("diagrams", diagramFileName)
				}
			}
		}
	}

	// Build the site
	uc.progressReporter.ReportProgress("Building site", len(systems), len(systems), "Generating HTML documentation...")
	err := uc.siteBuilder.BuildSite(ctx, project, systems, outputDir)
	if err != nil {
		uc.progressReporter.ReportError(fmt.Errorf("failed to build site: %w", err))
		return fmt.Errorf("failed to build site: %w", err)
	}

	uc.progressReporter.ReportSuccess(fmt.Sprintf("Documentation built successfully in %s", outputDir))
	return nil
}

// ExecuteWithFormats performs a documentation build with specified output formats.
func (uc *BuildDocs) ExecuteWithFormats(
	ctx context.Context,
	project *entities.Project,
	systems []*entities.System,
	outputDir string,
	options BuildDocsOptions,
) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	// Default to HTML if no formats specified
	formats := options.Formats
	if len(formats) == 0 {
		formats = []OutputFormat{FormatHTML}
	}

	// Check for required adapters
	for _, format := range formats {
		switch format {
		case FormatMarkdown:
			if uc.markdownBuilder == nil {
				return fmt.Errorf("markdown builder not configured")
			}
		case FormatPDF:
			if uc.pdfRenderer == nil {
				return fmt.Errorf("PDF renderer not configured")
			}
			if !uc.pdfRenderer.IsAvailable() {
				return fmt.Errorf("PDF renderer (veve-cli) not available")
			}
		}
	}

	// First, render diagrams (needed for HTML and PDF)
	needsDiagrams := containsFormat(formats, FormatHTML) || containsFormat(formats, FormatPDF)
	if needsDiagrams && len(systems) > 0 {
		if err := uc.renderDiagrams(ctx, systems, outputDir); err != nil {
			return err
		}
	}

	// Build each format
	for _, format := range formats {
		switch format {
		case FormatHTML:
			uc.progressReporter.ReportInfo("Building HTML documentation...")
			if err := uc.siteBuilder.BuildSite(ctx, project, systems, outputDir); err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to build HTML: %w", err))
				return fmt.Errorf("failed to build HTML: %w", err)
			}
			uc.progressReporter.ReportSuccess("HTML documentation built")

		case FormatMarkdown:
			uc.progressReporter.ReportInfo("Building Markdown documentation...")
			content, err := uc.markdownBuilder.BuildMarkdown(ctx, project, systems)
			if err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to build markdown: %w", err))
				return fmt.Errorf("failed to build markdown: %w", err)
			}

			// Write README.md
			readmePath := filepath.Join(outputDir, "README.md")
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write README.md: %w", err)
			}
			uc.progressReporter.ReportSuccess("Markdown documentation built: README.md")

		case FormatPDF:
			uc.progressReporter.ReportInfo("Building PDF documentation...")
			// PDF requires HTML to be built first
			htmlPath := filepath.Join(outputDir, "index.html")
			if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
				// Build HTML first if not already built
				if err := uc.siteBuilder.BuildSite(ctx, project, systems, outputDir); err != nil {
					return fmt.Errorf("failed to build HTML for PDF: %w", err)
				}
			}

			pdfPath := filepath.Join(outputDir, "architecture.pdf")
			if err := uc.pdfRenderer.RenderPDF(ctx, htmlPath, pdfPath); err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to build PDF: %w", err))
				return fmt.Errorf("failed to build PDF: %w", err)
			}
			uc.progressReporter.ReportSuccess("PDF documentation built: architecture.pdf")
		}
	}

	uc.progressReporter.ReportSuccess(fmt.Sprintf("All documentation built in %s", outputDir))
	return nil
}

// renderDiagrams renders all D2 diagrams to SVG files.
func (uc *BuildDocs) renderDiagrams(
	ctx context.Context,
	systems []*entities.System,
	outputDir string,
) error {
	// Count total diagrams
	totalDiagrams := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			totalDiagrams++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				totalDiagrams++
			}
			for _, component := range container.Components {
				if component.Diagram != nil {
					totalDiagrams++
				}
			}
		}
	}

	if totalDiagrams == 0 {
		return nil
	}

	uc.progressReporter.ReportInfo(fmt.Sprintf("Rendering %d diagrams...", totalDiagrams))

	diagramCount := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			diagramCount++
			diagramFileName := fmt.Sprintf("%s.svg", sys.ID)
			diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

			svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, sys.Diagram.Source)
			if err != nil {
				return fmt.Errorf("failed to render diagram for system %s: %w", sys.Name, err)
			}

			if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
				return fmt.Errorf("failed to create diagrams directory: %w", err)
			}
			if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
				return fmt.Errorf("failed to save diagram for system %s: %w", sys.Name, err)
			}
			sys.DiagramPath = filepath.Join("diagrams", diagramFileName)
		}

		for _, container := range sys.Containers {
			if container.Diagram != nil {
				diagramCount++
				diagramFileName := fmt.Sprintf("%s_%s.svg", sys.ID, container.ID)
				diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

				svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, container.Diagram.Source)
				if err != nil {
					return fmt.Errorf("failed to render diagram for container %s/%s: %w", sys.Name, container.Name, err)
				}

				if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
					return fmt.Errorf("failed to create diagrams directory: %w", err)
				}
				if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
					return fmt.Errorf("failed to save diagram for container %s/%s: %w", sys.Name, container.Name, err)
				}
				container.DiagramPath = filepath.Join("diagrams", diagramFileName)
			}

			enhancer := NewEnhanceComponentDiagram()
			for _, component := range container.Components {
				if component.Diagram != nil {
					diagramCount++
					enhancedD2Source, err := enhancer.Execute(component, container, sys)
					if err != nil {
						return fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					diagramFileName := fmt.Sprintf("%s_%s_%s.svg", sys.ID, container.ID, component.ID)
					diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

					svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, enhancedD2Source)
					if err != nil {
						return fmt.Errorf("failed to render diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
						return fmt.Errorf("failed to create diagrams directory: %w", err)
					}
					if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
						return fmt.Errorf("failed to save diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}
					component.DiagramPath = filepath.Join("diagrams", diagramFileName)
				}
			}
		}
	}

	uc.progressReporter.ReportProgress("Diagrams", diagramCount, totalDiagrams, "All diagrams rendered")
	return nil
}

// containsFormat checks if a format is in the list.
func containsFormat(formats []OutputFormat, format OutputFormat) bool {
	for _, f := range formats {
		if f == format {
			return true
		}
	}
	return false
}
