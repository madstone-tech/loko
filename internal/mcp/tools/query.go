package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryTool answers graph questions: dependents, dependencies, a path between
// two elements, orphans, and coupling (FR-003..FR-005).
type QueryTool struct {
	svc *usecases.AuthoringService
	enc usecases.OutputEncoder
}

// NewQueryTool returns the query tool.
func NewQueryTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder) *QueryTool {
	return &QueryTool{svc: svc, enc: enc}
}

// Name implements mcp.Tool.
func (t *QueryTool) Name() string { return "query" }

// Description implements mcp.Tool.
func (t *QueryTool) Description() string {
	return "Ask the architecture a question: who depends on an element (dependents), what it depends on (dependencies), the path between two elements, orphans, or coupling. An element includes everything inside it."
}

// InputSchema implements mcp.Tool.
func (t *QueryTool) InputSchema() map[string]any { return QuerySchema }

// Call implements mcp.Tool.
func (t *QueryTool) Call(ctx context.Context, args map[string]any) (any, error) {
	req, err := queryRequest(args)
	if err != nil {
		return nil, err
	}
	format, err := formatArg(args)
	if err != nil {
		return nil, err
	}
	res, err := usecases.Query(ctx, t.svc.Deps(), req)
	if err != nil {
		return nil, err
	}
	return encodeResult(t.enc, res, format)
}
