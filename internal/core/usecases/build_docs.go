package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

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

// diagramJob represents a single diagram rendering task.
type diagramJob struct {
	source   string // D2 source code to render
	fileName string // Output SVG filename (e.g., "sys-id.svg")
	label    string // Human-readable label for progress (e.g., "system PaymentService")
}

// diagramResult holds the outcome of a diagram rendering job.
type diagramResult struct {
	index      int
	svgContent string
	err        error
}

// renderDiagrams renders all D2 diagrams to SVG files using a parallel worker pool.
func (uc *BuildDocs) renderDiagrams(
	ctx context.Context,
	systems []*entities.System,
	outputDir string,
) error {
	// Collect all diagram jobs
	type pathSetter func(path string)
	var jobs []diagramJob
	var setters []pathSetter

	enhancer := NewEnhanceComponentDiagram()

	for _, sys := range systems {
		if sys.Diagram != nil {
			fileName := fmt.Sprintf("%s.svg", sys.ID)
			source := sys.Diagram.Source
			jobs = append(jobs, diagramJob{
				source:   source,
				fileName: fileName,
				label:    fmt.Sprintf("system %s", sys.Name),
			})
			s := sys // capture for closure
			setters = append(setters, func(path string) { s.DiagramPath = path })
		}

		for _, container := range sys.Containers {
			if container.Diagram != nil {
				fileName := fmt.Sprintf("%s_%s.svg", sys.ID, container.ID)
				jobs = append(jobs, diagramJob{
					source:   container.Diagram.Source,
					fileName: fileName,
					label:    fmt.Sprintf("container %s/%s", sys.Name, container.Name),
				})
				c := container
				setters = append(setters, func(path string) { c.DiagramPath = path })
			}

			for _, component := range container.Components {
				if component.Diagram != nil {
					enhancedSource, err := enhancer.Execute(component, container, sys)
					if err != nil {
						return fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w",
							sys.Name, container.Name, component.Name, err)
					}
					fileName := fmt.Sprintf("%s_%s_%s.svg", sys.ID, container.ID, component.ID)
					jobs = append(jobs, diagramJob{
						source:   enhancedSource,
						fileName: fileName,
						label:    fmt.Sprintf("component %s/%s/%s", sys.Name, container.Name, component.Name),
					})
					comp := component
					setters = append(setters, func(path string) { comp.DiagramPath = path })
				}
			}
		}
	}

	if len(jobs) == 0 {
		return nil
	}

	uc.progressReporter.ReportInfo(fmt.Sprintf("Rendering %d diagrams...", len(jobs)))

	// Create diagrams directory once
	diagramsDir := filepath.Join(outputDir, "diagrams")
	if err := os.MkdirAll(diagramsDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagrams directory: %w", err)
	}

	// Determine worker count
	numWorkers := min(8, len(jobs))

	// Channel-based worker pool
	jobCh := make(chan int, len(jobs))
	resultCh := make(chan diagramResult, len(jobs))

	// Start workers
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for idx := range jobCh {
				job := jobs[idx]
				svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, job.source)
				resultCh <- diagramResult{index: idx, svgContent: svgContent, err: err}
			}
		})
	}

	// Send all jobs
	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	completed := 0
	for result := range resultCh {
		completed++
		job := jobs[result.index]

		if result.err != nil {
			return fmt.Errorf("failed to render diagram for %s: %w", job.label, result.err)
		}

		uc.progressReporter.ReportProgress(
			fmt.Sprintf("Rendered %s", job.label),
			completed, len(jobs),
			fmt.Sprintf("Rendering diagrams (%d/%d)", completed, len(jobs)),
		)

		// Write SVG to disk
		diagramPath := filepath.Join(diagramsDir, job.fileName)
		if err := os.WriteFile(diagramPath, []byte(result.svgContent), 0644); err != nil {
			return fmt.Errorf("failed to save diagram for %s: %w", job.label, err)
		}

		// Set diagram path on entity
		setters[result.index](filepath.Join("diagrams", job.fileName))
	}

	uc.progressReporter.ReportProgress("Diagrams", len(jobs), len(jobs), "All diagrams rendered")
	return nil
}

// GenerateComponentTable generates a Markdown table of components in a container.
// Returns a table with columns: Name, Technology, Description.
// If container has no components, returns an empty string.
func GenerateComponentTable(container *entities.Container) string {
	if container == nil || container.Components == nil || len(container.Components) == 0 {
		return ""
	}

	// Get all components and sort by name
	components := container.ListComponents()

	// Sort components by name
	slices.SortFunc(components, func(a, b *entities.Component) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Build the markdown table
	var sb strings.Builder
	sb.WriteString("| Name | Technology | Description |\n")
	sb.WriteString("|------|------------|-------------|\n")

	for _, comp := range components {
		// Escape pipe characters in description to avoid breaking table
		description := strings.ReplaceAll(comp.Description, "|", "\\|")
		technology := strings.ReplaceAll(comp.Technology, "|", "\\|")
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", comp.Name, technology, description))
	}

	return sb.String()
}

// GenerateContainerTable generates a Markdown table of containers in a system.
// Returns a table with columns: Name, Technology, Description.
func GenerateContainerTable(system *entities.System) string {
	if system == nil || system.Containers == nil || len(system.Containers) == 0 {
		return ""
	}

	// Get all containers and sort by name
	containers := system.ListContainers()

	// Sort containers by name
	slices.SortFunc(containers, func(a, b *entities.Container) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Build the markdown table
	var sb strings.Builder
	sb.WriteString("| Name | Technology | Description |\n")
	sb.WriteString("|------|------------|-------------|\n")

	for _, cont := range containers {
		// Escape pipe characters in description to avoid breaking table
		description := strings.ReplaceAll(cont.Description, "|", "\\|")
		technology := strings.ReplaceAll(cont.Technology, "|", "\\|")
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", cont.Name, technology, description))
	}

	return sb.String()
}

// containsFormat checks if a format is in the list.
func containsFormat(formats []OutputFormat, format OutputFormat) bool {
	return slices.Contains(formats, format)
}
