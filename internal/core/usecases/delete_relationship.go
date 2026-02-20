package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// ErrRelationshipNotFound is returned when a relationship ID is not found.
var ErrRelationshipNotFound = fmt.Errorf("relationship not found")

// DeleteRelationshipRequest defines the input for the DeleteRelationship use case.
type DeleteRelationshipRequest struct {
	// ProjectRoot is the filesystem root of the loko project.
	ProjectRoot string

	// SystemID is the slugified system name owning the relationship.
	SystemID string

	// RelationshipID is the 8-hex-char deterministic ID to delete.
	RelationshipID string
}

// DeleteRelationship removes a relationship by ID and updates the D2 diagram.
type DeleteRelationship struct {
	repo RelationshipRepository
}

// NewDeleteRelationship creates a new DeleteRelationship use case.
func NewDeleteRelationship(repo RelationshipRepository) *DeleteRelationship {
	return &DeleteRelationship{repo: repo}
}

// Execute removes the relationship with the given ID, saves the updated list,
// and regenerates the D2 edges section. Returns ErrRelationshipNotFound if the
// ID does not exist.
func (uc *DeleteRelationship) Execute(
	ctx context.Context, req *DeleteRelationshipRequest,
) error {
	if req.RelationshipID == "" {
		return fmt.Errorf("relationship_id is required")
	}

	rels, err := uc.repo.LoadRelationships(ctx, req.ProjectRoot, req.SystemID)
	if err != nil {
		return fmt.Errorf("loading relationships: %w", err)
	}

	// Find the relationship to delete and build the remaining list.
	var deleted *entities.Relationship
	remaining := make([]entities.Relationship, 0, len(rels))
	for _, r := range rels {
		r := r
		if r.ID == req.RelationshipID {
			deleted = &r
			continue
		}
		remaining = append(remaining, r)
	}

	if deleted == nil {
		return fmt.Errorf("%w: %q", ErrRelationshipNotFound, req.RelationshipID)
	}

	if err := uc.repo.SaveRelationships(ctx, req.ProjectRoot, req.SystemID, remaining); err != nil {
		return fmt.Errorf("saving relationships after delete: %w", err)
	}

	// Update D2 diagram (best-effort).
	if err := updateD2File(req.ProjectRoot, req.SystemID, deleted, remaining); err != nil {
		_ = err // diagram update failure is non-fatal
	}

	return nil
}
