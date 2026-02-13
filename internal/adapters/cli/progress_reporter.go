package cli

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Compile-time interface check
var _ usecases.ProgressReporter = (*ProgressReporter)(nil)

// ProgressReporter implements ProgressReporter for console output.
type ProgressReporter struct{}

// NewProgressReporter creates a new ProgressReporter.
func NewProgressReporter() *ProgressReporter {
	return &ProgressReporter{}
}

// ReportProgress reports progress.
func (r *ProgressReporter) ReportProgress(step string, current int, total int, message string) {
	if total > 0 {
		percent := (current * 100) / total
		fmt.Printf("  [%3d%%] %s\n", percent, message)
	} else {
		fmt.Printf("  %s\n", message)
	}
}

// ReportError reports an error.
func (r *ProgressReporter) ReportError(err error) {
	fmt.Printf("  ✗ Error: %v\n", err)
}

// ReportSuccess reports success.
func (r *ProgressReporter) ReportSuccess(message string) {
	fmt.Printf("  ✓ %s\n", message)
}

// ReportInfo reports info.
func (r *ProgressReporter) ReportInfo(message string) {
	fmt.Printf("  ℹ %s\n", message)
}
