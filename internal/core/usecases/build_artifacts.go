package usecases

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// BuildRequest names what to build.
type BuildRequest struct {
	Root         string
	BuildVersion string
	// Formats is the user's format list, as typed: entries may be
	// comma-separated (FR-015).
	Formats []string
	OutDir  string
}

// BuildDeps are the ports a build uses. Backends is the registry of every
// available output format; the request selects from it.
type BuildDeps struct {
	Source   ArchitectureSource
	Prose    ProseReader
	Theme    ThemeSource
	Backends []Backend
	Store    ArtifactStore
}

// BuildResult is everything one build produced.
type BuildResult struct {
	// Artifacts is nil when compilation reported errors.
	Artifacts []viewmodel.Artifact
	Diags     arch.Diagnostics
	// Formats is the expanded format set that was rendered; AddedFormats the
	// ones added because another format requires them (research R5).
	Formats      []string
	AddedFormats []string
	// NothingToDraw is set when the architecture declares no elements.
	NothingToDraw bool
	// ProseFiles are the docs files the build referenced, for the watcher.
	ProseFiles []string
	// Sources are the project's source files, for the manifest notice.
	Sources []string
	Report  CommitReport
}

// ExitCode maps the result to a process exit code, exactly as compile does.
func (r BuildResult) ExitCode(strict bool) int { return r.Diags.ExitCode(strict) }

// ErrorCount and WarningCount report tallies without exposing entity types.
func (r BuildResult) ErrorCount() int   { return r.Diags.CountBySeverity(arch.SeverityError) }
func (r BuildResult) WarningCount() int { return r.Diags.CountBySeverity(arch.SeverityWarning) }

// SupportedFormats lists the formats the given backends provide, in canonical
// order. The CLI's default format set is derived from it.
func SupportedFormats(backends []Backend) []string {
	var out []string
	for _, f := range viewmodel.AllFormats {
		if backendFor(backends, f) != nil {
			out = append(out, string(f))
		}
	}
	return out
}

// BuildArtifacts compiles, projects and renders, returning the complete
// artifact set without writing it. Nothing is rendered when compilation
// reports errors (FR-024).
func BuildArtifacts(ctx context.Context, deps BuildDeps, req BuildRequest) (*BuildResult, error) {
	backends, formats, added, err := selectBackends(deps.Backends, req.Formats)
	if err != nil {
		return nil, err
	}
	compiled, err := CompileArchitecture(ctx, deps.Source, CompileRequest{Root: req.Root, BuildVersion: req.BuildVersion})
	if err != nil {
		return nil, err
	}
	res := &BuildResult{Diags: compiled.Diags, Formats: formats, AddedFormats: added}
	if compiled.HasErrors() {
		return res, nil
	}
	ir := BuildIR(compiled.Model, compiled.Resolved)
	prov := BuildProvenance(compiled.Model)
	res.Sources = prov.AllFiles()
	if len(ir.Elements) == 0 {
		res.NothingToDraw = true
		return res, nil
	}

	prose, err := readProse(ctx, deps.Prose, req.Root, ir)
	if err != nil {
		return nil, err
	}
	res.ProseFiles = proseFiles(prose)
	opts, err := loadRenderOptions(ctx, deps.Theme, req.Root)
	if err != nil {
		return nil, err
	}
	proj, viewDiags, err := Project(ir, prose, prov)
	if err != nil {
		return nil, err
	}
	res.Diags = append(res.Diags, viewDiags...)

	artifacts, err := renderAll(ctx, backends, proj, opts)
	var themeErr *ThemeError
	if errors.As(err, &themeErr) {
		res.Diags = append(res.Diags, themeDiagnostic(themeErr))
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	viewmodel.SortArtifacts(artifacts)
	if collisions := checkCollisions(artifacts, prov); len(collisions) > 0 {
		res.Diags = append(res.Diags, collisions...)
		return res, nil
	}
	res.Artifacts = artifacts
	return res, nil
}

func selectBackends(registry []Backend, raw []string) ([]Backend, []string, []string, error) {
	requested, err := viewmodel.ParseFormats(raw)
	if err != nil {
		return nil, nil, nil, err
	}
	expanded, added := viewmodel.Expand(requested)
	var backends []Backend
	for _, f := range expanded {
		b := backendFor(registry, f)
		if b == nil {
			return nil, nil, nil, fmt.Errorf("no backend is registered for format %q", f)
		}
		backends = append(backends, b)
	}
	return backends, formatNames(expanded), formatNames(added), nil
}

func backendFor(registry []Backend, f viewmodel.Format) Backend {
	for _, b := range registry {
		if b.Format() == f {
			return b
		}
	}
	return nil
}

func formatNames(fs []viewmodel.Format) []string {
	if len(fs) == 0 {
		return nil
	}
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = string(f)
	}
	return out
}

// renderAll runs every backend concurrently. Results are collected by index,
// so the combined order never depends on which backend finishes first.
func renderAll(ctx context.Context, backends []Backend, proj *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error) {
	results := make([][]viewmodel.Artifact, len(backends))
	errs := make([]error, len(backends))
	var wg sync.WaitGroup
	for i, b := range backends {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = b.Render(ctx, proj, opts)
		}()
	}
	wg.Wait()
	var all []viewmodel.Artifact
	for i, b := range backends {
		if errs[i] != nil {
			return nil, fmt.Errorf("render %s: %w", b.Format(), errs[i])
		}
		all = append(all, results[i]...)
	}
	return all, nil
}

// FormatRequires reports whether format f requires format g (research R5).
// Outer layers use it to explain an added format without touching entities.
func FormatRequires(f, g string) bool {
	for _, r := range viewmodel.Format(f).Requires() {
		if string(r) == g {
			return true
		}
	}
	return false
}

// themeDiagnostic reports a malformed theme override as a build error naming
// the file, instead of silently falling back to the built-in theme (FR-035).
func themeDiagnostic(e *ThemeError) arch.Diagnostic {
	return arch.Diagnostic{
		Severity: arch.SeverityError,
		Code:     arch.CodeThemeInvalid,
		Summary:  "Theme override is invalid",
		Detail:   e.Message,
		Range:    arch.SourceRange{File: e.File, StartLine: e.Line, StartColumn: 1},
	}
}
