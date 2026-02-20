package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// CreateRelationshipRequest defines the input for the CreateRelationship use case.
type CreateRelationshipRequest struct {
	// ProjectRoot is the filesystem root of the loko project.
	ProjectRoot string

	// SystemID is the slugified system name owning this relationship.
	SystemID string

	// Source is the slash-separated element path (e.g., "agwe/api-lambda").
	Source string

	// Target is the slash-separated element path (e.g., "agwe/sqs-queue").
	Target string

	// Label is a human-readable description of the relationship.
	Label string

	// Type is the communication pattern: "sync", "async", or "event" (optional).
	Type string

	// Technology is the free-text technology description (optional).
	Technology string

	// Direction is "forward" or "bidirectional" (optional).
	Direction string
}

// CreateRelationship creates a new C4 model relationship between two elements,
// persists it to relationships.toml, and updates the D2 diagram file.
//
// The operation is idempotent: if a relationship with the same source+target+label
// already exists (determined by ID hash), it is returned without error.
type CreateRelationship struct {
	repo RelationshipRepository
}

// NewCreateRelationship creates a new CreateRelationship use case.
func NewCreateRelationship(repo RelationshipRepository) *CreateRelationship {
	return &CreateRelationship{repo: repo}
}

// Execute creates and persists a relationship, then updates the D2 diagram.
// Returns the created (or existing duplicate) Relationship.
func (uc *CreateRelationship) Execute(
	ctx context.Context, req *CreateRelationshipRequest,
) (*entities.Relationship, error) {
	// 1. Validate and construct the entity.
	var opts []entities.RelationshipOption
	if req.Type != "" {
		opts = append(opts, entities.WithRelType(req.Type))
	}
	if req.Technology != "" {
		opts = append(opts, entities.WithRelTechnology(req.Technology))
	}
	if req.Direction != "" {
		opts = append(opts, entities.WithRelDirection(req.Direction))
	}

	rel, err := entities.NewRelationship(req.Source, req.Target, req.Label, opts...)
	if err != nil {
		return nil, err
	}

	// 2. Load existing relationships for the system.
	existing, err := uc.repo.LoadRelationships(ctx, req.ProjectRoot, req.SystemID)
	if err != nil {
		return nil, fmt.Errorf("loading relationships: %w", err)
	}

	// 3. Idempotency check: if same ID already exists, return it.
	for _, e := range existing {
		if e.ID == rel.ID {
			return &e, nil
		}
	}

	// 4. Append and save.
	updated := append(existing, *rel)
	if err := uc.repo.SaveRelationships(ctx, req.ProjectRoot, req.SystemID, updated); err != nil {
		return nil, fmt.Errorf("saving relationships: %w", err)
	}

	// 5. Update D2 diagram (best-effort — non-fatal if diagram doesn't exist yet).
	if err := updateD2File(req.ProjectRoot, req.SystemID, rel, updated); err != nil {
		// Log but don't fail — diagram can be updated manually later.
		_ = err
	}

	return rel, nil
}

// D2DiagramPath returns the path of the target D2 diagram file for a relationship.
// The rule (from R-002):
//   - Container→Container or cross-container component: system.d2
//   - Same-container component→component: container.d2
func D2DiagramPath(projectRoot, systemID string, rel *entities.Relationship) string {
	srcParts := strings.Split(rel.Source, "/")
	tgtParts := strings.Split(rel.Target, "/")

	// Same container (3-segment paths, same system + container)
	if len(srcParts) == 3 && len(tgtParts) == 3 &&
		srcParts[0] == tgtParts[0] && srcParts[1] == tgtParts[1] {
		return filepath.Join(projectRoot, "src", systemID, srcParts[1], "container.d2")
	}

	// Everything else → system.d2
	return filepath.Join(projectRoot, "src", systemID, "system.d2")
}

// updateD2File regenerates the edges section of the target D2 file from the
// full list of relationships. It is safe to call when the file doesn't exist
// (it will be created). If the file does not have a relationships section,
// edges are appended at the end.
func updateD2File(projectRoot, systemID string, newRel *entities.Relationship, allRels []entities.Relationship) error {
	d2Path := D2DiagramPath(projectRoot, systemID, newRel)

	// Generate all D2 edges for this system.
	var edgeLines strings.Builder
	edgeLines.WriteString("# relationships\n")
	for _, r := range allRels {
		edgeLines.WriteString(entities.RelationshipToD2Edge(r))
	}
	edgesContent := edgeLines.String()

	// Read existing file (if any).
	existing, err := os.ReadFile(d2Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading D2 file: %w", err)
		}
		// File doesn't exist — create parent dir and write fresh.
		if mkErr := os.MkdirAll(filepath.Dir(d2Path), 0o755); mkErr != nil {
			return fmt.Errorf("creating D2 directory: %w", mkErr)
		}
		return os.WriteFile(d2Path, []byte(edgesContent), 0o644)
	}

	// Remove existing "# relationships" section if present, then append fresh edges.
	content := string(existing)
	if idx := strings.Index(content, "# relationships\n"); idx >= 0 {
		content = strings.TrimRight(content[:idx], "\n") + "\n"
	}

	return os.WriteFile(d2Path, []byte(content+edgesContent), 0o644)
}
