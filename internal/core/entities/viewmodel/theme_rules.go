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

// OverridableThemeFiles lists every top-level theme file a project may
// replace (014 contracts/theme.md, amended by 023 for fonts.css).
var OverridableThemeFiles = []string{"custom.css", "element.gohtml", "fonts.css", "index.gohtml",
	"layout.gohtml", "partials.gohtml", "rail.gohtml", "site.js", "style.css", "view.gohtml"}

// themeAssetDir is the prefix of every asset Rel.
const themeAssetDir = "assets/"

// ValidateTheme rejects every override whose name is not an overridable
// theme file, so a typo such as sytle.css fails the build instead of being
// silently ignored (FR-035). It also rejects an asset whose path is not safe,
// a symbolic link, and two asset paths that differ only in letter case
// (feature 023, contract T-4). Problems are sorted by Origin.
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
	var assets []ThemeFile
	for _, f := range files {
		switch {
		case f.IsAsset():
			assets = append(assets, f)
		case !known[f.Name]:
			out = append(out, ThemeProblem{
				Origin: f.Origin,
				Message: f.Name + " is not an overridable theme file; overridable files are " +
					strings.Join(list, ", "),
			})
		}
	}
	out = append(out, assetProblems(assets)...)
	sort.Slice(out, func(i, j int) bool { return out[i].Origin < out[j].Origin })
	return out
}

// assetProblems gives one problem per asset that is a symbolic link, has an
// unsafe path, or repeats an earlier path in another letter case.
func assetProblems(assets []ThemeFile) []ThemeProblem {
	sort.Slice(assets, func(i, j int) bool { return assets[i].Rel < assets[j].Rel })
	var out []ThemeProblem
	seen := map[string]string{}
	for _, f := range assets {
		lower := strings.ToLower(f.Rel)
		switch {
		case f.Link:
			out = append(out, ThemeProblem{Origin: f.Origin,
				Message: "theme asset is a symbolic link. A theme asset must be a regular file."})
		case !strings.HasPrefix(f.Rel, themeAssetDir) || !SafePath(f.Rel):
			out = append(out, ThemeProblem{Origin: f.Origin,
				Message: "theme asset path is not safe. Use a relative path with forward slashes, with no \"..\", no drive letter and no control character."})
		case seen[lower] != "":
			out = append(out, ThemeProblem{Origin: f.Origin,
				Message: "theme asset path differs from " + seen[lower] + " only in letter case. Rename one of the two files."})
		default:
			seen[lower] = f.Origin
		}
	}
	return out
}
