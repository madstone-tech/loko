package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// writeSummary prints diagnostics grouped by code (feature 017): one line per
// code, so a model with dozens of warnings still fits on one screen. Text goes
// to stderr like every diagnostic; JSON goes to stdout for scripts.
func writeSummary(opts ValidateOptions, res *usecases.CompileResult, groups []usecases.DiagnosticGroup) error {
	if opts.Format == formatJSON {
		out, err := json.MarshalIndent(map[string]any{
			"errors": res.ErrorCount(), "warnings": res.WarningCount(), "ignored": res.Ignored, "summary": groups,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding summary: %w", err)
		}
		_, err = fmt.Fprintf(opts.Stdout, "%s\n", out)
		return err
	}
	head := fmt.Sprintf("%d errors, %d warnings", res.ErrorCount(), res.WarningCount())
	if res.Ignored > 0 {
		head += fmt.Sprintf(" (%d ignored)", res.Ignored)
	}
	if len(groups) == 0 {
		_, err := fmt.Fprintln(opts.Stderr, "No diagnostics.")
		return err
	}
	// Laid out in memory, then written once, so the one write error is the
	// only one to check.
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "%s\n\n", head) // writes to a strings.Builder cannot fail
	for _, g := range groups {
		addrs := strings.Join(g.Addresses, ", ")
		if g.More > 0 {
			addrs += fmt.Sprintf(", … (+%d more)", g.More)
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", g.Severity, g.Code, g.Count, addrs)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprint(opts.Stderr, b.String())
	return err
}
