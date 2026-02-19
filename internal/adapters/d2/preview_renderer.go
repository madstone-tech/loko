// Package d2 provides D2 diagram rendering and parsing adapters.
package d2

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// PreviewRenderer wraps the DiagramRenderer for terminal preview of D2 diagrams.
// It generates a simple component D2 snippet from component details and renders it.
type PreviewRenderer struct {
	renderer usecases.DiagramRenderer
}

// NewPreviewRenderer creates a new PreviewRenderer with the given DiagramRenderer.
func NewPreviewRenderer(renderer usecases.DiagramRenderer) *PreviewRenderer {
	return &PreviewRenderer{
		renderer: renderer,
	}
}

// GenerateComponentPreviewD2 creates a minimal D2 snippet for a component preview.
// It follows C4 component diagram conventions with basic styling.
func GenerateComponentPreviewD2(componentName, technology, containerName string) string {
	// Create a simple D2 diagram showing the component in its container context
	d2 := fmt.Sprintf(`direction: right

%s: %s {
  technology: %s
  shape: rectangle
}

style: {
  stroke: "#2563eb"
  fill: "#dbeafe"
  font-color: "#1e40af"
}`,
		componentName, componentName, technology)

	if containerName != "" {
		d2 = fmt.Sprintf(`direction: right

%s: {
  %s: %s {
    technology: %s
    shape: rectangle
  }
}

style: {
  stroke: "#2563eb"
  fill: "#dbeafe"
  font-color: "#1e40af"
}`,
			containerName, componentName, componentName, technology)
	}

	return d2
}

// RenderComponentPreview generates and renders a preview diagram for a component.
// It creates a minimal D2 snippet and renders it using the underlying DiagramRenderer.
func (pr *PreviewRenderer) RenderComponentPreview(ctx context.Context, componentName, technology, containerName string) (string, error) {
	if pr.renderer == nil {
		return "", fmt.Errorf("diagram renderer is nil")
	}

	if !pr.renderer.IsAvailable() {
		return "", fmt.Errorf("d2 binary not available for rendering")
	}

	// Generate the D2 source
	d2Source := GenerateComponentPreviewD2(componentName, technology, containerName)

	// Render to SVG
	svgContent, err := pr.renderer.RenderDiagram(ctx, d2Source)
	if err != nil {
		return "", fmt.Errorf("failed to render preview diagram: %w", err)
	}

	return svgContent, nil
}
