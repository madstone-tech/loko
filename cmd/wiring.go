package cmd

import (
	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/adapters/markdown"
	"github.com/madstone-tech/loko/internal/adapters/outputdir"
	"github.com/madstone-tech/loko/internal/adapters/projectfs"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

// newBuildDeps is the composition root for `build` and `serve`: the one
// place concrete adapters are chosen for the render ports. Constitution
// v1.4.0 (Principles I and II) places wiring in main.go or cmd/.
//
// Backends is the format registry. The default --format set is derived from
// it, so registering a backend here is all adding a format takes.
func newBuildDeps() usecases.BuildDeps {
	return usecases.BuildDeps{
		Source: hclsource.New(),
		Prose:  newProseReader(),
		Theme:  projectfs.Theme{},
		Backends: []usecases.Backend{
			d2.NewSourceBackend(),
			d2.NewSVGBackend(),
			markdown.New(),
			html.New(),
		},
		Store: outputdir.New(),
	}
}

// newProseReader returns the ProseReader for build, serve and the projection
// goldens.
func newProseReader() usecases.ProseReader { return projectfs.Prose{} }

// newAuthoringDeps is the composition root for reading and editing source:
// the MCP tools and `loko query`. The source adapter is both the compiler's
// input and, with overlays, how an edit is compiled before it is written.
func newAuthoringDeps(root string) usecases.AuthoringDeps {
	return usecases.AuthoringDeps{
		Root:         root,
		Source:       hclsource.New(),
		Editor:       hclsource.NewEditor(),
		BuildVersion: buildVersion(),
	}
}

// newMCPTools builds the five MCP tools (FR-027) around one authoring service,
// so every write to the project queues on the same mutex (FR-017).
func newMCPTools(root string) []mcp.Tool {
	svc := usecases.NewAuthoringService(newAuthoringDeps(root))
	enc := encoding.NewEncoder()
	return []mcp.Tool{
		tools.NewDescribeTool(svc, enc),
		tools.NewQueryTool(svc, enc),
		tools.NewValidateTool(svc, enc),
		tools.NewApplyEditTool(svc, enc),
		tools.NewMoveTool(svc, enc),
	}
}
