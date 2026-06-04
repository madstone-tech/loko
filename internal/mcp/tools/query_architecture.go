package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryArchitectureTool provides token-efficient architecture queries.
type QueryArchitectureTool struct {
	repo    usecases.ProjectRepository
	encoder usecases.OutputEncoder
}

// NewQueryArchitectureTool creates a new query_architecture tool.
func NewQueryArchitectureTool(repo usecases.ProjectRepository, encoder usecases.OutputEncoder) *QueryArchitectureTool {
	return &QueryArchitectureTool{repo: repo, encoder: encoder}
}

// Name returns the tool name.
func (t *QueryArchitectureTool) Name() string {
	return "query_architecture"
}

// Description returns the tool description.
func (t *QueryArchitectureTool) Description() string {
	return `Query architecture with configurable detail levels and output formats.

Supported detail levels:
- "summary": High-level overview (~200 tokens)
- "structure": System/container/component hierarchy (~500 tokens)  
- "full": Complete architecture with all metadata and relationships

Supported formats:
- "toon": Token-Optimized Object Notation (default, 30-40% fewer tokens)
- "json": Standard JSON for debugging or interoperability

Note: The custom 'compact' format from v0.1.0 is deprecated. Use 'toon' for token-efficient output.`
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
				"enum":        []string{"toon", "json"},
				"default":     "toon",
				"description": "Output format: 'toon' for token-efficient LLM output (default), 'json' for human-readable debugging",
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
		format = "toon"
	}
	if format == "text" || format == "compact" {
		format = "toon"
	}
	if format != "toon" && format != "json" {
		return nil, fmt.Errorf("invalid format \"%s\": expected \"toon\" or \"json\"", format)
	}

	uc := usecases.NewQueryArchitecture(t.repo)
	resp, err := uc.ExecuteWithFormat(ctx, projectRoot, detail, format)
	if err != nil {
		return nil, fmt.Errorf("failed to query architecture: %w", err)
	}
	return t.buildResponse(resp, targetSystem, format), nil
}

func (t *QueryArchitectureTool) buildResponse(resp *usecases.QueryArchitectureResponse, targetSystem, format string) map[string]any {
	result := map[string]any{
		"detail":         resp.Detail,
		"format":         resp.Format,
		"token_estimate": resp.TokenEstimate,
		"system_count":   len(resp.Systems),
		"_target_system": targetSystem,
	}
	if format == "json" {
		result["text"] = resp.Text
	} else {
		result["payload"] = resp.Text
	}
	return result
}
