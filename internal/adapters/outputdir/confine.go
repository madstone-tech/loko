package outputdir

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// defaultOut is the output directory a build uses when none is given.
const defaultOut = "dist"

// Confine resolves outDir against root and refuses it unless it stays inside
// the project, following symlinks. It is what lets an MCP client choose an
// output directory without being able to write anywhere else (feature 017).
//
// The root is made absolute before its symlinks are resolved: EvalSymlinks
// keeps a relative path relative, so "." would never prefix "dist" (the bug
// 1.0.2 fixed in hclsource.checkPath).
func (s *Store) Confine(_ context.Context, root, outDir string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", err
	}
	if outDir == "" {
		outDir = defaultOut
	}
	target := outDir
	if !filepath.IsAbs(target) {
		target = filepath.Join(absRoot, target)
	}
	target = filepath.Clean(target)
	real, err := resolveExisting(target)
	if err != nil {
		return "", err
	}
	if real != realRoot && !strings.HasPrefix(real, realRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("%s resolves outside the project", outDir)
	}
	return target, nil
}

// resolveExisting resolves symlinks in the longest existing prefix of p and
// appends the part that does not exist yet.
func resolveExisting(p string) (string, error) {
	var rest []string
	for {
		real, err := filepath.EvalSymlinks(p)
		if err == nil {
			return filepath.Join(append([]string{real}, rest...)...), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(p)
		if parent == p {
			return "", err
		}
		rest = append([]string{filepath.Base(p)}, rest...)
		p = parent
	}
}
