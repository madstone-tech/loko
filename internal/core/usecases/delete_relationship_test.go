package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestDeleteRelationship_HappyPath(t *testing.T) {
	repo := newMockRelationshipRepository()
	rel1 := makeTestRel(t, "sys/a", "sys/b", "link1")
	rel2 := makeTestRel(t, "sys/b", "sys/c", "link2")
	repo.seed("/tmp/proj", "sys", []entities.Relationship{rel1, rel2})

	uc := NewDeleteRelationship(repo)
	err := uc.Execute(context.Background(), &DeleteRelationshipRequest{
		ProjectRoot:    "/tmp/proj",
		SystemID:       "sys",
		RelationshipID: rel1.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only rel2 should remain.
	stored := repo.stored("/tmp/proj", "sys")
	if len(stored) != 1 {
		t.Fatalf("expected 1 remaining relationship, got %d", len(stored))
	}
	if stored[0].ID != rel2.ID {
		t.Errorf("wrong remaining relationship: got %q, want %q", stored[0].ID, rel2.ID)
	}
}

func TestDeleteRelationship_NotFound(t *testing.T) {
	repo := newMockRelationshipRepository()
	rel := makeTestRel(t, "sys/a", "sys/b", "link")
	repo.seed("/tmp/proj", "sys", []entities.Relationship{rel})

	uc := NewDeleteRelationship(repo)
	err := uc.Execute(context.Background(), &DeleteRelationshipRequest{
		ProjectRoot:    "/tmp/proj",
		SystemID:       "sys",
		RelationshipID: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected ErrRelationshipNotFound, got nil")
	}
	if !errors.Is(err, ErrRelationshipNotFound) {
		t.Errorf("expected ErrRelationshipNotFound, got: %v", err)
	}
}

func TestDeleteRelationship_EmptyID(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewDeleteRelationship(repo)
	err := uc.Execute(context.Background(), &DeleteRelationshipRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
	})
	if err == nil {
		t.Fatal("expected error for empty relationship_id")
	}
}

func TestDeleteRelationship_RepoSaveError(t *testing.T) {
	repo := newMockRelationshipRepository()
	rel := makeTestRel(t, "sys/a", "sys/b", "link")
	repo.seed("/tmp/proj", "sys", []entities.Relationship{rel})
	repo.SaveErr = errors.New("disk full")

	uc := NewDeleteRelationship(repo)
	err := uc.Execute(context.Background(), &DeleteRelationshipRequest{
		ProjectRoot:    "/tmp/proj",
		SystemID:       "sys",
		RelationshipID: rel.ID,
	})
	if err == nil {
		t.Fatal("expected error from repo Save failure")
	}
}
