package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/adapters/cli"
	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/encoding"
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
	projectRepo := filesystem.NewProjectRepository()
	project, err := projectRepo.LoadProject(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	c.setupTemplateEngine(project, projectRepo)

	systems, err := projectRepo.ListSystems(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to list systems: %w", err)
	}
	if len(systems) == 0 {
		fmt.Println("No systems found to build")
		return nil
	}

	outputFormats := c.parseFormats()
	if len(outputFormats) == 0 {
		outputFormats = []usecases.OutputFormat{usecases.FormatHTML}
	}

	buildDocs, err := c.createBuildUseCase(outputFormats)
	if err != nil {
		return err
	}

	startTime := time.Now()
	err = buildDocs.ExecuteWithFormats(ctx, project, systems, c.outputDir, usecases.BuildDocsOptions{Formats: outputFormats})
	elapsed := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if containsFormat(outputFormats, usecases.FormatHTML) {
		if err := c.renderMarkdown(ctx, project, systems); err != nil {
			return err
		}
	}

	fmt.Printf("✓ Build completed in %v\n", elapsed.Round(10*time.Millisecond))
	fmt.Printf("✓ Output: %s\n", c.outputDir)
	return nil
}

// setupTemplateEngine configures the template engine search paths on the repository.
func (c *BuildCommand) setupTemplateEngine(project *entities.Project, projectRepo *filesystem.ProjectRepository) {
	templateName := "standard-3layer"
	if project.Config != nil && project.Config.Template != "" {
		templateName = project.Config.Template
	}

	templateEngine := ason.NewTemplateEngine()
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		templateEngine.AddSearchPath(filepath.Join(exeDir, "..", "templates", templateName))
		templateEngine.AddSearchPath(filepath.Join(".", "templates", templateName))
	}
	if envDir, ok := os.LookupEnv("LOKO_TEMPLATE_DIR"); ok && envDir != "" {
		templateEngine.AddSearchPath(envDir)
	}
	projectRepo.SetTemplateEngine(templateEngine)
}

// createBuildUseCase creates and configures the BuildDocs use case with required adapters.
func (c *BuildCommand) createBuildUseCase(outputFormats []usecases.OutputFormat) (*usecases.BuildDocs, error) {
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return nil, fmt.Errorf("failed to create site builder: %w", err)
	}

	progressReporter := cli.NewProgressReporter()
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, progressReporter)

	if containsFormat(outputFormats, usecases.FormatMarkdown) {
		buildDocs.WithMarkdownBuilder(markdown.NewBuilder())
	}
	if containsFormat(outputFormats, usecases.FormatPDF) {
		pdfRenderer := pdf.NewRenderer()
		if !pdfRenderer.IsAvailable() {
			return nil, fmt.Errorf(`PDF output requested but veve-cli is not installed

veve-cli is required for PDF generation. Install it with:

  # macOS
  brew install terrastruct/tap/veve

  # Linux
  curl -fsSL https://github.com/terrastruct/veve/releases/latest/download/veve-linux-amd64 -o /usr/local/bin/veve-cli
  chmod +x /usr/local/bin/veve-cli

  # Windows
  scoop install veve

Or build HTML/Markdown only:
  loko build --format html,markdown

For more info: https://github.com/terrastruct/veve`)
		}
		buildDocs.WithPDFRenderer(pdfRenderer)
	}
	if containsFormat(outputFormats, usecases.FormatTOON) {
		encoder := encoding.NewEncoder()
		buildDocs.WithOutputEncoder(encoder)
	}

	return buildDocs, nil
}

// renderMarkdown renders markdown documentation files to HTML.
func (c *BuildCommand) renderMarkdown(ctx context.Context, project *entities.Project, systems []*entities.System) error {
	progressReporter := cli.NewProgressReporter()
	markdownRenderer := html.NewMarkdownRenderer("", "")
	renderMarkdownDocs := usecases.NewRenderMarkdownDocs(markdownRenderer, progressReporter)
	if err := renderMarkdownDocs.Execute(ctx, project, systems, c.outputDir); err != nil {
		return fmt.Errorf("markdown rendering failed: %w", err)
	}
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
		case "toon":
			format = usecases.FormatTOON
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
