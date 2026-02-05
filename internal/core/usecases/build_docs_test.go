package usecases

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// MockDiagramRenderer mocks the DiagramRenderer interface.
// Thread-safe for concurrent use in parallel rendering tests.
type MockDiagramRenderer struct {
	renderCount atomic.Int64
	err         error
}

func (m *MockDiagramRenderer) RenderDiagram(ctx context.Context, d2Source string) (string, error) {
	m.renderCount.Add(1)
	if m.err != nil {
		return "", m.err
	}
	return "<svg></svg>", nil
}

func (m *MockDiagramRenderer) RenderDiagramWithTimeout(ctx context.Context, d2Source string, timeoutSec int) (string, error) {
	return m.RenderDiagram(ctx, d2Source)
}

func (m *MockDiagramRenderer) IsAvailable() bool {
	return m.err == nil
}

// MockSiteBuilder mocks the SiteBuilder interface.
type MockSiteBuilder struct {
	buildCount    int
	lastProject   *entities.Project
	lastOutputDir string
	err           error
}

func (m *MockSiteBuilder) BuildSite(ctx context.Context, project *entities.Project, systems []*entities.System, outputDir string) error {
	m.buildCount++
	m.lastProject = project
	m.lastOutputDir = outputDir
	return m.err
}

func (m *MockSiteBuilder) BuildSystemPage(ctx context.Context, system *entities.System, containers []*entities.Container, outputDir string) error {
	return m.err
}

// MockProgressReporter mocks the ProgressReporter interface.
// Thread-safe for concurrent use in parallel rendering tests.
type MockProgressReporter struct {
	mu        sync.Mutex
	steps     []string
	errors    []error
	successes []string
	infos     []string
}

func (m *MockProgressReporter) ReportProgress(step string, current int, total int, message string) {
	m.mu.Lock()
	m.steps = append(m.steps, step)
	m.mu.Unlock()
}

func (m *MockProgressReporter) ReportError(err error) {
	m.mu.Lock()
	m.errors = append(m.errors, err)
	m.mu.Unlock()
}

func (m *MockProgressReporter) ReportSuccess(message string) {
	m.mu.Lock()
	m.successes = append(m.successes, message)
	m.mu.Unlock()
}

func (m *MockProgressReporter) ReportInfo(message string) {
	m.mu.Lock()
	m.infos = append(m.infos, message)
	m.mu.Unlock()
}

// MockProjectRepository mocks the ProjectRepository interface.
type MockBuildDocsProjectRepository struct {
	loadedProject *entities.Project
	savedProject  *entities.Project
	loadErr       error
	saveErr       error
}

func (m *MockBuildDocsProjectRepository) LoadProject(ctx context.Context, projectRoot string) (*entities.Project, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.loadedProject, nil
}

func (m *MockBuildDocsProjectRepository) SaveProject(ctx context.Context, project *entities.Project) error {
	m.savedProject = project
	return m.saveErr
}

func (m *MockBuildDocsProjectRepository) ListSystems(ctx context.Context, projectRoot string) ([]*entities.System, error) {
	if m.loadedProject == nil {
		return nil, entities.ErrProjectNotFound
	}
	systems := make([]*entities.System, 0)
	for _, sys := range m.loadedProject.Systems {
		systems = append(systems, sys)
	}
	return systems, nil
}

func (m *MockBuildDocsProjectRepository) LoadSystem(ctx context.Context, projectRoot, systemName string) (*entities.System, error) {
	if m.loadedProject == nil {
		return nil, entities.ErrProjectNotFound
	}
	if sys, ok := m.loadedProject.Systems[systemName]; ok {
		return sys, nil
	}
	return nil, entities.ErrSystemNotFound
}

func (m *MockBuildDocsProjectRepository) SaveSystem(ctx context.Context, projectRoot string, system *entities.System) error {
	return m.saveErr
}

func (m *MockBuildDocsProjectRepository) LoadContainer(ctx context.Context, projectRoot, systemName, containerName string) (*entities.Container, error) {
	if m.loadedProject == nil {
		return nil, entities.ErrProjectNotFound
	}
	sys, ok := m.loadedProject.Systems[systemName]
	if !ok {
		return nil, entities.ErrSystemNotFound
	}
	if container, ok := sys.Containers[containerName]; ok {
		return container, nil
	}
	return nil, entities.ErrContainerNotFound
}

