package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ValidateDiagramTool validates D2 diagram source code.
type ValidateDiagramTool struct {
	renderer usecases.DiagramRenderer
}

// NewValidateDiagramTool creates a new validate_diagram tool.
func NewValidateDiagramTool(renderer usecases.DiagramRenderer) *ValidateDiagramTool {
	return &ValidateDiagramTool{renderer: renderer}
}

func (t *ValidateDiagramTool) Name() string {
	return "validate_diagram"
}

func (t *ValidateDiagramTool) Description() string {
	return `Validate D2 diagram source code and report syntax errors.
This tool checks if D2 source code is syntactically valid and provides helpful error messages if there are issues.
It also provides recommendations for improving diagram structure and C4 Model compliance.`
}

func (t *ValidateDiagramTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"d2_source": map[string]any{
				"type":        "string",
				"description": "The D2 diagram source code to validate",
			},
			"level": map[string]any{
				"type":        "string",
				"enum":        []string{"system", "container", "component"},
				"description": "C4 Model level for context-aware validation",
			},
		},
		"required": []string{"d2_source"},
	}
}

// Call executes the validate diagram tool.
func (t *ValidateDiagramTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// 1. Parse and validate inputs
	d2Source, _ := args["d2_source"].(string)
	if d2Source == "" {
		return map[string]any{
			"valid":  false,
			"errors": []string{"d2_source cannot be empty"},
		}, nil
	}

	level, _ := args["level"].(string)

	// 2. Validate D2 syntax by attempting to render
	result := map[string]any{
		"valid":        true,
		"errors":       []string{},
		"warnings":     []string{},
		"suggestions":  []string{},
		"syntax_valid": false,
		"d2_available": t.renderer.IsAvailable(),
	}

	// Try to render the diagram
	if t.renderer.IsAvailable() {
		_, err := t.renderer.RenderDiagram(ctx, d2Source)
		if err != nil {
			result["valid"] = false
			result["errors"] = []string{fmt.Sprintf("D2 syntax error: %v", err)}
		} else {
			result["syntax_valid"] = true
		}
	} else {
		result["warnings"] = []string{"D2 CLI not available - syntax validation skipped. Install D2 from https://d2lang.com"}
	}

	// Perform structural validation
	warnings, suggestions := validateDiagramStructure(d2Source, level)
	if len(warnings) > 0 {
		existingWarnings := result["warnings"].([]string)
		result["warnings"] = append(existingWarnings, warnings...)
	}
	result["suggestions"] = suggestions

	// Overall validity check
	errors := result["errors"].([]string)
	result["valid"] = len(errors) == 0 && result["syntax_valid"].(bool)

	// 3. Format response
	return result, nil
}
