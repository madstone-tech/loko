package html

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// markdown converts CommonMark prose with GFM tables. Raw HTML in the prose is
// not passed through (goldmark omits it unless WithUnsafe is set), so a prose
// file cannot inject script into the site.
var markdown = goldmark.New(goldmark.WithExtensions(extension.Table))

// renderProse converts one element's prose to HTML. goldmark's output is
// escaped HTML, so it is marked safe for the template.
func renderProse(src string) template.HTML {
	var b bytes.Buffer
	if err := markdown.Convert([]byte(src), &b); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	return template.HTML(b.String()) //nolint:gosec // goldmark escapes raw HTML (WithUnsafe is off)
}
