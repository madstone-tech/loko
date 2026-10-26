package usecases

import (
	"context"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ignoreModel has two warnings, missing_docs and orphan_element, on one
// system, and a project block whose ignore_warnings the test sets.
func ignoreModel(ignore ...string) stubSource {
	return stubSource{model: &arch.SourceModel{
		Project: arch.ProjectDecl{Name: "p", Declared: true, IgnoreWarnings: ignore, IgnoreWarningsRange: at(1)},
		Elements: []arch.ElementDecl{
			{Kind: arch.KindPerson, Name: "lonely", Range: at(2)},
		},
	}}
}

func compileIgnoring(t *testing.T, src stubSource, ignore ...string) *CompileResult {
	t.Helper()
	res, err := CompileArchitecture(context.Background(), src, CompileRequest{BuildVersion: "1.0.0", Ignore: ignore})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestCompileIgnore(t *testing.T) {
	t.Parallel()
	base := compileIgnoring(t, ignoreModel())
	if !hasCode(base.Diags, arch.CodeMissingDocs) || !hasCode(base.Diags, arch.CodeOrphanElement) {
		t.Fatalf("fixture lacks its warnings: %v", codes(base.Diags))
	}

	perRun := compileIgnoring(t, ignoreModel(), arch.CodeMissingDocs)
	if hasCode(perRun.Diags, arch.CodeMissingDocs) || !hasCode(perRun.Diags, arch.CodeOrphanElement) || perRun.Ignored != 1 {
		t.Errorf("per-run ignore: %v, ignored %d", codes(perRun.Diags), perRun.Ignored)
	}
	inSource := compileIgnoring(t, ignoreModel(arch.CodeOrphanElement))
	if hasCode(inSource.Diags, arch.CodeOrphanElement) || !hasCode(inSource.Diags, arch.CodeMissingDocs) {
		t.Errorf("project ignore_warnings: %v", codes(inSource.Diags))
	}
	both := compileIgnoring(t, ignoreModel(arch.CodeOrphanElement), arch.CodeMissingDocs)
	if both.WarningCount() != 0 || both.ExitCode(true) != ExitSuccess || both.Ignored != 2 {
		t.Errorf("merged ignore: %v, exit %d, ignored %d", codes(both.Diags), both.ExitCode(true), both.Ignored)
	}
}

func TestCompileIgnoreInSource(t *testing.T) {
	t.Parallel()
	errCode := compileIgnoring(t, ignoreModel(arch.CodeSyntaxError))
	if !hasCode(errCode.Diags, arch.CodeInvalidAttributeValue) {
		t.Errorf("an error code in ignore_warnings: %v, want invalid_attribute_value", codes(errCode.Diags))
	}
	unknown := compileIgnoring(t, ignoreModel("no_such_code"))
	var named bool
	for _, d := range unknown.Diags {
		if d.Code == arch.CodeUnknownWarningCode && strings.Contains(d.Detail, "no_such_code") && d.Range == at(1) {
			named = true
		}
	}
	if !named || unknown.HasErrors() {
		t.Errorf("an unknown code: %v, want an unknown_warning_code warning naming it", codes(unknown.Diags))
	}
	quiet := compileIgnoring(t, ignoreModel("no_such_code", arch.CodeUnknownWarningCode))
	if hasCode(quiet.Diags, arch.CodeUnknownWarningCode) {
		t.Error("ignoring unknown_warning_code does not silence it")
	}
}

func TestCompileIgnoreRefusesPerRunErrors(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{arch.CodeSyntaxError, "no_such_code"} {
		_, err := CompileArchitecture(context.Background(), ignoreModel(), CompileRequest{Ignore: []string{bad}})
		if err == nil || !strings.Contains(err.Error(), bad) {
			t.Errorf("Ignore %q: %v, want a refusal naming it", bad, err)
		}
	}
}
