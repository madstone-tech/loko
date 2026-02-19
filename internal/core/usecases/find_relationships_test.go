package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestNewFindRelationships tests creating a FindRelationships use case.
func TestNewFindRelationships(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	uc := NewFindRelationships(mockRepo)

	if uc == nil {
		t.Error("NewFindRelationships() returned nil")
	}

	if uc.repo != mockRepo {
		t.Error("NewFindRelationships() did not set repo correctly")
	}

	if uc.buildGraph == nil {
		t.Error("NewFindRelationships() did not initialize buildGraph")
	}
}

// TestFindRelationshipsExecute tests the Execute method of FindRelationships.
func TestFindRelationshipsExecute(t *testing.T) {
	// Create test project
	project, err := entities.NewProject("test-project")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create test systems with relationships
	sys1, _ := entities.NewSystem("Backend")
	cont1, _ := entities.NewContainer("API")
	comp1, _ := entities.NewComponent("Auth Service")
	comp2, _ := entities.NewComponent("Database")

	// Add relationship between components
	comp1.AddRelationship(comp2.ID, "queries user data")

	cont1.AddComponent(comp1)
	cont1.AddComponent(comp2)
	sys1.AddContainer(cont1)

	tests := []struct {
		name       string
		request    entities.FindRelationshipsRequest
		setupMocks func(*MockProjectRepository)
		wantErr    bool
		validate   func(t *testing.T, result *entities.FindRelationshipsResponse)
	}{
		{
			name: "valid relationship search",
			request: entities.FindRelationshipsRequest{
				ProjectRoot:   "/test/project",
				SourcePattern: "*",
				Limit:         10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
					return project, nil
				}
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.FindRelationshipsResponse) {
				if result == nil {
					t.Fatal("result should not be nil")
				}
				if result.Message == "" {
					t.Error("expected non-empty message")
				}
			},
		},
		{
			name: "search with source pattern",
			request: entities.FindRelationshipsRequest{
				ProjectRoot:   "/test/project",
				SourcePattern: "*",
				Limit:         10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
					return project, nil
				}
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.FindRelationshipsResponse) {
				// Just verify it runs without error for now
				if result == nil {
					t.Error("result should not be nil")
				}
			},
		},
		{
			name: "search with target pattern",
			request: entities.FindRelationshipsRequest{
				ProjectRoot:   "/test/project",
				TargetPattern: "*",
				Limit:         10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
					return project, nil
				}
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.FindRelationshipsResponse) {
				// Just verify it runs without error for now
				if result == nil {
					t.Error("result should not be nil")
				}
			},
		},
		{
			name: "search with relationship type filter",
			request: entities.FindRelationshipsRequest{
				ProjectRoot:      "/test/project",
				SourcePattern:    "*",
				RelationshipType: "queries user data",
				Limit:            10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
					return project, nil
				}
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.FindRelationshipsResponse) {
				// Just verify it runs without error for now
				if result == nil {
					t.Error("result should not be nil")
				}
			},
		},
		{
			name: "no relationships found",
			request: entities.FindRelationshipsRequest{
				ProjectRoot:   "/test/project",
				SourcePattern: "nonexistent",
				Limit:         10,
			},
			setupMocks: func(m *MockProjectRepository) {
				m.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
					return project, nil
				}
				m.ListSystemsFunc = func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
					return []*entities.System{sys1}, nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result *entities.FindRelationshipsResponse) {
				if result.TotalMatched != 0 {
					t.Errorf("expected 0 total matched, got %d", result.TotalMatched)
				}
				if len(result.Relationships) != 0 {
					t.Errorf("expected 0 relationships, got %d", len(result.Relationships))
				}
				if result.Message != "No relationships found" {
					t.Errorf("expected 'No relationships found' message, got %q", result.Message)
				}
			},
		},
		{
			name:       "nil request validation",
			request:    entities.FindRelationshipsRequest{}, // Invalid: no ProjectRoot
			setupMocks: func(m *MockProjectRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProjectRepository{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			uc := NewFindRelationships(mockRepo)
			result, err := uc.Execute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestFindRelationshipsBuildMessage tests the buildMessage helper function.
func TestFindRelationshipsBuildMessage(t *testing.T) {
	uc := &FindRelationships{} // We only need the method, not full initialization

	tests := []struct {
		name         string
		totalMatched int
		returned     int
		request      entities.FindRelationshipsRequest
		expectedMsg  string
	}{
		{
			name:         "no relationships found",
			totalMatched: 0,
			returned:     0,
			request:      entities.FindRelationshipsRequest{},
			expectedMsg:  "No relationships found",
		},
		{
			name:         "all relationships returned",
			totalMatched: 5,
			returned:     5,
			request:      entities.FindRelationshipsRequest{},
			expectedMsg:  "Found 5 relationships",
		},
		{
			name:         "limited results with source filter",
			totalMatched: 10,
			returned:     5,
			request:      entities.FindRelationshipsRequest{SourcePattern: "auth"},
			expectedMsg:  "Showing 5 of 10 matching relationships (use limit parameter to adjust)",
		},
		{
			name:         "with source pattern filter",
			totalMatched: 3,
			returned:     3,
			request:      entities.FindRelationshipsRequest{SourcePattern: "auth"},
			expectedMsg:  "Found 3 relationships with filters: source=auth",
		},
		{
			name:         "with target pattern filter",
			totalMatched: 2,
			returned:     2,
			request:      entities.FindRelationshipsRequest{TargetPattern: "db"},
			expectedMsg:  "Found 2 relationships with filters: target=db",
		},
		{
			name:         "with relationship type filter",
			totalMatched: 1,
			returned:     1,
			request:      entities.FindRelationshipsRequest{RelationshipType: "calls"},
			expectedMsg:  "Found 1 relationships with filters: type=calls",
		},
		{
			name:         "with multiple filters",
			totalMatched: 4,
			returned:     4,
			request:      entities.FindRelationshipsRequest{SourcePattern: "api", TargetPattern: "db", RelationshipType: "queries"},
			expectedMsg:  "Found 4 relationships with filters: source=api, target=db, type=queries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := uc.buildMessage(tt.totalMatched, tt.returned, tt.request)
			if msg != tt.expectedMsg {
				t.Errorf("buildMessage() = %q, want %q", msg, tt.expectedMsg)
			}
		})
	}
}
