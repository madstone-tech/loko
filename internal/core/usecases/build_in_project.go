package usecases

import (
	"cmp"
	"context"
	"path"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// diagramDir is where every view's diagram is written; a view's name is its
// file's base name there.
const diagramDir = "diagrams/"

// ProjectBuildRequest is a build an MCP client asks for: confined to the
// project, with one view optionally returned inline (feature 017, R8).
type ProjectBuildRequest struct {
	Root         string
	BuildVersion string
	// OutDir is relative to the project root; "" means dist.
	OutDir  string
	Formats []string
	Ignore  []string
	// Strict makes remaining warnings fail the build, as loko build --strict.
	Strict bool
	// Inline names a view whose diagram is returned in the result.
	Inline       string
	InlineFormat string // d2 (default) or svg
}

// InlineView is one view's diagram, returned so a client can inspect it
// without reading files.
type InlineView struct {
	View    string `json:"view" toon:"view"`
	Format  string `json:"format" toon:"format"`
	Content string `json:"content" toon:"content"`
}

// ProjectBuildResult reports a build. OK is false when there are errors, or
// warnings under Strict; under Strict the files are still written, as the
// CLI does. Refusal means nothing was written.
type ProjectBuildResult struct {
	OK          bool             `json:"ok" toon:"ok"`
	OutDir      string           `json:"out" toon:"out"`
	Formats     []string         `json:"formats,omitempty" toon:"formats,omitempty"`
	Written     int              `json:"written" toon:"written"`
	Paths       []string         `json:"paths,omitempty" toon:"paths,omitempty"`
	Diags       arch.Diagnostics `json:"-" toon:"-"`
	Diagnostics []DiagnosticView `json:"diagnostics,omitempty" toon:"diagnostics,omitempty"`
	Inline      *InlineView      `json:"inline,omitempty" toon:"inline,omitempty"`
	InlineError *ReadError       `json:"inline_error,omitempty" toon:"inline_error,omitempty"`
	Refusal     *Refusal         `json:"refusal,omitempty" toon:"refusal,omitempty"`
}

// ShapeDiagnostics sets the encoded diagnostics; see EditResult.
func (r *ProjectBuildResult) ShapeDiagnostics(full bool) { r.Diagnostics = shapedOrNil(r.Diags, full) }

// BuildInProject confines the output directory to the project, builds into
// it, and picks the inline view from what was built (no second render).
func BuildInProject(ctx context.Context, deps BuildDeps, req ProjectBuildRequest) (*ProjectBuildResult, error) {
	out := &ProjectBuildResult{OutDir: cmp.Or(req.OutDir, "dist")}
	dir, err := deps.Store.Confine(ctx, req.Root, req.OutDir)
	if err != nil {
		out.Refusal = &Refusal{Reason: authoring.ReasonPathRefused, Detail: err.Error()}
		return out, nil
	}
	inlineFormat := cmp.Or(req.InlineFormat, string(viewmodel.FormatD2))
	formats := req.Formats
	if len(formats) == 0 {
		formats = SupportedFormats(deps.Backends)
	}
	if req.Inline != "" && !slices.Contains(formats, inlineFormat) {
		formats = append(slices.Clone(formats), inlineFormat)
	}
	res, err := Build(ctx, deps, BuildRequest{Root: req.Root, BuildVersion: req.BuildVersion,
		Formats: formats, OutDir: dir, Ignore: req.Ignore})
	if err != nil {
		return nil, err
	}
	out.Diags, out.Formats, out.Written = res.Diags, res.Formats, len(res.Report.Written)
	out.OK = res.ExitCode(req.Strict) == ExitSuccess
	for _, a := range res.Artifacts {
		out.Paths = append(out.Paths, path.Join(out.OutDir, a.Path))
	}
	if req.Inline != "" && res.Artifacts != nil {
		out.Inline, out.InlineError = PickInlineView(res.Artifacts, req.Inline, inlineFormat)
	}
	return out, nil
}

// PickInlineView returns the named view's diagram in the given format, or an
// unknown_view error listing the views that exist.
func PickInlineView(artifacts []viewmodel.Artifact, view, format string) (*InlineView, *ReadError) {
	want := diagramDir + view + "." + format
	var names []string
	for _, a := range artifacts {
		if a.Path == want {
			return &InlineView{View: view, Format: format, Content: string(a.Bytes)}, nil
		}
		if name, ok := strings.CutPrefix(a.Path, diagramDir); ok && path.Ext(name) == "."+format {
			names = append(names, strings.TrimSuffix(name, "."+format))
		}
	}
	return nil, &ReadError{Reason: "unknown_view", Detail: "no view named " + view, Suggestions: names}
}
