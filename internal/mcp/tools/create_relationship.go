package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateRelationshipTool creates a new C4 model relationship between two elements.
type CreateRelationshipTool struct {
	repo        usecases.RelationshipRepository
	projectRepo usecases.ProjectRepository
	graphCache  GraphCache
}

// NewCreateRelationshipTool creates a new create_relationship tool.
func NewCreateRelationshipTool(repo usecases.RelationshipRepository, projectRepo usecases.ProjectRepository, cache GraphCache) *CreateRelationshipTool {
	return &CreateRelationshipTool{repo: repo, projectRepo: projectRepo, graphCache: cache}
}

func (t *CreateRelationshipTool) Name() string { return "create_relationship" }
func (t *CreateRelationshipTool) Description() string {
	return "Create a directed relationship between two C4 elements (containers or components). Persists to relationships.toml and updates the D2 diagram."
}

func (t *CreateRelationshipTool) InputSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"project_root", "system_name", "source", "target", "label"},
		"properties": map[string]any{
			"project_root": map[string]any{"type": "string", "description": "Root directory of the loko project (default: '.')"},
			"system_name":  map[string]any{"type": "string", "description": "Name of the system owning this relationship (slugified or display name)"},
			"source":       map[string]any{"type": "string", "description": "Source element path, e.g. 'agwe/api-lambda'"},
			"target":       map[string]any{"type": "string", "description": "Target element path, e.g. 'agwe/sqs-queue'"},
			"label":        map[string]any{"type": "string", "description": "Human-readable description of the relationship"},
			"type": map[string]any{
				"type": "string", "enum": []string{"sync", "async", "event"},
				"description": "Communication type (default: 'sync')",
			},
			"technology": map[string]any{"type": "string", "description": "Technology used (e.g., 'AWS SDK SQS', 'gRPC')"},
			"direction": map[string]any{
				"type": "string", "enum": []string{"forward", "bidirectional"},
				"description": "Arrow direction (default: 'forward')",
			},
		},
	}
}

// Call executes the create_relationship tool.
func (t *CreateRelationshipTool) Call(ctx context.Context, args map[string]any) (any, error) {
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

	source := getString(args, "source")
	if source == "" {
		return nil, fmt.Errorf("source is required")
	}
	if _, err := validateElementPath(source); err != nil {
		return nil, err
	}

	target := getString(args, "target")
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}
	if _, err := validateElementPath(target); err != nil {
		return nil, err
	}

	label := getString(args, "label")
	if label == "" {
		return nil, fmt.Errorf("label is required")
	}

	uc := usecases.NewCreateRelationship(t.repo)
	rel, err := uc.Execute(ctx, &usecases.CreateRelationshipRequest{
		ProjectRoot: projectRoot,
		SystemID:    systemID,
		Source:      source,
		Target:      target,
		Label:       label,
		Type:        getString(args, "type"),
		Technology:  getString(args, "technology"),
		Direction:   getString(args, "direction"),
	})
	if err != nil {
		return nil, err
	}

	// Invalidate graph cache so the next query reflects the new relationship.
	if t.graphCache != nil {
		t.graphCache.Invalidate(projectRoot)
	}

	d2Path := usecases.D2DiagramPath(projectRoot, systemID, rel)

	return map[string]any{
		"relationship":    relationshipToMap(rel),
		"diagram_updated": true,
		"diagram_path":    d2Path,
	}, nil
}

// relationshipToMap converts a Relationship entity to a JSON-friendly map.
func relationshipToMap(rel *entities.Relationship) map[string]any {
	m := map[string]any{
		"id":     rel.ID,
		"source": rel.Source,
		"target": rel.Target,
		"label":  rel.Label,
	}
	if rel.Type != "" {
		m["type"] = rel.Type
	}
	if rel.Technology != "" {
		m["technology"] = rel.Technology
	}
	if rel.Direction != "" {
		m["direction"] = rel.Direction
	}
	return m
}
