// Package d2 provides a diagram renderer adapter that shells out to the d2 CLI.
// The d2 binary must be installed and available in the system PATH.
package d2

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Renderer implements the DiagramRenderer port by shelling out to the d2 CLI.
// It handles D2 source compilation to SVG with timeout support and graceful
// degradation if the d2 binary is not available.
type Renderer struct {
	d2Path string // Path to the d2 binary
	cache  map[string]string
	mu     sync.RWMutex
}

// NewRenderer creates a new D2 renderer.
// It attempts to locate the d2 binary in the system PATH.
func NewRenderer() *Renderer {
	d2Path, _ := exec.LookPath("d2")
	return &Renderer{
		d2Path: d2Path,
		cache:  make(map[string]string),
	}
}

// IsAvailable checks if the d2 binary is installed and accessible.
func (r *Renderer) IsAvailable() bool {
	return r.d2Path != ""
}

// RenderDiagram compiles D2 source code to SVG.
// Returns SVG content or error if d2 binary missing or compilation fails.
// Uses a default timeout of 30 seconds.
func (r *Renderer) RenderDiagram(ctx context.Context, d2Source string) (string, error) {
	return r.RenderDiagramWithTimeout(ctx, d2Source, 30)
}

// RenderDiagramWithTimeout compiles D2 source code to SVG with a specified timeout.
// timeoutSec specifies the maximum duration in seconds.
// Returns SVG content or error if d2 binary missing, compilation fails, or timeout occurs.
func (r *Renderer) RenderDiagramWithTimeout(ctx context.Context, d2Source string, timeoutSec int) (string, error) {
	// Validate input
	if d2Source == "" {
		return "", fmt.Errorf("d2 source cannot be empty")
	}

	trimmed := strings.TrimSpace(d2Source)
	if trimmed == "" {
		return "", fmt.Errorf("d2 source cannot be empty or whitespace-only")
	}

	// Check if d2 is available
	if !r.IsAvailable() {
		return "", fmt.Errorf("d2 binary not found in PATH")
	}

	// Check cache before rendering
	hash := ContentHash(d2Source)
	r.mu.RLock()
	if cached, ok := r.cache[hash]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
		defer cancel()
	}

	// Create temporary output file with unique name (safe for concurrent use)
	tmpFile, err := os.CreateTemp("", "loko-diagram-*.svg")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// Build the d2 command
	// d2 reads from stdin (-) and writes to the output file
	// Theme 0 = Neutral Default, Layout elk = ELK graph layout
	cmd := exec.CommandContext(ctx, r.d2Path,
		"--layout", "elk",
		"--theme", "0",
		"-",
		tmpPath,
	)

	// Pass D2 source via stdin
	cmd.Stdin = strings.NewReader(d2Source)

	// Capture stderr for error messages
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return "", fmt.Errorf("d2 compilation failed: %w\nstderr: %s", err, errMsg)
		}
		return "", fmt.Errorf("d2 compilation failed: %w", err)
	}

	// Read the rendered SVG
	svgContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read rendered SVG: %w", err)
	}

	// Store in cache for future use
	r.mu.Lock()
	r.cache[hash] = string(svgContent)
	r.mu.Unlock()

	return string(svgContent), nil
}

// ClearCache removes all cached diagram renders.
func (r *Renderer) ClearCache() {
	r.mu.Lock()
	r.cache = make(map[string]string)
	r.mu.Unlock()
}

// ContentHash computes the SHA256 hash of the D2 source code.
// This is used for cache invalidation in incremental builds.
func ContentHash(d2Source string) string {
	hash := sha256.Sum256([]byte(d2Source))
	return fmt.Sprintf("%x", hash)
}

// RenderToFile renders D2 source code directly to a file.
// This is a convenience method for writing SVG output to disk.
func (r *Renderer) RenderToFile(ctx context.Context, d2Source string, outputPath string, timeoutSec int) error {
	svgContent, err := r.RenderDiagramWithTimeout(ctx, d2Source, timeoutSec)
	if err != nil {
		return fmt.Errorf("failed to render diagram: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write SVG to file
	if err := os.WriteFile(outputPath, []byte(svgContent), 0644); err != nil {
		return fmt.Errorf("failed to write SVG file: %w", err)
	}

	return nil
}

// RenderToWriter renders D2 source code and writes SVG to the provided writer.
// This is useful for streaming output without intermediate files.
func (r *Renderer) RenderToWriter(ctx context.Context, d2Source string, w io.Writer, timeoutSec int) error {
	svgContent, err := r.RenderDiagramWithTimeout(ctx, d2Source, timeoutSec)
	if err != nil {
		return fmt.Errorf("failed to render diagram: %w", err)
	}

	if _, err := io.WriteString(w, svgContent); err != nil {
		return fmt.Errorf("failed to write SVG: %w", err)
	}

	return nil
}
