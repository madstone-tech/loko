package viewmodel

import (
	"sort"
	"strings"
)

// ThemeProblem is one override the theme rules reject.
type ThemeProblem struct {
	Origin  string
	Message string
}

// ValidateTheme rejects every override whose name is not an overridable
// theme file, so a typo such as sytle.css fails the build instead of being
// silently ignored (FR-035). Problems are sorted by Origin.
//
// Template parse errors and unknown {{define}} blocks are the html backend's
// to report: only a template parser can see them.
func ValidateTheme(files []ThemeFile, overridable []string) []ThemeProblem {
	known := make(map[string]bool, len(overridable))
	for _, n := range overridable {
		known[n] = true
	}
	list := append([]string(nil), overridable...)
	sort.Strings(list)
	var out []ThemeProblem
	for _, f := range files {
		if !known[f.Name] {
			out = append(out, ThemeProblem{
				Origin: f.Origin,
				Message: f.Name + " is not an overridable theme file; overridable files are " +
					strings.Join(list, ", "),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Origin < out[j].Origin })
	return out
}
