// Package usecases contains business logic for loko operations.
package usecases

import (
	"context"
	"fmt"
	"strings"
)

// RenderDiagramPreviewRequest holds the input for rendering a D2 diagram preview.
type RenderDiagramPreviewRequest struct {
	// D2Source is the raw D2 diagram source code to render.
	D2Source string
}

// RenderDiagramPreviewResult holds the output from rendering a D2 diagram preview.
type RenderDiagramPreviewResult struct {
	// Preview contains the rendered SVG content or ASCII representation.
	Preview string
	// Format indicates the format of the preview ("svg" or "ascii").
	Format string
}

// RenderDiagramPreview is the use case for rendering D2 diagram previews.
// It takes D2 source code and returns a preview suitable for terminal display.
type RenderDiagramPreview struct {
	renderer DiagramRenderer
}

// NewRenderDiagramPreview creates a new RenderDiagramPreview use case.
func NewRenderDiagramPreview(renderer DiagramRenderer) *RenderDiagramPreview {
	return &RenderDiagramPreview{
		renderer: renderer,
	}
}

// Execute renders the D2 source and returns the preview.
// Empty D2 source returns empty result (no error).
// If the renderer is unavailable or returns an error, that error is propagated.
func (uc *RenderDiagramPreview) Execute(ctx context.Context, req *RenderDiagramPreviewRequest) (*RenderDiagramPreviewResult, error) {
	if uc.renderer == nil {
		return nil, fmt.Errorf("diagram renderer is nil")
	}

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Handle empty source gracefully
	if strings.TrimSpace(req.D2Source) == "" {
		return &RenderDiagramPreviewResult{
			Preview: "",
			Format:  "svg",
		}, nil
	}

	// Render the diagram
	svgContent, err := uc.renderer.RenderDiagram(ctx, req.D2Source)
	if err != nil {
		return nil, fmt.Errorf("failed to render diagram: %w", err)
	}

	// For now, we return SVG as the preview format
	// In the future, we could convert to ASCII if needed
	return &RenderDiagramPreviewResult{
		Preview: svgContent,
		Format:  "svg",
	}, nil
}
