package usecases

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ValidateResult is a compile's diagnostics, with exact source ranges, and the
// revision they describe (FR-006).
type ValidateResult struct {
	OK       bool             `json:"ok" toon:"ok"`
	Errors   int              `json:"errors" toon:"errors"`
	Warnings int              `json:"warnings" toon:"warnings"`
	Diags    arch.Diagnostics `json:"diagnostics" toon:"diagnostics"`
	Revision string           `json:"revision" toon:"revision"`
}

// Validate compiles the project and reports what the compiler found.
func Validate(ctx context.Context, deps AuthoringDeps) (*ValidateResult, error) {
	res, err := compileOverlay(ctx, deps, nil)
	if err != nil {
		return nil, err
	}
	rev, err := currentRevision(ctx, deps)
	if err != nil {
		return nil, err
	}
	diags := res.Diags.SortedForOutput()
	if diags == nil {
		diags = arch.Diagnostics{}
	}
	return &ValidateResult{
		OK:       !res.HasErrors(),
		Errors:   res.ErrorCount(),
		Warnings: res.WarningCount(),
		Diags:    diags,
		Revision: rev,
	}, nil
}
