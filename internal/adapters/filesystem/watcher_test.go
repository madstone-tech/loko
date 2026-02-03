package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// stopWatcher is a helper to properly close a watcher in tests.
func stopWatcher(t *testing.T, fw *FileWatcher) {
	if err := fw.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

// TestNewFileWatcher tests watcher initialization.
func TestNewFileWatcher(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	if fw == nil {
		t.Error("NewFileWatcher returned nil")
	}

	// Clean up
	if err := fw.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

// TestWatchInvalidPath tests error handling for invalid paths.
func TestWatchInvalidPath(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer func() {
		if err := fw.Stop(); err != nil {
			t.Errorf("Stop failed: %v", err)
		}
	}()

	ctx := context.Background()
	_, err = fw.Watch(ctx, "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}
}

// TestWatchStoppedWatcher tests error when watching after stop.
func TestWatchStoppedWatcher(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	if err := fw.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	ctx := context.Background()
	tmpDir := t.TempDir()
	_, watchErr := fw.Watch(ctx, tmpDir)
	if watchErr == nil {
		t.Error("expected error when watching after stop, got nil")
	}
}

// TestWatchMarkdownFile tests detecting markdown file changes.
func TestWatchMarkdownFile(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a markdown file
	mdFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case evt := <-events:
		if evt.Path != "test.md" {
			t.Errorf("expected path 'test.md', got '%s'", evt.Path)
		}
		if evt.Op != "create" && evt.Op != "write" {
			t.Errorf("expected 'create' or 'write', got '%s'", evt.Op)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

// TestWatchD2File tests detecting D2 file changes.
func TestWatchD2File(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a D2 file
	d2File := filepath.Join(tmpDir, "diagram.d2")
	if err := os.WriteFile(d2File, []byte("x -> y"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case evt := <-events:
		if evt.Path != "diagram.d2" {
			t.Errorf("expected path 'diagram.d2', got '%s'", evt.Path)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

// TestWatchIgnoresNonMarkdownFiles tests that non-markdown files are ignored.
func TestWatchIgnoresNonMarkdownFiles(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a non-markdown file
	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should not receive event
	select {
	case evt := <-events:
		t.Errorf("unexpected event for non-markdown file: %v", evt)
	case <-time.After(500 * time.Millisecond):
		// Expected: no event
	}
}

// TestWatchIgnoresGitDirectory tests that .git directory is ignored.
func TestWatchIgnoresGitDirectory(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a markdown file in .git (should be ignored)
	mdFile := filepath.Join(gitDir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should not receive event
	select {
	case evt := <-events:
		t.Errorf("unexpected event from .git directory: %v", evt)
	case <-time.After(500 * time.Millisecond):
		// Expected: no event
	}
}

// TestWatchIgnoresDistDirectory tests that dist directory is ignored.
func TestWatchIgnoresDistDirectory(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create dist directory
	distDir := filepath.Join(tmpDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatalf("failed to create dist directory: %v", err)
	}

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a markdown file in dist (should be ignored)
	mdFile := filepath.Join(distDir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should not receive event
	select {
	case evt := <-events:
		t.Errorf("unexpected event from dist directory: %v", evt)
	case <-time.After(500 * time.Millisecond):
		// Expected: no event
	}
}

// TestWatchSubdirectory tests watching files in subdirectories.
func TestWatchSubdirectory(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "src", "systems")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a markdown file in subdirectory
	mdFile := filepath.Join(subDir, "system.md")
	if err := os.WriteFile(mdFile, []byte("# System"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case evt := <-events:
		expectedPath := filepath.Join("src", "systems", "system.md")
		expectedPath = filepath.ToSlash(expectedPath)
		if evt.Path != expectedPath {
			t.Errorf("expected path '%s', got '%s'", expectedPath, evt.Path)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

// TestWatchDebouncing tests that rapid events are debounced.
func TestWatchDebouncing(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a file and write to it multiple times rapidly
	mdFile := filepath.Join(tmpDir, "test.md")
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(mdFile, []byte("# Test "+string(rune(i))), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Collect events within debounce window
	eventCount := 0
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case <-events:
			eventCount++
		case <-timeout:
			goto done
		}
	}
done:

	// Should receive fewer events than writes due to debouncing
	if eventCount > 3 {
		t.Errorf("expected debounced events (<=3), got %d", eventCount)
	}
}

// TestWatchContextCancellation tests that context cancellation stops watching.
func TestWatchContextCancellation(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Cancel context
	cancel()

	// Create a file (should not be received due to cancelled context)
	mdFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should not receive event
	select {
	case <-events:
		t.Error("unexpected event after context cancellation")
	case <-time.After(500 * time.Millisecond):
		// Expected: no event
	}
}

// TestWatchPathNormalization tests that paths are normalized to lowercase.
func TestWatchPathNormalization(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a file with uppercase extension
	mdFile := filepath.Join(tmpDir, "TEST.MD")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case evt := <-events:
		if evt.Path != "test.md" {
			t.Errorf("expected normalized path 'test.md', got '%s'", evt.Path)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

// TestWatchFileRemoval tests detecting file removal.
func TestWatchFileRemoval(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a markdown file after starting watch
	mdFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for creation event
	select {
	case <-events:
		// Got creation event
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for creation event")
		return
	}

	// Remove the file
	if err := os.Remove(mdFile); err != nil {
		t.Fatalf("failed to remove test file: %v", err)
	}

	// Wait for removal event
	select {
	case evt := <-events:
		if evt.Op != "remove" {
			t.Errorf("expected 'remove' operation, got '%s'", evt.Op)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for removal event")
	}
}

// TestStopClosesChannel tests that Stop closes the event channel.
func TestStopClosesChannel(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Stop the watcher
	if err := fw.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Channel should be closed
	select {
	case _, ok := <-events:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for channel close")
	}
}

// TestStopIdempotent tests that Stop can be called multiple times.
func TestStopIdempotent(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	tmpDir := t.TempDir()
	ctx := context.Background()

	_, err = fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Stop multiple times
	if err := fw.Stop(); err != nil {
		t.Fatalf("first Stop failed: %v", err)
	}

	if err := fw.Stop(); err != nil {
		t.Fatalf("second Stop failed: %v", err)
	}
}

// TestWatchNewDirectoryCreation tests that newly created directories are watched.
func TestWatchNewDirectoryCreation(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a new subdirectory
	newDir := filepath.Join(tmpDir, "newsrc")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Wait a bit for directory to be added to watcher
	time.Sleep(200 * time.Millisecond)

	// Create a markdown file in the new directory
	mdFile := filepath.Join(newDir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case evt := <-events:
		expectedPath := filepath.Join("newsrc", "test.md")
		expectedPath = filepath.ToSlash(expectedPath)
		if evt.Path != expectedPath {
			t.Errorf("expected path '%s', got '%s'", expectedPath, evt.Path)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

// TestWatchForwardSlashes tests that paths use forward slashes on all platforms.
func TestWatchForwardSlashes(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create nested directories
	nestedDir := filepath.Join(tmpDir, "src", "systems", "auth")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a markdown file
	mdFile := filepath.Join(nestedDir, "system.md")
	if err := os.WriteFile(mdFile, []byte("# System"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case evt := <-events:
		// Check that path uses forward slashes
		if !containsOnlyForwardSlashes(evt.Path) {
			t.Errorf("path contains backslashes: %s", evt.Path)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

// Helper function to check if path uses only forward slashes
func containsOnlyForwardSlashes(path string) bool {
	for _, ch := range path {
		if ch == '\\' {
			return false
		}
	}
	return true
}

// TestWatchMultipleFiles tests watching multiple file changes.
func TestWatchMultipleFiles(t *testing.T) {
	fw, err := NewFileWatcher()
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer stopWatcher(t, fw)

	tmpDir := t.TempDir()
	ctx := context.Background()

	events, err := fw.Watch(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create multiple files
	files := []string{"file1.md", "file2.d2", "file3.md"}
	for _, file := range files {
		filePath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Collect events
	receivedPaths := make(map[string]bool)
	timeout := time.After(2 * time.Second)
	for {
		select {
		case evt := <-events:
			receivedPaths[evt.Path] = true
		case <-timeout:
			goto done
		}
	}
done:

	// Check that we received events for all files
	for _, file := range files {
		if !receivedPaths[file] {
			t.Errorf("did not receive event for file: %s", file)
		}
	}
}
