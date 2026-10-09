package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ValidateTool compiles the project and returns its diagnostics with exact
// source ranges (FR-006).
type ValidateTool struct {
	svc *usecases.AuthoringService
	enc usecases.OutputEncoder
}

// NewValidateTool returns the validate tool.
func NewValidateTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder) *ValidateTool {
	return &ValidateTool{svc: svc, enc: enc}
}

// Name implements mcp.Tool.
func (t *ValidateTool) Name() string { return "validate" }

// Description implements mcp.Tool.
func (t *ValidateTool) Description() string {
	return "Compile the architecture and return every error and warning with its file and position, plus the current revision."
}

// InputSchema implements mcp.Tool.
func (t *ValidateTool) InputSchema() map[string]any { return ValidateSchema }

// Call implements mcp.Tool.
func (t *ValidateTool) Call(ctx context.Context, args map[string]any) (any, error) {
	format, err := formatArg(args)
	if err != nil {
		return nil, err
	}
	res, err := usecases.Validate(ctx, t.svc.Deps())
	if err != nil {
		return nil, err
	}
	return encodeResult(t.enc, res, format)
}
