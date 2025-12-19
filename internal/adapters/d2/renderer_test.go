package d2

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestNewRenderer tests renderer initialization.
func TestNewRenderer(t *testing.T) {
	r := NewRenderer()
	if r == nil {
		t.Error("NewRenderer returned nil")
	}
}

// TestIsAvailable tests d2 binary availability detection.
func TestIsAvailable(t *testing.T) {
	r := NewRenderer()
	// This test will pass if d2 is installed, skip if not
	available := r.IsAvailable()
	if !available {
		t.Skip("d2 binary not found in PATH, skipping availability test")
	}
}

// TestRenderDiagramEmptySource tests error handling for empty source.
func TestRenderDiagramEmptySource(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	_, err := r.RenderDiagram(ctx, "")
	if err == nil {
		t.Error("expected error for empty source, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' in error message, got: %v", err)
	}
}

// TestRenderDiagramWhitespaceOnly tests error handling for whitespace-only source.
func TestRenderDiagramWhitespaceOnly(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	_, err := r.RenderDiagram(ctx, "   \n   \n   ")
	if err == nil {
		t.Error("expected error for whitespace-only source, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' in error message, got: %v", err)
	}
}

// TestRenderDiagramSimple tests rendering a simple valid diagram.
func TestRenderDiagramSimple(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := "my-shape: {\n  shape: rectangle\n}"

	svg, err := r.RenderDiagram(ctx, d2Source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svg == "" {
		t.Error("expected non-empty SVG content")
	}

	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG content to contain <svg tag")
	}
}

// TestRenderDiagramWithConnections tests rendering a diagram with connections.
func TestRenderDiagramWithConnections(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := `
user: User
api: API Server {
  shape: rectangle
}
db: Database {
  shape: cylinder
}

user -> api: "HTTP"
api -> db: "SQL"
`

	svg, err := r.RenderDiagram(ctx, d2Source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svg == "" {
		t.Error("expected non-empty SVG content")
	}

	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG content to contain <svg tag")
	}
}

// TestRenderDiagramInvalidSyntax tests error handling for invalid D2 syntax.
func TestRenderDiagramInvalidSyntax(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	// Use an invalid shape name to trigger a compilation error
	d2Source := "a: {\n  shape: nonexistent_shape\n}"

	_, err := r.RenderDiagram(ctx, d2Source)
	if err == nil {
		t.Error("expected error for invalid syntax, got nil")
	}
	if !strings.Contains(err.Error(), "compilation failed") {
		t.Errorf("expected 'compilation failed' in error message, got: %v", err)
	}
}

// TestRenderDiagramWithTimeout tests timeout handling.
func TestRenderDiagramWithTimeout(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := "my-shape: {\n  shape: rectangle\n}"

	// Use a reasonable timeout (should succeed)
	svg, err := r.RenderDiagramWithTimeout(ctx, d2Source, 10)
	if err != nil {
		t.Fatalf("unexpected error with 10s timeout: %v", err)
	}

	if svg == "" {
		t.Error("expected non-empty SVG content")
	}
}

// TestRenderDiagramContextTimeout tests context timeout handling.
func TestRenderDiagramContextTimeout(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d2Source := "my-shape: {\n  shape: rect\n}"

	_, err := r.RenderDiagram(ctx, d2Source)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

// TestRenderDiagramWithTimeoutContextDeadline tests that existing context deadline is respected.
func TestRenderDiagramWithTimeoutContextDeadline(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	// Create a context with a deadline
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	d2Source := "my-shape: {\n  shape: rectangle\n}"

	svg, err := r.RenderDiagram(ctx, d2Source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svg == "" {
		t.Error("expected non-empty SVG content")
	}
}

// TestContentHash tests SHA256 hash computation.
func TestContentHash(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "simple_diagram",
			source: "my-shape: {\n  shape: rect\n}",
		},
		{
			name:   "empty_string",
			source: "",
		},
		{
			name:   "whitespace",
			source: "   \n   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := ContentHash(tt.source)
			if hash == "" {
				t.Error("expected non-empty hash")
			}

			// Hash should be consistent
			hash2 := ContentHash(tt.source)
			if hash != hash2 {
				t.Error("hash should be consistent for same input")
			}

			// Different inputs should produce different hashes
			if tt.source != "" {
				otherHash := ContentHash(tt.source + "x")
				if hash == otherHash {
					t.Error("different inputs should produce different hashes")
				}
			}
		})
	}
}

// TestRenderToFile tests rendering to a file.
func TestRenderToFile(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := "my-shape: {\n  shape: rectangle\n}"

	// Create a temporary directory
	tmpDir := t.TempDir()
	outputPath := tmpDir + "/test.svg"

	err := r.RenderToFile(ctx, d2Source, outputPath, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected non-empty file content")
	}

	if !strings.Contains(string(content), "<svg") {
		t.Error("expected SVG content in file")
	}
}

// TestRenderToFileInvalidSource tests error handling in RenderToFile.
func TestRenderToFileInvalidSource(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()
	outputPath := tmpDir + "/test.svg"

	err := r.RenderToFile(ctx, "", outputPath, 10)
	if err == nil {
		t.Error("expected error for empty source, got nil")
	}
}

// TestRenderToFileCreateDirectory tests that RenderToFile creates output directories.
func TestRenderToFileCreateDirectory(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := "my-shape: {\n  shape: rectangle\n}"

	tmpDir := t.TempDir()
	outputPath := tmpDir + "/nested/dir/test.svg"

	err := r.RenderToFile(ctx, d2Source, outputPath, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created in nested directory
	_, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
}

// TestRenderToWriter tests rendering to a writer.
func TestRenderToWriter(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := "my-shape: {\n  shape: rectangle\n}"

	var buf strings.Builder
	err := r.RenderToWriter(ctx, d2Source, &buf, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := buf.String()
	if content == "" {
		t.Error("expected non-empty content")
	}

	if !strings.Contains(content, "<svg") {
		t.Error("expected SVG content")
	}
}

// TestRenderToWriterInvalidSource tests error handling in RenderToWriter.
func TestRenderToWriterInvalidSource(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	var buf strings.Builder

	err := r.RenderToWriter(ctx, "", &buf, 10)
	if err == nil {
		t.Error("expected error for empty source, got nil")
	}
}

// TestRenderDiagramMultipleCalls tests that multiple renders work correctly.
func TestRenderDiagramMultipleCalls(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	sources := []string{
		"shape1: {\n  shape: rectangle\n}",
		"shape2: {\n  shape: circle\n}",
		"shape3: {\n  shape: diamond\n}",
	}

	for i, source := range sources {
		svg, err := r.RenderDiagram(ctx, source)
		if err != nil {
			t.Fatalf("render %d failed: %v", i, err)
		}

		if svg == "" {
			t.Errorf("render %d returned empty SVG", i)
		}

		if !strings.Contains(svg, "<svg") {
			t.Errorf("render %d missing SVG tag", i)
		}
	}
}

// TestRenderDiagramConcurrent tests concurrent rendering (thread safety).
func TestRenderDiagramConcurrent(t *testing.T) {
	r := NewRenderer()
	if !r.IsAvailable() {
		t.Skip("d2 binary not found in PATH")
	}

	ctx := context.Background()
	d2Source := "my-shape: {\n  shape: rectangle\n}"

	// Run multiple renders concurrently
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := r.RenderDiagram(ctx, d2Source)
			done <- err
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		err := <-done
		if err != nil {
			t.Fatalf("concurrent render %d failed: %v", i, err)
		}
	}
}
