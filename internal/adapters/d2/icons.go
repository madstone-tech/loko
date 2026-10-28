package d2

import (
	"embed"
	"strings"
	"sync"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// The icons are Lucide's, copied unmodified from lucide-static v1.62.0
// (https://unpkg.com/lucide-static@1.62.0/icons/) under the ISC licence; see
// icons/LICENSE, which ships as LICENSE-lucide (feature 020, research R7).
//
//go:embed icons/*.svg
var iconFiles embed.FS

// The icon table is the projection's (viewmodel.IconFor), shared with the
// HTML site so a page and its diagram draw the same icon (feature 021). The
// alias keeps the package's own references short.
var iconFor = vm.IconFor

// iconSymbol is the icon's drawing as a <symbol>, defined once per diagram
// and drawn with <use>. Lucide's stroke attributes wrap the drawing inside
// the symbol, once per diagram rather than once per card; the stroke is
// currentColor, so each icon takes the color set on its <use>.
func iconSymbol(name string) string { return iconSymbols()[name] }

// lucideStroke is Lucide's own root attributes, which the vendored files set
// on their <svg> element.
const lucideStroke = `fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"`

var iconSymbols = sync.OnceValue(func() map[string]string {
	entries, err := iconFiles.ReadDir("icons")
	if err != nil {
		panic(err)
	}
	out := make(map[string]string, len(entries))
	for _, e := range entries {
		data, err := iconFiles.ReadFile("icons/" + e.Name())
		if err != nil {
			panic(err)
		}
		name := strings.TrimSuffix(e.Name(), ".svg")
		out[name] = `<symbol id="loko-icon-` + name + `" viewBox="0 0 24 24"><g ` + lucideStroke + `>` +
			svgInner(string(data)) + `</g></symbol>`
	}
	return out
})

// svgInner is the content of a Lucide file's root <svg> element, one line
// per drawn element joined without whitespace.
func svgInner(file string) string {
	start := strings.Index(file, "<svg")
	end := strings.LastIndex(file, "</svg>")
	if start < 0 || end < start {
		panic("icons: not an SVG file")
	}
	open := start + strings.Index(file[start:], ">")
	var b strings.Builder
	for line := range strings.Lines(file[open+1 : end]) {
		b.WriteString(strings.TrimSpace(line))
	}
	return b.String()
}
