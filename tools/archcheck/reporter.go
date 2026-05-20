package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// WriteText writes a human-readable violation report to w. Each violation
// produces one line in the format defined by ci-step.contract.md §3.2. When
// annotateGitHub is true, a GitHub workflow annotation line is also emitted per
// violation. A summary line is always written last.
func WriteText(w io.Writer, report *Report, annotateGitHub bool) {
	// Sort violations: kind order (layer, file-size, function-size), then file ASC,
	// then line ASC, then subject ASC.
	sorted := sortedViolations(report.Violations)

	for _, v := range sorted {
		_, _ = fmt.Fprintln(w, v.Message)
		if annotateGitHub {
			_, _ = fmt.Fprintf(w, "::error file=%s,line=%d::[%s] %s: %d > %d\n",
				v.File, v.Line, v.Rule, v.Subject, v.Actual, v.Limit)
		}
	}

	_, _ = fmt.Fprintf(w, "archcheck: %d file(s) scanned, %d function(s) scanned, %d violation(s) found.\n",
		report.TotalFilesScanned, report.TotalFunctionsScanned, len(report.Violations))
}

// WriteJSON writes the report as indented JSON to w.
func WriteJSON(w io.Writer, report *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encoding report as JSON: %w", err)
	}
	return nil
}

// sortedViolations returns a copy of violations sorted by kind order (layer,
// file-size, function-size), then file ASC, line ASC, subject ASC.
func sortedViolations(violations []Violation) []Violation {
	out := make([]Violation, len(violations))
	copy(out, violations)
	kindRank := map[string]int{
		"layer":         0,
		"file-size":     1,
		"function-size": 2,
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Kind != b.Kind {
			ra, okA := kindRank[a.Kind]
			rb, okB := kindRank[b.Kind]
			if okA && okB {
				return ra < rb
			}
			if okA != okB {
				return okA
			}
			return a.Kind < b.Kind
		}
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Subject < b.Subject
	})
	return out
}
