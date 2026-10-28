package html

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// fontFace describes one bundled font file (feature 021, research R6).
type fontFace struct {
	file, family, weight, style string
}

// fontFaces is every file in theme/fonts, in file-name order. fonts.go fails
// the build if a file is missing, and TestAssetsPublished if one is unlisted.
var fontFaces = []fontFace{
	{"audiowide-400.woff2", "Audiowide", "400", "normal"},
	{"jetbrains-mono-400.woff2", "JetBrains Mono", "400", "normal"},
	{"jetbrains-mono-600.woff2", "JetBrains Mono", "600", "normal"},
	{"orbitron-600.woff2", "Orbitron", "600", "normal"},
	{"orbitron-700.woff2", "Orbitron", "700", "normal"},
	{"source-sans-3-400-italic.woff2", "Source Sans 3", "400", "italic"},
	{"source-sans-3-400.woff2", "Source Sans 3", "400", "normal"},
	{"source-sans-3-600.woff2", "Source Sans 3", "600", "normal"},
}

// brandArtifacts are the site's generated font, glyph and licence files. Each
// opens with the generated-file notice like every other output (014 FR-020):
// the fonts travel inside fonts.css as base64, because a WOFF2 file cannot
// carry a notice.
func brandArtifacts(notice string) ([]vm.Artifact, error) {
	css, err := fontsCSS(notice)
	if err != nil {
		return nil, err
	}
	eye, err := fs.ReadFile(builtinTheme, "theme/brand/madstone-eye.svg")
	if err != nil {
		return nil, err
	}
	licences, err := licencesText(notice)
	if err != nil {
		return nil, err
	}
	return []vm.Artifact{
		{Path: "assets/fonts.css", Format: vm.FormatHTML, Bytes: css},
		{Path: "assets/brand/madstone-eye.svg", Format: vm.FormatHTML,
			Bytes: append([]byte("<!-- "+notice+" -->\n"), withSVGNamespace(eye)...)},
		{Path: "assets/LICENSES.txt", Format: vm.FormatHTML, Bytes: licences},
	}, nil
}

func fontsCSS(notice string) ([]byte, error) {
	var b strings.Builder
	b.WriteString("/* " + notice + " */\n")
	for _, f := range fontFaces {
		data, err := fs.ReadFile(builtinTheme, "theme/fonts/"+f.file)
		if err != nil {
			return nil, fmt.Errorf("bundled font: %w", err)
		}
		fmt.Fprintf(&b, "@font-face {\n  font-family: %q;\n  font-weight: %s;\n  font-style: %s;\n  font-display: swap;\n"+
			"  src: url(data:font/woff2;base64,%s) format(\"woff2\");\n}\n",
			f.family, f.weight, f.style, base64.StdEncoding.EncodeToString(data))
	}
	return []byte(b.String()), nil
}

// licencesText concatenates every licence and notice file under the theme's
// subdirectories, in path order, each unmodified under a heading.
func licencesText(notice string) ([]byte, error) {
	var files []string
	err := fs.WalkDir(builtinTheme, "theme", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := path.Base(p)
		if !d.IsDir() && path.Dir(p) != "theme" && (strings.HasSuffix(name, ".txt") || name == "LICENSE") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var b strings.Builder
	b.WriteString(notice + "\n")
	for _, f := range files {
		data, err := fs.ReadFile(builtinTheme, f)
		if err != nil {
			return nil, err
		}
		b.WriteString("\n== " + strings.TrimPrefix(f, "theme/") + " ==\n")
		b.Write(data)
	}
	return []byte(b.String()), nil
}

// withSVGNamespace adds the SVG namespace to a root <svg> that lacks it. The
// brand glyph omits it, which is fine inline but stops a browser drawing the
// file through <img> (feature 021).
func withSVGNamespace(svg []byte) []byte {
	i := bytes.Index(svg, []byte("<svg"))
	if i < 0 || bytes.Contains(svg, []byte(`xmlns="http://www.w3.org/2000/svg"`)) {
		return svg
	}
	out := append([]byte(nil), svg[:i+len("<svg")]...)
	out = append(out, ` xmlns="http://www.w3.org/2000/svg"`...)
	return append(out, svg[i+len("<svg"):]...)
}
