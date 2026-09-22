package arch

import (
	"path/filepath"
	"strconv"
	"strings"
)

// SourceRange locates a span of authored source. Every diagnostic carries one
// (FR-030), and the IR records them so later stages can point back at the
// declaration that produced a node.
//
// This duplicates the parser library's range type on purpose. internal/core
// may not import HCL (FR-044), so ranges cross the adapter boundary as this
// plain struct. Conversion happens once, in internal/adapters/hclsource.
//
// File is always project-relative and always uses forward slashes, on every
// platform. Absolute paths would leak a machine-specific string into the
// export and break byte-identical output (FR-036c, SC-003).
type SourceRange struct {
	File        string `json:"file" toon:"file"`
	StartLine   int    `json:"startLine" toon:"startLine"`
	StartColumn int    `json:"startColumn" toon:"startColumn"`
	StartByte   int    `json:"startByte,omitempty" toon:"startByte,omitempty"`
	EndLine     int    `json:"endLine" toon:"endLine"`
	EndColumn   int    `json:"endColumn" toon:"endColumn"`
	EndByte     int    `json:"endByte,omitempty" toon:"endByte,omitempty"`
}

// NormalizeFile converts a path to the project-relative, forward-slash form
// SourceRange requires. A path outside root, or one that cannot be made
// relative, is returned cleaned and slash-separated rather than rejected: a
// diagnostic with an imperfect path is more useful than no diagnostic.
func NormalizeFile(root, path string) string {
	if root != "" {
		if rel, err := filepath.Rel(root, path); err == nil {
			path = rel
		}
	}
	return filepath.ToSlash(filepath.Clean(path))
}

// IsZero reports whether the range carries no position. Some diagnostics are
// project-wide rather than file-scoped — no_source_found, for instance, has
// nothing to point at.
func (r SourceRange) IsZero() bool {
	return r.File == "" && r.StartLine == 0 && r.StartColumn == 0
}

// Loc renders "file:line:column", the form editors and CI annotations expect.
// A zero range renders as "<project>".
//
// Deliberately NOT named String: implementing fmt.Stringer makes encoders that
// prefer that interface collapse the whole range to one string, dropping the
// end position and byte offsets. The JSON and TOON exports must carry
// equivalent information (FR-036), so the range stays a structured value and
// this convenience is opt-in.
func (r SourceRange) Loc() string {
	if r.IsZero() {
		return "<project>"
	}
	var b strings.Builder
	b.WriteString(r.File)
	if r.StartLine > 0 {
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(r.StartLine))
		if r.StartColumn > 0 {
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(r.StartColumn))
		}
	}
	return b.String()
}

// Compare orders ranges by file, then by start byte, then by start line and
// column. It is the primary key of the deterministic diagnostic ordering
// required by FR-033.
func (r SourceRange) Compare(o SourceRange) int {
	if c := strings.Compare(r.File, o.File); c != 0 {
		return c
	}
	if r.StartByte != o.StartByte {
		return cmpInt(r.StartByte, o.StartByte)
	}
	if r.StartLine != o.StartLine {
		return cmpInt(r.StartLine, o.StartLine)
	}
	return cmpInt(r.StartColumn, o.StartColumn)
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
