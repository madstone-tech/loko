package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// MockRelationshipRepository is an in-memory implementation of RelationshipRepository
// for use in use case unit tests. Not thread-safe (tests are single-threaded).
type MockRelationshipRepository struct {
	// data maps "projectRoot|systemID" -> []Relationship
	data map[string][]entities.Relationship

	// Recorded calls for assertion
	LoadCalls   []mockRelLoad
	SaveCalls   []mockRelSave
	DeleteCalls []mockRelDelete

	// Injected errors (set per key to simulate failures)
	LoadErr   error
	SaveErr   error
	DeleteErr error
}

type mockRelLoad struct {
	ProjectRoot string
	SystemID    string
}

type mockRelSave struct {
	ProjectRoot string
	SystemID    string
	Rels        []entities.Relationship
}

type mockRelDelete struct {
	ProjectRoot string
	SystemID    string
	ElementPath string
}

func newMockRelationshipRepository() *MockRelationshipRepository {
	return &MockRelationshipRepository{
		data: make(map[string][]entities.Relationship),
	}
}

func (m *MockRelationshipRepository) key(projectRoot, systemID string) string {
	return projectRoot + "|" + systemID
}

func (m *MockRelationshipRepository) LoadRelationships(
	ctx context.Context, projectRoot, systemID string,
) ([]entities.Relationship, error) {
	m.LoadCalls = append(m.LoadCalls, mockRelLoad{projectRoot, systemID})
	if m.LoadErr != nil {
		return nil, m.LoadErr
	}
	rels := m.data[m.key(projectRoot, systemID)]
	if rels == nil {
		return []entities.Relationship{}, nil
	}
	// Return a copy to prevent callers from mutating internal state.
	cp := make([]entities.Relationship, len(rels))
	copy(cp, rels)
	return cp, nil
}

func (m *MockRelationshipRepository) SaveRelationships(
	ctx context.Context, projectRoot, systemID string, rels []entities.Relationship,
) error {
	m.SaveCalls = append(m.SaveCalls, mockRelSave{projectRoot, systemID, rels})
	if m.SaveErr != nil {
		return m.SaveErr
	}
	cp := make([]entities.Relationship, len(rels))
	copy(cp, rels)
	m.data[m.key(projectRoot, systemID)] = cp
	return nil
}

func (m *MockRelationshipRepository) DeleteElement(
	ctx context.Context, projectRoot, systemID, elementPath string,
) error {
	m.DeleteCalls = append(m.DeleteCalls, mockRelDelete{projectRoot, systemID, elementPath})
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	rels := m.data[m.key(projectRoot, systemID)]
	filtered := make([]entities.Relationship, 0, len(rels))
	for _, r := range rels {
		if r.Source != elementPath && r.Target != elementPath {
			filtered = append(filtered, r)
		}
	}
	m.data[m.key(projectRoot, systemID)] = filtered
	return nil
}

// seed directly sets relationships for a system (test helper, bypasses SaveCalls tracking).
func (m *MockRelationshipRepository) seed(projectRoot, systemID string, rels []entities.Relationship) {
	m.data[m.key(projectRoot, systemID)] = rels
}

// stored returns the current stored relationships (test assertion helper).
func (m *MockRelationshipRepository) stored(projectRoot, systemID string) []entities.Relationship {
	return m.data[m.key(projectRoot, systemID)]
}

// mockD2Writer is a minimal D2 file writer for use case tests.
// It records write calls but does not touch the filesystem.
type mockD2Writer struct {
	Written  map[string]string // path -> content
	WriteErr error
}

func newMockD2Writer() *mockD2Writer {
	return &mockD2Writer{Written: make(map[string]string)}
}

func (m *mockD2Writer) write(path, content string) error {
	if m.WriteErr != nil {
		return fmt.Errorf("mock d2 write error: %w", m.WriteErr)
	}
	m.Written[path] = content
	return nil
}
