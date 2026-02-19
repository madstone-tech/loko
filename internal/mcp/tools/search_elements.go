package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// SearchElementsTool searches for architecture elements by pattern and filters.
type SearchElementsTool struct {
	useCase *usecases.SearchElements
}

// NewSearchElementsTool creates a new search_elements tool.
func NewSearchElementsTool(repo usecases.ProjectRepository) *SearchElementsTool {
	return &SearchElementsTool{
		useCase: usecases.NewSearchElements(repo),
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
		},
		"required": []string{"project_root", "query"},
	}
}

func (t *SearchElementsTool) Call(ctx context.Context, arguments map[string]any) (any, error) {
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
	return t.useCase.Execute(ctx, req)
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
