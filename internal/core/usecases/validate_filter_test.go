package usecases

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func warn(code, addr, file string, line int) arch.Diagnostic {
	return arch.Diagnostic{Severity: arch.SeverityWarning, Code: code, Summary: "s", Address: arch.Address(addr),
		Range: arch.SourceRange{File: file, StartLine: line, StartColumn: 3, EndLine: line, EndColumn: 9}}
}

func TestSummarize(t *testing.T) {
	t.Parallel()
	var d arch.Diagnostics
	for i := range 25 {
		d = append(d, warn(arch.CodeMissingDocs, fmt.Sprintf("container.c%02d", 24-i), "a.loko.hcl", i+1))
	}
	d = append(d, warn(arch.CodeOrphanElement, "container.o", "a.loko.hcl", 30),
		arch.Diagnostic{Severity: arch.SeverityError, Code: arch.CodeUnresolvedReference, Address: "container.x"})
	groups := Summarize(d)
	if len(groups) != 3 || groups[0].Code != arch.CodeUnresolvedReference || groups[1].Code != arch.CodeMissingDocs {
		t.Fatalf("groups %+v, want the error first, then warnings by code", groups)
	}
	docs := groups[1]
	if docs.Count != 25 || len(docs.Addresses) != 20 || docs.More != 5 || !slices.IsSorted(docs.Addresses) {
		t.Errorf("missing_docs group: count %d, %d addresses, more %d", docs.Count, len(docs.Addresses), docs.More)
	}
	if docs.Addresses[0] != "container.c00" {
		t.Errorf("addresses not sorted: %v", docs.Addresses[:3])
	}
	if got := Summarize(nil); len(got) != 0 {
		t.Errorf("Summarize(nil) = %v, want no groups", got)
	}
}

func TestFilterSeverity(t *testing.T) {
	t.Parallel()
	d := arch.Diagnostics{warn(arch.CodeMissingDocs, "system.s", "a", 1),
		{Severity: arch.SeverityError, Code: arch.CodeSyntaxError}}
	errs, err := FilterSeverity(d, "error")
	if err != nil || len(errs) != 1 || errs[0].Code != arch.CodeSyntaxError {
		t.Errorf("error filter: %v %v", errs, err)
	}
	if all, _ := FilterSeverity(d, ""); len(all) != 2 {
		t.Errorf("no filter dropped diagnostics: %v", all)
	}
	if _, err := FilterSeverity(d, "info"); err == nil || !strings.Contains(err.Error(), "warning") {
		t.Errorf("FilterSeverity accepted info, or its message does not name warning: %v", err)
	}
	withInfo := append(slices.Clone(d), arch.Diagnostic{Severity: arch.SeverityInfo, Code: arch.CodeWideView})
	shown, err := FilterSeverity(withInfo, "warning")
	if err != nil || len(shown) != 2 || slices.ContainsFunc(shown, func(x arch.Diagnostic) bool { return x.Severity == arch.SeverityInfo }) {
		t.Errorf("warning filter: %v %v", shown, err)
	}
}

func TestSummarizeOrdersBySeverity(t *testing.T) {
	t.Parallel()
	d := arch.Diagnostics{
		{Severity: arch.SeverityInfo, Code: arch.CodeDocsRootRelative},
		warn(arch.CodeOrphanElement, "container.o", "a", 1),
		{Severity: arch.SeverityInfo, Code: arch.CodeWideView},
		{Severity: arch.SeverityError, Code: arch.CodeUnresolvedReference},
		warn(arch.CodeMissingDocs, "container.m", "a", 2),
	}
	var got []string
	for _, g := range Summarize(d) {
		got = append(got, g.Code)
	}
	want := []string{arch.CodeUnresolvedReference, arch.CodeMissingDocs, arch.CodeOrphanElement,
		arch.CodeDocsRootRelative, arch.CodeWideView}
	if !slices.Equal(got, want) {
		t.Errorf("groups %v, want %v (errors, warnings, information)", got, want)
	}
}

func TestDiagnosticViews(t *testing.T) {
	t.Parallel()
	d := arch.Diagnostics{warn(arch.CodeMissingDocs, "system.s", "arch/a.loko.hcl", 12),
		{Severity: arch.SeverityError, Code: arch.CodeNoSourceFound, Summary: "none"}}
	brief := DiagnosticViews(d, false)
	if brief[0].At != "arch/a.loko.hcl:12" || brief[0].Range != nil || brief[0].Summary != "" {
		t.Errorf("compact: %+v", brief[0])
	}
	if brief[1].At != "" {
		t.Errorf("a diagnostic with no file has at %q", brief[1].At)
	}
	// Full views encode exactly as 1.0.2's diagnostics did.
	full, _ := json.Marshal(DiagnosticViews(d, true))
	old, _ := json.Marshal(d)
	if string(full) != string(old) {
		t.Errorf("full views differ from 1.0.2:\n%s\n%s", full, old)
	}
	compact, _ := json.Marshal(brief)
	if strings.Contains(string(compact), "startColumn") {
		t.Errorf("compact view carries a range: %s", compact)
	}
}
