package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func projectBuild(t *testing.T, f *buildFixture, req ProjectBuildRequest) *ProjectBuildResult {
	t.Helper()
	req.Root = "root"
	res, err := BuildInProject(t.Context(), f.deps, req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestBuildInProject(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	f.svg.artifacts = []viewmodel.Artifact{{Path: "diagrams/landscape.svg", Bytes: []byte("<svg/>")}}
	res := projectBuild(t, f, ProjectBuildRequest{Inline: "landscape", InlineFormat: "svg"})
	if !res.OK || res.OutDir != "dist" || f.store.outDir != "root/dist" || f.store.commits != 1 {
		t.Fatalf("result %+v, store %+v", res, f.store)
	}
	if res.Inline == nil || res.Inline.Content != "<svg/>" || res.Inline.View != "landscape" {
		t.Errorf("inline %+v", res.Inline)
	}
	if len(res.Paths) != 2 || res.Paths[0] != "dist/diagrams/landscape.svg" {
		t.Errorf("paths %v", res.Paths)
	}
}

func TestBuildInProjectRefusesOutsideTheRoot(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	f.store.refuse = true
	res := projectBuild(t, f, ProjectBuildRequest{OutDir: "../x"})
	if res.OK || res.Refusal == nil || res.Refusal.Reason != "path_refused" || f.store.commits != 0 {
		t.Errorf("result %+v, commits %d", res, f.store.commits)
	}
}

func TestBuildInProjectUnknownView(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	res := projectBuild(t, f, ProjectBuildRequest{Formats: []string{"svg"}, Inline: "nope", InlineFormat: "svg"})
	if res.InlineError == nil || res.InlineError.Reason != "unknown_view" || len(res.InlineError.Suggestions) != 2 {
		t.Errorf("inline error %+v", res.InlineError)
	}
}

// TestBuildInProjectStrict: as loko build --strict, files are written and the
// remaining warnings make the result not OK.
func TestBuildInProjectStrict(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel()) // has missing_docs warnings
	res := projectBuild(t, f, ProjectBuildRequest{Strict: true})
	if res.OK || f.store.commits != 1 {
		t.Errorf("strict with warnings: ok %v, commits %d", res.OK, f.store.commits)
	}
	ignored := newBuildFixture(smallModel())
	res = projectBuild(t, ignored, ProjectBuildRequest{Strict: true, Ignore: arch.IgnorableCodes})
	if !res.OK {
		t.Errorf("strict with every warning ignored: %+v", res.Diags)
	}
}

func TestBuildInProjectCompileErrors(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(brokenModel())
	res := projectBuild(t, f, ProjectBuildRequest{})
	if res.OK || len(res.Diags) == 0 || f.store.commits != 0 {
		t.Errorf("broken model: ok %v, diags %d, commits %d", res.OK, len(res.Diags), f.store.commits)
	}
}
