package usecases

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestDetectDrift_NoDrift tests the case where there are no drift issues.
func TestDetectDrift_NoDrift(t *testing.T) {
	// Create test project with consistent components
	_, _ = entities.NewProject("test-project")

	// Create system with container and components
	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")

	// Create components with consistent relationships
	auth, _ := entities.NewComponent("Auth")
	auth.SetDescription("Authentication service")

	db, _ := entities.NewComponent("Database")
	db.SetDescription("Database service")

	// Add relationship between existing components
	auth.AddRelationship(db.ID, "queries user data")

	container.AddComponent(auth)
	container.AddComponent(db)
	system.AddContainer(container)

	// Create use case with mock repository
	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}

	uc := NewDetectDrift(mockRepo)
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/test",
		Systems:     []*entities.System{system},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify no drift issues
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d: %+v", len(result.Issues), result.Issues)
	}

	if result.HasErrors {
		t.Error("expected HasErrors to be false")
	}

	if result.HasWarnings {
		t.Error("expected HasWarnings to be false")
	}

	if result.ComponentsChecked == 0 {
		t.Error("expected ComponentsChecked to be > 0")
	}
}

// TestDetectDrift_OrphanedRelationship tests DriftOrphanedRelationship error.
func TestDetectDrift_OrphanedRelationship(t *testing.T) {
	// Create test project
	_, _ = entities.NewProject("test-project")

	// Create system with container and components
	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")

	// Create component with relationship to non-existent component
	auth, _ := entities.NewComponent("Auth")
	auth.AddRelationship("deleted-comp", "calls") // This component doesn't exist

	container.AddComponent(auth)
	system.AddContainer(container)

	// Create mock repository
	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}

	uc := NewDetectDrift(mockRepo)
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/test",
		Systems:     []*entities.System{system},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check that we have an orphaned relationship issue
	foundOrphaned := false
	for _, issue := range result.Issues {
		if issue.Type == entities.DriftOrphanedRelationship {
			foundOrphaned = true
			break
		}
	}

	if !foundOrphaned {
		t.Error("expected DriftOrphanedRelationship issue, but none found")
	}

	if !result.HasErrors {
		t.Error("expected HasErrors to be true")
	}
}

// TestDetectDrift_MultipleDriftTypes tests multiple drift types in one run.
func TestDetectDrift_MultipleDriftTypes(t *testing.T) {
	// Create test project
	_, _ = entities.NewProject("test-project")

	// Create system with container and components
	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")

	// Create component with orphaned relationship
	auth, _ := entities.NewComponent("Auth")
	auth.AddRelationship("deleted-comp", "calls") // Orphaned relationship

	// Create another component with another orphaned relationship
	api, _ := entities.NewComponent("API")
	api.AddRelationship("missing-service", "depends on") // Another orphaned relationship

	container.AddComponent(auth)
	container.AddComponent(api)
	system.AddContainer(container)

	// Create mock repository
	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}

	uc := NewDetectDrift(mockRepo)
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/test",
		Systems:     []*entities.System{system},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should have 2 orphaned relationship issues
	orphanedCount := 0
	for _, issue := range result.Issues {
		if issue.Type == entities.DriftOrphanedRelationship {
			orphanedCount++
		}
	}

	if orphanedCount != 2 {
		t.Errorf("expected 2 DriftOrphanedRelationship issues, got %d", orphanedCount)
	}

	if !result.HasErrors {
		t.Error("expected HasErrors to be true")
	}

	if result.ComponentsChecked != 2 {
		t.Errorf("expected ComponentsChecked to be 2, got %d", result.ComponentsChecked)
	}
}

// TestNewDetectDriftWithD2 tests that NewDetectDriftWithD2 wires the D2 parser.
func TestNewDetectDriftWithD2(t *testing.T) {
	mockRepo := &mockDriftProjectRepository{}
	mockParser := &mockDriftD2Parser{}
	uc := NewDetectDriftWithD2(mockRepo, mockParser)
	if uc == nil {
		t.Fatal("expected non-nil DetectDrift")
	}
	if uc.d2Parser == nil {
		t.Error("expected D2 parser to be set")
	}
}

// TestDetectDrift_WithRepo tests Execute() when Systems is nil (loads from repo).
func TestDetectDrift_WithRepo(t *testing.T) {
	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")
	comp, _ := entities.NewComponent("Auth")
	container.AddComponent(comp)
	system.AddContainer(container)

	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}

	uc := NewDetectDrift(mockRepo)
	// Don't provide Systems in req — force repo.ListSystems call
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/test",
		Systems:     nil,
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ComponentsChecked != 1 {
		t.Errorf("expected 1 component checked, got %d", result.ComponentsChecked)
	}
}

// TestDetectDrift_RepoError tests Execute() when repo.ListSystems fails.
func TestDetectDrift_RepoError(t *testing.T) {
	mockRepo := &mockDriftProjectRepository{
		err: fmt.Errorf("repo error"),
	}

	uc := NewDetectDrift(mockRepo)
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/test",
		Systems:     nil,
	}

	_, err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

