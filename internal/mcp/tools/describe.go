package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// DescribeTool answers "what is in this architecture?" at three levels of
// detail, optionally scoped to one element (FR-001, FR-002).
type DescribeTool struct {
	svc *usecases.AuthoringService
	enc usecases.OutputEncoder
}

// NewDescribeTool returns the describe tool.
func NewDescribeTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder) *DescribeTool {
	return &DescribeTool{svc: svc, enc: enc}
}

// Name implements mcp.Tool.
func (t *DescribeTool) Name() string { return "describe" }

// Description implements mcp.Tool.
func (t *DescribeTool) Description() string {
	return "Describe the architecture: summary (counts and top-level elements), structure (the tree down to containers) or full (everything), optionally scoped to one element. Returns the revision to base edits on."
}

// InputSchema implements mcp.Tool.
func (t *DescribeTool) InputSchema() map[string]any { return DescribeSchema }

// Call implements mcp.Tool.
func (t *DescribeTool) Call(ctx context.Context, args map[string]any) (any, error) {
	level, err := optStr(args, "level", "summary")
	if err != nil {
		return nil, err
	}
	address, err := optStr(args, "address", "")
	if err != nil {
		return nil, err
	}
	format, err := formatArg(args)
	if err != nil {
		return nil, err
	}
	res, err := usecases.Describe(ctx, t.svc.Deps(), usecases.DescribeRequest{Level: level, Address: address})
	if err != nil {
		return nil, err
	}
	return encodeResult(t.enc, res, format)
}
