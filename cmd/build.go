package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// BuildOptions carries the parsed flags.
type BuildOptions struct {
	Root    string
	Formats []string
	Out     string
	Strict  bool
	Stdout  io.Writer
	Stderr  io.Writer
}

// runBuildWith renders the project into the output directory and returns the
// process exit code. It parses input, calls usecases.Build, and formats
// output — nothing else (Principle III, FR-039).
func runBuildWith(ctx context.Context, opts BuildOptions) (int, error) {
	deps := newBuildDeps()
	formats := opts.Formats
	if len(formats) == 0 {
		formats = usecases.SupportedFormats(deps.Backends)
	}
	res, err := usecases.Build(ctx, deps, usecases.BuildRequest{
		Root:         opts.Root,
		BuildVersion: buildVersion(),
		Formats:      formats,
		OutDir:       buildOutDir(opts),
	})
	if err != nil {
		return usecases.ExitErrors, err
	}
	if err := writeBuildDiagnostics(opts, res); err != nil {
		return usecases.ExitErrors, err
	}
	if res.ErrorCount() == 0 {
		if err := printBuildSummary(opts.Stdout, buildOutDir(opts), res); err != nil {
			return usecases.ExitErrors, err
		}
	}
	return res.ExitCode(opts.Strict), nil
}

// buildOutDir: an explicit --out is used as typed, relative to the working
// directory as `loko export --out` is; the default lives in the project.
func buildOutDir(opts BuildOptions) string {
	if opts.Out != "" {
		return opts.Out
	}
	return filepath.Join(opts.Root, "dist")
}

func writeBuildDiagnostics(opts BuildOptions, res *usecases.BuildResult) error {
	if len(res.Diags) == 0 {
		return nil
	}
	renderer, err := hclsource.NewRendererForRoot(opts.Root, useColour(opts.Stderr))
	if err != nil {
		return err
	}
	_, _, err = renderer.Write(opts.Stderr, res.Diags)
	return err
}

func printBuildSummary(w io.Writer, out string, res *usecases.BuildResult) error {
	var b strings.Builder
	if res.NothingToDraw {
		b.WriteString("nothing to draw: the architecture declares no elements\n")
	} else {
		for _, f := range res.AddedFormats {
			fmt.Fprintf(&b, "added %s (required by %s)\n", f, requiredBy(f, res.Formats))
		}
		fmt.Fprintf(&b, "built %s: %d written, %d unchanged, %d removed\n",
			out, len(res.Report.Written), len(res.Report.Unchanged), len(res.Report.Removed))
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		return fmt.Errorf("writing summary: %w", err)
	}
	return nil
}

// requiredBy names the requested formats that pulled in f.
func requiredBy(f string, formats []string) string {
	var by []string
	for _, g := range formats {
		if g != f && usecases.FormatRequires(g, f) {
			by = append(by, g)
		}
	}
	return strings.Join(by, ", ")
}
