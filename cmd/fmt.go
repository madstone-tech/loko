package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// FmtOptions carries the parsed flags.
type FmtOptions struct {
	Root   string
	Check  bool
	Stdout io.Writer
	Stderr io.Writer
}

// runFmtWith formats or checks the project, returning the process exit code.
func runFmtWith(ctx context.Context, opts FmtOptions) (int, error) {
	result, err := usecases.FormatSources(ctx, hclsource.NewFormatter(), usecases.FormatRequest{
		Root:  opts.Root,
		Check: opts.Check,
	})
	if err != nil {
		return 1, err
	}

	if len(result.Diags) > 0 {
		renderer, rErr := hclsource.NewRendererForRoot(opts.Root, useColour(opts.Stderr))
		if rErr != nil {
			return 1, rErr
		}
		if _, _, wErr := renderer.Write(opts.Stderr, result.Diags); wErr != nil {
			return 1, wErr
		}
	}

	// The paths go to stdout in both modes: in check mode they are the answer,
	// and in write mode they are a record of what changed. Either way a script
	// can consume them.
	for _, path := range result.Changed {
		if _, err := fmt.Fprintln(opts.Stdout, path); err != nil {
			return 1, err
		}
	}

	if opts.Check && len(result.Changed) > 0 {
		if _, err := fmt.Fprintf(opts.Stderr,
			"\n%d file(s) are not canonically formatted. Run `loko fmt` to fix them.\n",
			len(result.Changed)); err != nil {
			return 1, err
		}
	}

	return result.ExitCode(), nil
}
