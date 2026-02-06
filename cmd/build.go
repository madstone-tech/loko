package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/adapters/markdown"
	"github.com/madstone-tech/loko/internal/adapters/pdf"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// BuildCommand builds the documentation.
type BuildCommand struct {
	projectRoot string
	clean       bool
	outputDir   string
	formats     []string // Output formats: html, markdown, pdf
}

// NewBuildCommand creates a new build command.
func NewBuildCommand(projectRoot string) *BuildCommand {
	return &BuildCommand{
		projectRoot: projectRoot,
		clean:       false,
		outputDir:   "dist",
		formats:     []string{"html"}, // Default to HTML only
	}
}

// WithClean sets whether to rebuild everything (ignore cache).
func (c *BuildCommand) WithClean(clean bool) *BuildCommand {
	c.clean = clean
	return c
}

// WithOutputDir sets the output directory.
func (c *BuildCommand) WithOutputDir(dir string) *BuildCommand {
	c.outputDir = dir
	return c
}

// WithFormats sets the output formats (html, markdown, pdf).
func (c *BuildCommand) WithFormats(formats []string) *BuildCommand {
	if len(formats) > 0 {
		c.formats = formats
	}
	return c
}

// WithFormat adds a single output format.
func (c *BuildCommand) WithFormat(format string) *BuildCommand {
	format = strings.ToLower(strings.TrimSpace(format))
	if format != "" {
		c.formats = append(c.formats, format)
	}
	return c
}

// Execute runs the build command.
func (c *BuildCommand) Execute(ctx context.Context) error {
	// Load the project first to get template configuration
	projectRepo := filesystem.NewProjectRepository()
	project, err := projectRepo.LoadProject(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine template name from config (default: standard-3layer)
	templateName := "standard-3layer"
	if project.Config != nil && project.Config.Template != "" {
		templateName = project.Config.Template
	}

	// Create template engine and add search path
	templateEngine := ason.NewTemplateEngine()

	// Find template directory relative to binary location
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		templateDir := filepath.Join(exeDir, "..", "templates", templateName)
		templateEngine.AddSearchPath(templateDir)

		// Also try relative to current directory
		templateEngine.AddSearchPath(filepath.Join(".", "templates", templateName))
	}

	// Allow override via environment variable
	if envTemplateDir, ok := os.LookupEnv("LOKO_TEMPLATE_DIR"); ok && envTemplateDir != "" {
		templateEngine.AddSearchPath(envTemplateDir)
	}

	// Set template engine on repository
	projectRepo.SetTemplateEngine(templateEngine)

	// List all systems
	systems, err := projectRepo.ListSystems(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to list systems: %w", err)
	}

	if len(systems) == 0 {
		fmt.Println("No systems found to build")
		return nil
	}

	// Parse output formats
	outputFormats := c.parseFormats()
	if len(outputFormats) == 0 {
		outputFormats = []usecases.OutputFormat{usecases.FormatHTML}
	}

	// Print what we're building
	formatNames := make([]string, len(outputFormats))
	for i, f := range outputFormats {
		formatNames[i] = string(f)
	}
	fmt.Printf("Building documentation: %s\n", strings.Join(formatNames, ", "))

	// Create adapters
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return fmt.Errorf("failed to create site builder: %w", err)
	}

	// Create progress reporter (simple console output)
	progressReporter := &simpleProgressReporter{}

	// Create build use case with optional adapters
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, progressReporter)

	// Add markdown builder if needed
	if containsFormat(outputFormats, usecases.FormatMarkdown) {
		markdownBuilder := markdown.NewBuilder()
		buildDocs.WithMarkdownBuilder(markdownBuilder)
	}

	// Add PDF renderer if needed
	if containsFormat(outputFormats, usecases.FormatPDF) {
		pdfRenderer := pdf.NewRenderer()
		if !pdfRenderer.IsAvailable() {
			return fmt.Errorf("PDF output requested but veve-cli is not installed")
		}
		buildDocs.WithPDFRenderer(pdfRenderer)
	}

	startTime := time.Now()

	// Build with specified formats
	options := usecases.BuildDocsOptions{
		Formats: outputFormats,
	}
	err = buildDocs.ExecuteWithFormats(ctx, project, systems, c.outputDir, options)
	elapsed := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Render markdown files to HTML (after diagrams are built) - only if HTML is enabled
	if containsFormat(outputFormats, usecases.FormatHTML) {
		fmt.Println("\n📄 Rendering markdown documentation...")
		markdownRenderer := html.NewMarkdownRenderer("", "")
		renderMarkdownDocs := usecases.NewRenderMarkdownDocs(markdownRenderer, progressReporter)
		err = renderMarkdownDocs.Execute(ctx, project, systems, c.outputDir)
		if err != nil {
			return fmt.Errorf("markdown rendering failed: %w", err)
		}
	}

	fmt.Printf("✓ Build completed in %v\n", elapsed.Round(10*time.Millisecond))
	fmt.Printf("✓ Output: %s\n", c.outputDir)

	return nil
}

// parseFormats converts string format names to OutputFormat constants.
func (c *BuildCommand) parseFormats() []usecases.OutputFormat {
	var formats []usecases.OutputFormat
	seen := make(map[usecases.OutputFormat]bool)

	for _, f := range c.formats {
		var format usecases.OutputFormat
		switch strings.ToLower(strings.TrimSpace(f)) {
		case "html":
			format = usecases.FormatHTML
		case "markdown", "md":
			format = usecases.FormatMarkdown
		case "pdf":
			format = usecases.FormatPDF
		default:
			fmt.Printf("Warning: unknown format %q, skipping\n", f)
			continue
		}

		if !seen[format] {
			seen[format] = true
			formats = append(formats, format)
		}
	}

	return formats
}

// containsFormat checks if a format is in the list.
func containsFormat(formats []usecases.OutputFormat, format usecases.OutputFormat) bool {
	for _, f := range formats {
		if f == format {
			return true
		}
	}
	return false
}

// simpleProgressReporter implements ProgressReporter for console output.
type simpleProgressReporter struct {
}

// ReportProgress reports progress.
func (r *simpleProgressReporter) ReportProgress(step string, current int, total int, message string) {
	if total > 0 {
		percent := (current * 100) / total
		fmt.Printf("  [%3d%%] %s\n", percent, message)
	} else {
		fmt.Printf("  %s\n", message)
	}
}

// ReportError reports an error.
func (r *simpleProgressReporter) ReportError(err error) {
	fmt.Printf("  ✗ Error: %v\n", err)
}

// ReportSuccess reports success.
func (r *simpleProgressReporter) ReportSuccess(message string) {
	fmt.Printf("  ✓ %s\n", message)
}

// ReportInfo reports info.
func (r *simpleProgressReporter) ReportInfo(message string) {
	fmt.Printf("  ℹ %s\n", message)
}
