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

// MockMarkdownBuilder mocks the MarkdownBuilder interface.
type MockMarkdownBuilder struct {
	buildMarkdownFunc       func(ctx context.Context, project *entities.Project, systems []*entities.System) (string, error)
	buildSystemMarkdownFunc func(ctx context.Context, system *entities.System, containers []*entities.Container) (string, error)
}

func (m *MockMarkdownBuilder) BuildMarkdown(ctx context.Context, project *entities.Project, systems []*entities.System) (string, error) {
	if m.buildMarkdownFunc != nil {
		return m.buildMarkdownFunc(ctx, project, systems)
	}
	return "# Test Markdown", nil
}

func (m *MockMarkdownBuilder) BuildSystemMarkdown(ctx context.Context, system *entities.System, containers []*entities.Container) (string, error) {
	if m.buildSystemMarkdownFunc != nil {
		return m.buildSystemMarkdownFunc(ctx, system, containers)
	}
	return "# Test System Markdown", nil
}

// MockPDFRenderer mocks the PDFRenderer interface.
type MockPDFRenderer struct {
	renderPDFFunc   func(ctx context.Context, htmlPath string, outputPath string) error
	isAvailableFunc func() bool
}

func (m *MockPDFRenderer) RenderPDF(ctx context.Context, htmlPath string, outputPath string) error {
	if m.renderPDFFunc != nil {
		return m.renderPDFFunc(ctx, htmlPath, outputPath)
	}
	return nil
}

func (m *MockPDFRenderer) IsAvailable() bool {
	if m.isAvailableFunc != nil {
		return m.isAvailableFunc()
	}
	return true
}

// MockOutputEncoder mocks the OutputEncoder interface.
type MockOutputEncoder struct {
	encodeJSONFunc func(value any) ([]byte, error)
	encodeTOONFunc func(value any) ([]byte, error)
	decodeJSONFunc func(data []byte, value any) error
	decodeTOONFunc func(data []byte, value any) error
}

func (m *MockOutputEncoder) EncodeJSON(value any) ([]byte, error) {
	if m.encodeJSONFunc != nil {
		return m.encodeJSONFunc(value)
	}
	return []byte("{}"), nil
}

func (m *MockOutputEncoder) EncodeTOON(value any) ([]byte, error) {
	if m.encodeTOONFunc != nil {
		return m.encodeTOONFunc(value)
	}
	return []byte(""), nil
}

func (m *MockOutputEncoder) DecodeJSON(data []byte, value any) error {
	if m.decodeJSONFunc != nil {
		return m.decodeJSONFunc(data, value)
	}
	return nil
}

func (m *MockOutputEncoder) DecodeTOON(data []byte, value any) error {
	if m.decodeTOONFunc != nil {
		return m.decodeTOONFunc(data, value)
	}
	return nil
}

// TestDefaultBuildDocsOptions tests the DefaultBuildDocsOptions function.
func TestDefaultBuildDocsOptions(t *testing.T) {
	opts := DefaultBuildDocsOptions()

	if len(opts.Formats) != 1 {
		t.Errorf("Expected 1 format, got %d", len(opts.Formats))
	}

	if opts.Formats[0] != FormatHTML {
		t.Errorf("Expected FormatHTML, got %v", opts.Formats[0])
	}
}

// TestBuildDocsWithOptions tests the With* methods for setting optional builders.
func TestBuildDocsWithOptions(t *testing.T) {
	mockRenderer := &MockDiagramRenderer{}
	mockSiteBuilder := &MockSiteBuilder{}
	mockProgress := &MockProgressReporter{}
	mockMarkdownBuilder := &MockMarkdownBuilder{}
	mockPDFRenderer := &MockPDFRenderer{}
	mockOutputEncoder := &MockOutputEncoder{}

	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)
	uc = uc.WithMarkdownBuilder(mockMarkdownBuilder)
	uc = uc.WithPDFRenderer(mockPDFRenderer)
	uc = uc.WithOutputEncoder(mockOutputEncoder)

	// Verify the builders were set correctly by checking they're not nil
	// We can't directly access the fields, but we can test that the methods work
	// by calling ExecuteWithFormats with the corresponding formats
}

