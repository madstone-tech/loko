package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// FilesystemRelationshipRepository implements the RelationshipRepository port
// using the local file system.
//
// Relationships for a system are stored in:
//
//	<projectRoot>/src/<systemID>/relationships.toml
//
// Writes are atomic: content is first written to a .tmp file, then
// renamed to the target path (POSIX rename is atomic on the same filesystem).
type FilesystemRelationshipRepository struct{}

// NewFilesystemRelationshipRepository creates a new FilesystemRelationshipRepository.
func NewFilesystemRelationshipRepository() *FilesystemRelationshipRepository {
	return &FilesystemRelationshipRepository{}
}

// relationshipsPath returns the canonical path for a system's relationships.toml.
func relationshipsPath(projectRoot, systemID string) string {
	return filepath.Join(projectRoot, "src", systemID, "relationships.toml")
}

// LoadRelationships reads all relationships for a system from relationships.toml.
// Returns an empty slice (not an error) if the file does not exist yet.
func (r *FilesystemRelationshipRepository) LoadRelationships(
	ctx context.Context, projectRoot, systemID string,
) ([]entities.Relationship, error) {
	path := relationshipsPath(projectRoot, systemID)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Not an error — the system simply has no relationships yet.
			return []entities.Relationship{}, nil
		}
		return nil, fmt.Errorf("reading relationships.toml for system %q: %w", systemID, err)
	}

	var file entities.RelationshipsFile
	if err := toml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing relationships.toml for system %q: %w", systemID, err)
	}

	if file.Relationships == nil {
		return []entities.Relationship{}, nil
	}

	return file.Relationships, nil
}

// SaveRelationships atomically overwrites relationships.toml for a system.
//
// The write sequence is:
//  1. Marshal to TOML bytes
//  2. Write to <path>.tmp
//  3. os.Rename(<path>.tmp, <path>)   — atomic on POSIX filesystems
//
// If any step fails, the .tmp file is cleaned up and the original file is untouched.
func (r *FilesystemRelationshipRepository) SaveRelationships(
	ctx context.Context, projectRoot, systemID string, rels []entities.Relationship,
) error {
	path := relationshipsPath(projectRoot, systemID)

	// Ensure the parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating directory for relationships.toml: %w", err)
	}

	file := entities.RelationshipsFile{Relationships: rels}
	data, err := toml.Marshal(file)
	if err != nil {
		return fmt.Errorf("marshaling relationships for system %q: %w", systemID, err)
	}

	tmpPath := path + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temporary relationships file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Best-effort cleanup of the temporary file.
		_ = os.Remove(tmpPath)
		return fmt.Errorf("atomically replacing relationships.toml: %w", err)
	}

	return nil
}

// DeleteElement removes all relationships where source or target equals elementPath.
// This is called when a container or component is deleted, to clean up dangling edges.
// If relationships.toml does not exist, the method is a no-op (no error).
func (r *FilesystemRelationshipRepository) DeleteElement(
	ctx context.Context, projectRoot, systemID, elementPath string,
) error {
	rels, err := r.LoadRelationships(ctx, projectRoot, systemID)
	if err != nil {
		return err
	}
	if len(rels) == 0 {
		return nil // Nothing to remove.
	}

	filtered := make([]entities.Relationship, 0, len(rels))
	for _, rel := range rels {
		if rel.Source == elementPath || rel.Target == elementPath {
			continue // Drop relationships involving the deleted element.
		}
		filtered = append(filtered, rel)
	}

	// No relationships were removed — skip the write.
	if len(filtered) == len(rels) {
		return nil
	}

	return r.SaveRelationships(ctx, projectRoot, systemID, filtered)
}
