package d2

import (
	"context"
	"testing"
)

// mockDiagramRenderer is a test double for DiagramRenderer.
type mockDiagramRenderer struct {
	renderDiagramFunc func(ctx context.Context, d2Source string) (string, error)
	isAvailableFunc   func() bool
}

func (m *mockDiagramRenderer) RenderDiagram(ctx context.Context, d2Source string) (string, error) {
	if m.renderDiagramFunc != nil {
		return m.renderDiagramFunc(ctx, d2Source)
	}
	return "<svg>test diagram</svg>", nil
}

func (m *mockDiagramRenderer) RenderDiagramWithTimeout(ctx context.Context, d2Source string, timeoutSec int) (string, error) {
	return m.RenderDiagram(ctx, d2Source)
}

func (m *mockDiagramRenderer) IsAvailable() bool {
	if m.isAvailableFunc != nil {
		return m.isAvailableFunc()
	}
	return true
}

// TestNewPreviewRenderer tests creating a PreviewRenderer.
func TestNewPreviewRenderer(t *testing.T) {
	mockRenderer := &mockDiagramRenderer{}
	pr := NewPreviewRenderer(mockRenderer)

	if pr == nil {
		t.Error("NewPreviewRenderer() returned nil")
	}

	if pr.renderer != mockRenderer {
		t.Error("NewPreviewRenderer() did not set renderer correctly")
	}
}

// TestGenerateComponentPreviewD2 tests the D2 source generation.
func TestGenerateComponentPreviewD2(t *testing.T) {
	// Test with container name
	d2Source := GenerateComponentPreviewD2("Auth Service", "Go", "Backend")
	if d2Source == "" {
		t.Error("GenerateComponentPreviewD2() returned empty string")
	}

	// Test without container name
	d2Source = GenerateComponentPreviewD2("Database", "PostgreSQL", "")
	if d2Source == "" {
		t.Error("GenerateComponentPreviewD2() returned empty string for component without container")
	}
}

// TestRenderComponentPreview tests rendering a component preview.
func TestRenderComponentPreview(t *testing.T) {
	mockRenderer := &mockDiagramRenderer{}
	pr := NewPreviewRenderer(mockRenderer)

	svg, err := pr.RenderComponentPreview(context.Background(), "Auth Service", "Go", "Backend")
	if err != nil {
		t.Errorf("RenderComponentPreview() error = %v", err)
	}

	if svg != "<svg>test diagram</svg>" {
		t.Errorf("RenderComponentPreview() unexpected SVG content: %s", svg)
	}
}

// TestRenderComponentPreviewNilRenderer tests nil safety.
func TestRenderComponentPreviewNilRenderer(t *testing.T) {
	pr := NewPreviewRenderer(nil)

	_, err := pr.RenderComponentPreview(context.Background(), "Auth Service", "Go", "Backend")
	if err == nil {
		t.Error("RenderComponentPreview() expected error with nil renderer")
	}
}

// TestRenderComponentPreviewUnavailableRenderer tests unavailable renderer.
func TestRenderComponentPreviewUnavailableRenderer(t *testing.T) {
	mockRenderer := &mockDiagramRenderer{
		isAvailableFunc: func() bool {
			return false
		},
	}
	pr := NewPreviewRenderer(mockRenderer)

	_, err := pr.RenderComponentPreview(context.Background(), "Auth Service", "Go", "Backend")
	if err == nil {
		t.Error("RenderComponentPreview() expected error with unavailable renderer")
	}
}
