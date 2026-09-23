package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// FormatRequest names what to format and in which mode.
type FormatRequest struct {
	Root string
	// Check reports which files are not canonical without writing anything.
	Check bool
}

// FormatResult lists the files that changed, or would change in check mode.
type FormatResult struct {
	// Changed holds project-relative paths, sorted.
	Changed []string
	Diags   arch.Diagnostics
	Checked bool
}

// ExitCode maps the result to a process exit code.
//
// Check mode reuses the "errors present" code rather than introducing a fourth
// (FR-035a, FR-038): a CI pipeline branching on exit codes should not have to
// learn a new one just because the formatter grew a flag.
func (r FormatResult) ExitCode() int {
	if r.Diags.HasErrors() {
		return arch.ExitErrors
	}
	if r.Checked && len(r.Changed) > 0 {
		return arch.ExitErrors
	}
	return arch.ExitSuccess
}

// FormatSources canonicalises the project's source files (FR-035).
//
// A file that does not parse is reported and left byte-unchanged; the command
// does not partially format a project it cannot fully read.
func FormatSources(ctx context.Context, f SourceFormatter, req FormatRequest) (*FormatResult, error) {
	if f == nil {
		return nil, fmt.Errorf("format: no formatter configured")
	}

	var (
		changed []string
		diags   arch.Diagnostics
		err     error
	)
	if req.Check {
		changed, diags, err = f.Check(ctx, req.Root)
	} else {
		changed, diags, err = f.Format(ctx, req.Root)
	}
	if err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}

	return &FormatResult{Changed: changed, Diags: diags, Checked: req.Check}, nil
}
