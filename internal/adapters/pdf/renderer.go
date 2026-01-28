// Package pdf provides a PDF renderer adapter that shells out to veve-cli.
// It implements the PDFRenderer interface for generating PDF documents from HTML.
package pdf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrPDFNotAvailable indicates the veve-cli binary is not installed.
var ErrPDFNotAvailable = fmt.Errorf("veve-cli is not installed or not in PATH")

// Renderer implements the PDFRenderer interface by shelling out to veve-cli.
type Renderer struct {
	vevePath string // Path to veve-cli binary
}

// NewRenderer creates a new PDF renderer.
// It checks if veve-cli is available in PATH.
func NewRenderer() *Renderer {
	vevePath, _ := exec.LookPath("veve-cli")
	return &Renderer{
		vevePath: vevePath,
	}
}

// RenderPDF converts HTML to PDF using veve-cli.
// Returns ErrPDFNotAvailable if veve-cli is not installed.
func (r *Renderer) RenderPDF(ctx context.Context, htmlPath string, outputPath string) error {
	if !r.IsAvailable() {
		return ErrPDFNotAvailable
	}

	// Verify input file exists
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		return fmt.Errorf("HTML file does not exist: %s", htmlPath)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build veve-cli command
	// veve-cli html-to-pdf <input.html> <output.pdf>
	cmd := exec.CommandContext(ctx, r.vevePath, "html-to-pdf", htmlPath, outputPath)

	// Capture stderr for error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("veve-cli failed: %w\nOutput: %s", err, string(output))
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF file was not created: %s", outputPath)
	}

	return nil
}

// IsAvailable checks if the veve-cli binary is installed and accessible.
func (r *Renderer) IsAvailable() bool {
	return r.vevePath != ""
}

// Version returns the veve-cli version if available.
func (r *Renderer) Version() (string, error) {
	if !r.IsAvailable() {
		return "", ErrPDFNotAvailable
	}

	cmd := exec.Command(r.vevePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get veve-cli version: %w", err)
	}

	return string(output), nil
}