// TestBuildDocsExecuteWithFormats tests the ExecuteWithFormats method.
func TestBuildDocsExecuteWithFormats(t *testing.T) {
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
	mockMarkdownBuilder := &MockMarkdownBuilder{}
	mockPDFRenderer := &MockPDFRenderer{}
	mockOutputEncoder := &MockOutputEncoder{}

	// Test HTML format (default)
	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)
	opts := BuildDocsOptions{}
	err := uc.ExecuteWithFormats(ctx, project, systems, t.TempDir(), opts)
	if err != nil {
		t.Errorf("ExecuteWithFormats() HTML error = %v", err)
	}

	// Test with explicit HTML format
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatHTML},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, t.TempDir(), opts)
	if err != nil {
		t.Errorf("ExecuteWithFormats() explicit HTML error = %v", err)
	}

	// Test with Markdown format
	uc = uc.WithMarkdownBuilder(mockMarkdownBuilder)
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatMarkdown},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, t.TempDir(), opts)
	if err != nil {
		t.Errorf("ExecuteWithFormats() Markdown error = %v", err)
	}

	// Test with PDF format
	uc = uc.WithPDFRenderer(mockPDFRenderer)
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatPDF},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, t.TempDir(), opts)
	if err != nil {
		t.Errorf("ExecuteWithFormats() PDF error = %v", err)
	}

	// Test with TOON format
	uc = uc.WithOutputEncoder(mockOutputEncoder)
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatTOON},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, t.TempDir(), opts)
	if err != nil {
		t.Errorf("ExecuteWithFormats() TOON error = %v", err)
	}

	// Test with multiple formats
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatHTML, FormatMarkdown, FormatTOON},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, t.TempDir(), opts)
	if err != nil {
		t.Errorf("ExecuteWithFormats() multiple formats error = %v", err)
	}

	// Test with nil project
	err = uc.ExecuteWithFormats(ctx, nil, systems, t.TempDir(), opts)
	if err == nil {
		t.Error("ExecuteWithFormats() expected error for nil project")
	}
}

// TestBuildDocsExecuteWithFormatsErrors tests error cases for ExecuteWithFormats.
func TestBuildDocsExecuteWithFormatsErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project := &entities.Project{
		Name: "test-project",
		Systems: map[string]*entities.System{
			"PaymentService": {
				ID:   "PaymentService",
				Name: "Payment Service",
			},
		},
	}

	systems := []*entities.System{
		{
			ID:   "PaymentService",
			Name: "Payment Service",
		},
	}

	mockRenderer := &MockDiagramRenderer{}
	mockSiteBuilder := &MockSiteBuilder{}
	mockProgress := &MockProgressReporter{}

	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)
	tempDir := t.TempDir()

	// Test Markdown without builder
	opts := BuildDocsOptions{
		Formats: []OutputFormat{FormatMarkdown},
	}
	err := uc.ExecuteWithFormats(ctx, project, systems, tempDir, opts)
	if err == nil {
		t.Error("ExecuteWithFormats() expected error for Markdown without builder")
	}

	// Test PDF without renderer
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatPDF},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, tempDir, opts)
	if err == nil {
		t.Error("ExecuteWithFormats() expected error for PDF without renderer")
	}

	// Test TOON without encoder
	opts = BuildDocsOptions{
		Formats: []OutputFormat{FormatTOON},
	}
	err = uc.ExecuteWithFormats(ctx, project, systems, tempDir, opts)
	if err == nil {
		t.Error("ExecuteWithFormats() expected error for TOON without encoder")
	}
}

// TestBuildDocsWithDiagramErrors tests error handling in diagram rendering.
func TestBuildDocsWithDiagramErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a project with a system that has a diagram
	project := &entities.Project{
		Name: "test-project",
		Systems: map[string]*entities.System{
			"PaymentService": {
				ID:   "PaymentService",
				Name: "Payment Service",
				Diagram: &entities.Diagram{
					Source: "shape: rect",
				},
			},
		},
	}

	systems := []*entities.System{
		{
			ID:   "PaymentService",
			Name: "Payment Service",
			Diagram: &entities.Diagram{
				Source: "shape: rect",
			},
		},
	}

	// Test with renderer that returns an error
	mockRenderer := &MockDiagramRenderer{}
	mockRenderer.err = fmt.Errorf("diagram rendering failed")
	mockSiteBuilder := &MockSiteBuilder{}
	mockProgress := &MockProgressReporter{}

	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)
	err := uc.Execute(ctx, project, systems, t.TempDir())
	if err == nil {
		t.Error("Execute() expected error when diagram rendering fails")
	}
}

