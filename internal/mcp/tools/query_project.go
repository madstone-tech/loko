package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryProjectTool returns metadata about the current project.
type QueryProjectTool struct {
	repo usecases.ProjectRepository
}

// NewQueryProjectTool creates a new query_project tool.
func NewQueryProjectTool(repo usecases.ProjectRepository) *QueryProjectTool {
	return &QueryProjectTool{repo: repo}
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
func (t *QueryProjectTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project (defaults to current)",
			},
		},
	}
}

// Call executes the tool.
func (t *QueryProjectTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}

	// Load project
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// Load systems
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}

	return map[string]interface{}{
		"project": map[string]interface{}{
			"name":        project.Name,
			"description": project.Description,
			"version":     project.Version,
		},
		"stats": map[string]interface{}{
			"systems":    len(systems),
			"containers": project.ContainerCount(),
			"components": project.ComponentCount(),
		},
	}, nil
}
