package usecases

import "context"

// ValidateRequest selects what validate reports (feature 017). The zero
// value reports every diagnostic, compactly.
type ValidateRequest struct {
	// Ignore lists warning codes to leave out, on top of ignore_warnings.
	Ignore []string
	// Severity is "" for every diagnostic, or "error".
	Severity string
	// Summary groups diagnostics by code instead of listing them.
	Summary bool
	// FullRanges gives each diagnostic's full source range instead of file:line.
	FullRanges bool
}

// ValidateResult is a compile's diagnostics and the revision they describe
// (FR-006). Errors and Warnings count what is left after Ignore, before
// Severity. With Summary, Diags is empty and Summary holds the groups.
type ValidateResult struct {
	OK       bool              `json:"ok" toon:"ok"`
	Errors   int               `json:"errors" toon:"errors"`
	Warnings int               `json:"warnings" toon:"warnings"`
	Ignored  int               `json:"ignored,omitempty" toon:"ignored,omitempty"`
	Diags    []DiagnosticView  `json:"diagnostics" toon:"diagnostics"`
	Summary  []DiagnosticGroup `json:"summary,omitempty" toon:"summary,omitempty"`
	Revision string            `json:"revision" toon:"revision"`
}

// Validate compiles the project and reports what the compiler found.
func Validate(ctx context.Context, deps AuthoringDeps, req ValidateRequest) (*ValidateResult, error) {
	res, err := CompileArchitecture(ctx, overlaySource{src: deps.Source},
		CompileRequest{Root: deps.Root, BuildVersion: deps.BuildVersion, Ignore: req.Ignore})
	if err != nil {
		return nil, err
	}
	shown, err := FilterSeverity(res.Diags.SortedForOutput(), req.Severity)
	if err != nil {
		return nil, err
	}
	rev, err := currentRevision(ctx, deps)
	if err != nil {
		return nil, err
	}
	out := &ValidateResult{
		OK:       !res.HasErrors(),
		Errors:   res.ErrorCount(),
		Warnings: res.WarningCount(),
		Ignored:  res.Ignored,
		Diags:    []DiagnosticView{},
		Revision: rev,
	}
	if req.Summary {
		out.Summary = Summarize(shown)
	} else {
		out.Diags = DiagnosticViews(shown, req.FullRanges)
	}
	return out, nil
}
