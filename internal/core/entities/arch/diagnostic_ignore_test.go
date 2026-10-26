package arch

import (
	"slices"
	"strings"
	"testing"
)

func TestWarningCodes(t *testing.T) {
	t.Parallel()
	last := -1
	for _, c := range WarningCodes {
		i := slices.Index(AllCodes, c)
		if i < 0 {
			t.Errorf("warning code %q is not in AllCodes", c)
		}
		if i < last {
			t.Errorf("WarningCodes is not in AllCodes order at %q", c)
		}
		last = i
	}
	for _, c := range []string{CodeMissingDocs, CodeUnboundInstance, CodePlannedInstanceBound, CodeUnknownWarningCode} {
		if !slices.Contains(WarningCodes, c) {
			t.Errorf("WarningCodes lacks %q", c)
		}
	}
	for _, c := range []string{CodeSyntaxError, CodeUnresolvedReference, CodeInvalidAttributeValue} {
		if slices.Contains(WarningCodes, c) {
			t.Errorf("WarningCodes contains error code %q", c)
		}
	}
}

func TestCheckIgnore(t *testing.T) {
	t.Parallel()
	if err := CheckIgnore([]string{CodeMissingDocs, CodeOrphanElement}); err != nil {
		t.Errorf("warning codes refused: %v", err)
	}
	for _, bad := range []string{CodeSyntaxError, "no_such_code"} {
		err := CheckIgnore([]string{CodeMissingDocs, bad})
		if err == nil || !strings.Contains(err.Error(), bad) || !strings.Contains(err.Error(), CodeOrphanElement) {
			t.Errorf("CheckIgnore(%q) = %v, want an error naming it and listing the warning codes", bad, err)
		}
	}
}

func TestDiagnosticsWithout(t *testing.T) {
	t.Parallel()
	d := Diagnostics{
		{Severity: SeverityWarning, Code: CodeMissingDocs},
		{Severity: SeverityError, Code: CodeSyntaxError},
		{Severity: SeverityWarning, Code: CodeOrphanElement},
		{Severity: SeverityWarning, Code: CodeMissingDocs},
	}
	kept, ignored := d.Without([]string{CodeMissingDocs, CodeSyntaxError})
	if ignored != 2 || len(kept) != 2 || kept[0].Code != CodeSyntaxError || kept[1].Code != CodeOrphanElement {
		t.Errorf("kept %v, ignored %d", kept, ignored)
	}
	if kept, n := d.Without(nil); n != 0 || len(kept) != len(d) {
		t.Errorf("Without(nil) removed %d", n)
	}
}
