package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func compileWith(t *testing.T, src stubSource, version string) *CompileResult {
	t.Helper()
	res, err := CompileArchitecture(context.Background(), src,
		CompileRequest{Root: "", BuildVersion: version})
	if err != nil {
		t.Fatalf("CompileArchitecture: %v", err)
	}
	return res
}

// TestCompileAccumulatesEveryStage is FR-031 at the orchestration level: no
// stage returns early, so diagnostics from parsing, resolution, structure and
// warnings all reach the caller from one invocation.
func TestCompileAccumulatesEveryStage(t *testing.T) {
	t.Parallel()

	// The stub supplies the two diagnostics only a parser can produce; the
	// rest must come from the stages the orchestrator runs.
	src := stubSource{
		diags: arch.Diagnostics{
			{Severity: arch.SeverityError, Code: arch.CodeUnknownBlock, Summary: "from parser", Range: at(1)},
			{Severity: arch.SeverityError, Code: arch.CodeUnknownFunction, Summary: "from parser", Range: at(2)},
		},
		model: &arch.SourceModel{
			Project: arch.ProjectDecl{Name: "p", Declared: true, Version: ">= 99.0", VersionRange: at(3)},
			Elements: []arch.ElementDecl{
				elem(arch.KindSystem, "s", 10),
				// Unresolvable relationship target.
				{Kind: arch.KindContainer, Name: "api", Range: at(20), Parent: ref("system.s", 21),
					Relations: []arch.RelationDecl{uses("x", "container.nope", 22)}},
				// Container with no system at all.
				elem(arch.KindContainer, "loose", 30),
			},
		},
	}

	res := compileWith(t, src, "1.0.0")

	for _, want := range []string{
		arch.CodeUnknownBlock,        // parser
		arch.CodeUnknownFunction,     // parser
		arch.CodeVersionUnsatisfied,  // version check
		arch.CodeUnresolvedReference, // resolution
		arch.CodeWrongParentKind,     // structure
		arch.CodeMissingDocs,         // warnings
	} {
		if !hasCode(res.Diags, want) {
			t.Errorf("%s missing; a stage returned early. Got %v", want, codes(res.Diags))
		}
	}
}

func TestCompileExitCodes(t *testing.T) {
	t.Parallel()

	clean := compileWith(t, stubSource{model: &arch.SourceModel{}}, "1.0.0")
	if got := clean.ExitCode(false); got != ExitSuccess {
		t.Errorf("clean lenient exit = %d, want %d", got, ExitSuccess)
	}

	warned := compileWith(t, stubSource{model: &arch.SourceModel{
		Elements: []arch.ElementDecl{elem(arch.KindSystem, "s", 1)},
	}}, "1.0.0")
	if warned.HasErrors() {
		t.Fatalf("unexpected errors: %v", codes(warned.Diags))
	}
	if got := warned.ExitCode(false); got != ExitSuccess {
		t.Errorf("warnings lenient exit = %d, want %d", got, ExitSuccess)
	}
	if got := warned.ExitCode(true); got != ExitWarnings {
		t.Errorf("warnings strict exit = %d, want %d", got, ExitWarnings)
	}
	if warned.WarningCount() == 0 {
		t.Error("expected warnings for an empty, undocumented, unconnected system")
	}

	broken := compileWith(t, stubSource{model: &arch.SourceModel{
		Elements: []arch.ElementDecl{
			{Kind: arch.KindContainer, Name: "a", Range: at(1), Parent: ref("system.nope", 2)},
		},
	}}, "1.0.0")
	if got := broken.ExitCode(false); got != ExitErrors {
		t.Errorf("error exit = %d, want %d", got, ExitErrors)
	}
	if broken.ErrorCount() == 0 {
		t.Error("ErrorCount reported none")
	}
}

func TestCompileSourceFailureIsAnError(t *testing.T) {
	t.Parallel()

	_, err := CompileArchitecture(context.Background(),
		stubSource{err: errors.New("root unreadable")}, CompileRequest{Root: "/nope"})
	if err == nil {
		t.Error("a source failure returned no error")
	}

	if _, err := CompileArchitecture(context.Background(), nil, CompileRequest{}); err == nil {
		t.Error("a nil source returned no error")
	}
}

// TestCompileHandlesNilModel: a source may report a fatal parse and hand back
// nothing; the pipeline must still produce diagnostics rather than panic.
func TestCompileHandlesNilModel(t *testing.T) {
	t.Parallel()

	res := compileWith(t, stubSource{
		model: nil,
		diags: arch.Diagnostics{{Severity: arch.SeverityError, Code: arch.CodeSyntaxError, Range: at(1)}},
	}, "1.0.0")

	if !hasCode(res.Diags, arch.CodeSyntaxError) {
		t.Error("parser diagnostics lost when the model was nil")
	}
	if res.Model == nil {
		t.Error("CompileResult.Model is nil; callers would have to nil-check it")
	}
}

// TestCompileDeterministicOrder covers FR-033 at the pipeline level.
func TestCompileDeterministicOrder(t *testing.T) {
	t.Parallel()

	src := stubSource{model: &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "z", 30),
		elem(arch.KindSystem, "a", 10),
		elem(arch.KindSystem, "m", 20),
	}}}

	first := renderOrder(compileWith(t, src, "1.0.0"))
	for i := 0; i < 5; i++ {
		got := renderOrder(compileWith(t, src, "1.0.0"))
		if got != first {
			t.Fatalf("run %d differs:\n%s\nvs\n%s", i, got, first)
		}
	}
}

func renderOrder(res *CompileResult) string {
	out := ""
	for _, d := range res.Diags.SortedForOutput() {
		out += d.Code + "@" + d.Range.Loc() + "\n"
	}
	return out
}
