package hclsource

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Renderer writes diagnostics for a human reader, with the offending line and
// a caret under the span.
//
// Showing the source line is most of what makes a compiler's errors usable.
// The renderer holds the file contents itself rather than reaching back into
// the parser, so rendering stays independent of how the diagnostics were
// produced — machine-readable output uses the same Diagnostics with no
// renderer at all.
type Renderer struct {
	// sources maps project-relative path to file contents.
	sources map[string][]byte
	// colour is disabled when the destination is not a terminal or NO_COLOR is
	// set, and always in machine-readable output.
	colour bool
	width  int
}

// NewRenderer builds a renderer over the files the parser read.
func NewRenderer(sources map[string][]byte, colour bool) *Renderer {
	return &Renderer{sources: sources, colour: colour, width: 78}
}

// NewRendererForRoot reads the discovered files so a caller that did not run
// the parser itself can still render snippets.
func NewRendererForRoot(root string, colour bool) (*Renderer, error) {
	files, _, err := Discover(root)
	if err != nil {
		return nil, err
	}
	sources := make(map[string][]byte, len(files))
	for _, f := range files {
		b, readErr := os.ReadFile(f.Abs)
		if readErr != nil {
			// A file we cannot read has already produced its own diagnostic;
			// rendering simply omits its snippet.
			continue
		}
		sources[f.Rel] = b
	}
	return NewRenderer(sources, colour), nil
}

// ANSI codes, used only when colour is enabled.
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiDim    = "\033[2m"
)

// Write renders every diagnostic in deterministic order and returns the number
// of errors and warnings written.
func (r *Renderer) Write(w io.Writer, diags arch.Diagnostics) (errors, warnings int, err error) {
	tw := &trackedWriter{w: w}
	for _, d := range diags.SortedForOutput() {
		r.writeOne(tw, d)
		if d.Severity == arch.SeverityError {
			errors++
			continue
		}
		warnings++
	}
	r.writeSummary(tw, errors, warnings)
	return errors, warnings, tw.err
}

func (r *Renderer) writeOne(w *trackedWriter, d arch.Diagnostic) {
	label, colour := "Error", ansiRed
	if d.Severity == arch.SeverityWarning {
		label, colour = "Warning", ansiYellow
	}

	w.printf("%s: %s\n", r.paint(label, colour+ansiBold), d.Summary)

	if !d.Range.IsZero() {
		w.printf("  on %s:\n", r.paint(d.Range.String(), ansiBold))
		r.writeSnippet(w, d.Range, colour)
	}
	if d.Detail != "" {
		for _, line := range wrap(d.Detail, r.width-2) {
			w.printf("  %s\n", line)
		}
	}
	for _, rel := range d.Related {
		msg := rel.Message
		if msg == "" {
			msg = "related"
		}
		w.printf("  %s\n", r.paint(fmt.Sprintf("%s: %s", msg, rel.Range), ansiDim))
	}
	w.println("")
}

// writeSnippet prints the offending line with a caret under the span.
func (r *Renderer) writeSnippet(w *trackedWriter, rng arch.SourceRange, colour string) {
	src, ok := r.sources[rng.File]
	if !ok || rng.StartLine < 1 {
		return
	}
	lines := strings.Split(string(src), "\n")
	if rng.StartLine > len(lines) {
		return
	}

	line := strings.ReplaceAll(lines[rng.StartLine-1], "\t", "    ")
	gutter := fmt.Sprintf("%d", rng.StartLine)
	w.printf("  %s | %s\n", gutter, line)

	// Only underline a span that lies on one line; a multi-line span gets the
	// start position alone, which is still precise enough to navigate to.
	start := rng.StartColumn
	end := rng.EndColumn
	if rng.EndLine != rng.StartLine || end <= start {
		end = start + 1
	}
	if start < 1 {
		return
	}
	pad := strings.Repeat(" ", len(gutter)+3+adjustForTabs(lines[rng.StartLine-1], start-1))
	caret := strings.Repeat("^", maxInt(1, end-start))
	w.printf("  %s%s\n", pad, r.paint(caret, colour))
}

// adjustForTabs accounts for tabs having been expanded to four spaces, so the
// caret lands under the right character in a tab-indented file.
func adjustForTabs(line string, col int) int {
	if col > len(line) {
		col = len(line)
	}
	width := 0
	for i := 0; i < col; i++ {
		if line[i] == '\t' {
			width += 4
			continue
		}
		width++
	}
	return width
}

func (r *Renderer) writeSummary(w *trackedWriter, errors, warnings int) {
	switch {
	case errors == 0 && warnings == 0:
		w.println(r.paint("No problems found.", ansiBold))
	default:
		w.printf("%s\n", r.paint(
			fmt.Sprintf("%s, %s", plural(errors, "error"), plural(warnings, "warning")), ansiBold))
	}
}

func (r *Renderer) paint(s, colour string) string {
	if !r.colour {
		return s
	}
	return colour + s + ansiReset
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// wrap breaks text on word boundaries at width columns.
func wrap(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var (
		out  []string
		line strings.Builder
	)
	for _, word := range words {
		if line.Len() > 0 && line.Len()+1+len(word) > width {
			out = append(out, line.String())
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteByte(' ')
		}
		line.WriteString(word)
	}
	if line.Len() > 0 {
		out = append(out, line.String())
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// trackedWriter records the first write error so the renderer can report a
// failed write once, at the end, instead of checking every call site or
// silently discarding it.
type trackedWriter struct {
	w   io.Writer
	err error
}

func (t *trackedWriter) printf(format string, args ...any) {
	if t.err != nil {
		return
	}
	_, t.err = fmt.Fprintf(t.w, format, args...)
}

func (t *trackedWriter) println(s string) {
	if t.err != nil {
		return
	}
	_, t.err = fmt.Fprintln(t.w, s)
}
