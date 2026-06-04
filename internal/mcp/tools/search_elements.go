package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// SearchElementsTool searches for architecture elements by pattern and filters.
type SearchElementsTool struct {
	useCase *usecases.SearchElements
	encoder usecases.OutputEncoder
}

// NewSearchElementsTool creates a new search_elements tool.
func NewSearchElementsTool(repo usecases.ProjectRepository, encoder usecases.OutputEncoder) *SearchElementsTool {
	return &SearchElementsTool{
		useCase: usecases.NewSearchElements(repo),
		encoder: encoder,
	}
}

func (t *SearchElementsTool) Name() string {
	return "search_elements"
}

func (t *SearchElementsTool) Description() string {
	return "Search architecture elements by name pattern, type, technology, or tags"
}

func (t *SearchElementsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{"type": "string", "description": "Project root directory"},
			"query":        map[string]any{"type": "string", "description": "Search pattern (supports glob: *, ?)"},
			"type":         map[string]any{"type": "string", "description": "Filter by type: system, container, component"},
			"technology":   map[string]any{"type": "string", "description": "Filter by technology (e.g., Go, Python)"},
			"tag":          map[string]any{"type": "string", "description": "Filter by tag (e.g., critical, production)"},
			"limit":        map[string]any{"type": "number", "description": "Max results (default: 20, max: 100)"},
			"format": map[string]any{
				"type":        "string",
				"enum":        []string{"toon", "json"},
				"default":     "toon",
				"description": "Output format: 'toon' for token-efficient LLM output (default), 'json' for human-readable debugging",
			},
		},
		"required": []string{"project_root", "query"},
	}
}

func (t *SearchElementsTool) Call(ctx context.Context, arguments map[string]any) (any, error) {
	format, err := getFormat(arguments)
	if err != nil {
		return nil, err
	}

	// Parse arguments to request
	req := entities.SearchElementsRequest{
		ProjectRoot: getString(arguments, "project_root"),
		Query:       getString(arguments, "query"),
		Type:        getString(arguments, "type"),
		Technology:  getString(arguments, "technology"),
		Tag:         getString(arguments, "tag"),
		Limit:       getInt(arguments, "limit"),
	}

	// Call use case
	resp, err := t.useCase.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// For JSON format, return the raw response for backward compatibility
	if format == "json" {
		return resp, nil
	}

	// For TOON format, wrap in a map
	result := map[string]any{
		"query":   req.Query,
		"results": resp.Results,
		"count":   len(resp.Results),
		"total":   resp.TotalMatched,
		"message": resp.Message,
	}

	return formatResponse(result, format, t.encoder)
}

// Helper functions for argument extraction
func getString(args map[string]any, key string) string {
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(args map[string]any, key string) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}
