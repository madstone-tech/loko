package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryOptions carries one `loko query` invocation.
type QueryOptions struct {
	Root    string
	Format  string // text | json | toon
	Request usecases.QueryRequest
	Stdout  io.Writer
	Stderr  io.Writer
}

// runQueryWith answers a graph question with usecases.Query, the function the
// MCP query tool calls, so both front ends give the same answer (FR-029). It
// returns the process exit code.
func runQueryWith(ctx context.Context, opts QueryOptions) (int, error) {
	res, err := usecases.Query(ctx, newAuthoringDeps(opts.Root), opts.Request)
	if err != nil {
		return usecases.ExitErrors, err
	}
	switch {
	case len(res.Diags) > 0 && !res.OK:
		return writeQueryDiagnostics(opts, res)
	case res.Error != nil:
		_, err := fmt.Fprintln(opts.Stderr, queryErrorText(opts.Request, res.Error))
		return usecases.ExitErrors, err
	case opts.Format == formatJSON || opts.Format == formatTOON:
		return writeEncoded(opts, res)
	}
	return usecases.ExitSuccess, writeQueryText(opts.Stdout, res)
}

// writeQueryDiagnostics reports a project that does not compile, as validate does (FR-007).
func writeQueryDiagnostics(opts QueryOptions, res *usecases.QueryResult) (int, error) {
	renderer, err := hclsource.NewRendererForRoot(opts.Root, useColour(opts.Stderr))
	if err != nil {
		return usecases.ExitErrors, err
	}
	if _, _, err := renderer.Write(opts.Stderr, res.Diags); err != nil {
		return usecases.ExitErrors, err
	}
	return usecases.ExitErrors, nil
}

// queryErrorText phrases a query that could not be answered (FR-030).
func queryErrorText(req usecases.QueryRequest, e *usecases.ReadError) string {
	if e.Reason != "not_found" {
		return e.Detail
	}
	addr := req.Address
	if !strings.HasSuffix(e.Detail, " "+addr) {
		addr = req.To
	}
	msg := fmt.Sprintf("unknown address %q", addr)
	if len(e.Suggestions) > 0 {
		msg += ": did you mean " + strings.Join(e.Suggestions, " or ") + "?"
	}
	return msg
}

// writeEncoded prints the result as the MCP tool would return it, plus a
// final newline.
func writeEncoded(opts QueryOptions, res *usecases.QueryResult) (int, error) {
	enc := encoding.NewEncoder()
	encode := enc.EncodeJSON
	if opts.Format == formatTOON {
		encode = enc.EncodeTOON
	}
	b, err := encode(res)
	if err != nil {
		return usecases.ExitErrors, fmt.Errorf("encoding result: %w", err)
	}
	if _, err := fmt.Fprintf(opts.Stdout, "%s\n", b); err != nil {
		return usecases.ExitErrors, err
	}
	return usecases.ExitSuccess, nil
}

// writeQueryText prints a compact table for people.
func writeQueryText(w io.Writer, res *usecases.QueryResult) error {
	var rows strings.Builder
	switch res.Kind {
	case "path":
		if res.Found == nil || !*res.Found {
			rows.WriteString("no path\n")
		}
		for i, s := range res.Path {
			fmt.Fprintf(&rows, "%d.\t%s → %s\t(%s)\n", i+1, s.From, s.To, s.Relationship)
		}
	case "coupling":
		rows.WriteString("ADDRESS\tFAN-IN\tFAN-OUT\n")
		for _, c := range res.Coupling {
			fmt.Fprintf(&rows, "%s\t%d\t%d\n", c.Address, c.FanIn, c.FanOut)
		}
	default:
		rows.WriteString("ADDRESS\tKIND\tDISTANCE\n")
		for _, e := range res.Elements {
			fmt.Fprintf(&rows, "%s\t%s\t%d\n", e.Address, e.Kind, e.Distance)
		}
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := io.WriteString(tw, rows.String()); err != nil {
		return err
	}
	return tw.Flush()
}

const formatTOON = "toon"
