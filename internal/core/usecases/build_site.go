package usecases

import (
	"context"
	"fmt"
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
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

// loadRenderOptions reads theme overrides. Whether they are well formed is
// the html backend's to judge; BuildArtifacts turns its ThemeError into a
// theme_invalid diagnostic.
func loadRenderOptions(ctx context.Context, theme ThemeSource, root string) (RenderOptions, error) {
	if theme == nil {
		return RenderOptions{}, nil
	}
	files, err := theme.LoadTheme(ctx, root)
	if err != nil {
		return RenderOptions{}, fmt.Errorf("loading theme: %w", err)
	}
	return RenderOptions{Theme: files}, nil
}
