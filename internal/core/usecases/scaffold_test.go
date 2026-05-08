package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// mockTemplateEngine is a test double for TemplateEngine.
type mockTemplateEngine struct {
	renderTemplateFunc func(ctx context.Context, templateName string, variables map[string]string) (string, error)
	listTemplatesFunc  func(ctx context.Context) ([]string, error)
	addSearchPathFunc  func(path string)
}

func (m *mockTemplateEngine) RenderTemplate(ctx context.Context, templateName string, variables map[string]string) (string, error) {
	if m.renderTemplateFunc != nil {
		return m.renderTemplateFunc(ctx, templateName, variables)
	}
	return "rendered content", nil
}

func (m *mockTemplateEngine) ListTemplates(ctx context.Context) ([]string, error) {
	if m.listTemplatesFunc != nil {
		return m.listTemplatesFunc(ctx)
	}
	return []string{"default"}, nil
}

func (m *mockTemplateEngine) AddSearchPath(path string) {
	if m.addSearchPathFunc != nil {
		m.addSearchPathFunc(path)
	}
}

// mockDiagramGenerator is a test double for DiagramGenerator.
type mockDiagramGenerator struct {
	generateSystemContextDiagramFunc func(system *entities.System) (string, error)
	generateContainerDiagramFunc     func(system *entities.System) (string, error)
	generateComponentDiagramFunc     func(container *entities.Container) (string, error)
}

func (m *mockDiagramGenerator) GenerateSystemContextDiagram(system *entities.System) (string, error) {
	if m.generateSystemContextDiagramFunc != nil {
		return m.generateSystemContextDiagramFunc(system)
	}
	return "system context diagram", nil
}

func (m *mockDiagramGenerator) GenerateContainerDiagram(system *entities.System) (string, error) {
	if m.generateContainerDiagramFunc != nil {
		return m.generateContainerDiagramFunc(system)
	}
	return "container diagram", nil
}

func (m *mockDiagramGenerator) GenerateComponentDiagram(container *entities.Container) (string, error) {
	if m.generateComponentDiagramFunc != nil {
		return m.generateComponentDiagramFunc(container)
	}
	return "component diagram", nil
}

// mockLogger is a test double for Logger.
type mockLogger struct {
	debugFunc       func(msg string, keysAndValues ...any)
	infoFunc        func(msg string, keysAndValues ...any)
	warnFunc        func(msg string, keysAndValues ...any)
	errorFunc       func(msg string, err error, keysAndValues ...any)
	withContextFunc func(ctx context.Context) Logger
	withFieldsFunc  func(keysAndValues ...any) Logger
}

func (m *mockLogger) Debug(msg string, keysAndValues ...any) {
	if m.debugFunc != nil {
		m.debugFunc(msg, keysAndValues...)
	}
}

func (m *mockLogger) Info(msg string, keysAndValues ...any) {
	if m.infoFunc != nil {
		m.infoFunc(msg, keysAndValues...)
	}
}

func (m *mockLogger) Warn(msg string, keysAndValues ...any) {
	if m.warnFunc != nil {
		m.warnFunc(msg, keysAndValues...)
	}
}

func (m *mockLogger) Error(msg string, err error, keysAndValues ...any) {
	if m.errorFunc != nil {
		m.errorFunc(msg, err, keysAndValues...)
	}
}

func (m *mockLogger) WithContext(ctx context.Context) Logger {
	if m.withContextFunc != nil {
		return m.withContextFunc(ctx)
	}
	return m
}

func (m *mockLogger) WithFields(keysAndValues ...any) Logger {
	if m.withFieldsFunc != nil {
		return m.withFieldsFunc(keysAndValues...)
	}
	return m
}

// TestNewScaffoldEntity tests creating a ScaffoldEntity use case.
func TestNewScaffoldEntity(t *testing.T) {
	mockRepo := &MockProjectRepository{}

	uc := NewScaffoldEntity(mockRepo)

	if uc == nil {
		t.Error("NewScaffoldEntity() returned nil")
	}

	if uc.projectRepo != mockRepo {
		t.Error("NewScaffoldEntity() did not set projectRepo correctly")
	}

	if uc.templateEngine != nil {
		t.Error("NewScaffoldEntity() should not set templateEngine by default")
	}

	if uc.diagramGenerator != nil {
		t.Error("NewScaffoldEntity() should not set diagramGenerator by default")
	}

	if uc.logger != nil {
		t.Error("NewScaffoldEntity() should not set logger by default")
	}
}

