package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ApplyEditTool adds, updates and removes declarations in the HCL source,
// compiling the result before anything is written (FR-010..FR-021).
type ApplyEditTool struct {
	svc *usecases.AuthoringService
	enc usecases.OutputEncoder
}

// NewApplyEditTool returns the apply_edit tool.
func NewApplyEditTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder) *ApplyEditTool {
	return &ApplyEditTool{svc: svc, enc: enc}
}

// Name implements mcp.Tool.
func (t *ApplyEditTool) Name() string { return "apply_edit" }

// Description implements mcp.Tool.
func (t *ApplyEditTool) Description() string {
	return "Edit the architecture source: add, update or remove elements, relationships, environments, node groups, instances and bindings. " +
		"Edits apply in order, compile once, and save all-or-nothing; a refusal (ok: false) changes nothing and says why. " +
		"Pass the revision from your last read as base_revision. Use preview to see the diff first."
}

// InputSchema implements mcp.Tool.
func (t *ApplyEditTool) InputSchema() map[string]any { return ApplyEditSchema }

// Call implements mcp.Tool.
func (t *ApplyEditTool) Call(ctx context.Context, args map[string]any) (any, error) {
	base, err := str(args, "base_revision")
	if err != nil {
		return nil, err
	}
	edits, err := decodeEdits(args)
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
	res, err := t.svc.Apply(ctx, usecases.ApplyRequest{Edits: edits, BaseRevision: base, Preview: preview})
	if err != nil {
		return nil, err
	}
	return encodeResult(t.enc, res, format)
}
