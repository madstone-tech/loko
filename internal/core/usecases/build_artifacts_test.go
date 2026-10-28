package usecases

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

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

// TestBuildArtifactsWideViewHint: the hint is added after rendering, is
// removed by the run's and the project's ignore lists, never changes the exit
// code, and needs a laid-out diagram (feature 018, FR-012 to FR-014).
func TestBuildArtifactsWideViewHint(t *testing.T) {
	t.Parallel()
	wide := func(m *arch.SourceModel) *buildFixture {
		f := newBuildFixture(m)
		f.svg.artifacts = []viewmodel.Artifact{svgOf("system-s", 4000, 1000)}
		return f
	}
	hints := func(r *BuildResult) int {
		n := 0
		for _, d := range r.Diags {
			if d.Code == arch.CodeWideView {
				n++
			}
		}
		return n
	}
	f := wide(smallModel())
	res, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}})
	if err != nil {
		t.Fatal(err)
	}
	without, _ := res.Diags.Without([]string{arch.CodeWideView})
	if hints(res) != 1 || res.ExitCode(true) != without.ExitCode(true) {
		t.Errorf("wide view: %d hints, exit %d (%d without the hint)", hints(res), res.ExitCode(true), without.ExitCode(true))
	}
	res, _ = BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}, Ignore: []string{arch.CodeWideView}})
	if hints(res) != 0 {
		t.Error("--ignore wide_view left the hint")
	}
	ignoring := smallModel()
	ignoring.Project.IgnoreWarnings = []string{arch.CodeWideView}
	g := wide(ignoring)
	if res, _ := BuildArtifacts(context.Background(), g.deps, BuildRequest{Root: ".", Formats: []string{"svg"}}); hints(res) != 0 {
		t.Error("ignore_warnings = [wide_view] left the hint")
	}
	d2 := newBuildFixture(smallModel())
	d2.deps.Backends = append(d2.deps.Backends, &fakeBackend{format: viewmodel.FormatD2,
		artifacts: []viewmodel.Artifact{{Path: "diagrams/system-s.d2"}}})
	if res, _ := BuildArtifacts(context.Background(), d2.deps, BuildRequest{Root: ".", Formats: []string{"d2"}}); hints(res) != 0 {
		t.Error("a D2-only build gave a hint")
	}
}

// TestRenderAllStagesRequiredFormats is research R1: a backend whose format
// requires another receives that format's artifacts through Upstream.
func TestRenderAllStagesRequiredFormats(t *testing.T) {
	svgBytes := []byte("<svg/>")
	for range 20 {
		svg := &fakeBackend{format: viewmodel.FormatSVG,
			artifacts: []viewmodel.Artifact{{Path: "diagrams/a.svg", Format: viewmodel.FormatSVG, Bytes: svgBytes}}}
		d2 := &fakeBackend{format: viewmodel.FormatD2,
			artifacts: []viewmodel.Artifact{{Path: "diagrams/a.d2", Format: viewmodel.FormatD2}}}
		html := &fakeBackend{format: viewmodel.FormatHTML,
			artifacts: []viewmodel.Artifact{{Path: "index.html", Format: viewmodel.FormatHTML}}}
		got, err := renderAll(context.Background(), []Backend{html, d2, svg}, &viewmodel.Projection{}, RenderOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if html.opts.Upstream == nil {
			t.Fatal("html got no Upstream")
		}
		up, err := html.opts.Upstream()
		if err != nil || len(up) != 1 || string(up["diagrams/a.svg"]) != string(svgBytes) {
			t.Fatalf("html Upstream = %v, %v; want only diagrams/a.svg", up, err)
		}
		if svg.opts.Upstream != nil || d2.opts.Upstream != nil {
			t.Fatal("a backend that requires nothing got an Upstream")
		}
		var paths []string
		for _, a := range got {
			paths = append(paths, a.Path)
		}
		if strings.Join(paths, " ") != "diagrams/a.d2 diagrams/a.svg index.html" {
			t.Fatalf("order = %v, want stage 1 (registry order) then stage 2", paths)
		}
	}
}

// TestOverviewDiagnostics is feature 021 research R12: the index counts the
// open warnings and infos known before rendering, after ignore filtering.
func TestOverviewDiagnostics(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel()) // no relationships: orphan warnings
	res, err := BuildArtifacts(context.Background(), f.deps, BuildRequest{Root: ".", Formats: []string{"html"}})
	if err != nil {
		t.Fatal(err)
	}
	got := f.html.got.Overview
	if got.Warnings == 0 || got.Warnings != res.WarningCount() {
		t.Errorf("overview warnings = %d, build warnings = %d", got.Warnings, res.WarningCount())
	}
	if !slices.Equal(f.html.opts.Formats, []viewmodel.Format{viewmodel.FormatSVG, viewmodel.FormatHTML}) {
		t.Errorf("RenderOptions.Formats = %v, want the formats being built", f.html.opts.Formats)
	}

	g := newBuildFixture(smallModel())
	ignored, err := BuildArtifacts(context.Background(), g.deps, BuildRequest{Root: ".", Formats: []string{"html"},
		Ignore: []string{arch.CodeOrphanElement}})
	if err != nil {
		t.Fatal(err)
	}
	if n := g.html.got.Overview.Warnings; n >= got.Warnings || n != ignored.WarningCount() {
		t.Errorf("with orphan_element ignored: %d open warnings, build reports %d (was %d)", n, ignored.WarningCount(), got.Warnings)
	}
}

// gatedBackend renders only once release is closed, and closes started when
// it begins, so a test can tell whether two backends overlap.
type gatedBackend struct {
	format           viewmodel.Format
	started, release chan struct{}
	upstream         func() (map[string][]byte, error)
}

func (b *gatedBackend) Format() viewmodel.Format { return b.format }

func (b *gatedBackend) Render(_ context.Context, _ *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error) {
	close(b.started)
	<-b.release
	b.upstream = opts.Upstream
	return []viewmodel.Artifact{{Path: "x." + string(b.format), Format: b.format, Bytes: []byte(b.format)}}, nil
}

// TestRenderAllOverlapsStages: the dependent backend starts while the one it
// requires is still rendering, so the two are not serialised (research R1).
func TestRenderAllOverlapsStages(t *testing.T) {
	t.Parallel()
	svg := &gatedBackend{format: viewmodel.FormatSVG, started: make(chan struct{}), release: make(chan struct{})}
	html := &gatedBackend{format: viewmodel.FormatHTML, started: make(chan struct{}), release: make(chan struct{})}
	done := make(chan error, 1)
	go func() {
		_, err := renderAll(context.Background(), []Backend{svg, html}, &viewmodel.Projection{}, RenderOptions{})
		done <- err
	}()
	for _, b := range []*gatedBackend{svg, html} {
		select {
		case <-b.started:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s did not start while the other backend was still rendering", b.format)
		}
	}
	close(svg.release)
	close(html.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	up, err := html.upstream()
	if err != nil || string(up["x.svg"]) != "svg" {
		t.Errorf("Upstream after both finished = %v, %v", up, err)
	}
}
