package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func newTestRepo() *FilesystemRelationshipRepository {
	return NewFilesystemRelationshipRepository()
}

// makeRel is a test helper that creates a Relationship via the entity constructor.
func makeRel(t *testing.T, source, target, label string, opts ...entities.RelationshipOption) entities.Relationship {
	t.Helper()
	rel, err := entities.NewRelationship(source, target, label, opts...)
	if err != nil {
		t.Fatalf("makeRel(%q, %q, %q): %v", source, target, label, err)
	}
	return *rel
}

// TestLoadRelationships_AbsentFile verifies that a missing relationships.toml
// returns an empty slice, not an error.
func TestLoadRelationships_AbsentFile(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()

	rels, err := repo.LoadRelationships(ctx, dir, "my-system")
	if err != nil {
		t.Fatalf("expected no error for absent file, got: %v", err)
	}
	if len(rels) != 0 {
		t.Errorf("expected empty slice, got %d relationships", len(rels))
	}
}

// TestSaveAndLoadRelationships verifies a round-trip: save → load returns same data.
func TestSaveAndLoadRelationships(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()
	systemID := "test-system"

	rels := []entities.Relationship{
		makeRel(t, "test-system/api", "test-system/db", "Reads data"),
		makeRel(t, "test-system/worker", "test-system/queue", "Dequeues jobs",
			entities.WithRelType("async"),
			entities.WithRelTechnology("AWS SQS"),
		),
	}

	if err := repo.SaveRelationships(ctx, dir, systemID, rels); err != nil {
		t.Fatalf("SaveRelationships: %v", err)
	}

	loaded, err := repo.LoadRelationships(ctx, dir, systemID)
	if err != nil {
		t.Fatalf("LoadRelationships: %v", err)
	}

	if len(loaded) != len(rels) {
		t.Fatalf("expected %d relationships, got %d", len(rels), len(loaded))
	}

	for i, want := range rels {
		got := loaded[i]
		if got.ID != want.ID {
			t.Errorf("[%d] ID: got %q, want %q", i, got.ID, want.ID)
		}
		if got.Source != want.Source {
			t.Errorf("[%d] Source: got %q, want %q", i, got.Source, want.Source)
		}
		if got.Target != want.Target {
			t.Errorf("[%d] Target: got %q, want %q", i, got.Target, want.Target)
		}
		if got.Label != want.Label {
			t.Errorf("[%d] Label: got %q, want %q", i, got.Label, want.Label)
		}
		if got.Type != want.Type {
			t.Errorf("[%d] Type: got %q, want %q", i, got.Type, want.Type)
		}
		if got.Technology != want.Technology {
			t.Errorf("[%d] Technology: got %q, want %q", i, got.Technology, want.Technology)
		}
	}
}

// TestSaveRelationships_CreatesDirectory verifies that SaveRelationships creates
// the src/<systemID>/ directory if it does not exist yet.
func TestSaveRelationships_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()
	systemID := "brand-new-system"

	// Confirm directory does NOT exist yet
	sysDir := filepath.Join(dir, "src", systemID)
	if _, err := os.Stat(sysDir); !os.IsNotExist(err) {
		t.Fatalf("expected directory to not exist, but it does")
	}

	rels := []entities.Relationship{makeRel(t, "brand-new-system/a", "brand-new-system/b", "test")}
	if err := repo.SaveRelationships(ctx, dir, systemID, rels); err != nil {
		t.Fatalf("SaveRelationships: %v", err)
	}

	tomlPath := filepath.Join(sysDir, "relationships.toml")
	if _, err := os.Stat(tomlPath); err != nil {
		t.Errorf("expected relationships.toml to exist: %v", err)
	}
}

