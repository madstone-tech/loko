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
	// FormatTOON generates TOON (Token-Optimized Object Notation) format for LLM consumption.
	FormatTOON OutputFormat = "toon"
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

// Public entry point: (uc *BuildDocs) ExecuteWithFormats(ctx, project, systems, outputDir, options)
// is the primary method called by cmd/build.go. The legacy Execute method (single-format HTML)
// is retained for backward compatibility. Full split of this file into per-format files is
// deferred to US4 tasks T062–T066.
//
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
	outputEncoder    OutputEncoder
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

// WithOutputEncoder sets the output encoder for TOON/JSON output.
func (uc *BuildDocs) WithOutputEncoder(oe OutputEncoder) *BuildDocs {
	uc.outputEncoder = oe
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

	// Render all diagrams in parallel
	if err := uc.renderDiagrams(ctx, systems, outputDir); err != nil {
		return err
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
		case FormatTOON:
			if uc.outputEncoder == nil {
				return fmt.Errorf("output encoder not configured")
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

		case FormatTOON:
			uc.progressReporter.ReportInfo("Building TOON documentation...")
			// Build architecture graph for TOON export
			graphBuilder := NewBuildArchitectureGraph()
			graph, err := graphBuilder.Execute(ctx, project, systems)
			if err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to build architecture graph: %w", err))
				return fmt.Errorf("failed to build architecture graph: %w", err)
			}

			// Encode architecture to TOON format
			toonData, err := uc.outputEncoder.EncodeTOON(graph)
			if err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to encode TOON: %w", err))
				return fmt.Errorf("failed to encode TOON: %w", err)
			}

			// Write architecture.toon
			toonPath := filepath.Join(outputDir, "architecture.toon")
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			if err := os.WriteFile(toonPath, toonData, 0644); err != nil {
				return fmt.Errorf("failed to write architecture.toon: %w", err)
			}
			uc.progressReporter.ReportSuccess("TOON documentation built: architecture.toon")
		}
	}

	uc.progressReporter.ReportSuccess(fmt.Sprintf("All documentation built in %s", outputDir))
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
