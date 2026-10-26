package tools

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// BuildTool renders the project, as loko build does, into a directory inside
// it, and can return one view's diagram inline (feature 017, FR-015). Its
// backends are wired in cmd, so this package imports no render adapter.
type BuildTool struct {
	svc  *usecases.AuthoringService
	deps usecases.BuildDeps
	enc  usecases.OutputEncoder
}

// NewBuildTool returns the build tool. The service supplies the project root
// and build version.
func NewBuildTool(svc *usecases.AuthoringService, deps usecases.BuildDeps, enc usecases.OutputEncoder) *BuildTool {
	return &BuildTool{svc: svc, deps: deps, enc: enc}
}

// Name implements mcp.Tool.
func (t *BuildTool) Name() string { return "build" }

// Description implements mcp.Tool.
func (t *BuildTool) Description() string {
	return "Render diagrams, markdown and a site, as loko build does, into out (default dist) inside the project. Returns the written paths; inline returns one view's diagram (d2 or svg) so you can check it without reading files."
}

// InputSchema implements mcp.Tool.
func (t *BuildTool) InputSchema() map[string]any { return BuildSchema }

// Call implements mcp.Tool.
func (t *BuildTool) Call(ctx context.Context, args map[string]any) (any, error) {
	req, err := projectBuildRequest(args)
	if err != nil {
		return nil, err
	}
	deps := t.svc.Deps()
	req.Root, req.BuildVersion = deps.Root, deps.BuildVersion
	res, err := usecases.BuildInProject(ctx, t.deps, req)
	if err != nil {
		return nil, err
	}
	return encodeShaped(t.enc, res, args)
}
