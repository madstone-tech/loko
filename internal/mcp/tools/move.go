package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// MoveTool renames an element, rewriting every reference and recording the
// rename in a moved block so history survives (FR-022..FR-025).
type MoveTool struct {
	svc *usecases.AuthoringService
	enc usecases.OutputEncoder
}

// NewMoveTool returns the move tool.
func NewMoveTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder) *MoveTool {
	return &MoveTool{svc: svc, enc: enc}
}

// Name implements mcp.Tool.
func (t *MoveTool) Name() string { return "move" }

// Description implements mcp.Tool.
func (t *MoveTool) Description() string {
	return "Rename an element: its name, its kind, or both. Every reference in every file follows, and a moved block records the old address. " +
		"A kind change that needs a different parent is an apply_edit batch: the rename, then an update setting the new parent."
}

// InputSchema implements mcp.Tool.
func (t *MoveTool) InputSchema() map[string]any { return MoveSchema }

// Call implements mcp.Tool.
func (t *MoveTool) Call(ctx context.Context, args map[string]any) (any, error) {
	from, err := str(args, "from")
	if err != nil {
		return nil, err
	}
	to, err := str(args, "to")
	if err != nil {
		return nil, err
	}
	base, err := str(args, "base_revision")
	if err != nil {
		return nil, err
	}
	preview, err := boolArg(args, "preview")
	if err != nil {
		return nil, err
	}
	format, err := formatArg(args)
	if err != nil {
		return nil, err
	}
	res, err := t.svc.Move(ctx, from, to, base, preview)
	if err != nil {
		return nil, err
	}
	return encodeResult(t.enc, res, format)
}
