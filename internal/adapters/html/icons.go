package html

import (
	"encoding/base64"
	"io/fs"
	"path"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// iconsCSS is the generated assets/icons.css (feature 021): one rule per
// Lucide icon in theme/icons, drawn as a CSS mask so it takes the text colour.
// The files are copies of the diagram's icons (TestIconCopiesMatch), so a
// page and its diagram draw the same glyph. Data URIs keep the site working
// from file:// and offline, where an external <use> would be blocked.
func iconsCSS(notice string) (vm.Artifact, error) {
	files, err := fs.Glob(builtinTheme, "theme/icons/*.svg")
	if err != nil {
		return vm.Artifact{}, err
	}
	var b strings.Builder
	b.WriteString("/* " + notice + " */\n")
	b.WriteString(".icon { display: inline-block; width: 1em; height: 1em; flex: none; background-color: currentColor;\n" +
		"  -webkit-mask: var(--icon) center / contain no-repeat; mask: var(--icon) center / contain no-repeat; }\n")
	for _, f := range files {
		data, err := fs.ReadFile(builtinTheme, f)
		if err != nil {
			return vm.Artifact{}, err
		}
		name := strings.TrimSuffix(path.Base(f), ".svg")
		b.WriteString(`.icon[data-icon="` + name + `"] { --icon: url("data:image/svg+xml;base64,` +
			base64.StdEncoding.EncodeToString(data) + `"); }` + "\n")
	}
	return vm.Artifact{Path: "assets/icons.css", Format: vm.FormatHTML, Bytes: []byte(b.String())}, nil
}
