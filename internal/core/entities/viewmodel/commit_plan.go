package viewmodel

import (
	"bytes"
	"path"
	"sort"
	"strings"
)

// ManifestFile is the name of the ownership manifest in the output root.
const ManifestFile = ".loko-manifest"

// Manifest lists every path the last build wrote. The tool owns exactly these
// paths in the output directory, and nothing else (FR-023).
type Manifest struct {
	Paths []string
}

// NewManifest builds a manifest for an artifact set.
func NewManifest(artifacts []Artifact) Manifest {
	ps := make([]string, 0, len(artifacts))
	for _, a := range artifacts {
		ps = append(ps, a.Path)
	}
	sort.Strings(ps)
	return Manifest{Paths: dedupe(ps)}
}

// Encode serialises the manifest: the notice as a comment, then one path per
// line.
func (m Manifest) Encode(sources []string) []byte {
	var b bytes.Buffer
	b.WriteString("# " + NoticeText(sources) + "\n")
	for _, p := range m.Paths {
		b.WriteString(p + "\n")
	}
	return b.Bytes()
}

// DecodeManifest parses a manifest. It never widens ownership: comment lines,
// blank lines and any path that is absolute or escapes the output root are
// dropped, so a corrupt manifest can only make the tool prune less.
func DecodeManifest(data []byte) Manifest {
	var ps []string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !SafePath(line) {
			continue
		}
		ps = append(ps, line)
	}
	sort.Strings(ps)
	return Manifest{Paths: dedupe(ps)}
}

// SafePath reports whether p is a clean, forward-slash path that stays inside
// the output root: not absolute, no "..", no backslash, drive letter or
// control character.
func SafePath(p string) bool {
	if strings.HasPrefix(p, "/") || strings.ContainsAny(p, "\\:") {
		return false
	}
	for i := 0; i < len(p); i++ {
		if p[i] < 0x20 || p[i] == 0x7f {
			return false
		}
	}
	c := path.Clean(p)
	return c == p && c != "." && c != ".." && !strings.HasPrefix(c, "../")
}

// CommitPlan is what a commit must do. Every list is sorted.
type CommitPlan struct {
	Write     []string
	Unchanged []string
	Remove    []string
}

// PlanCommit decides which artifacts to write, which to leave alone because
// the bytes on disk already match (so their modification times survive,
// FR-022), and which previously owned paths to remove (FR-023). A path the old
// manifest does not list is never removed. onDisk reports a file's current
// bytes.
func PlanCommit(old Manifest, next []Artifact, onDisk func(path string) ([]byte, bool)) CommitPlan {
	var plan CommitPlan
	keep := make(map[string]bool, len(next))
	sorted := append([]Artifact(nil), next...)
	SortArtifacts(sorted)
	for _, a := range sorted {
		keep[a.Path] = true
		if cur, ok := onDisk(a.Path); ok && bytes.Equal(cur, a.Bytes) {
			plan.Unchanged = append(plan.Unchanged, a.Path)
			continue
		}
		plan.Write = append(plan.Write, a.Path)
	}
	for _, p := range old.Paths {
		if !keep[p] {
			plan.Remove = append(plan.Remove, p)
		}
	}
	sort.Strings(plan.Remove)
	return plan
}
