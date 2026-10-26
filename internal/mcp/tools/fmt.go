package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// FmtTool formats the project's source as `loko fmt` does, previewed or
// written with the same revision check as apply_edit (feature 017, FR-019).
type FmtTool struct {
	svc *usecases.AuthoringService
	enc usecases.OutputEncoder
}

// NewFmtTool returns the fmt tool.
func NewFmtTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder) *FmtTool {
	return &FmtTool{svc: svc, enc: enc}
}

// Name implements mcp.Tool.
func (t *FmtTool) Name() string { return "fmt" }

// Description implements mcp.Tool.
func (t *FmtTool) Description() string {
	return "Format every *.loko.hcl file canonically, as loko fmt does. preview: true returns the diffs and writes nothing. A write is refused if a file changed since base_revision."
}

// InputSchema implements mcp.Tool.
func (t *FmtTool) InputSchema() map[string]any { return FmtSchema }

// Call implements mcp.Tool.
func (t *FmtTool) Call(ctx context.Context, args map[string]any) (any, error) {
	base, err := str(args, "base_revision")
	if err != nil {
		return nil, err
	}
	preview, err := boolArg(args, "preview")
	if err != nil {
		return nil, err
	}
	res, err := t.svc.Format(ctx, base, preview)
	if err != nil {
		return nil, err
	}
	return encodeShaped(t.enc, res, args)
}
