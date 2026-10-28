package tools

import (
	"cmp"
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// DescribeTool answers "what is in this architecture?" at three levels of
// detail, optionally scoped to one element (FR-001, FR-002). Its responses
// echo the server's identity (ServerInfo) so an agent can tell which version
// is serving and which project is open.
type DescribeTool struct {
	svc    *usecases.AuthoringService
	enc    usecases.OutputEncoder
	server ServerInfo
}

// ServerInfo is the operating context every read tool echoes, so a client can
// detect a stale server or a misdirected --project before an edit lands.
type ServerInfo struct {
	// Version is the serving binary's version, as the MCP handshake reports
	// it: "" means a development build and is echoed as "dev".
	Version string
	// Root is the resolved absolute project root the tools write inside.
	Root string
}

// NewDescribeTool returns the describe tool. ServerInfo carries the serving
// binary's identity into every response.
func NewDescribeTool(svc *usecases.AuthoringService, enc usecases.OutputEncoder, server ServerInfo) *DescribeTool {
	return &DescribeTool{svc: svc, enc: enc, server: server}
}

// Name implements mcp.Tool.
func (t *DescribeTool) Name() string { return "describe" }

// Description implements mcp.Tool.
func (t *DescribeTool) Description() string {
	return "Describe the architecture: summary (counts and top-level elements), structure (the tree down to containers) or full (everything), optionally scoped to one element. Returns the revision to base edits on, plus server_version, the resolved project root and the source files the server reads."
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
	res, err := usecases.Describe(ctx, t.svc.Deps(), usecases.DescribeRequest{Level: level, Address: address})
	if err != nil {
		return nil, err
	}
	res.ServerVersion, res.Root = cmp.Or(t.server.Version, "dev"), t.server.Root
	return encodeShaped(t.enc, res, args)
}
