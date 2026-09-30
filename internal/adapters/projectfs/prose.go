package projectfs

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Prose is the ProseReader over the local file system.
type Prose struct{}

// ReadProse reads docs relative to root. A missing file, a directory, or a
// path that escapes the root is reported as not found (FR-026) — the same rule
// the docs_not_found warning applies, so the build never shows prose the
// compiler said was missing.
func (Prose) ReadProse(_ context.Context, root, docs string) (string, bool, error) {
	abs, ok := contained(root, docs)
	if !ok {
		return "", false, nil
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return "", false, nil
	}
	data, err := os.ReadFile(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

// contained resolves rel against root and reports whether it stays inside.
func contained(root, rel string) (string, bool) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	abs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	r, err := filepath.Rel(absRoot, abs)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", false
	}
	return abs, true
}
