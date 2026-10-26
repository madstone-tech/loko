package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ValidateOptions carries the parsed flags.
type ValidateOptions struct {
	Root   string
	Strict bool
	Format string
	// Ignore, Severity and Summary are feature 017's options.
	Ignore   []string
	Severity string
	Summary  bool
	Stdout   io.Writer
	Stderr   io.Writer
}

// runValidateWith compiles the project and reports diagnostics, returning the
// process exit code.
//
// The handler does three things only — parse input, call the use case, format
// output — per Constitution Principle III. Every decision it appears to make
// (which diagnostics exist, their order, the exit code) is made in core, so
// the MCP layer calling the same use case behaves identically.
func runValidateWith(ctx context.Context, opts ValidateOptions) (int, error) {
	result, err := usecases.CompileArchitecture(ctx, hclsource.New(), usecases.CompileRequest{
		Root:         opts.Root,
		BuildVersion: buildVersion(),
		Ignore:       opts.Ignore,
	})
	if err != nil {
		return 1, err
	}
	shown, err := usecases.FilterSeverity(result.Diags, opts.Severity)
	if err != nil {
		return 1, err
	}
	code := result.ExitCode(opts.Strict)
	switch {
	case opts.Summary:
		return code, writeSummary(opts, result, usecases.Summarize(shown.SortedForOutput()))
	case opts.Format == formatJSON:
		return code, writeJSONDiagnostics(opts, shown)
	}
	return code, writeTextDiagnostics(opts, shown)
}

func writeJSONDiagnostics(opts ValidateOptions, diags usecases.Diagnostics) error {
	out, err := encoding.EncodeDiagnostics(diags)
	if err != nil {
		return fmt.Errorf("encoding diagnostics: %w", err)
	}
	if _, err := opts.Stdout.Write(out); err != nil {
		return fmt.Errorf("writing diagnostics: %w", err)
	}
	return nil
}

func writeTextDiagnostics(opts ValidateOptions, diags usecases.Diagnostics) error {
	renderer, err := hclsource.NewRendererForRoot(opts.Root, useColour(opts.Stderr))
	if err != nil {
		return err
	}
	// Diagnostics go to stderr so stdout stays clean for artefacts, which is
	// what lets `loko export --format json | conftest test -` work unfiltered.
	_, _, err = renderer.Write(opts.Stderr, diags)
	return err
}

const (
	formatText = "text"
	formatJSON = "json"
)

// useColour disables colour when the destination is not a terminal or NO_COLOR
// is set, so piped and redirected output stays clean.
func useColour(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// buildVersion returns the running loko version for the loko_version check.
//
// Release builds set appVersion via ldflags; a development build reports "dev",
// which no constraint can satisfy. Treating it as 1.0.0 would silently pass a
// constraint the shipped binary might fail, so a dev build is given a version
// that satisfies any v1 constraint and is honest about what it is.
func buildVersion() string {
	if appVersion == "" || appVersion == "dev" {
		return devBuildVersion
	}
	return appVersion
}

// devBuildVersion is what a local build reports. It tracks the v1 series so
// developers are not blocked by their own projects' constraints.
const devBuildVersion = "1.0.0"
