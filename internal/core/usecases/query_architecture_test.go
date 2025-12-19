package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestQueryArchitectureEmpty tests querying an empty project.
func TestQueryArchitectureEmpty(t *testing.T) {
	emptyProject, _ := entities.NewProject("Empty")
	repo := &MockProjectRepository{
		LoadProjectFunc: func(ctx context.Context, projectRoot string) (*entities.Project, error) {
			return emptyProject, nil
		},
		ListSystemsFunc: func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
			return []*entities.System{}, nil
		},
	}

	uc := NewQueryArchitecture(repo)

	resp, err := uc.Execute(context.Background(), "test", "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("response should not be nil")
	}

	if resp.Detail != "summary" {
		t.Errorf("expected detail='summary', got '%s'", resp.Detail)
	}

	// Empty project should have short response
	if resp.TokenEstimate > 100 {
		t.Errorf("empty project token estimate should be <100, got %d", resp.TokenEstimate)
	}
}

// TestQueryArchitectureSummary tests summary detail level.
func TestQueryArchitectureSummary(t *testing.T) {
	project := createTestProject("TestProj", 3, 2) // 3 systems, 2 containers each
	systems, _ := getTestSystems(project)

	repo := &MockProjectRepository{
		LoadProjectFunc: func(ctx context.Context, projectRoot string) (*entities.Project, error) {
			return project, nil
		},
		ListSystemsFunc: func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
			return systems, nil
		},
	}

	uc := NewQueryArchitecture(repo)

	resp, err := uc.Execute(context.Background(), "test", "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Summary should be brief (3 systems with descriptions)
	if resp.TokenEstimate < 15 || resp.TokenEstimate > 100 {
		t.Errorf("summary token estimate should be ~50, got %d", resp.TokenEstimate)
	}

	if resp.Text == "" {
		t.Fatal("response text should not be empty")
	}
}

// TestQueryArchitectureStructure tests structure detail level.
func TestQueryArchitectureStructure(t *testing.T) {
	project := createTestProject("TestProj", 3, 2)
	systems, _ := getTestSystems(project)

	repo := &MockProjectRepository{
		LoadProjectFunc: func(ctx context.Context, projectRoot string) (*entities.Project, error) {
			return project, nil
		},
		ListSystemsFunc: func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
			return systems, nil
		},
	}

	uc := NewQueryArchitecture(repo)

	resp, err := uc.Execute(context.Background(), "test", "structure")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Structure should be larger than summary
	if resp.TokenEstimate < 50 || resp.TokenEstimate > 200 {
		t.Errorf("structure token estimate should be ~100, got %d", resp.TokenEstimate)
	}

	// Structure should include systems
	if len(resp.Systems) == 0 {
		t.Error("structure response should include systems")
	}
}

// TestQueryArchitectureFull tests full detail level.
func TestQueryArchitectureFull(t *testing.T) {
	project := createTestProject("TestProj", 2, 3) // 2 systems, 3 containers each
	systems, _ := getTestSystems(project)

	repo := &MockProjectRepository{
		LoadProjectFunc: func(ctx context.Context, projectRoot string) (*entities.Project, error) {
			return project, nil
		},
		ListSystemsFunc: func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
			return systems, nil
		},
	}

	uc := NewQueryArchitecture(repo)

	resp, err := uc.Execute(context.Background(), "test", "full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Full should include all systems
	if len(resp.Systems) != 2 {
		t.Errorf("expected 2 systems in full response, got %d", len(resp.Systems))
	}

	// Full response should be substantial
	if resp.TokenEstimate < 100 {
		t.Errorf("full response should have substantial tokens, got %d", resp.TokenEstimate)
	}
}

// TestQueryArchitectureInvalidDetail tests invalid detail level.
func TestQueryArchitectureInvalidDetail(t *testing.T) {
	repo := &MockProjectRepository{
		LoadProjectFunc: func(ctx context.Context, projectRoot string) (*entities.Project, error) {
			proj, _ := entities.NewProject("Test")
			return proj, nil
		},
	}

	uc := NewQueryArchitecture(repo)

	_, err := uc.Execute(context.Background(), "test", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid detail level")
	}
}

// TestTokenCountingAccuracy verifies token estimates are reasonable.
func TestTokenCountingAccuracy(t *testing.T) {
	project := createTestProject("TestProj", 5, 4) // Larger project
	systems, _ := getTestSystems(project)

	repo := &MockProjectRepository{
		LoadProjectFunc: func(ctx context.Context, projectRoot string) (*entities.Project, error) {
			return project, nil
		},
		ListSystemsFunc: func(ctx context.Context, projectRoot string) ([]*entities.System, error) {
			return systems, nil
		},
	}

	uc := NewQueryArchitecture(repo)

	// Test each detail level (tokens scale with project size: 5 systems, 4 containers each)
	tests := []struct {
		detail    string
		minTokens int
		maxTokens int
	}{
		{"summary", 10, 100}, // Small project
		{"structure", 50, 300},
		{"full", 100, 600},
	}

	for _, tt := range tests {
		resp, err := uc.Execute(context.Background(), "test", tt.detail)
		if err != nil {
			t.Fatalf("detail=%s: unexpected error: %v", tt.detail, err)
		}

		if resp.TokenEstimate < tt.minTokens || resp.TokenEstimate > tt.maxTokens {
			t.Errorf("detail=%s: token estimate %d outside range [%d, %d]",
				tt.detail, resp.TokenEstimate, tt.minTokens, tt.maxTokens)
		}
	}
}

// Helper functions

func createTestProject(name string, numSystems, numContainers int) *entities.Project {
	project, _ := entities.NewProject(name)

	for i := 0; i < numSystems; i++ {
		sysName := "System" + string(rune(i+65))
		system, _ := entities.NewSystem(sysName)
		system.Description = "Test system " + string(rune(i+65))

		for j := 0; j < numContainers; j++ {
			contName := "Container" + string(rune(j+65))
			container, _ := entities.NewContainer(contName)
			container.Description = "Test container " + string(rune(j+65))
			container.Technology = "Go"
			system.AddContainer(container)
		}

		project.AddSystem(system)
	}

	return project
}

func getTestSystems(project *entities.Project) ([]*entities.System, error) {
	systems := make([]*entities.System, 0, len(project.Systems))
	for _, sys := range project.Systems {
		systems = append(systems, sys)
	}
	return systems, nil
}
