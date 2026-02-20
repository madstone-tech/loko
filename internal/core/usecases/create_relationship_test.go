package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestCreateRelationship_HappyPath(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewCreateRelationship(repo)
	ctx := context.Background()

	req := &CreateRelationshipRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "my-system",
		Source:      "my-system/api",
		Target:      "my-system/db",
		Label:       "Reads data",
		Type:        "sync",
	}

	rel, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel == nil {
		t.Fatal("expected non-nil relationship")
	}
	if rel.ID == "" {
		t.Error("expected non-empty ID")
	}
	if rel.Source != req.Source {
		t.Errorf("Source: got %q, want %q", rel.Source, req.Source)
	}
	if rel.Target != req.Target {
		t.Errorf("Target: got %q, want %q", rel.Target, req.Target)
	}

	// Verify it was persisted.
	if len(repo.SaveCalls) != 1 {
		t.Errorf("expected 1 Save call, got %d", len(repo.SaveCalls))
	}
	stored := repo.stored("/tmp/proj", "my-system")
	if len(stored) != 1 {
		t.Errorf("expected 1 stored relationship, got %d", len(stored))
	}
}

func TestCreateRelationship_Idempotent(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewCreateRelationship(repo)
	ctx := context.Background()

	req := &CreateRelationshipRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
		Source:      "sys/a",
		Target:      "sys/b",
		Label:       "link",
	}

	rel1, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// Second call with identical args — must return existing without re-saving.
	rel2, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if rel1.ID != rel2.ID {
		t.Errorf("expected same ID on duplicate: %q vs %q", rel1.ID, rel2.ID)
	}

	// Only 1 Save should have happened (first call only).
	saveCalls := 0
	for _, c := range repo.SaveCalls {
		if c.SystemID == "sys" {
			saveCalls++
		}
	}
	if saveCalls != 1 {
		t.Errorf("expected 1 Save call for idempotent duplicate, got %d", saveCalls)
	}

	// Only 1 relationship stored.
	stored := repo.stored("/tmp/proj", "sys")
	if len(stored) != 1 {
		t.Errorf("expected 1 stored relationship, got %d", len(stored))
	}
}

func TestCreateRelationship_ValidationError(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewCreateRelationship(repo)
	ctx := context.Background()

	// Source == Target — validation error.
	req := &CreateRelationshipRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
		Source:      "sys/a",
		Target:      "sys/a",
		Label:       "self",
	}

	_, err := uc.Execute(ctx, req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	// No save should have occurred.
	if len(repo.SaveCalls) != 0 {
		t.Errorf("expected 0 Save calls on validation error, got %d", len(repo.SaveCalls))
	}
}

func TestCreateRelationship_WithOptions(t *testing.T) {
	repo := newMockRelationshipRepository()
	uc := NewCreateRelationship(repo)
	ctx := context.Background()

	req := &CreateRelationshipRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
		Source:      "sys/producer",
		Target:      "sys/consumer",
		Label:       "publishes events",
		Type:        "event",
		Technology:  "AWS EventBridge",
		Direction:   "forward",
	}

	rel, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.Type != "event" {
		t.Errorf("Type: got %q, want %q", rel.Type, "event")
	}
	if rel.Technology != "AWS EventBridge" {
		t.Errorf("Technology: got %q, want %q", rel.Technology, "AWS EventBridge")
	}
}

func TestCreateRelationship_RepoLoadError(t *testing.T) {
	repo := newMockRelationshipRepository()
	repo.LoadErr = errors.New("disk failure")
	uc := NewCreateRelationship(repo)
	ctx := context.Background()

	req := &CreateRelationshipRequest{
		ProjectRoot: "/tmp/proj",
		SystemID:    "sys",
		Source:      "sys/a",
		Target:      "sys/b",
		Label:       "link",
	}

	_, err := uc.Execute(ctx, req)
	if err == nil {
		t.Fatal("expected error from repo failure, got nil")
	}
}

func TestD2DiagramPath_SameContainer(t *testing.T) {
	rel := &entities.Relationship{
		Source: "sys/api/auth",
		Target: "sys/api/db",
	}
	path := D2DiagramPath("/root", "sys", rel)
	expected := "/root/src/sys/api/container.d2"
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestD2DiagramPath_CrossContainer(t *testing.T) {
	rel := &entities.Relationship{
		Source: "sys/api",
		Target: "sys/worker",
	}
	path := D2DiagramPath("/root", "sys", rel)
	expected := "/root/src/sys/system.d2"
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
