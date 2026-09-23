package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// FindRelationshipsTool finds relationships between architecture elements.
type FindRelationshipsTool struct {
	useCase *usecases.FindRelationships
}

// NewFindRelationshipsTool creates a new find_relationships tool.
func NewFindRelationshipsTool(repo usecases.ProjectRepository) *FindRelationshipsTool {
	return &FindRelationshipsTool{
		useCase: usecases.NewFindRelationships(repo),
	}
}

func (t *FindRelationshipsTool) Name() string {
	return "find_relationships"
}

func (t *FindRelationshipsTool) Description() string {
	return "Search the architecture graph for relationships derived from component .md frontmatter (legacy source). Supports glob patterns. Use 'list_relationships' instead for relationships created via create_relationship (stored in relationships.toml)."
}

func (t *FindRelationshipsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root":      map[string]any{"type": "string", "description": "Project root directory"},
			"source_pattern":    map[string]any{"type": "string", "description": "Source element pattern (supports glob: *, ?)"},
			"target_pattern":    map[string]any{"type": "string", "description": "Target element pattern (supports glob: *, ?)"},
			"relationship_type": map[string]any{"type": "string", "description": "Filter by relationship type (e.g., depends-on, uses)"},
			"limit":             map[string]any{"type": "number", "description": "Max results (default: 20, max: 100)"},
		},
		"required": []string{"project_root"},
	}
}

func (t *FindRelationshipsTool) Call(ctx context.Context, arguments map[string]any) (any, error) {
	// Parse arguments to request
	req := entities.FindRelationshipsRequest{
		ProjectRoot:      getString(arguments, "project_root"),
		SourcePattern:    getString(arguments, "source_pattern"),
		TargetPattern:    getString(arguments, "target_pattern"),
		RelationshipType: getString(arguments, "relationship_type"),
		Limit:            getInt(arguments, "limit"),
	}

	// Call use case
	return t.useCase.Execute(ctx, req)
}
