package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryProjectTool returns metadata about the current project.
type QueryProjectTool struct {
	repo    usecases.ProjectRepository
	encoder usecases.OutputEncoder
}

// NewQueryProjectTool creates a new query_project tool.
func NewQueryProjectTool(repo usecases.ProjectRepository, encoder usecases.OutputEncoder) *QueryProjectTool {
	return &QueryProjectTool{repo: repo, encoder: encoder}
}

// Name returns the tool name.
func (t *QueryProjectTool) Name() string {
	return "query_project"
}

// Description returns the tool description.
func (t *QueryProjectTool) Description() string {
	return "Query current project metadata, systems, containers, and overall architecture summary"
}

// InputSchema returns the JSON schema for tool inputs.
func (t *QueryProjectTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project (defaults to current)",
			},
			"format": map[string]any{
				"type":        "string",
				"enum":        []string{"toon", "json"},
				"default":     "toon",
				"description": "Output format: 'toon' for token-efficient LLM output (default), 'json' for human-readable debugging",
			},
		},
	}
}

// Call executes the tool.
func (t *QueryProjectTool) Call(ctx context.Context, args map[string]any) (any, error) {
	format, err := getFormat(args)
	if err != nil {
		return nil, err
	}
	result, err := t.loadAndBuild(ctx, args)
	if err != nil {
		return nil, err
	}
	return formatResponse(result, format, t.encoder)
}

func (t *QueryProjectTool) loadAndBuild(ctx context.Context, args map[string]any) (map[string]any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}

	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}
	return map[string]any{
		"project": map[string]any{
			"name":        project.Name,
			"description": project.Description,
			"version":     project.Version,
		},
		"stats": map[string]any{
			"systems":    len(systems),
			"containers": project.ContainerCount(),
			"components": project.ComponentCount(),
		},
	}, nil
}
