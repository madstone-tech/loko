package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ExportOptions carries the parsed flags.
type ExportOptions struct {
	Root   string
	Format string
	Out    string
	Stdout io.Writer
	Stderr io.Writer
}

// runExportWith compiles and writes the IR, returning the process exit code.
func runExportWith(ctx context.Context, opts ExportOptions) (int, error) {
	result, err := usecases.ExportIR(ctx, hclsource.New(), encoding.NewEncoder(),
		usecases.ExportRequest{
			Root:         opts.Root,
			BuildVersion: buildVersion(),
			Format:       usecases.ExportFormat(opts.Format),
		})
	if err != nil {
		return 1, err
	}

	// No artefact when compilation reported errors (FR-037). Any --out file is
	// left exactly as it was rather than truncated, so a broken project cannot
	// destroy the last good export.
	if result.Artifact == nil {
		if renderErr := renderDiagnostics(opts, result); renderErr != nil {
			return 1, renderErr
		}
		return result.ExitCode(false), nil
	}

	if err := writeArtifact(opts, result.Artifact); err != nil {
		return 1, err
	}
	// Warnings go to stderr even on success, so stdout carries only the
	// artefact and `loko export --format json | conftest test -` works.
	if result.Diags != nil {
		if renderErr := renderDiagnostics(opts, result); renderErr != nil {
			return 1, renderErr
		}
	}
	return result.ExitCode(false), nil
}

func writeArtifact(opts ExportOptions, artifact []byte) error {
	if opts.Out == "" {
		if _, err := opts.Stdout.Write(artifact); err != nil {
			return fmt.Errorf("writing artefact: %w", err)
		}
		return nil
	}
	if dir := filepath.Dir(opts.Out); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(opts.Out, artifact, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", opts.Out, err)
	}
	return nil
}

func renderDiagnostics(opts ExportOptions, result *usecases.ExportResult) error {
	if len(result.Diags) == 0 {
		return nil
	}
	renderer, err := hclsource.NewRendererForRoot(opts.Root, useColour(opts.Stderr))
	if err != nil {
		return err
	}
	_, _, writeErr := renderer.Write(opts.Stderr, result.Diags)
	return writeErr
}
