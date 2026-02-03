package pdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRenderer_IsAvailable(t *testing.T) {
	renderer := NewRenderer()

	// This test just verifies the method works without panicking
	// The result depends on whether veve-cli is installed
	available := renderer.IsAvailable()
	t.Logf("veve-cli available: %v", available)
}

func TestRenderer_RenderPDF_NotAvailable(t *testing.T) {
	// Create a renderer with no veve-cli path
	renderer := &Renderer{vevePath: ""}

	ctx := context.Background()
	err := renderer.RenderPDF(ctx, "input.html", "output.pdf")

	if err != ErrPDFNotAvailable {
		t.Errorf("Expected ErrPDFNotAvailable, got: %v", err)
	}
}

func TestRenderer_RenderPDF_MissingInputFile(t *testing.T) {
	renderer := NewRenderer()
	if !renderer.IsAvailable() {
		t.Skip("veve-cli not available")
	}

	ctx := context.Background()
	err := renderer.RenderPDF(ctx, "/nonexistent/file.html", "output.pdf")

	if err == nil {
		t.Error("Expected error for missing input file")
	}
}

func TestRenderer_RenderPDF_Integration(t *testing.T) {
	renderer := NewRenderer()
	if !renderer.IsAvailable() {
		t.Skip("veve-cli not available - skipping integration test")
	}

	// Create a temporary HTML file
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")
	pdfPath := filepath.Join(tmpDir, "output.pdf")

	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>Test Document</h1><p>This is a test.</p></body>
</html>`

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	ctx := context.Background()
	err := renderer.RenderPDF(ctx, htmlPath, pdfPath)

	if err != nil {
		t.Fatalf("RenderPDF failed: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Error("PDF file was not created")
	}
}

func TestRenderer_Version(t *testing.T) {
	renderer := NewRenderer()
	if !renderer.IsAvailable() {
		t.Skip("veve-cli not available")
	}

	version, err := renderer.Version()
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}

	t.Logf("veve-cli version: %s", version)
}

func TestRenderer_Version_NotAvailable(t *testing.T) {
	renderer := &Renderer{vevePath: ""}

	_, err := renderer.Version()

	if err != ErrPDFNotAvailable {
		t.Errorf("Expected ErrPDFNotAvailable, got: %v", err)
	}
}
