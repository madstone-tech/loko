package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func makeTestRel(t *testing.T, source, target, label string) entities.Relationship {
	t.Helper()
	rel, err := entities.NewRelationship(source, target, label)
	if err != nil {
		t.Fatalf("makeTestRel: %v", err)
	}
	return *rel
}

func TestListRelationships_EmptyProject(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewListRelationships(repo)
	ctx := context.Background()

	result, err := uc.Execute(ctx, &ListRelationshipsRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestListRelationships_NoFilter(t *testing.T) {
	repo := newMockRelationshipRepository()
	rels := []entities.Relationship{
		makeTestRel(t, "sys/a", "sys/b", "link1"),
		makeTestRel(t, "sys/b", "sys/c", "link2"),
		makeTestRel(t, "sys/c", "sys/a", "link3"),
	}
	repo.seed("/tmp/proj", "sys", rels)

	uc := NewListRelationships(repo)
	result, err := uc.Execute(context.Background(), &ListRelationshipsRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 relationships, got %d", len(result))
	}
}

func TestListRelationships_SourceFilter(t *testing.T) {
	repo := newMockRelationshipRepository()
	rels := []entities.Relationship{
		makeTestRel(t, "sys/api", "sys/db", "reads"),
		makeTestRel(t, "sys/api", "sys/cache", "caches"),
		makeTestRel(t, "sys/worker", "sys/queue", "enqueues"),
	}
	repo.seed("/tmp/proj", "sys", rels)

	uc := NewListRelationships(repo)
	result, err := uc.Execute(context.Background(), &ListRelationshipsRequest{
		ProjectRoot:  "/tmp/proj",
		SystemID:     "sys",
		FilterSource: "sys/api",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 relationships from sys/api, got %d", len(result))
	}
	for _, r := range result {
		if r.Source != "sys/api" {
			t.Errorf("expected Source=sys/api, got %q", r.Source)
		}
	}
}

func TestListRelationships_TargetFilter(t *testing.T) {
	repo := newMockRelationshipRepository()
	rels := []entities.Relationship{
		makeTestRel(t, "sys/api", "sys/db", "reads"),
		makeTestRel(t, "sys/worker", "sys/db", "writes"),
		makeTestRel(t, "sys/api", "sys/cache", "caches"),
	}
	repo.seed("/tmp/proj", "sys", rels)

	uc := NewListRelationships(repo)
	result, err := uc.Execute(context.Background(), &ListRelationshipsRequest{
		ProjectRoot:  "/tmp/proj",
		SystemID:     "sys",
		FilterTarget: "sys/db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 relationships to sys/db, got %d", len(result))
	}
}

func TestListRelationships_MissingRequired(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewListRelationships(repo)
	ctx := context.Background()

	_, err := uc.Execute(ctx, &ListRelationshipsRequest{SystemID: "sys"})
	if err == nil {
		t.Error("expected error for missing project_root")
	}

	_, err = uc.Execute(ctx, &ListRelationshipsRequest{ProjectRoot: "/tmp"})
	if err == nil {
		t.Error("expected error for missing system_id")
	}
}