// TestNewScaffoldEntityWithOptions tests creating a ScaffoldEntity with options.
func TestNewScaffoldEntityWithOptions(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	mockTE := &mockTemplateEngine{}
	mockDG := &mockDiagramGenerator{}
	mockL := &mockLogger{}

	uc := NewScaffoldEntity(mockRepo,
		WithTemplateEngine(mockTE),
		WithDiagramGenerator(mockDG),
		WithLogger(mockL))

	if uc == nil {
		t.Error("NewScaffoldEntity() returned nil")
	}

	if uc.projectRepo != mockRepo {
		t.Error("NewScaffoldEntity() did not set projectRepo correctly")
	}

	if uc.templateEngine != mockTE {
		t.Error("NewScaffoldEntity() did not set templateEngine correctly")
	}

	if uc.diagramGenerator != mockDG {
		t.Error("NewScaffoldEntity() did not set diagramGenerator correctly")
	}

	if uc.logger != mockL {
		t.Error("NewScaffoldEntity() did not set logger correctly")
	}
}

// TestScaffoldEntityWithLogger tests scaffolding with logging.
func TestScaffoldEntityWithLogger(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}
	mockRepo.SaveSystemFunc = func(ctx context.Context, projectRoot string, system *entities.System) error {
		return nil
	}

	mockL := &mockLogger{}
	infoCalled := false
	mockL.infoFunc = func(msg string, keysAndValues ...any) {
		infoCalled = true
	}

	uc := NewScaffoldEntity(mockRepo, WithLogger(mockL))

	req := &ScaffoldEntityRequest{
		ProjectRoot: "/test/project",
		EntityType:  "system",
		Name:        "Test System",
	}

	_, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !infoCalled {
		t.Error("expected logger Info method to be called")
	}
}

// TestScaffoldEntityNilRequest tests error handling for nil request.
func TestScaffoldEntityNilRequest(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	uc := NewScaffoldEntity(mockRepo)

	_, err := uc.Execute(context.Background(), nil)
	if err == nil {
		t.Error("Execute() expected error for nil request")
	}
}

// TestScaffoldEntityWithTemplate tests scaffolding with template rendering.
func TestScaffoldEntityWithTemplate(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}
	mockRepo.SaveSystemFunc = func(ctx context.Context, projectRoot string, system *entities.System) error {
		// Set a temporary path to avoid file system issues
		system.Path = t.TempDir()
		return nil
	}

	mockTE := &mockTemplateEngine{}
	renderCalled := false
	mockTE.renderTemplateFunc = func(ctx context.Context, templateName string, variables map[string]string) (string, error) {
		renderCalled = true
		return "rendered template content", nil
	}

	uc := NewScaffoldEntity(mockRepo, WithTemplateEngine(mockTE))

	req := &ScaffoldEntityRequest{
		ProjectRoot: t.TempDir(),
		EntityType:  "system",
		Name:        "Test System",
		Template:    "test-template",
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !renderCalled {
		t.Error("expected template engine RenderTemplate to be called")
	}

	if len(result.FilesCreated) < 1 { // Should have at least system.toml
		t.Errorf("expected at least 1 file created, got %d", len(result.FilesCreated))
	}
}

// TestScaffoldEntityWithDiagramGenerator tests scaffolding with diagram generation.
func TestScaffoldEntityWithDiagramGenerator(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}
	mockRepo.SaveSystemFunc = func(ctx context.Context, projectRoot string, system *entities.System) error {
		// Set a temporary path to avoid file system issues
		system.Path = t.TempDir()
		return nil
	}

	mockDG := &mockDiagramGenerator{}
	diagramCalled := false
	mockDG.generateSystemContextDiagramFunc = func(system *entities.System) (string, error) {
		diagramCalled = true
		return "generated diagram content", nil
	}

	uc := NewScaffoldEntity(mockRepo, WithDiagramGenerator(mockDG))

	req := &ScaffoldEntityRequest{
		ProjectRoot: t.TempDir(),
		EntityType:  "system",
		Name:        "Test System",
	}

	_, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !diagramCalled {
		t.Error("expected diagram generator to be called")
	}
}

// TestScaffoldEntityInvalidEntityType tests error handling for invalid entity types.
func TestScaffoldEntityInvalidEntityType(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	mockRepo := &MockProjectRepository{}
	mockRepo.LoadProjectFunc = func(ctx context.Context, projectRoot string) (*entities.Project, error) {
		return project, nil
	}

	uc := NewScaffoldEntity(mockRepo)

	req := &ScaffoldEntityRequest{
		ProjectRoot: "/test/project",
		EntityType:  "invalid-type",
		Name:        "Test Entity",
	}

	_, err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Error("Execute() expected error for invalid entity type")
	}
}
