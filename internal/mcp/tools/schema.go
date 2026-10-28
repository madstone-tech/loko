package tools

import (
	"cmp"
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// SchemaTool describes the architecture language: blocks, attributes, allowed
// values, apply_edit addresses and the docs path rule (feature 017, FR-001).
type SchemaTool struct {
	enc     usecases.OutputEncoder
	version string
}

// NewSchemaTool returns the schema tool. It reads no project, so it needs no
// authoring service; version is the serving binary's version echoed in the
// result so a client can detect a stale server.
func NewSchemaTool(enc usecases.OutputEncoder, version string) *SchemaTool {
	return &SchemaTool{enc: enc, version: version}
}

// Name implements mcp.Tool.
func (t *SchemaTool) Name() string { return "schema" }

// Description implements mcp.Tool.
func (t *SchemaTool) Description() string {
	return "Read-only. Describe the architecture language: every block, where it nests, its attributes with types, required flags and allowed values, the address forms apply_edit accepts, and one example per block. The view block lists the reserved view names the compiler derives automatically. Call it before your first edit; pass block for one block only."
}

// InputSchema implements mcp.Tool.
func (t *SchemaTool) InputSchema() map[string]any { return SchemaSchema }

// Call implements mcp.Tool.
func (t *SchemaTool) Call(_ context.Context, args map[string]any) (any, error) {
	format, err := formatArg(args)
	if err != nil {
		return nil, err
	}
	block, err := optStr(args, "block", "")
	if err != nil {
		return nil, err
	}
	res := usecases.DescribeLanguage(block)
	res.ServerVersion = cmp.Or(t.version, "dev")
	return encodeResult(t.enc, res, format)
}
