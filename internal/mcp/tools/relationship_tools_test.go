package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mock infrastructure
// ─────────────────────────────────────────────────────────────────────────────

// mockRelRepo implements usecases.RelationshipRepository for tool tests.
type mockRelRepo struct {
	data      map[string][]entities.Relationship
	loadErr   error
	saveErr   error
	deleteErr error
}

func newMockRelRepo() *mockRelRepo {
	return &mockRelRepo{data: make(map[string][]entities.Relationship)}
}

func (m *mockRelRepo) key(root, sys string) string { return root + "|" + sys }

func (m *mockRelRepo) LoadRelationships(_ context.Context, root, sys string) ([]entities.Relationship, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	rels := m.data[m.key(root, sys)]
	cp := make([]entities.Relationship, len(rels))
	copy(cp, rels)
	return cp, nil
}

func (m *mockRelRepo) SaveRelationships(_ context.Context, root, sys string, rels []entities.Relationship) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	cp := make([]entities.Relationship, len(rels))
	copy(cp, rels)
	m.data[m.key(root, sys)] = cp
	return nil
}

func (m *mockRelRepo) DeleteElement(_ context.Context, root, sys, elem string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	rels := m.data[m.key(root, sys)]
	filtered := rels[:0]
	for _, r := range rels {
		if r.Source != elem && r.Target != elem {
			filtered = append(filtered, r)
		}
	}
	m.data[m.key(root, sys)] = filtered
	return nil
}

func (m *mockRelRepo) stored(root, sys string) []entities.Relationship {
	return m.data[m.key(root, sys)]
}

func (m *mockRelRepo) seed(root, sys string, rels []entities.Relationship) {
	m.data[m.key(root, sys)] = rels
}

// mockCache implements the tools.GraphCache interface for tool tests.
type mockCache struct {
	invalidated []string
}

func (m *mockCache) Get(_ string) (*entities.ArchitectureGraph, bool) { return nil, false }
func (m *mockCache) Set(_ string, _ *entities.ArchitectureGraph)      {}
func (m *mockCache) Invalidate(root string)                           { m.invalidated = append(m.invalidated, root) }

// ─────────────────────────────────────────────────────────────────────────────
// CreateRelationshipTool tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateRelationshipTool_HappyPath(t *testing.T) {
	repo := newMockRelRepo()
	cache := &mockCache{}
	tool := NewCreateRelationshipTool(repo, nil, cache)
	ctx := context.Background()

	result, err := tool.Call(ctx, map[string]any{
		"project_root": "/tmp/proj",
		"system_name":  "My System",
		"source":       "my-system/api",
		"target":       "my-system/worker",
		"label":        "Dispatches job",
		"type":         "async",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map response, got %T", result)
	}
	if m["relationship"] == nil {
		t.Error("expected 'relationship' key in response")
	}
	rel, _ := m["relationship"].(map[string]any)
	if rel["id"] == nil || rel["id"] == "" {
		t.Error("expected non-empty 'id' in relationship")
	}
	if m["diagram_updated"] != true {
		t.Error("expected diagram_updated: true")
	}

	// Cache must have been invalidated.
	if len(cache.invalidated) != 1 || cache.invalidated[0] != "/tmp/proj" {
		t.Errorf("expected cache invalidation for /tmp/proj, got %v", cache.invalidated)
	}
}

