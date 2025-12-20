package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// BuildCommand builds the documentation.
type BuildCommand struct {
	projectRoot string
	clean       bool
	outputDir   string
}

// NewBuildCommand creates a new build command.
func NewBuildCommand(projectRoot string) *BuildCommand {
	return &BuildCommand{
		projectRoot: projectRoot,
		clean:       false,
		outputDir:   "dist",
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

// Execute runs the build command.
func (c *BuildCommand) Execute(ctx context.Context) error {
	// Create template engine and add search path
	templateEngine := ason.NewTemplateEngine()

	// Find template directory relative to binary location
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		templateDir := filepath.Join(exeDir, "..", "templates", "standard-3layer")
		templateEngine.AddSearchPath(templateDir)

		// Also try relative to current directory
		templateEngine.AddSearchPath(filepath.Join(".", "templates", "standard-3layer"))

		// Try absolute path for development
		templateEngine.AddSearchPath("/Users/andhi/code/mdstn/loko/templates/standard-3layer")
	}

	// Load the project
	projectRepo := filesystem.NewProjectRepository()
	projectRepo.SetTemplateEngine(templateEngine)

	project, err := projectRepo.LoadProject(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// List all systems
	systems, err := projectRepo.ListSystems(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to list systems: %w", err)
	}

	if len(systems) == 0 {
		fmt.Println("No systems found to build")
		return nil
	}

	// Create adapters
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return fmt.Errorf("failed to create site builder: %w", err)
	}
	markdownRenderer := html.NewMarkdownRenderer("", "")

	// Create progress reporter (simple console output)
	progressReporter := &simpleProgressReporter{}

	// Create and execute build use case
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, progressReporter)

	startTime := time.Now()
	err = buildDocs.Execute(ctx, project, systems, c.outputDir)
	elapsed := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Render markdown files to HTML (after diagrams are built)
	fmt.Println("\n📄 Rendering markdown documentation...")
	renderMarkdownDocs := usecases.NewRenderMarkdownDocs(markdownRenderer, progressReporter)
	err = renderMarkdownDocs.Execute(ctx, project, systems, c.outputDir)
	if err != nil {
		return fmt.Errorf("markdown rendering failed: %w", err)
	}

	fmt.Printf("✓ Build completed in %v\n", elapsed.Round(10*time.Millisecond))
	fmt.Printf("✓ Output: %s\n", c.outputDir)

	return nil
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
