package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ListRelationshipsTool lists relationships for a system with optional filtering.
type ListRelationshipsTool struct {
	repo        usecases.RelationshipRepository
	projectRepo usecases.ProjectRepository
}

// NewListRelationshipsTool creates a new list_relationships tool.
func NewListRelationshipsTool(repo usecases.RelationshipRepository, projectRepo usecases.ProjectRepository) *ListRelationshipsTool {
	return &ListRelationshipsTool{repo: repo, projectRepo: projectRepo}
}

// Name returns the tool name.
func (t *ListRelationshipsTool) Name() string { return "list_relationships" }

// Description returns the tool description.
func (t *ListRelationshipsTool) Description() string {
	return "List C4 model relationships for a system from relationships.toml (the authoritative source). Use this — not find_relationships — to query relationships created via create_relationship. Optionally filter by source or target element path."
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *ListRelationshipsTool) InputSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"project_root", "system_name"},
		"properties": map[string]any{
			"project_root": map[string]any{"type": "string", "description": "Root directory of the loko project"},
			"system_name":  map[string]any{"type": "string", "description": "System to list relationships for"},
			"source":       map[string]any{"type": "string", "description": "Optional: filter by source element path"},
			"target":       map[string]any{"type": "string", "description": "Optional: filter by target element path"},
		},
	}
}

// Call executes the list_relationships tool.
func (t *ListRelationshipsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot := getString(args, "project_root")
	if projectRoot == "" {
		projectRoot = "."
	}
	systemID, err := t.resolveSystem(ctx, projectRoot, getString(args, "system_name"))
	if err != nil {
		return nil, err
	}
	uc := usecases.NewListRelationships(t.repo)
	rels, err := uc.Execute(ctx, &usecases.ListRelationshipsRequest{
		ProjectRoot: projectRoot, SystemID: systemID,
		FilterSource: getString(args, "source"), FilterTarget: getString(args, "target"),
	})
	if err != nil {
		return nil, err
	}
	relMaps := make([]map[string]any, 0, len(rels))
	for _, r := range rels {
		r := r
		relMaps = append(relMaps, relationshipToMap(&r))
	}
	return map[string]any{"system": systemID, "count": len(relMaps), "relationships": relMaps}, nil
}

func (t *ListRelationshipsTool) resolveSystem(ctx context.Context, projectRoot, systemName string) (string, error) {
	if systemName == "" {
		return "", fmt.Errorf("system_name is required")
	}
	systemID := entities.NormalizeName(systemName)
	if t.projectRepo != nil {
		if _, err := t.projectRepo.LoadSystem(ctx, projectRoot, systemID); err != nil {
			graph, _ := getGraphFromProject(ctx, t.projectRepo, projectRoot)
			return "", notFoundError("system", systemName, suggestSlugID(systemName, graph))
		}
	}
	return systemID, nil
}