func TestCreateRelationshipTool_MissingRequired(t *testing.T) {
	repo := newMockRelRepo()
	cache := &mockCache{}
	tool := NewCreateRelationshipTool(repo, nil, cache)
	ctx := context.Background()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"missing system_name", map[string]any{"project_root": "/tmp", "source": "s/a", "target": "s/b", "label": "l"}},
		{"missing source", map[string]any{"project_root": "/tmp", "system_name": "sys", "target": "s/b", "label": "l"}},
		{"missing target", map[string]any{"project_root": "/tmp", "system_name": "sys", "source": "s/a", "label": "l"}},
		{"missing label", map[string]any{"project_root": "/tmp", "system_name": "sys", "source": "s/a", "target": "s/b"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Call(ctx, tc.args)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestCreateRelationshipTool_IdempotentDuplicate(t *testing.T) {
	repo := newMockRelRepo()
	cache := &mockCache{}
	tool := NewCreateRelationshipTool(repo, nil, cache)
	ctx := context.Background()

	args := map[string]any{
		"project_root": "/tmp/proj",
		"system_name":  "sys",
		"source":       "sys/a",
		"target":       "sys/b",
		"label":        "link",
	}

	r1, _ := tool.Call(ctx, args)
	r2, _ := tool.Call(ctx, args)

	m1 := r1.(map[string]any)["relationship"].(map[string]any)
	m2 := r2.(map[string]any)["relationship"].(map[string]any)

	if m1["id"] != m2["id"] {
		t.Errorf("duplicate call returned different IDs: %v vs %v", m1["id"], m2["id"])
	}
	// Only 1 relationship stored.
	if len(repo.stored("/tmp/proj", "sys")) != 1 {
		t.Errorf("expected 1 stored relationship after duplicate, got %d", len(repo.stored("/tmp/proj", "sys")))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ListRelationshipsTool tests
// ─────────────────────────────────────────────────────────────────────────────

func TestListRelationshipsTool_EmptySystem(t *testing.T) {
	repo := newMockRelRepo()
	tool := NewListRelationshipsTool(repo, nil)

	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": "/tmp/proj",
		"system_name":  "empty-system",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["count"] != 0 {
		t.Errorf("expected count 0, got %v", m["count"])
	}
	rels, _ := m["relationships"].([]map[string]any)
	if len(rels) != 0 {
		t.Errorf("expected empty relationships, got %d", len(rels))
	}
}

func TestListRelationshipsTool_WithRelationships(t *testing.T) {
	repo := newMockRelRepo()
	rel, _ := entities.NewRelationship("sys/api", "sys/worker", "link")
	repo.seed("/tmp/proj", "sys", []entities.Relationship{*rel})

	tool := NewListRelationshipsTool(repo, nil)
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": "/tmp/proj",
		"system_name":  "sys",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["count"] != 1 {
		t.Errorf("expected count 1, got %v", m["count"])
	}
	if m["system"] != "sys" {
		t.Errorf("expected system 'sys', got %v", m["system"])
	}
}

func TestListRelationshipsTool_MissingSystemName(t *testing.T) {
	repo := newMockRelRepo()
	tool := NewListRelationshipsTool(repo, nil)

	_, err := tool.Call(context.Background(), map[string]any{"project_root": "/tmp"})
	if err == nil {
		t.Error("expected error for missing system_name")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteRelationshipTool tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteRelationshipTool_HappyPath(t *testing.T) {
	repo := newMockRelRepo()
	rel, _ := entities.NewRelationship("sys/a", "sys/b", "link")
	repo.seed("/tmp/proj", "sys", []entities.Relationship{*rel})
	cache := &mockCache{}
	tool := NewDeleteRelationshipTool(repo, nil, cache)

	result, err := tool.Call(context.Background(), map[string]any{
		"project_root":    "/tmp/proj",
		"system_name":     "sys",
		"relationship_id": rel.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["deleted"] != true {
		t.Error("expected deleted: true")
	}
	if m["relationship_id"] != rel.ID {
		t.Errorf("expected relationship_id %q, got %v", rel.ID, m["relationship_id"])
	}

	// Cache must have been invalidated.
	if len(cache.invalidated) != 1 {
		t.Errorf("expected 1 cache invalidation, got %d", len(cache.invalidated))
	}

	// Relationship removed from storage.
	if len(repo.stored("/tmp/proj", "sys")) != 0 {
		t.Error("expected 0 stored relationships after delete")
	}
}

func TestDeleteRelationshipTool_NotFound(t *testing.T) {
	repo := newMockRelRepo()
	cache := &mockCache{}
	tool := NewDeleteRelationshipTool(repo, nil, cache)

	_, err := tool.Call(context.Background(), map[string]any{
		"project_root":    "/tmp/proj",
		"system_name":     "sys",
		"relationship_id": "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for non-existent relationship_id")
	}
	if !errors.Is(err, usecases.ErrRelationshipNotFound) {
		t.Errorf("expected ErrRelationshipNotFound, got: %v", err)
	}

	// Cache must NOT be invalidated on error.
	if len(cache.invalidated) != 0 {
		t.Errorf("expected no cache invalidation on error, got %d", len(cache.invalidated))
	}
}

func TestDeleteRelationshipTool_MissingRequired(t *testing.T) {
	repo := newMockRelRepo()
	cache := &mockCache{}
	tool := NewDeleteRelationshipTool(repo, nil, cache)
	ctx := context.Background()

	_, err := tool.Call(ctx, map[string]any{"project_root": "/tmp", "relationship_id": "abc"})
	if err == nil {
		t.Error("expected error for missing system_name")
	}

	_, err = tool.Call(ctx, map[string]any{"project_root": "/tmp", "system_name": "sys"})
	if err == nil {
		t.Error("expected error for missing relationship_id")
	}
}
