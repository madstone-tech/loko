package arch

import (
	"fmt"
	"slices"
	"strings"
)

// IgnorableCodes lists every warning and information code, in AllCodes order:
// the codes a project or a run may choose to ignore (feature 017, FR-011;
// feature 018, FR-014). Errors are never ignorable.
var IgnorableCodes = []string{
	CodeOrphanElement,
	CodeEmptySystem,
	CodeMissingDocs,
	CodeDocsNotFound,
	CodeUnboundInstance,
	CodeSelfRelationship,
	CodeViewEmpty,
	CodeViewShadowed,
	CodePlannedInstanceBound,
	CodeUnknownWarningCode,
	CodeDocsAmbiguous,
	CodeDocsRootRelative,
	CodeWideView,
}

// CheckIgnore refuses a code that is not ignorable: an error code, or one loko
// does not emit. The message lists the ignorable codes, so a typo is corrected
// rather than silently ignoring nothing.
func CheckIgnore(codes []string) error {
	for _, c := range codes {
		if slices.Contains(IgnorableCodes, c) {
			continue
		}
		why := "is not a code loko emits"
		if slices.Contains(AllCodes, c) {
			why = "is an error, and errors cannot be ignored"
		}
		return fmt.Errorf("%q %s; ignorable codes are: %s", c, why, strings.Join(IgnorableCodes, ", "))
	}
	return nil
}

// Without returns the diagnostics with warnings and information of the given
// codes removed, and how many were removed. Errors are kept whatever their
// code.
func (d Diagnostics) Without(codes []string) (Diagnostics, int) {
	if len(codes) == 0 {
		return d, 0
	}
	kept := make(Diagnostics, 0, len(d))
	for _, diag := range d {
		if diag.Severity != SeverityError && slices.Contains(codes, diag.Code) {
			continue
		}
		kept = append(kept, diag)
	}
	return kept, len(d) - len(kept)
}
