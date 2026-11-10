package usecases

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// Build renders the architecture and commits the artifacts to req.OutDir. It
// is the single call `loko build` makes.
//
// The store is reached only when there are no errors, so a broken project
// leaves the previous output exactly as it was (FR-024).
func Build(ctx context.Context, deps BuildDeps, req BuildRequest) (*BuildResult, error) {
	if deps.Store == nil {
		return nil, fmt.Errorf("build: no artifact store configured")
	}
	res, err := BuildArtifacts(ctx, deps, req)
	if err != nil || res.Diags.HasErrors() || res.NothingToDraw {
		return res, err
	}
	report, err := deps.Store.Commit(ctx, req.OutDir, res.Artifacts, res.Sources)
	if err != nil {
		return nil, err
	}
	res.Report = report
	return res, nil
}

// readProse reads every distinct docs reference once, in address order. A
// missing file is recorded as not found — the compiler has already warned
// with docs_not_found, so no second diagnostic is emitted.
func readProse(ctx context.Context, reader ProseReader, root string, ir *arch.IR) (ProseSet, error) {
	set := ProseSet{}
	if reader == nil {
		return set, nil
	}
	for _, e := range ir.Elements {
		if e.Docs == "" {
			continue
		}
		if _, done := set[e.Docs]; done {
			continue
		}
		text, found, err := reader.ReadProse(ctx, root, e.Docs)
		if err != nil {
			return nil, fmt.Errorf("reading prose %s: %w", e.Docs, err)
		}
		set[e.Docs] = ProseDoc{Text: text, Found: found}
	}
	return set, nil
}

func proseFiles(set ProseSet) []string {
	out := make([]string, 0, len(set))
	for docs := range set {
		out = append(out, docs)
	}
	sort.Strings(out)
	return out
}

// loadRenderOptions reads the theme and checks its file names and assets
// before any backend renders, so an invalid theme fails every build, html or
// not, and nothing is written (contract T-4). Template parse errors and
// unknown blocks are the html backend's to judge. BuildArtifacts turns its
// ThemeError into a theme_invalid diagnostic.
func loadRenderOptions(ctx context.Context, theme ThemeSource, root string) (RenderOptions, arch.Diagnostics, error) {
	if theme == nil {
		return RenderOptions{}, nil, nil
	}
	files, err := theme.LoadTheme(ctx, root)
	if err != nil {
		return RenderOptions{}, nil, fmt.Errorf("loading theme: %w", err)
	}
	var diags arch.Diagnostics
	for _, p := range viewmodel.ValidateTheme(files, viewmodel.OverridableThemeFiles) {
		diags = append(diags, themeDiagnostic(&ThemeError{File: p.Origin, Message: p.Message}))
	}
	return RenderOptions{Theme: files, ThemeReport: themeReport(files)}, diags, nil
}

// themeReport lists the top-level theme files, sorted by name, with no blocks
// yet, and counts the assets. It is nil when the theme has no file.
func themeReport(files []viewmodel.ThemeFile) *viewmodel.ThemeReport {
	if len(files) == 0 {
		return nil
	}
	r := &viewmodel.ThemeReport{Dir: "templates/"}
	for _, f := range files {
		if f.IsAsset() {
			r.Assets++
			continue
		}
		r.Files = append(r.Files, viewmodel.ThemeFileReport{Name: f.Name})
	}
	slices.SortFunc(r.Files, func(a, b viewmodel.ThemeFileReport) int { return cmp.Compare(a.Name, b.Name) })
	return r
}
