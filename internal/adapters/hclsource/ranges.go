package hclsource

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// converter turns HCL's own positions and diagnostics into the plain core
// types. It is the single crossing point for the FR-044 boundary: nothing
// downstream of here sees an hcl.Range or an hcl.Diagnostic.
type converter struct {
	// root is the project root, used to make every path project-relative.
	root string
}

// rng converts an hcl.Range. A nil range (HCL uses nil for diagnostics with no
// position) yields the zero SourceRange, which renders as "<project>".
func (c converter) rng(r *hcl.Range) arch.SourceRange {
	if r == nil {
		return arch.SourceRange{}
	}
	return arch.SourceRange{
		File:        arch.NormalizeFile(c.root, r.Filename),
		StartLine:   r.Start.Line,
		StartColumn: r.Start.Column,
		StartByte:   r.Start.Byte,
		EndLine:     r.End.Line,
		EndColumn:   r.End.Column,
		EndByte:     r.End.Byte,
	}
}

// diags converts HCL diagnostics, assigning each the given code.
//
// The code is supplied by the caller rather than inferred from the message:
// HCL's summaries are prose and subject to change upstream, and `code` is the
// stable identifier consumers match on.
func (c converter) diags(in hcl.Diagnostics, code string) arch.Diagnostics {
	if len(in) == 0 {
		return nil
	}
	out := make(arch.Diagnostics, 0, len(in))
	for _, d := range in {
		out = append(out, c.diag(d, code))
	}
	return out
}

func (c converter) diag(d *hcl.Diagnostic, code string) arch.Diagnostic {
	severity := arch.SeverityError
	if d.Severity == hcl.DiagWarning {
		severity = arch.SeverityWarning
	}
	out := arch.Diagnostic{
		Severity: severity,
		Code:     code,
		Summary:  d.Summary,
		Detail:   d.Detail,
		Range:    c.rng(d.Subject),
	}
	// HCL's Context is the wider construct the subject sits in — the whole
	// block for an attribute error. Keeping it as a related range means the
	// reader can find the declaration without losing the precise position.
	if d.Context != nil && d.Subject != nil && *d.Context != *d.Subject {
		out.Related = []arch.RelatedRange{{
			Message: "in this block",
			Range:   c.rng(d.Context),
		}}
	}
	return out
}

// errorf builds an error diagnostic at a converted range.
func (c converter) errorf(r *hcl.Range, code, summary, detail string) arch.Diagnostic {
	return arch.Diagnostic{
		Severity: arch.SeverityError,
		Code:     code,
		Summary:  summary,
		Detail:   detail,
		Range:    c.rng(r),
	}
}
