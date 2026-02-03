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
	return `Query architecture with configurable detail levels and output formats.

Detail levels:
- summary: ~200 tokens - project overview with system counts
- structure: ~500 tokens - systems and their containers
- full: complete details - all systems, containers, components

Output formats:
- text: human-readable markdown (default)
- json: structured JSON
- toon: Token-Optimized Object Notation (30-40% fewer tokens than JSON)`
}

// InputSchema returns the JSON schema for tool inputs.
func (t *QueryArchitectureTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"detail": map[string]any{
				"type":        "string",
				"enum":        []string{"summary", "structure", "full"},
				"description": "Detail level: summary (~200 tokens), structure (~500 tokens), or full",
			},
			"format": map[string]any{
				"type":        "string",
				"enum":        []string{"text", "json", "toon"},
				"description": "Output format: text (markdown), json (structured), or toon (Token-Optimized, 30-40% fewer tokens)",
				"default":     "text",
			},
			"target_system": map[string]any{
				"type":        "string",
				"description": "Optional: focus on a specific system",
			},
		},
		"required": []string{"project_root", "detail"},
	}
}

// Call executes the tool.
func (t *QueryArchitectureTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	detail, _ := args["detail"].(string)
	format, _ := args["format"].(string)
	targetSystem, _ := args["target_system"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	if detail == "" {
		detail = "structure"
	}

	if format == "" {
		format = "text"
	}

	// Use QueryArchitecture use case with format
	uc := usecases.NewQueryArchitecture(t.repo)
	resp, err := uc.ExecuteWithFormat(ctx, projectRoot, detail, format)
	if err != nil {
		return nil, fmt.Errorf("failed to query architecture: %w", err)
	}

	return map[string]any{
		"text":           resp.Text,
		"detail":         resp.Detail,
		"format":         resp.Format,
		"token_estimate": resp.TokenEstimate,
		"system_count":   len(resp.Systems),
		"_target_system": targetSystem, // For future targeted query filtering
	}, nil
}
