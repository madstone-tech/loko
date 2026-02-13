package cli

import (
	"fmt"
	"time"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Compile-time interface check
var _ usecases.ReportFormatter = (*ReportFormatter)(nil)

// ReportFormatter implements the usecases.ReportFormatter interface
// for CLI output formatting.
type ReportFormatter struct{}

// NewReportFormatter creates a new ReportFormatter instance.
func NewReportFormatter() *ReportFormatter {
	return &ReportFormatter{}
}

// PrintValidationReport prints validation errors to stdout.
func (f *ReportFormatter) PrintValidationReport(errors []usecases.ValidationError) {
	if len(errors) == 0 {
		fmt.Println("✓ No validation errors found!")
		return
	}

	for _, err := range errors {
		if err.Line > 0 {
			fmt.Printf("  [%s] %s:%d — %s\n", err.Code, err.Path, err.Line, err.Message)
		} else {
			fmt.Printf("  [%s] %s — %s\n", err.Code, err.Path, err.Message)
		}
	}

	fmt.Printf("\nTotal errors: %d\n", len(errors))
}

// PrintBuildReport prints build statistics to stdout.
func (f *ReportFormatter) PrintBuildReport(stats usecases.BuildStats) {
	fmt.Printf("Build complete (%s):\n", stats.Format)
	fmt.Printf("  Files generated: %d\n", stats.FilesGenerated)
	fmt.Printf("  Diagrams: %d\n", stats.DiagramCount)
	fmt.Printf("  Duration: %s\n", stats.Duration.Round(time.Millisecond))
}
