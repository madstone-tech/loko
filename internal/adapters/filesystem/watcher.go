// Package filesystem provides file system implementations of the core ports.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// FileWatcher monitors the file system for changes to .md and .d2 files.
// It filters out unwanted directories and debounces rapid events.
type FileWatcher struct {
	watcher *fsnotify.Watcher
	events  chan usecases.FileChangeEvent
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	stopped bool
}

// NewFileWatcher creates a new file system watcher.
func NewFileWatcher() (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &FileWatcher{
		watcher: w,
		events:  make(chan usecases.FileChangeEvent, 10),
		done:    make(chan struct{}),
	}, nil
}

// Watch starts monitoring a directory for changes.
// Returns a read-only channel of FileChangeEvent; returns error if setup fails.
// The channel is closed when Stop() is called.
func (fw *FileWatcher) Watch(ctx context.Context, rootPath string) (<-chan usecases.FileChangeEvent, error) {
	fw.mu.Lock()
	if fw.stopped {
		fw.mu.Unlock()
		return nil, fmt.Errorf("watcher already stopped")
	}
	fw.mu.Unlock()

	// Validate root path exists
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("invalid root path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory")
	}

	// Add root and all subdirectories to watcher
	if err := fw.addRecursive(rootPath); err != nil {
		return nil, fmt.Errorf("failed to add watch paths: %w", err)
	}

	// Start background event processor
	fw.wg.Add(1)
	go func() {
		defer fw.wg.Done()
		fw.processEvents(ctx, rootPath)
	}()

	return fw.events, nil
}

// Stop halts file watching and closes all channels.
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	if fw.stopped {
		fw.mu.Unlock()
		return nil
	}
	fw.stopped = true
	fw.mu.Unlock()

	// Signal the goroutine to stop
	close(fw.done)

	// Close the underlying watcher to unblock the goroutine
	err := fw.watcher.Close()

	// Wait for the goroutine to exit before closing the events channel
	fw.wg.Wait()
	close(fw.events)

	if err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	return nil
}

// addRecursive adds the root path and all subdirectories to the watcher.
func (fw *FileWatcher) addRecursive(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}

		if !info.IsDir() {
			return nil
		}

		// Skip ignored directories
		if fw.shouldIgnoreDir(path, rootPath) {
			return filepath.SkipDir
		}

		if err := fw.watcher.Add(path); err != nil {
			// Log but don't fail; some directories may be inaccessible
			return nil
		}

		return nil
	})
}

// shouldIgnoreDir returns true if the directory should not be watched.
func (fw *FileWatcher) shouldIgnoreDir(path, rootPath string) bool {
	rel, err := filepath.Rel(rootPath, path)
	if err != nil {
		return true
	}

	// Normalize to forward slashes for consistent comparison
	rel = filepath.ToSlash(rel)

	// List of directories to ignore
	ignoredDirs := map[string]bool{
		"dist":          true,
		".git":          true,
		".loko":         true,
		"node_modules":  true,
		".venv":         true,
		"venv":          true,
		"__pycache__":   true,
		".pytest_cache": true,
		"build":         true,
		"target":        true,
	}

	// Check if any path component matches ignored directories
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if ignoredDirs[part] {
			return true
		}
	}

	return false
}

// shouldProcessFile returns true if the file should trigger a change event.
func (fw *FileWatcher) shouldProcessFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".d2"
}

// processEvents reads from fsnotify and sends debounced events.
func (fw *FileWatcher) processEvents(ctx context.Context, rootPath string) {
	// Debounce timer: batch events within 100ms
	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // Drain initial tick

	// Track pending events to debounce
	pendingEvents := make(map[string]usecases.FileChangeEvent)
	var mu sync.Mutex

	for {
		select {
		case <-fw.done:
			return

		case <-ctx.Done():
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Handle new directory creation (add to watcher)
			if event.Op&fsnotify.Create == fsnotify.Create {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if !fw.shouldIgnoreDir(event.Name, rootPath) {
						_ = fw.watcher.Add(event.Name)
					}
				}
			}

			// Only process .md and .d2 files
			if !fw.shouldProcessFile(event.Name) {
				continue
			}

			// Convert absolute path to relative
			relPath, err := filepath.Rel(rootPath, event.Name)
			if err != nil {
				continue
			}

			// Normalize to forward slashes and lowercase
			relPath = filepath.ToSlash(relPath)
			relPath = strings.ToLower(relPath)

			// Map fsnotify operation to event operation
			op := fw.mapOperation(event.Op)

			mu.Lock()
			pendingEvents[relPath] = usecases.FileChangeEvent{
				Path: relPath,
				Op:   op,
			}
			mu.Unlock()

			// Reset debounce timer
			debounceTimer.Reset(100 * time.Millisecond)

		case <-debounceTimer.C:
			// Send all pending events
			mu.Lock()
			for _, evt := range pendingEvents {
				select {
				case fw.events <- evt:
				case <-fw.done:
					mu.Unlock()
					return
				case <-ctx.Done():
					mu.Unlock()
					return
				}
			}
			pendingEvents = make(map[string]usecases.FileChangeEvent)
			mu.Unlock()

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err
		}
	}
}

// mapOperation converts fsnotify.Op to FileChangeEvent operation string.
func (fw *FileWatcher) mapOperation(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "create"
	case op&fsnotify.Write == fsnotify.Write:
		return "write"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "remove"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	default:
		return "write" // Default to write for unknown operations
	}
}