// TestBuildDocsWithWorkerPool tests the parallel worker pool functionality.
func TestBuildDocsWithWorkerPool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a project with multiple systems/containers/components that have diagrams
	project := &entities.Project{
		Name: "test-project",
		Systems: map[string]*entities.System{
			"System1": {
				ID:   "System1",
				Name: "System 1",
				Diagram: &entities.Diagram{
					Source: "shape: rect",
				},
				Containers: map[string]*entities.Container{
					"Container1": {
						ID:   "Container1",
						Name: "Container 1",
						Diagram: &entities.Diagram{
							Source: "shape: rect",
						},
					},
				},
			},
			"System2": {
				ID:   "System2",
				Name: "System 2",
				Diagram: &entities.Diagram{
					Source: "shape: rect",
				},
				Containers: map[string]*entities.Container{
					"Container2": {
						ID:   "Container2",
						Name: "Container 2",
						Diagram: &entities.Diagram{
							Source: "shape: rect",
						},
					},
				},
			},
		},
	}

	systems := []*entities.System{
		{
			ID:   "System1",
			Name: "System 1",
			Diagram: &entities.Diagram{
				Source: "shape: rect",
			},
			Containers: map[string]*entities.Container{
				"Container1": {
					ID:   "Container1",
					Name: "Container 1",
					Diagram: &entities.Diagram{
						Source: "shape: rect",
					},
				},
			},
		},
		{
			ID:   "System2",
			Name: "System 2",
			Diagram: &entities.Diagram{
				Source: "shape: rect",
			},
			Containers: map[string]*entities.Container{
				"Container2": {
					ID:   "Container2",
					Name: "Container 2",
					Diagram: &entities.Diagram{
						Source: "shape: rect",
					},
				},
			},
		},
	}

	mockRenderer := &MockDiagramRenderer{}
	mockSiteBuilder := &MockSiteBuilder{}
	mockProgress := &MockProgressReporter{}

	uc := NewBuildDocs(mockRenderer, mockSiteBuilder, mockProgress)

	// This should trigger the worker pool since we have multiple diagrams
	err := uc.Execute(ctx, project, systems, t.TempDir())
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify that multiple diagrams were rendered
	if mockRenderer.renderCount.Load() != 4 { // 2 system diagrams + 2 container diagrams
		t.Errorf("Expected 4 diagram renders, got %d", mockRenderer.renderCount.Load())
	}
}

// TestContainsFormat tests the containsFormat helper function.
func TestContainsFormat(t *testing.T) {
	formats := []OutputFormat{FormatHTML, FormatMarkdown}

	if !containsFormat(formats, FormatHTML) {
		t.Error("Expected to find FormatHTML in formats")
	}

	if !containsFormat(formats, FormatMarkdown) {
		t.Error("Expected to find FormatMarkdown in formats")
	}

	if containsFormat(formats, FormatPDF) {
		t.Error("Did not expect to find FormatPDF in formats")
	}

	// Test with empty slice
	if containsFormat([]OutputFormat{}, FormatHTML) {
		t.Error("Did not expect to find FormatHTML in empty formats")
	}
}

// TestGenerateComponentTable_Empty tests GenerateComponentTable with an empty container.
func TestGenerateComponentTable_Empty(t *testing.T) {
	container := &entities.Container{
		ID:         "test-container",
		Name:       "Test Container",
		Components: make(map[string]*entities.Component),
	}

	result := GenerateComponentTable(container)

	// Should return empty string for empty container
	if result != "" {
		t.Errorf("Expected empty string for empty container, got: %q", result)
	}
}

// TestGenerateComponentTable_Single tests GenerateComponentTable with a single component.
func TestGenerateComponentTable_Single(t *testing.T) {
	container := &entities.Container{
		ID:         "test-container",
		Name:       "Test Container",
		Components: make(map[string]*entities.Component),
	}

	component, _ := entities.NewComponent("Auth Service")
	component.Technology = "Go + Gin"
	component.Description = "Handles authentication and authorization"
	container.AddComponent(component)

	result := GenerateComponentTable(container)

	expected := "| Name | Technology | Description |\n|------|------------|-------------|\n| Auth Service | Go + Gin | Handles authentication and authorization |\n"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestGenerateComponentTable_Multiple tests GenerateComponentTable with multiple components.
func TestGenerateComponentTable_Multiple(t *testing.T) {
	container := &entities.Container{
		ID:         "test-container",
		Name:       "Test Container",
		Components: make(map[string]*entities.Component),
	}

	// Add components in random order to test sorting
	component1, _ := entities.NewComponent("Z Service")
	component1.Technology = "Node.js"
	component1.Description = "Handles Z operations"
	container.AddComponent(component1)

	component2, _ := entities.NewComponent("A Service")
	component2.Technology = "Python"
	component2.Description = "Handles A operations"
	container.AddComponent(component2)

	component3, _ := entities.NewComponent("M Service")
	component3.Technology = "Java"
	component3.Description = "Handles M operations"
	container.AddComponent(component3)

	result := GenerateComponentTable(container)

	expected := `| Name | Technology | Description |
|------|------------|-------------|
| A Service | Python | Handles A operations |
| M Service | Java | Handles M operations |
| Z Service | Node.js | Handles Z operations |
`
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestGenerateComponentTable_NilComponents tests GenerateComponentTable with nil components map.
func TestGenerateComponentTable_NilComponents(t *testing.T) {
	container := &entities.Container{
		ID:         "test-container",
		Name:       "Test Container",
		Components: nil,
	}

	result := GenerateComponentTable(container)

	// Should return empty string for nil components
	if result != "" {
		t.Errorf("Expected empty string for nil components, got: %q", result)
	}
}
