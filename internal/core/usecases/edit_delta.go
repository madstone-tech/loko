package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// DiagnosticCounts counts diagnostics by severity. A count of 0 is omitted
// when encoded (feature 023, contract W-2).
type DiagnosticCounts struct {
	Errors   int `json:"errors,omitempty" toon:"errors,omitempty"`
	Warnings int `json:"warnings,omitempty" toon:"warnings,omitempty"`
	Info     int `json:"info,omitempty" toon:"info,omitempty"`
}

// add counts one diagnostic.
func (c *DiagnosticCounts) add(d arch.Diagnostic) {
	switch d.Severity {
	case arch.SeverityError:
		c.Errors++
	case arch.SeverityWarning:
		c.Warnings++
	default:
		c.Info++
	}
}

// orNil returns nil when every count is 0, so the encoded field is omitted.
func (c DiagnosticCounts) orNil() *DiagnosticCounts {
	if c == (DiagnosticCounts{}) {
		return nil
	}
	return &c
}

// deltaKey identifies a diagnostic across a change. Positions are not part of
// it: an inserted line moves every diagnostic below it (research R7).
func deltaKey(d arch.Diagnostic) string {
	return fmt.Sprintf("%s|%d|%s|%s|%s", d.Code, d.Severity, d.Address, d.Range.File, d.Detail)
}

// diagnosticDelta compares the diagnostics before and after a change.
// Diagnostics with equal keys pair in source order. A diagnostic in after
// with no partner is added, one in before with no partner is removed, and
// the matched pairs are counted by severity.
func diagnosticDelta(before, after arch.Diagnostics) (added, removed arch.Diagnostics, unchanged DiagnosticCounts) {
	pending := make(map[string][]int, len(before))
	for i, d := range before {
		k := deltaKey(d)
		pending[k] = append(pending[k], i)
	}
	matched := make([]bool, len(before))
	for _, d := range after {
		k := deltaKey(d)
		if idx := pending[k]; len(idx) > 0 {
			matched[idx[0]] = true
			pending[k] = idx[1:]
			unchanged.add(d)
			continue
		}
		added = append(added, d)
	}
	for i, d := range before {
		if !matched[i] {
			removed = append(removed, d)
		}
	}
	return added, removed, unchanged
}
