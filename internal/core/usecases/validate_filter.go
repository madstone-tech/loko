package usecases

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// summaryAddresses caps the addresses listed per code, so a summary of a
// large model still fits on one screen (feature 017, R6).
const summaryAddresses = 20

// DiagnosticGroup is every diagnostic of one code.
type DiagnosticGroup struct {
	Code      string   `json:"code" toon:"code"`
	Severity  string   `json:"severity" toon:"severity"`
	Count     int      `json:"count" toon:"count"`
	Addresses []string `json:"addresses,omitempty" toon:"addresses,omitempty"`
	// More counts the addresses beyond the first summaryAddresses.
	More int `json:"more,omitempty" toon:"more,omitempty"`
}

// Summarize groups diagnostics by code: errors first, then by code, each with
// its sorted addresses.
func Summarize(d arch.Diagnostics) []DiagnosticGroup {
	byCode := map[string]*DiagnosticGroup{}
	for _, diag := range d {
		g, ok := byCode[diag.Code]
		if !ok {
			g = &DiagnosticGroup{Code: diag.Code, Severity: diag.Severity.String()}
			byCode[diag.Code] = g
		}
		g.Count++
		if diag.Address != "" {
			g.Addresses = append(g.Addresses, string(diag.Address))
		}
	}
	groups := make([]DiagnosticGroup, 0, len(byCode))
	for _, g := range byCode {
		slices.Sort(g.Addresses)
		g.Addresses = slices.Compact(g.Addresses)
		if n := len(g.Addresses); n > summaryAddresses {
			g.Addresses, g.More = g.Addresses[:summaryAddresses], n-summaryAddresses
		}
		groups = append(groups, *g)
	}
	slices.SortFunc(groups, func(a, b DiagnosticGroup) int {
		ae, be := a.Severity == arch.SeverityError.String(), b.Severity == arch.SeverityError.String()
		if ae != be {
			if ae {
				return -1
			}
			return 1
		}
		return cmp.Compare(a.Code, b.Code)
	})
	return groups
}

// FilterSeverity keeps only errors when severity is "error", and everything
// when it is empty. It changes what is shown, never the exit code.
func FilterSeverity(d arch.Diagnostics, severity string) (arch.Diagnostics, error) {
	switch severity {
	case "":
		return d, nil
	case arch.SeverityError.String():
		return slices.DeleteFunc(slices.Clone(d), func(x arch.Diagnostic) bool {
			return x.Severity != arch.SeverityError
		}), nil
	}
	return nil, fmt.Errorf("severity %q: use error, or omit it for every diagnostic", severity)
}

// DiagnosticView is a diagnostic as an MCP client sees it. The compact form
// gives the location as At (file:line) and leaves out Summary, which only
// restates Code; the full form gives Summary, Range and Related and encodes
// exactly as arch.Diagnostic does (feature 017, R7).
type DiagnosticView struct {
	Severity arch.Severity       `json:"severity" toon:"severity"`
	Code     string              `json:"code" toon:"code"`
	Summary  string              `json:"summary,omitempty" toon:"summary,omitempty"`
	Detail   string              `json:"detail,omitempty" toon:"detail,omitempty"`
	Address  arch.Address        `json:"address,omitempty" toon:"address,omitempty"`
	At       string              `json:"at,omitempty" toon:"at,omitempty"`
	Range    *arch.SourceRange   `json:"range,omitempty" toon:"range,omitempty"`
	Related  []arch.RelatedRange `json:"related,omitempty" toon:"related,omitempty"`
}

// DiagnosticViews returns the compact form of each diagnostic, or the full
// form when full is set. The result is never nil, so it encodes as a list.
func DiagnosticViews(d arch.Diagnostics, full bool) []DiagnosticView {
	out := make([]DiagnosticView, 0, len(d))
	for _, diag := range d {
		v := DiagnosticView{Severity: diag.Severity, Code: diag.Code, Detail: diag.Detail, Address: diag.Address}
		switch {
		case full:
			v.Summary, v.Range, v.Related = diag.Summary, &diag.Range, diag.Related
		case diag.Range.File != "" && diag.Range.StartLine > 0:
			v.At = fmt.Sprintf("%s:%d", diag.Range.File, diag.Range.StartLine)
		default:
			v.At = diag.Range.File
		}
		out = append(out, v)
	}
	return out
}

// ShapeDiagnostics sets the encoded diagnostics from the raw ones, compact or
// full. Front ends call it just before encoding a result.
func (r *EditResult) ShapeDiagnostics(full bool) { r.Diagnostics = shapedOrNil(r.Diags, full) }

// ShapeDiagnostics sets the encoded diagnostics; see EditResult.
func (r *DescribeResult) ShapeDiagnostics(full bool) { r.Diagnostics = shapedOrNil(r.Diags, full) }

// ShapeDiagnostics sets the encoded diagnostics; see EditResult.
func (r *QueryResult) ShapeDiagnostics(full bool) { r.Diagnostics = shapedOrNil(r.Diags, full) }

// shapedOrNil keeps an empty set omitted, as the raw field was.
func shapedOrNil(d arch.Diagnostics, full bool) []DiagnosticView {
	if len(d) == 0 {
		return nil
	}
	return DiagnosticViews(d, full)
}
