package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// DeleteRelationshipTool removes a relationship by ID.
type DeleteRelationshipTool struct {
	repo        usecases.RelationshipRepository
	projectRepo usecases.ProjectRepository
	graphCache  GraphCache
}

// NewDeleteRelationshipTool creates a new delete_relationship tool.
func NewDeleteRelationshipTool(repo usecases.RelationshipRepository, projectRepo usecases.ProjectRepository, cache GraphCache) *DeleteRelationshipTool {
	return &DeleteRelationshipTool{repo: repo, projectRepo: projectRepo, graphCache: cache}
}

func (t *DeleteRelationshipTool) Name() string { return "delete_relationship" }
func (t *DeleteRelationshipTool) Description() string {
	return "Delete a C4 model relationship by ID. Updates the D2 diagram and invalidates the graph cache."
}

func (t *DeleteRelationshipTool) InputSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"project_root", "system_name", "relationship_id"},
		"properties": map[string]any{
			"project_root":    map[string]any{"type": "string", "description": "Root directory of the loko project"},
			"system_name":     map[string]any{"type": "string", "description": "System owning this relationship"},
			"relationship_id": map[string]any{"type": "string", "description": "ID of the relationship to delete"},
		},
	}
}

// Call executes the delete_relationship tool.
func (t *DeleteRelationshipTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot := getString(args, "project_root")
	if projectRoot == "" {
		projectRoot = "."
	}

	systemName := getString(args, "system_name")
	if systemName == "" {
		return nil, fmt.Errorf("system_name is required")
	}
	systemID := entities.NormalizeName(systemName)

	// Validate system exists — provide slug suggestion on mismatch.
	if t.projectRepo != nil {
		if _, err := t.projectRepo.LoadSystem(ctx, projectRoot, systemID); err != nil {
			graph, _ := getGraphFromProject(ctx, t.projectRepo, projectRoot)
			return nil, notFoundError("system", systemName, suggestSlugID(systemName, graph))
		}
	}

	relID := getString(args, "relationship_id")
	if relID == "" {
		return nil, fmt.Errorf("relationship_id is required")
	}

	uc := usecases.NewDeleteRelationship(t.repo)
	if err := uc.Execute(ctx, &usecases.DeleteRelationshipRequest{
		ProjectRoot:    projectRoot,
		SystemID:       systemID,
		RelationshipID: relID,
	}); err != nil {
		return nil, err
	}

	// Invalidate graph cache so the next query reflects the deletion.
	if t.graphCache != nil {
		t.graphCache.Invalidate(projectRoot)
	}

	// diagram_path: best-effort — use system.d2 as the default.
	d2Path := fmt.Sprintf("src/%s/system.d2", systemID)

	return map[string]any{
		"deleted":         true,
		"relationship_id": relID,
		"diagram_updated": true,
		"diagram_path":    d2Path,
	}, nil
}
