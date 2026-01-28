// Package ui provides styled terminal output using lipgloss.
// It implements consistent formatting for CLI messages, errors, and progress.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorPrimary   = lipgloss.Color("#2563eb")
	colorSuccess   = lipgloss.Color("#10b981")
	colorWarning   = lipgloss.Color("#f59e0b")
	colorError     = lipgloss.Color("#ef4444")
	colorMuted     = lipgloss.Color("#6b7280")
	colorHighlight = lipgloss.Color("#8b5cf6")
)

// Styles
var (
	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(colorHighlight)

	// Box styles
	InfoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	ErrorBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorError).
		Padding(0, 1)

	SuccessBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSuccess).
		Padding(0, 1)
)

// Output handles styled terminal output.
type Output struct {
	writer    io.Writer
	errWriter io.Writer
	verbose   bool
}

// NewOutput creates a new Output with default writers.
func NewOutput() *Output {
	return &Output{
		writer:    os.Stdout,
		errWriter: os.Stderr,
		verbose:   false,
	}
}

// WithVerbose enables verbose output.
func (o *Output) WithVerbose(verbose bool) *Output {
	o.verbose = verbose
	return o
}

// WithWriter sets the output writer.
func (o *Output) WithWriter(w io.Writer) *Output {
	o.writer = w
	return o
}

// WithErrWriter sets the error writer.
func (o *Output) WithErrWriter(w io.Writer) *Output {
	o.errWriter = w
	return o
}

// Title prints a title message.
func (o *Output) Title(msg string) {
	fmt.Fprintln(o.writer, TitleStyle.Render(msg))
}

// Subtitle prints a subtitle message.
func (o *Output) Subtitle(msg string) {
	fmt.Fprintln(o.writer, SubtitleStyle.Render(msg))
}

// Success prints a success message with checkmark.
func (o *Output) Success(msg string) {
	fmt.Fprintln(o.writer, SuccessStyle.Render("✓ "+msg))
}

// Warning prints a warning message.
func (o *Output) Warning(msg string) {
	fmt.Fprintln(o.errWriter, WarningStyle.Render("⚠ "+msg))
}

// Error prints an error message.
func (o *Output) Error(msg string) {
	fmt.Fprintln(o.errWriter, ErrorStyle.Render("✗ "+msg))
}

// ErrorWithDetails prints an error with additional details.
func (o *Output) ErrorWithDetails(msg string, details string) {
	fmt.Fprintln(o.errWriter, ErrorStyle.Render("✗ "+msg))
	if details != "" {
		fmt.Fprintln(o.errWriter, MutedStyle.Render("  "+details))
	}
}

// Info prints an info message.
func (o *Output) Info(msg string) {
	fmt.Fprintln(o.writer, "ℹ "+msg)
}

// Debug prints a debug message (only in verbose mode).
func (o *Output) Debug(msg string) {
	if o.verbose {
		fmt.Fprintln(o.writer, MutedStyle.Render("› "+msg))
	}
}

// Progress prints a progress message with percentage.
func (o *Output) Progress(current, total int, msg string) {
	if total <= 0 {
		fmt.Fprintf(o.writer, "  %s\n", msg)
		return
	}
	percent := (current * 100) / total
	bar := o.renderProgressBar(percent)
	fmt.Fprintf(o.writer, "  %s %3d%% %s\n", bar, percent, msg)
}

// renderProgressBar creates a visual progress bar.
func (o *Output) renderProgressBar(percent int) string {
	width := 20
	filled := (percent * width) / 100
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return MutedStyle.Render("[") + SuccessStyle.Render(bar[:filled]) + MutedStyle.Render(bar[filled:]) + MutedStyle.Render("]")
}

// List prints a list of items.
func (o *Output) List(items []string) {
	for _, item := range items {
		fmt.Fprintln(o.writer, "  • "+item)
	}
}

// Table prints a simple table.
func (o *Output) Table(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	headerLine := ""
	separatorLine := ""
	for i, h := range headers {
		headerLine += fmt.Sprintf("%-*s  ", widths[i], h)
		separatorLine += strings.Repeat("─", widths[i]) + "  "
	}
	fmt.Fprintln(o.writer, TitleStyle.Render(headerLine))
	fmt.Fprintln(o.writer, MutedStyle.Render(separatorLine))

	// Print rows
	for _, row := range rows {
		line := ""
		for i, cell := range row {
			if i < len(widths) {
				line += fmt.Sprintf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(o.writer, line)
	}
}

// Box prints a message in a box.
func (o *Output) Box(msg string) {
	fmt.Fprintln(o.writer, InfoBox.Render(msg))
}

// ErrorBox prints an error in a box.
func (o *Output) ErrorBoxMsg(msg string) {
	fmt.Fprintln(o.errWriter, ErrorBox.Render(msg))
}

// SuccessBox prints a success message in a box.
func (o *Output) SuccessBoxMsg(msg string) {
	fmt.Fprintln(o.writer, SuccessBox.Render(msg))
}

// Divider prints a horizontal divider.
func (o *Output) Divider() {
	fmt.Fprintln(o.writer, MutedStyle.Render(strings.Repeat("─", 40)))
}

// Newline prints a blank line.
func (o *Output) Newline() {
	fmt.Fprintln(o.writer)
}

// KeyValue prints a key-value pair.
func (o *Output) KeyValue(key, value string) {
	fmt.Fprintf(o.writer, "%s: %s\n", MutedStyle.Render(key), value)
}

// Highlight prints highlighted text.
func (o *Output) Highlight(msg string) {
	fmt.Fprintln(o.writer, HighlightStyle.Render(msg))
}

// FormatError formats an error message for display.
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	return ErrorStyle.Render("Error: " + err.Error())
}

// FormatSuccess formats a success message.
func FormatSuccess(msg string) string {
	return SuccessStyle.Render("✓ " + msg)
}

// FormatWarning formats a warning message.
func FormatWarning(msg string) string {
	return WarningStyle.Render("⚠ " + msg)
}
