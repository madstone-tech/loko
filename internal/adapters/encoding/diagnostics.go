package encoding

import (
	"encoding/json"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// DiagnosticsSchemaVersion is the shape of machine-readable diagnostic output.
// It is versioned separately from the IR export: the two are consumed by
// different tools and change for different reasons.
const DiagnosticsSchemaVersion = 1

// diagnosticsDoc mirrors contracts/diagnostics.schema.json.
//
// A distinct wire type rather than tagging the entity: the entity is free to
// gain fields without silently widening a published contract, and `severity`
// serialises as its lowercase name here rather than as the entity's integer.
type diagnosticsDoc struct {
	SchemaVersion int              `json:"schemaVersion"`
	Diagnostics   []diagnosticWire `json:"diagnostics"`
	Summary       summaryWire      `json:"summary"`
}

type diagnosticWire struct {
	Severity string        `json:"severity"`
	Code     string        `json:"code"`
	Summary  string        `json:"summary"`
	Detail   string        `json:"detail,omitempty"`
	Address  string        `json:"address,omitempty"`
	Range    rangeWire     `json:"range"`
	Related  []relatedWire `json:"related,omitempty"`
}

type relatedWire struct {
	Message string    `json:"message,omitempty"`
	Range   rangeWire `json:"range"`
}

type rangeWire struct {
	File        string `json:"file"`
	StartLine   int    `json:"startLine"`
	StartColumn int    `json:"startColumn"`
	StartByte   int    `json:"startByte,omitempty"`
	EndLine     int    `json:"endLine"`
	EndColumn   int    `json:"endColumn"`
	EndByte     int    `json:"endByte,omitempty"`
}

type summaryWire struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

// EncodeDiagnostics renders diagnostics as JSON conforming to
// contracts/diagnostics.schema.json (FR-030a, FR-034a).
//
// Output order is the deterministic order of FR-033, identical to the text
// rendering, so a reader comparing the two sees the same sequence.
func EncodeDiagnostics(diags arch.Diagnostics) ([]byte, error) {
	sorted := diags.SortedForOutput()

	doc := diagnosticsDoc{
		SchemaVersion: DiagnosticsSchemaVersion,
		// Non-nil so an empty run emits [] rather than null: a consumer should
		// not have to special-case the clean case.
		Diagnostics: make([]diagnosticWire, 0, len(sorted)),
		Summary: summaryWire{
			Errors:   diags.CountBySeverity(arch.SeverityError),
			Warnings: diags.CountBySeverity(arch.SeverityWarning),
		},
	}

	for _, d := range sorted {
		w := diagnosticWire{
			Severity: d.Severity.String(),
			Code:     d.Code,
			Summary:  d.Summary,
			Detail:   d.Detail,
			Address:  string(d.Address),
			Range:    toRangeWire(d.Range),
		}
		for _, rel := range d.Related {
			w.Related = append(w.Related, relatedWire{
				Message: rel.Message,
				Range:   toRangeWire(rel.Range),
			})
		}
		doc.Diagnostics = append(doc.Diagnostics, w)
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func toRangeWire(r arch.SourceRange) rangeWire {
	return rangeWire{
		File:        r.File,
		StartLine:   r.StartLine,
		StartColumn: r.StartColumn,
		StartByte:   r.StartByte,
		EndLine:     r.EndLine,
		EndColumn:   r.EndColumn,
		EndByte:     r.EndByte,
	}
}
