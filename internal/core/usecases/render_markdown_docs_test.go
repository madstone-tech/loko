package usecases

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// mockMarkdownRenderer is a test double for MarkdownRenderer.
type mockMarkdownRenderer struct {
	renderMarkdownToHTMLFunc func(markdown string) string
}

func (m *mockMarkdownRenderer) RenderMarkdownToHTML(markdown string) string {
	if m.renderMarkdownToHTMLFunc != nil {
		return m.renderMarkdownToHTMLFunc(markdown)
	}
	return "<p>rendered html</p>"
}

// mockProgressReporter is a test double for ProgressReporter.
type mockProgressReporter struct {
	reportProgressFunc func(step string, current int, total int, message string)
	reportErrorFunc    func(err error)
	reportSuccessFunc  func(message string)
	reportInfoFunc     func(message string)
}

func (m *mockProgressReporter) ReportProgress(step string, current int, total int, message string) {
	if m.reportProgressFunc != nil {
		m.reportProgressFunc(step, current, total, message)
	}
}

func (m *mockProgressReporter) ReportError(err error) {
	if m.reportErrorFunc != nil {
		m.reportErrorFunc(err)
	}
}

func (m *mockProgressReporter) ReportSuccess(message string) {
	if m.reportSuccessFunc != nil {
		m.reportSuccessFunc(message)
	}
}

func (m *mockProgressReporter) ReportInfo(message string) {
	if m.reportInfoFunc != nil {
		m.reportInfoFunc(message)
	}
}

// TestNewRenderMarkdownDocs tests creating a RenderMarkdownDocs use case.
func TestNewRenderMarkdownDocs(t *testing.T) {
	mockRenderer := &mockMarkdownRenderer{}
	mockProgressReporter := &mockProgressReporter{}

	uc := NewRenderMarkdownDocs(mockRenderer, mockProgressReporter)

	if uc == nil {
		t.Error("NewRenderMarkdownDocs() returned nil")
	}

	if uc.markdownRenderer != mockRenderer {
		t.Error("NewRenderMarkdownDocs() did not set markdownRenderer correctly")
	}

	if uc.progressReporter != mockProgressReporter {
		t.Error("NewRenderMarkdownDocs() did not set progressReporter correctly")
	}
}

// TestRenderMarkdownDocsExecute tests the Execute method.
func TestRenderMarkdownDocsExecute(t *testing.T) {
	project, _ := entities.NewProject("test-project")
	system, _ := entities.NewSystem("Test System")
	container, _ := entities.NewContainer("Test Container")
	component, _ := entities.NewComponent("Test Component")

	container.AddComponent(component)
	system.AddContainer(container)

	mockRenderer := &mockMarkdownRenderer{}
	mockProgressReporter := &mockProgressReporter{}

	uc := NewRenderMarkdownDocs(mockRenderer, mockProgressReporter)

	// Test with valid inputs
	err := uc.Execute(context.Background(), project, []*entities.System{system}, t.TempDir())
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Test with nil project
	err = uc.Execute(context.Background(), nil, []*entities.System{system}, t.TempDir())
	if err == nil {
		t.Error("Execute() expected error for nil project")
	}

	// Test with empty systems
	err = uc.Execute(context.Background(), project, []*entities.System{}, t.TempDir())
	if err != nil {
		t.Errorf("Execute() error with empty systems = %v", err)
	}
}

// TestRenderSystemMarkdown tests the renderSystemMarkdown method.
func TestRenderSystemMarkdown(t *testing.T) {
	system, _ := entities.NewSystem("Test System")
	system.Path = t.TempDir()

	mockRenderer := &mockMarkdownRenderer{}
	mockRenderer.renderMarkdownToHTMLFunc = func(markdown string) string {
		return "<h1>Test System</h1>"
	}

	mockProgressReporter := &mockProgressReporter{}
	uc := NewRenderMarkdownDocs(mockRenderer, mockProgressReporter)

	// Test with non-existent markdown file (should not error)
	err := uc.renderSystemMarkdown(context.Background(), system, t.TempDir())
	if err != nil {
		t.Errorf("renderSystemMarkdown() error = %v", err)
	}

	// Note: Testing with actual files would require more complex setup
	// For now, we're just verifying the method doesn't panic
}

// TestRenderContainerMarkdown tests the renderContainerMarkdown method.
func TestRenderContainerMarkdown(t *testing.T) {
	system, _ := entities.NewSystem("Test System")
	container, _ := entities.NewContainer("Test Container")
	container.Path = t.TempDir()

	mockRenderer := &mockMarkdownRenderer{}
	mockProgressReporter := &mockProgressReporter{}
	uc := NewRenderMarkdownDocs(mockRenderer, mockProgressReporter)

	// Test with non-existent markdown file (should not error)
	err := uc.renderContainerMarkdown(context.Background(), system, container, t.TempDir())
	if err != nil {
		t.Errorf("renderContainerMarkdown() error = %v", err)
	}
}

// TestRenderComponentMarkdown tests the renderComponentMarkdown method.
func TestRenderComponentMarkdown(t *testing.T) {
	system, _ := entities.NewSystem("Test System")
	container, _ := entities.NewContainer("Test Container")
	component, _ := entities.NewComponent("Test Component")
	component.Path = t.TempDir()

	mockRenderer := &mockMarkdownRenderer{}
	mockProgressReporter := &mockProgressReporter{}
	uc := NewRenderMarkdownDocs(mockRenderer, mockProgressReporter)

	// Test with non-existent markdown file (should not error)
	err := uc.renderComponentMarkdown(context.Background(), system, container, component, t.TempDir())
	if err != nil {
		t.Errorf("renderComponentMarkdown() error = %v", err)
	}
}
