package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
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

// TestNewRenderDiagramPreview tests creating a RenderDiagramPreview use case.
func TestNewRenderDiagramPreview(t *testing.T) {
	mockRenderer := &mockDiagramRenderer{}

	uc := NewRenderDiagramPreview(mockRenderer)

	if uc == nil {
		t.Error("NewRenderDiagramPreview() returned nil")
	}

	if uc.renderer != mockRenderer {
		t.Error("NewRenderDiagramPreview() did not set renderer correctly")
	}
}

// TestRenderDiagramPreviewExecute tests the Execute method with valid input.
func TestRenderDiagramPreviewExecute(t *testing.T) {
	mockRenderer := &mockDiagramRenderer{}
	uc := NewRenderDiagramPreview(mockRenderer)

	// Test with valid D2 source
	req := &RenderDiagramPreviewRequest{
		D2Source: "x -> y: relation",
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if result == nil {
		t.Error("Execute() returned nil result")
	}

	if result.Preview != "<svg>test diagram</svg>" {
		t.Errorf("Execute() unexpected preview content: %s", result.Preview)
	}

	if result.Format != "svg" {
		t.Errorf("Execute() unexpected format: %s", result.Format)
	}
}

// TestRenderDiagramPreviewExecuteEmptySource tests with empty D2 source.
func TestRenderDiagramPreviewExecuteEmptySource(t *testing.T) {
	mockRenderer := &mockDiagramRenderer{}
	uc := NewRenderDiagramPreview(mockRenderer)

	// Test with empty D2 source
	req := &RenderDiagramPreviewRequest{
		D2Source: "",
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Errorf("Execute() error with empty source = %v", err)
	}

	if result.Preview != "" {
		t.Errorf("Execute() expected empty preview for empty source, got: %s", result.Preview)
	}

	if result.Format != "svg" {
		t.Errorf("Execute() unexpected format: %s", result.Format)
	}
}

// TestRenderDiagramPreviewExecuteRendererError tests error handling from renderer.
func TestRenderDiagramPreviewExecuteRendererError(t *testing.T) {
	expectedErr := entities.ErrInvalidD2
	mockRenderer := &mockDiagramRenderer{
		renderDiagramFunc: func(ctx context.Context, d2Source string) (string, error) {
			return "", expectedErr
		},
	}
	uc := NewRenderDiagramPreview(mockRenderer)

	req := &RenderDiagramPreviewRequest{
		D2Source: "invalid d2 source",
	}

	_, err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Error("Execute() expected error but got nil")
		return
	}

	// Check if the error is wrapped correctly
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() expected error %v, got %v", expectedErr, err)
	}
}

// TestRenderDiagramPreviewExecuteNilRenderer tests nil safety.
func TestRenderDiagramPreviewExecuteNilRenderer(t *testing.T) {
	uc := NewRenderDiagramPreview(nil)

	req := &RenderDiagramPreviewRequest{
		D2Source: "x -> y: relation",
	}

	_, err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Error("Execute() expected error with nil renderer")
	}
}