func (m *MockBuildDocsProjectRepository) SaveContainer(ctx context.Context, projectRoot, systemName string, container *entities.Container) error {
	return m.saveErr
}

// TestBuildDocsUseCase tests the BuildDocs use case.
func TestBuildDocsUseCase(t *testing.T) {
	tests := []struct {
		name        string
		project     *entities.Project
		systems     []*entities.System
		wantErr     bool
		wantRenders int
	}{
		{
			name: "successful_single_system_build",
			project: &entities.Project{
				Name:        "test-project",
				Description: "A test project",
				Systems: map[string]*entities.System{
					"PaymentService": {
						ID:   "PaymentService",
						Name: "Payment Service",
						Containers: map[string]*entities.Container{
							"API": {
								ID:   "API",
								Name: "API Service",
							},
						},
						Diagram: &entities.Diagram{
							Source: "shape: rect",
						},
					},
				},
			},
			systems: []*entities.System{
				{
					ID:   "PaymentService",
					Name: "Payment Service",
					Containers: map[string]*entities.Container{
						"API": {
							ID:   "API",
							Name: "API Service",
						},
					},
					Diagram: &entities.Diagram{
						Source: "shape: rect",
					},
				},
			},
			wantErr:     false,
			wantRenders: 1,
		},
		{
			name: "multiple_systems_with_diagrams",
			project: &entities.Project{
				Name: "multi-system",
				Systems: map[string]*entities.System{
					"System1": {
						ID:   "System1",
						Name: "System 1",
						Containers: map[string]*entities.Container{
							"Container1": {
								ID:   "Container1",
								Name: "Container 1",
								Diagram: &entities.Diagram{
									Source: "shape: rect",
								},
							},
						},
						Diagram: &entities.Diagram{
							Source: "shape: rect",
						},
					},
					"System2": {
						ID:   "System2",
						Name: "System 2",
						Containers: map[string]*entities.Container{
							"Container2": {
								ID:   "Container2",
								Name: "Container 2",
								Diagram: &entities.Diagram{
									Source: "shape: rect",
								},
							},
						},
						Diagram: &entities.Diagram{
							Source: "shape: rect",
						},
					},
				},
			},
			systems: []*entities.System{
				{
					ID:   "System1",
					Name: "System 1",
					Containers: map[string]*entities.Container{
						"Container1": {
							ID:   "Container1",
							Name: "Container 1",
							Diagram: &entities.Diagram{
								Source: "shape: rect",
							},
						},
					},
					Diagram: &entities.Diagram{
						Source: "shape: rect",
					},
				},
				{
					ID:   "System2",
					Name: "System 2",
					Containers: map[string]*entities.Container{
						"Container2": {
							ID:   "Container2",
							Name: "Container 2",
							Diagram: &entities.Diagram{
								Source: "shape: rect",
							},
						},
					},
					Diagram: &entities.Diagram{
						Source: "shape: rect",
					},
				},
			},
			wantErr:     false,
			wantRenders: 4, // 2 systems + 2 containers
		},
		{
			name: "project_without_diagrams",
			project: &entities.Project{
				Name: "no-diagrams",
				Systems: map[string]*entities.System{
					"System1": {
						ID:   "System1",
						Name: "System 1",
						Containers: map[string]*entities.Container{
							"Container1": {
								ID:   "Container1",
								Name: "Container 1",
							},
						},
					},
				},
			},
			systems: []*entities.System{
				{
					ID:   "System1",
					Name: "System 1",
					Containers: map[string]*entities.Container{
						"Container1": {
							ID:   "Container1",
							Name: "Container 1",
						},
					},
				},
			},
			wantErr:     false,
			wantRenders: 0,
		},
		{
			name: "diagram_rendering_error",
			project: &entities.Project{
				Name: "bad-diagram",
				Systems: map[string]*entities.System{
					"System1": {
						ID:   "System1",
						Name: "System 1",
						Containers: map[string]*entities.Container{
							"Container1": {
								ID:   "Container1",
								Name: "Container 1",
							},
						},
						Diagram: &entities.Diagram{
							Source: "invalid d2",
						},
					},
				},
			},
			systems: []*entities.System{
				{
					ID:   "System1",
					Name: "System 1",
					Containers: map[string]*entities.Container{
						"Container1": {
							ID:   "Container1",
							Name: "Container 1",
						},
					},
					Diagram: &entities.Diagram{
						Source: "invalid d2",
					},
				},
			},
			wantErr:     true,
			wantRenders: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			mockRenderer := &MockDiagramRenderer{}
			mockSiteBuilder := &MockSiteBuilder{}
			mockProgress := &MockProgressReporter{}

			// Inject error for rendering error test
			if tt.wantErr && tt.wantRenders > 0 {
				mockRenderer.err = fmt.Errorf("%w", entities.ErrInvalidD2)
			}

			// Create BuildDocs use case with mocks
			uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)

			err := uc.Execute(ctx, tt.project, tt.systems, "/tmp/output")

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if int(mockRenderer.renderCount.Load()) != tt.wantRenders {
				t.Errorf("expected %d renders, got %d", tt.wantRenders, mockRenderer.renderCount.Load())
			}

			if !tt.wantErr && mockSiteBuilder.buildCount != 1 {
				t.Errorf("expected 1 site build, got %d", mockSiteBuilder.buildCount)
			}
		})
	}
}

