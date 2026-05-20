package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/core/entities"
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

// Name returns the tool name.
func (t *BuildDocsTool) Name() string { return "build_docs" }

// Description returns the tool description.
func (t *BuildDocsTool) Description() string {
	return "Build HTML documentation for the project"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *BuildDocsTool) InputSchema() map[string]any { return Schemas["build_docs"].(map[string]any) }

// Call executes the build docs tool by delegating to the BuildDocsUseCase.
func (t *BuildDocsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	outputDir, _ := args["output_dir"].(string)
	if outputDir == "" {
		return nil, fmt.Errorf("output_dir is required")
	}
	project, systems, err := t.loadProjectData(ctx, projectRoot)
	if err != nil {
		return nil, err
	}
	if err := t.runBuild(ctx, project, systems, outputDir); err != nil {
		return nil, err
	}
	return map[string]any{
		"success": true,
		"message": fmt.Sprintf("Documentation built successfully in %s", outputDir),
		"output":  outputDir, "systems": len(systems),
		"files": map[string]any{
			"index": "index.html", "systems": len(systems), "diagrams": countDiagrams(systems),
		},
	}, nil
}

func (t *BuildDocsTool) loadProjectData(ctx context.Context, projectRoot string) (*entities.Project, []*entities.System, error) {
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load project: %w", err)
	}
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list systems: %w", err)
	}
	return project, systems, nil
}

func (t *BuildDocsTool) runBuild(ctx context.Context, project *entities.Project, systems []*entities.System, outputDir string) error {
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return fmt.Errorf("failed to create site builder: %w", err)
	}
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, &mcpProgressReporter{})
	if err := buildDocs.Execute(ctx, project, systems, outputDir); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}

// mcpProgressReporter implements ProgressReporter for MCP tool context.
type mcpProgressReporter struct{}

// ReportProgress reports progress (silent in MCP context).
func (r *mcpProgressReporter) ReportProgress(step string, current int, total int, message string) {}

// ReportError reports an error (silent in MCP context; errors returned directly).
func (r *mcpProgressReporter) ReportError(err error) {}

// ReportSuccess reports success (silent in MCP context; success implicit in return value).
func (r *mcpProgressReporter) ReportSuccess(message string) {}

// ReportInfo reports info (silent in MCP context; info implicit in return value).
func (r *mcpProgressReporter) ReportInfo(message string) {}
