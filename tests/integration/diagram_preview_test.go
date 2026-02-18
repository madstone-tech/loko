package integration

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenderDiagramPreviewUseCase tests the RenderDiagramPreview use case.
func TestRenderDiagramPreviewUseCase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a mock renderer that always returns a fixed SVG
	mockRenderer := &mockDiagramRenderer{
		svgContent: `<svg><rect x="10" y="10" width="100" height="50" fill="blue"/></svg>`,
	}

	// Create the use case
	uc := usecases.NewRenderDiagramPreview(mockRenderer)

	// Test with valid D2 source
	req := &usecases.RenderDiagramPreviewRequest{
		D2Source: "x -> y: relation",
	}

	result, err := uc.Execute(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockRenderer.svgContent, result.Preview)
	assert.Equal(t, "svg", result.Format)

	// Test with empty source
	reqEmpty := &usecases.RenderDiagramPreviewRequest{
		D2Source: "",
	}

	resultEmpty, err := uc.Execute(context.Background(), reqEmpty)
	require.NoError(t, err)
	assert.NotNil(t, resultEmpty)
	assert.Equal(t, "", resultEmpty.Preview)
	assert.Equal(t, "svg", resultEmpty.Format)
}

// TestPreviewRendererAdapter tests the PreviewRenderer adapter.
func TestPreviewRendererAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a mock renderer
	mockRenderer := &mockDiagramRenderer{
		svgContent: `<svg><circle cx="50" cy="50" r="40" stroke="green" stroke-width="4" fill="yellow"/></svg>`,
		available:  true,
	}

	// Create preview renderer
	previewRenderer := d2.NewPreviewRenderer(mockRenderer)

	// Test generating D2 source
	d2Source := d2.GenerateComponentPreviewD2("Auth Service", "Go", "Backend")
	assert.NotEmpty(t, d2Source)
	assert.Contains(t, d2Source, "Auth Service")
	assert.Contains(t, d2Source, "Go")
	assert.Contains(t, d2Source, "Backend")

	// Test rendering preview
	svgContent, err := previewRenderer.RenderComponentPreview(context.Background(), "Auth Service", "Go", "Backend")
	require.NoError(t, err)
	assert.Equal(t, mockRenderer.svgContent, svgContent)
}

// mockDiagramRenderer implements DiagramRenderer for testing.
type mockDiagramRenderer struct {
	svgContent string
	available  bool
}

func (m *mockDiagramRenderer) RenderDiagram(_ context.Context, _ string) (string, error) {
	return m.svgContent, nil
}

func (m *mockDiagramRenderer) RenderDiagramWithTimeout(_ context.Context, d2Source string, _ int) (string, error) {
	return m.RenderDiagram(context.Background(), d2Source)
}

func (m *mockDiagramRenderer) IsAvailable() bool {
	return m.available
}