// TestBuildDocsProgressReporting tests that progress is correctly reported.
func TestBuildDocsProgressReporting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project := &entities.Project{
		Name: "test-project",
		Systems: map[string]*entities.System{
			"System1": {
				ID:   "System1",
				Name: "System 1",
				Containers: map[string]*entities.Container{
					"Container1": {
						ID:   "Container1",
						Name: "Container 1",
						Diagram: &entities.Diagram{
							Source: "shape: rect",
						},
					},
				},
				Diagram: &entities.Diagram{
					Source: "shape: rect",
				},
			},
		},
	}

	systems := []*entities.System{
		{
			ID:   "System1",
			Name: "System 1",
			Containers: map[string]*entities.Container{
				"Container1": {
					ID:   "Container1",
					Name: "Container 1",
					Diagram: &entities.Diagram{
						Source: "shape: rect",
					},
				},
			},
			Diagram: &entities.Diagram{
				Source: "shape: rect",
			},
		},
	}

	mockRenderer := &MockDiagramRenderer{}
	mockSiteBuilder := &MockSiteBuilder{}
	mockProgress := &MockProgressReporter{}

	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)
	err := uc.Execute(ctx, project, systems, "/tmp/output")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify progress was reported
	if len(mockProgress.steps) == 0 {
		t.Error("expected progress steps to be reported")
	}

	if len(mockProgress.successes) == 0 {
		t.Error("expected success messages to be reported")
	}
}

// TestBuildDocsSiteBuilderCalled tests that the site builder is called with correct data.
func TestBuildDocsSiteBuilderCalled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project := &entities.Project{
		Name: "test-project",
		Systems: map[string]*entities.System{
			"PaymentService": {
				ID:   "PaymentService",
				Name: "Payment Service",
				Containers: map[string]*entities.Container{
					"API": {
						ID:   "API",
						Name: "API Service",
					},
				},
			},
		},
	}

	systems := []*entities.System{
		{
			ID:   "PaymentService",
			Name: "Payment Service",
			Containers: map[string]*entities.Container{
				"API": {
					ID:   "API",
					Name: "API Service",
				},
			},
		},
	}

	mockRenderer := &MockDiagramRenderer{}
	mockSiteBuilder := &MockSiteBuilder{}
	mockProgress := &MockProgressReporter{}

	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)
	outputDir := "/tmp/output"
	err := uc.Execute(ctx, project, systems, outputDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockSiteBuilder.buildCount != 1 {
		t.Errorf("expected site builder to be called once, got %d times", mockSiteBuilder.buildCount)
	}

	if mockSiteBuilder.lastProject.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %s", mockSiteBuilder.lastProject.Name)
	}

	if mockSiteBuilder.lastOutputDir != outputDir {
		t.Errorf("expected output dir %s, got %s", outputDir, mockSiteBuilder.lastOutputDir)
	}
}