// TestDetectDrift_WithD2Parser tests Execute() with D2 parser enabled but no .d2 files.
func TestDetectDrift_WithD2Parser(t *testing.T) {
	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")

	comp, _ := entities.NewComponent("Auth")
	comp.Path = t.TempDir() // existing dir but no .d2 files
	container.AddComponent(comp)
	system.AddContainer(container)

	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}
	mockParser := &mockDriftD2Parser{}

	uc := NewDetectDriftWithD2(mockRepo, mockParser)
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/test",
		Systems:     []*entities.System{system},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ComponentsChecked != 1 {
		t.Errorf("expected 1 component checked, got %d", result.ComponentsChecked)
	}
}

// TestDetectDrift_WithD2ParserAndD2File tests Execute() with D2 parser and a .d2 file present.
func TestDetectDrift_WithD2ParserAndD2File(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a minimal .d2 file in the temp dir
	d2Content := "auth -> db: queries\n"
	d2File := tmpDir + "/component.d2"
	if err := os.WriteFile(d2File, []byte(d2Content), 0644); err != nil {
		t.Fatalf("failed to create .d2 file: %v", err)
	}

	system, _ := entities.NewSystem("Backend")
	container, _ := entities.NewContainer("API")
	comp, _ := entities.NewComponent("Auth")
	comp.Path = tmpDir
	container.AddComponent(comp)
	system.AddContainer(container)

	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}
	rel, _ := entities.NewD2Relationship("auth", "db", "queries")
	mockParser := &mockDriftD2Parser{
		relationships: []entities.D2Relationship{*rel},
	}

	uc := NewDetectDriftWithD2(mockRepo, mockParser)
	req := &DetectDriftRequest{
		Systems: []*entities.System{system},
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ComponentsChecked != 1 {
		t.Errorf("expected 1 component checked, got %d", result.ComponentsChecked)
	}
	// checkMissingComponents is a no-op placeholder, so no issues expected from D2
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues (checkMissingComponents is a no-op), got %d", len(result.Issues))
	}
}

// mockDriftD2Parser is a test double for the D2Parser port.
type mockDriftD2Parser struct {
	relationships []entities.D2Relationship
	err           error
}

func (m *mockDriftD2Parser) ParseRelationships(_ context.Context, _ string) ([]entities.D2Relationship, error) {
	return m.relationships, m.err
}

// BenchmarkDetectDrift benchmarks the DetectDrift use case with varying numbers of components.
func BenchmarkDetectDrift(b *testing.B) {
	// Create test project with many components
	_, _ = entities.NewProject("benchmark-project")

	system, _ := entities.NewSystem("BenchmarkSystem")
	container, _ := entities.NewContainer("BenchmarkContainer")

	// Create 100 components with some orphaned relationships
	for i := 0; i < 100; i++ {
		comp, _ := entities.NewComponent(fmt.Sprintf("component-%d", i))
		if i%10 == 0 { // Every 10th component has an orphaned relationship
			comp.AddRelationship("non-existent-component", "calls")
		}
		container.AddComponent(comp)
	}
	system.AddContainer(container)

	// Create mock repository
	mockRepo := &mockDriftProjectRepository{
		systems: []*entities.System{system},
	}

	uc := NewDetectDrift(mockRepo)
	req := &DetectDriftRequest{
		ProjectRoot: "/tmp/benchmark",
		Systems:     []*entities.System{system},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := uc.Execute(context.Background(), req)
		if err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}

// mockDriftProjectRepository is a test double for the ProjectRepository port.
type mockDriftProjectRepository struct {
	systems []*entities.System
	err     error
}

func (m *mockDriftProjectRepository) LoadProject(_ context.Context, _ string) (*entities.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	project, _ := entities.NewProject("test-project")
	return project, nil
}

func (m *mockDriftProjectRepository) SaveProject(_ context.Context, _ *entities.Project) error {
	return m.err
}

func (m *mockDriftProjectRepository) ListSystems(_ context.Context, _ string) ([]*entities.System, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.systems, nil
}

func (m *mockDriftProjectRepository) LoadSystem(_ context.Context, _, _ string) (*entities.System, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.systems) > 0 {
		return m.systems[0], nil
	}
	return nil, nil
}

func (m *mockDriftProjectRepository) SaveSystem(_ context.Context, _ string, _ *entities.System) error {
	return m.err
}

func (m *mockDriftProjectRepository) LoadContainer(_ context.Context, _, _, _ string) (*entities.Container, error) {
	return nil, nil
}

func (m *mockDriftProjectRepository) SaveContainer(_ context.Context, _, _ string, _ *entities.Container) error {
	return m.err
}

func (m *mockDriftProjectRepository) LoadComponent(_ context.Context, _, _, _, _ string) (*entities.Component, error) {
	return nil, nil
}

func (m *mockDriftProjectRepository) SaveComponent(_ context.Context, _, _, _ string, _ *entities.Component) error {
	return m.err
}
