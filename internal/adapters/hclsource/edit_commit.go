package hclsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// Editor implements the usecases.SourceEditor port: it plans edits on
// in-memory copies of source files and commits them atomically. It writes
// nothing but *.loko.hcl files inside the project root (FR-026).
type Editor struct{}

// NewEditor returns the HCL source editor.
func NewEditor() *Editor { return &Editor{} }

// Revision hashes every discovered source file.
func (e *Editor) Revision(_ context.Context, root string) (authoring.Revision, error) {
	files, _, err := Discover(root)
	if err != nil {
		return authoring.Revision{}, err
	}
	hashes := make([]authoring.FileHash, 0, len(files))
	for _, f := range files {
		h, err := hashFile(f.Abs)
		if err != nil {
			return authoring.Revision{}, fmt.Errorf("hashing %s: %w", f.Rel, err)
		}
		hashes = append(hashes, authoring.FileHash{Path: f.Rel, SHA256: h})
	}
	return authoring.NewRevision(hashes), nil
}

func hashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hashBytes(b), nil
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// checkPath refuses anything but a *.loko.hcl file inside root (FR-020,
// FR-026), whatever the caller validated before (research R12).
func checkPath(root, rel string) error {
	clean := path.Clean(rel)
	switch {
	case !strings.HasSuffix(rel, ".loko.hcl"):
		return refuse(authoring.ReasonPathRefused, "%s is not an architecture source file (*.loko.hcl)", rel)
	case rel == "" || path.IsAbs(rel) || filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" ||
		strings.Contains(rel, `\`) || clean != rel || clean == ".." || strings.HasPrefix(clean, "../"):
		return refuse(authoring.ReasonPathRefused, "%s is not a clean path inside the project", rel)
	}
	// A symlinked directory must not lead outside the root either.
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, filepath.FromSlash(path.Dir(rel)))
	for {
		real, err := filepath.EvalSymlinks(dir)
		if err == nil {
			if real != realRoot && !strings.HasPrefix(real, realRoot+string(filepath.Separator)) {
				return refuse(authoring.ReasonPathRefused, "%s resolves outside the project", rel)
			}
			return nil
		}
		if !os.IsNotExist(err) || dir == root {
			return err
		}
		dir = filepath.Dir(dir)
	}
}
