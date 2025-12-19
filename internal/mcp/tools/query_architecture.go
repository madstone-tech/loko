package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryArchitectureTool provides token-efficient architecture queries.
type QueryArchitectureTool struct {
	repo usecases.ProjectRepository
}

// NewQueryArchitectureTool creates a new query_architecture tool.
func NewQueryArchitectureTool(repo usecases.ProjectRepository) *QueryArchitectureTool {
	return &QueryArchitectureTool{repo: repo}
}

// Name returns the tool name.
func (t *QueryArchitectureTool) Name() string {
	return "query_architecture"
}

// Description returns the tool description.
func (t *QueryArchitectureTool) Description() string {
	return "Query architecture with configurable detail levels (summary ~200 tokens, structure ~500 tokens, full for complete details)"
}

// InputSchema returns the JSON schema for tool inputs.
func (t *QueryArchitectureTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"detail": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"summary", "structure", "full"},
				"description": "Detail level: summary (~200 tokens), structure (~500 tokens), or full",
			},
			"target_system": map[string]interface{}{
				"type":        "string",
				"description": "Optional: focus on a specific system",
			},
		},
		"required": []string{"project_root", "detail"},
	}
}

// Call executes the tool.
func (t *QueryArchitectureTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	detail, _ := args["detail"].(string)
	targetSystem, _ := args["target_system"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	if detail == "" {
		detail = "structure"
	}

	// Use QueryArchitecture use case
	uc := usecases.NewQueryArchitecture(t.repo)
	resp, err := uc.Execute(ctx, projectRoot, detail)
	if err != nil {
		return nil, fmt.Errorf("failed to query architecture: %w", err)
	}

	return map[string]interface{}{
		"text":           resp.Text,
		"detail":         resp.Detail,
		"token_estimate": resp.TokenEstimate,
		"system_count":   len(resp.Systems),
		"_target_system": targetSystem, // For future targeted query filtering
	}, nil
}
