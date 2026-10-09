package authoring

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

// FileHash is one source file's SHA-256, keyed by its project-relative path.
type FileHash struct {
	Path   string
	SHA256 string
}

// Revision fingerprints a project's source: every file's hash, sorted by path.
// A write quotes the revision it was based on, so a file changed underneath it
// is detected rather than overwritten (FR-016).
type Revision struct {
	files []FileHash
}

// NewRevision builds a revision from file hashes in any order.
func NewRevision(files []FileHash) Revision {
	files = append([]FileHash{}, files...)
	slices.SortFunc(files, func(a, b FileHash) int { return strings.Compare(a.Path, b.Path) })
	return Revision{files: files}
}

// IsZero reports whether the revision was never set.
func (r Revision) IsZero() bool { return r.files == nil }

// Files returns the hashes, sorted by path.
func (r Revision) Files() []FileHash { return slices.Clone(r.files) }

// Hash returns the hash recorded for path, or false if the file was not part
// of the project at that revision.
func (r Revision) Hash(path string) (string, bool) {
	i, ok := slices.BinarySearchFunc(r.files, path, func(f FileHash, p string) int {
		return strings.Compare(f.Path, p)
	})
	if !ok {
		return "", false
	}
	return r.files[i].SHA256, true
}

// Token is the short, opaque string handed to callers: a digest of every
// path and hash. It is deterministic (equal source, equal token) and small
// enough to carry on every result, whatever the project's size; the full
// per-file hashes stay with whoever issued it (Principle VI).
func (r Revision) Token() string {
	h := sha256.New()
	for _, f := range r.files {
		h.Write([]byte(f.Path))
		h.Write([]byte{0})
		h.Write([]byte(f.SHA256))
		h.Write([]byte{'\n'})
	}
	return "r1-" + hex.EncodeToString(h.Sum(nil))[:16]
}
