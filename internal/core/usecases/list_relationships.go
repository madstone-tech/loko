package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// ListRelationshipsRequest defines the input for the ListRelationships use case.
type ListRelationshipsRequest struct {
	// ProjectRoot is the filesystem root of the loko project.
	ProjectRoot string

	// SystemID is the slugified system name to list relationships for.
	SystemID string

	// FilterSource, if non-empty, restricts results to relationships where
	// Source exactly matches this value.
	FilterSource string

	// FilterTarget, if non-empty, restricts results to relationships where
	// Target exactly matches this value.
	FilterTarget string
}

// ListRelationships loads and optionally filters the relationships for a system.
type ListRelationships struct {
	repo RelationshipRepository
}

// NewListRelationships creates a new ListRelationships use case.
func NewListRelationships(repo RelationshipRepository) *ListRelationships {
	return &ListRelationships{repo: repo}
}

// Execute returns all relationships for the given system, applying optional filters.
// Returns an empty slice (not an error) if no relationships exist.
func (uc *ListRelationships) Execute(
	ctx context.Context, req *ListRelationshipsRequest,
) ([]entities.Relationship, error) {
	if req.ProjectRoot == "" {
		return nil, fmt.Errorf("project_root is required")
	}
	if req.SystemID == "" {
		return nil, fmt.Errorf("system_id is required")
	}

	rels, err := uc.repo.LoadRelationships(ctx, req.ProjectRoot, req.SystemID)
	if err != nil {
		return nil, fmt.Errorf("loading relationships: %w", err)
	}

	// Apply filters.
	if req.FilterSource == "" && req.FilterTarget == "" {
		return rels, nil
	}

	filtered := make([]entities.Relationship, 0, len(rels))
	for _, r := range rels {
		if req.FilterSource != "" && r.Source != req.FilterSource {
			continue
		}
		if req.FilterTarget != "" && r.Target != req.FilterTarget {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered, nil
}