// TestSaveRelationships_IsAtomic verifies the write-to-.tmp-then-rename pattern:
// no .tmp file should remain after a successful save.
func TestSaveRelationships_IsAtomic(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()
	systemID := "atomic-test"

	rels := []entities.Relationship{makeRel(t, "atomic-test/a", "atomic-test/b", "writes")}
	if err := repo.SaveRelationships(ctx, dir, systemID, rels); err != nil {
		t.Fatalf("SaveRelationships: %v", err)
	}

	tmpPath := filepath.Join(dir, "src", systemID, "relationships.toml.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("expected .tmp file to be removed after successful save")
	}
}

// TestSaveRelationships_EmptySlice verifies that saving an empty slice creates
// a valid (empty) relationships.toml, and loading it returns an empty slice.
func TestSaveRelationships_EmptySlice(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()
	systemID := "empty-system"

	if err := repo.SaveRelationships(ctx, dir, systemID, []entities.Relationship{}); err != nil {
		t.Fatalf("SaveRelationships with empty slice: %v", err)
	}

	loaded, err := repo.LoadRelationships(ctx, dir, systemID)
	if err != nil {
		t.Fatalf("LoadRelationships: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected empty slice after saving empty, got %d", len(loaded))
	}
}

// TestDeleteElement_RemovesMatchingRelationships verifies that DeleteElement
// removes only relationships where source or target matches elementPath.
func TestDeleteElement_RemovesMatchingRelationships(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()
	systemID := "sys"

	rels := []entities.Relationship{
		makeRel(t, "sys/api", "sys/db", "Reads"),
		makeRel(t, "sys/worker", "sys/api", "Calls"),     // target matches
		makeRel(t, "sys/api", "sys/cache", "Caches"),     // source matches
		makeRel(t, "sys/worker", "sys/queue", "Enqueue"), // no match
	}

	if err := repo.SaveRelationships(ctx, dir, systemID, rels); err != nil {
		t.Fatalf("SaveRelationships: %v", err)
	}

	// Delete element "sys/api" — should remove first three entries
	if err := repo.DeleteElement(ctx, dir, systemID, "sys/api"); err != nil {
		t.Fatalf("DeleteElement: %v", err)
	}

	loaded, err := repo.LoadRelationships(ctx, dir, systemID)
	if err != nil {
		t.Fatalf("LoadRelationships after delete: %v", err)
	}

	// Only "sys/worker → sys/queue" should remain
	if len(loaded) != 1 {
		t.Fatalf("expected 1 relationship after delete, got %d: %+v", len(loaded), loaded)
	}
	if loaded[0].Source != "sys/worker" || loaded[0].Target != "sys/queue" {
		t.Errorf("unexpected remaining relationship: %+v", loaded[0])
	}
}

// TestDeleteElement_NoMatchIsNoOp verifies that deleting an element with no
// matching relationships leaves the file unchanged (no error).
func TestDeleteElement_NoMatchIsNoOp(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()
	systemID := "sys"

	rels := []entities.Relationship{
		makeRel(t, "sys/a", "sys/b", "link"),
	}
	if err := repo.SaveRelationships(ctx, dir, systemID, rels); err != nil {
		t.Fatalf("SaveRelationships: %v", err)
	}

	if err := repo.DeleteElement(ctx, dir, systemID, "sys/nonexistent"); err != nil {
		t.Fatalf("DeleteElement with no match: %v", err)
	}

	loaded, err := repo.LoadRelationships(ctx, dir, systemID)
	if err != nil {
		t.Fatalf("LoadRelationships: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("expected 1 relationship (unchanged), got %d", len(loaded))
	}
}

// TestDeleteElement_AbsentFile verifies that DeleteElement on a system with no
// relationships.toml is a no-op (no error).
func TestDeleteElement_AbsentFile(t *testing.T) {
	dir := t.TempDir()
	repo := newTestRepo()
	ctx := context.Background()

	err := repo.DeleteElement(ctx, dir, "nonexistent-system", "sys/element")
	if err != nil {
		t.Errorf("DeleteElement on absent file should be no-op, got: %v", err)
	}
}
