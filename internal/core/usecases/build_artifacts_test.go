package usecases

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// smallModel is a valid source model: one system with one container.
func smallModel() *arch.SourceModel {
	return &arch.SourceModel{
		Project: arch.ProjectDecl{Name: "p", Declared: true, Range: at(1)},
		Files:   []string{"arch.loko.hcl"},
		Elements: []arch.ElementDecl{
			{Kind: arch.KindSystem, Name: "s", Range: at(2)},
			{Kind: arch.KindContainer, Name: "c", Range: at(3), Parent: ref("system.s", 4)},
		},
	}
}

func brokenModel() *arch.SourceModel {
	m := smallModel()
	m.Elements[1].Parent = ref("system.missing", 4)
	return m
}

type buildFixture struct {
	svg, html *fakeBackend
	store     *fakeStore
	deps      BuildDeps
}

func newBuildFixture(model *arch.SourceModel) *buildFixture {
	f := &buildFixture{
		svg:   &fakeBackend{format: viewmodel.FormatSVG, artifacts: []viewmodel.Artifact{{Path: "diagrams/b.svg"}, {Path: "diagrams/a.svg"}}},
		html:  &fakeBackend{format: viewmodel.FormatHTML, artifacts: []viewmodel.Artifact{{Path: "index.html"}}},
		store: &fakeStore{},
	}
	f.deps = BuildDeps{
		Source:   stubSource{model: model},
		Prose:    &fakeProse{},
		Theme:    &fakeTheme{},
		Backends: []Backend{f.svg, f.html},
		Store:    f.store,
	}
	return f
}

func TestBuildArtifacts(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	res, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"html"}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res.AddedFormats, []string{"svg"}) {
		t.Errorf("AddedFormats = %v, want [svg] (html requires svg)", res.AddedFormats)
	}
	if f.svg.runs != 1 || f.html.runs != 1 {
		t.Errorf("backend runs svg=%d html=%d, want 1 each", f.svg.runs, f.html.runs)
	}
	var paths []string
	for _, a := range res.Artifacts {
		paths = append(paths, a.Path)
	}
	if want := []string{"diagrams/a.svg", "diagrams/b.svg", "index.html"}; !reflect.DeepEqual(paths, want) {
		t.Errorf("artifacts = %v, want sorted %v", paths, want)
	}
	if f.svg.got == nil || len(f.svg.got.Views) == 0 || f.svg.got != f.html.got {
		t.Error("every backend must receive the same projection")
	}
	if res.NothingToDraw || res.ExitCode(false) != ExitSuccess {
		t.Errorf("result = %+v", res)
	}
	if f.store.commits != 0 {
		t.Error("BuildArtifacts must not commit; Build does")
	}
}

func TestBuildArtifactsCompileErrorsRenderNothing(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(brokenModel())
	res, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Diags.HasErrors() || res.Artifacts != nil || f.svg.runs != 0 {
		t.Fatalf("compile errors must stop before any backend runs: %+v", res)
	}
	if res.ExitCode(false) != ExitErrors {
		t.Errorf("exit = %d", res.ExitCode(false))
	}
}

func TestBuildArtifactsNothingToDraw(t *testing.T) {
	t.Parallel()
	model := &arch.SourceModel{Project: arch.ProjectDecl{Name: "p", Declared: true, Range: at(1)}, Files: []string{"arch.loko.hcl"}}
	f := newBuildFixture(model)
	res, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.NothingToDraw || len(res.Artifacts) != 0 || f.svg.runs != 0 {
		t.Errorf("empty architecture: %+v runs=%d", res, f.svg.runs)
	}
}

func TestBuildArtifactsUnknownFormat(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	_, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"pdf"}})
	want := `unsupported format "pdf": supported formats are d2, svg, md, html`
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}

func TestBuildArtifactsMissingBackend(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	_, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"d2"}})
	if err == nil {
		t.Fatal("a format with no registered backend must be an error")
	}
}

func TestBuildArtifactsBackendError(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	f.svg.err = errors.New("boom")
	if _, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}}); err == nil {
		t.Fatal("a backend failure must surface as an error")
	}
}

func TestSupportedFormats(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	if got := SupportedFormats(f.deps.Backends); !reflect.DeepEqual(got, []string{"svg", "html"}) {
		t.Errorf("SupportedFormats = %v", got)
	}
}

func TestBuildCommitsOnlyOnSuccess(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	res, err := Build(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}, OutDir: "out"})
	if err != nil {
		t.Fatal(err)
	}
	if f.store.commits != 1 || f.store.outDir != "out" || len(res.Report.Written) != 2 {
		t.Errorf("commit: n=%d dir=%q report=%+v", f.store.commits, f.store.outDir, res.Report)
	}

	broken := newBuildFixture(brokenModel())
	if _, err := Build(context.Background(), broken.deps, BuildRequest{Root: ".", Formats: []string{"svg"}, OutDir: "out"}); err != nil {
		t.Fatal(err)
	}
	if broken.store.commits != 0 {
		t.Error("a build with errors must never commit (FR-024)")
	}
}

func TestBuildArtifactsTheme(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	theme := []viewmodel.ThemeFile{{Name: "style.css", Origin: "templates/style.css"}}
	f.deps.Theme = &fakeTheme{files: theme}
	if _, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"html"}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.html.opts.Theme, theme) {
		t.Errorf("the theme did not reach the backend: %+v", f.html.opts)
	}
}

func TestBuildThemeErrorIsADiagnostic(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	f.html.err = &ThemeError{File: "templates/partials.gohtml", Line: 3, Message: `defines unknown block "nope"`}
	res, err := Build(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"html"}, OutDir: "out"})
	if err != nil {
		t.Fatalf("a malformed theme is a diagnostic, not an operational error: %v", err)
	}
	var theme []arch.Diagnostic
	for _, d := range res.Diags {
		if d.Code == arch.CodeThemeInvalid {
			theme = append(theme, d)
		}
	}
	if len(theme) != 1 || theme[0].Severity != arch.SeverityError ||
		theme[0].Range.File != "templates/partials.gohtml" || theme[0].Range.StartLine != 3 {
		t.Fatalf("theme diagnostics = %+v", theme)
	}
	if res.Artifacts != nil || f.store.commits != 0 {
		t.Error("nothing may be committed with a malformed theme")
	}
}
