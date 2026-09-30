package viewmodel

import (
	"fmt"
	"sort"
	"strings"
)

// Format names an output format.
type Format string

const (
	FormatD2   Format = "d2"
	FormatSVG  Format = "svg"
	FormatMD   Format = "md"
	FormatHTML Format = "html"
)

// AllFormats lists every supported format in canonical order.
var AllFormats = []Format{FormatD2, FormatSVG, FormatMD, FormatHTML}

func (f Format) rank() int {
	for i, g := range AllFormats {
		if g == f {
			return i
		}
	}
	return len(AllFormats)
}

// Requires names the formats whose output this format embeds by reference
// (research R5). Backends never call each other; the build produces the
// required format too.
func (f Format) Requires() []Format {
	switch f {
	case FormatHTML, FormatMD:
		return []Format{FormatSVG}
	default:
		return nil
	}
}

func supportedList() string {
	names := make([]string, len(AllFormats))
	for i, f := range AllFormats {
		names[i] = string(f)
	}
	return strings.Join(names, ", ")
}

// ParseFormats reads a user-supplied format list. Entries may be
// comma-separated, repeated, and in any order. The result is de-duplicated and
// in canonical order. An unknown name is rejected by name alongside the
// supported list (FR-015).
func ParseFormats(raw []string) ([]Format, error) {
	set := map[Format]bool{}
	for _, entry := range raw {
		for part := range strings.SplitSeq(entry, ",") {
			name := strings.TrimSpace(part)
			if name == "" {
				continue
			}
			f := Format(name)
			if f.rank() == len(AllFormats) {
				return nil, fmt.Errorf("unsupported format %q: supported formats are %s", name, supportedList())
			}
			set[f] = true
		}
	}
	if len(set) == 0 {
		return nil, fmt.Errorf("no format given: supported formats are %s", supportedList())
	}
	return sortedFormats(set), nil
}

// Expand adds every required format to the set and reports which were added.
func Expand(formats []Format) (expanded, added []Format) {
	set := map[Format]bool{}
	for _, f := range formats {
		set[f] = true
	}
	extra := map[Format]bool{}
	for _, f := range formats {
		for _, r := range f.Requires() {
			if !set[r] {
				extra[r] = true
			}
		}
	}
	for f := range extra {
		set[f] = true
	}
	if len(extra) > 0 {
		added = sortedFormats(extra)
	}
	return sortedFormats(set), added
}

func sortedFormats(set map[Format]bool) []Format {
	out := make([]Format, 0, len(set))
	for f := range set {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rank() < out[j].rank() })
	return out
}

// Artifact is one generated file. Bytes already carry the generated-file
// notice (FR-020).
type Artifact struct {
	// Path is relative to the output root, forward slashes.
	Path   string
	Format Format
	Bytes  []byte
	// Owner is the address of what the artifact depicts, used to name both
	// sides of an output path collision (FR-028). Empty for site-wide files.
	Owner string
}

// SortArtifacts orders artifacts by Path.
func SortArtifacts(as []Artifact) {
	sort.SliceStable(as, func(i, j int) bool { return as[i].Path < as[j].Path })
}

// ThemeFile is one user override from <project root>/templates/.
type ThemeFile struct {
	Name   string // base name, e.g. "layout.gohtml"
	Bytes  []byte
	Origin string // project-relative path, for error messages
}
