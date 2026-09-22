package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ExportFormat names an output encoding.
type ExportFormat string

const (
	ExportJSON ExportFormat = "json"
	ExportTOON ExportFormat = "toon"
)

// ValidExportFormat reports whether s names a supported encoding.
func ValidExportFormat(s string) bool {
	return ExportFormat(s) == ExportJSON || ExportFormat(s) == ExportTOON
}

// IREncoder serialises a compiled IR. Implemented in the encoding adapter.
type IREncoder interface {
	EncodeIR(ir *arch.IR, format ExportFormat) ([]byte, error)
}

// ExportRequest names what to export and how.
type ExportRequest struct {
	Root         string
	BuildVersion string
	Format       ExportFormat
}

// ExportResult carries the artefact, or the diagnostics that prevented one.
type ExportResult struct {
	// Artifact is nil when compilation reported errors.
	Artifact []byte
	Diags    arch.Diagnostics
	IR       *arch.IR
}

// ExportIR compiles the project and encodes the IR.
//
// No artefact is produced when compilation reports errors (FR-037). The
// caller is left with diagnostics and nothing to write, so a broken project
// cannot silently overwrite a good export with a partial one — which is the
// failure that would make a committed artefact untrustworthy.
func ExportIR(ctx context.Context, src ArchitectureSource, enc IREncoder,
	req ExportRequest) (*ExportResult, error) {

	if enc == nil {
		return nil, fmt.Errorf("export: no encoder configured")
	}
	if !ValidExportFormat(string(req.Format)) {
		return nil, fmt.Errorf("export: unsupported format %q: use json or toon", req.Format)
	}

	compiled, err := CompileArchitecture(ctx, src, CompileRequest{
		Root:         req.Root,
		BuildVersion: req.BuildVersion,
	})
	if err != nil {
		return nil, err
	}

	result := &ExportResult{Diags: compiled.Diags}
	if compiled.HasErrors() {
		return result, nil
	}

	result.IR = BuildIR(compiled.Model, compiled.Resolved)
	artifact, err := enc.EncodeIR(result.IR, req.Format)
	if err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}
	result.Artifact = artifact
	return result, nil
}

// ExitCode maps the result to a process exit code.
func (r ExportResult) ExitCode(strict bool) int { return r.Diags.ExitCode(strict) }
