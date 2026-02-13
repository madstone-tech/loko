package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// BuildDocsTool triggers documentation build.
type BuildDocsTool struct {
	repo usecases.ProjectRepository
}

// NewBuildDocsTool creates a new build_docs tool.
func NewBuildDocsTool(repo usecases.ProjectRepository) *BuildDocsTool {
	return &BuildDocsTool{repo: repo}
}

func (t *BuildDocsTool) Name() string {
	return "build_docs"
}

func (t *BuildDocsTool) Description() string {
	return "Build HTML documentation for the project"
}

func (t *BuildDocsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"output_dir": map[string]any{
				"type":        "string",
				"description": "Output directory for HTML files",
			},
		},
		"required": []string{"project_root", "output_dir"},
	}
}

// Call executes the build docs tool by delegating to the BuildDocsUseCase.
func (t *BuildDocsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// 1. Parse and validate inputs
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}

	outputDir, _ := args["output_dir"].(string)
	if outputDir == "" {
		return nil, fmt.Errorf("output_dir is required")
	}

	// 2. Load project and systems
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}

	// 3. Call BuildDocsUseCase
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return nil, fmt.Errorf("failed to create site builder: %w", err)
	}

	// Create progress reporter (simple in-memory reporter)
	progressReporter := &mcpProgressReporter{}

	// Create and execute build use case
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, progressReporter)

	err = buildDocs.Execute(ctx, project, systems, outputDir)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// 4. Format response
	return map[string]any{
		"success": true,
		"message": fmt.Sprintf("Documentation built successfully in %s", outputDir),
		"output":  outputDir,
		"systems": len(systems),
		"files": map[string]any{
			"index":    "index.html",
			"systems":  len(systems),
			"diagrams": countDiagrams(systems),
		},
	}, nil
}

// mcpProgressReporter implements ProgressReporter for MCP tool context.
type mcpProgressReporter struct {
}

// ReportProgress reports progress.
func (r *mcpProgressReporter) ReportProgress(step string, current int, total int, message string) {
	// Silent in MCP context; progress is implicit in tool execution
}

// ReportError reports an error.
func (r *mcpProgressReporter) ReportError(err error) {
	// Silent in MCP context; errors are returned directly
}

// ReportSuccess reports success.
func (r *mcpProgressReporter) ReportSuccess(message string) {
	// Silent in MCP context; success is implicit in return value
}

// ReportInfo reports info.
func (r *mcpProgressReporter) ReportInfo(message string) {
	// Silent in MCP context; info is implicit in return value
}
