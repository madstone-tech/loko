package arch

import "testing"

func rng(file string, line, col, byteOff int) SourceRange {
	return SourceRange{
		File:        file,
		StartLine:   line,
		StartColumn: col,
		StartByte:   byteOff,
		EndLine:     line,
		EndColumn:   col + 1,
		EndByte:     byteOff + 1,
	}
}

// TestSortedForOutput covers FR-033: diagnostics must be emitted in a
// deterministic order for identical input, and the text and JSON renderings
// must agree on that order. Sort key is file, then start byte, then code.
func TestSortedForOutput(t *testing.T) {
	t.Parallel()

	in := Diagnostics{
		{Severity: SeverityWarning, Code: CodeOrphanElement, Range: rng("b.loko.hcl", 1, 1, 0)},
		{Severity: SeverityError, Code: CodeUnresolvedReference, Range: rng("a.loko.hcl", 9, 3, 200)},
		{Severity: SeverityError, Code: CodeUnknownBlock, Range: rng("a.loko.hcl", 2, 1, 10)},
		// Same file and offset: `code` breaks the tie so the order is total.
		{Severity: SeverityWarning, Code: CodeMissingDocs, Range: rng("a.loko.hcl", 2, 1, 10)},
	}

	got := in.SortedForOutput()

	want := []struct {
		file string
		code string
	}{
		{"a.loko.hcl", CodeMissingDocs},
		{"a.loko.hcl", CodeUnknownBlock},
		{"a.loko.hcl", CodeUnresolvedReference},
		{"b.loko.hcl", CodeOrphanElement},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d diagnostics, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Range.File != w.file || got[i].Code != w.code {
			t.Errorf("position %d: got %s/%s, want %s/%s",
				i, got[i].Range.File, got[i].Code, w.file, w.code)
		}
	}
}

// TestSortedForOutputDoesNotMutate matters because callers render the same
// Diagnostics twice (text to the terminal, JSON to a file) and must get the
// same sequence both times.
func TestSortedForOutputDoesNotMutate(t *testing.T) {
	t.Parallel()

	in := Diagnostics{
		{Severity: SeverityError, Code: CodeUnknownBlock, Range: rng("z.loko.hcl", 1, 1, 50)},
		{Severity: SeverityError, Code: CodeUnknownBlock, Range: rng("a.loko.hcl", 1, 1, 0)},
	}
	first := in[0].Range.File

	_ = in.SortedForOutput()

	if in[0].Range.File != first {
		t.Errorf("SortedForOutput mutated the receiver: in[0] is now %q, was %q",
			in[0].Range.File, first)
	}
}

// TestExitCode covers FR-038: exactly three exit codes. A fourth would be a
// breaking change for any CI pipeline branching on them.
func TestExitCode(t *testing.T) {
	t.Parallel()

	var (
		none     = Diagnostics{}
		warnOnly = Diagnostics{{Severity: SeverityWarning, Code: CodeMissingDocs}}
		withErr  = Diagnostics{{Severity: SeverityError, Code: CodeUnknownBlock}}
		both     = Diagnostics{
			{Severity: SeverityWarning, Code: CodeMissingDocs},
			{Severity: SeverityError, Code: CodeUnknownBlock},
		}
	)

	tests := []struct {
		name   string
		diags  Diagnostics
		strict bool
		want   int
	}{
		{"clean", none, false, ExitSuccess},
		{"clean strict", none, true, ExitSuccess},
		{"warnings lenient", warnOnly, false, ExitSuccess},
		{"warnings strict", warnOnly, true, ExitWarnings},
		{"errors lenient", withErr, false, ExitErrors},
		{"errors strict", withErr, true, ExitErrors},
		{"errors outrank warnings", both, true, ExitErrors},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.diags.ExitCode(tt.strict); got != tt.want {
				t.Errorf("ExitCode(strict=%v) = %d, want %d", tt.strict, got, tt.want)
			}
		})
	}
}

func TestHasErrorsAndCounts(t *testing.T) {
	t.Parallel()

	d := Diagnostics{
		{Severity: SeverityWarning, Code: CodeMissingDocs},
		{Severity: SeverityError, Code: CodeUnknownBlock},
		{Severity: SeverityWarning, Code: CodeOrphanElement},
	}

	if !d.HasErrors() {
		t.Error("HasErrors() = false, want true")
	}
	if got, want := d.CountBySeverity(SeverityError), 1; got != want {
		t.Errorf("errors = %d, want %d", got, want)
	}
	if got, want := d.CountBySeverity(SeverityWarning), 2; got != want {
		t.Errorf("warnings = %d, want %d", got, want)
	}

	var empty Diagnostics
	if empty.HasErrors() {
		t.Error("zero-value Diagnostics reports errors; the zero value must be usable")
	}
	if got := empty.ExitCode(true); got != ExitSuccess {
		t.Errorf("zero-value ExitCode(true) = %d, want %d", got, ExitSuccess)
	}
}

// TestAllCodesAreKnown guards the contract that `code` is the stable machine
// identifier consumers match on. Every constant must be listed in AllCodes so
// the rule-coverage test (T098) can enumerate them.
func TestAllCodesAreKnown(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool, len(AllCodes))
	for _, c := range AllCodes {
		if c == "" {
			t.Error("AllCodes contains an empty code")
		}
		if seen[c] {
			t.Errorf("AllCodes contains duplicate %q", c)
		}
		seen[c] = true
	}

	// The codes in contracts/diagnostics.schema.json that something emits.
	// Feature 014 reserves four render-stage codes in the schema and adds each
	// here in the task that first produces it.
	if got, want := len(AllCodes), 29; got != want {
		t.Errorf("len(AllCodes) = %d, want %d — keep it in step with contracts/diagnostics.schema.json", got, want)
	}

	for _, required := range []string{CodeSyntaxError, CodeNoSourceFound, CodeUnresolvedReference, CodeSelfRelationship} {
		if !seen[required] {
			t.Errorf("AllCodes is missing %q", required)
		}
	}
}

func TestSeverityString(t *testing.T) {
	t.Parallel()

	if got, want := SeverityError.String(), "error"; got != want {
		t.Errorf("SeverityError.String() = %q, want %q", got, want)
	}
	if got, want := SeverityWarning.String(), "warning"; got != want {
		t.Errorf("SeverityWarning.String() = %q, want %q", got, want)
	}
}
