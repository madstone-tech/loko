package html

import (
	"bytes"
	"html"
	"html/template"
)

// inlineSVG prepares a rendered diagram for embedding in a page (feature 021,
// research R2, R3). It drops the XML declaration and the generated-file
// comment, labels the outer <svg> by its figure caption (and, when given, the
// element list that describes it), and adds the loko-focus class to the one
// card whose data-loko-address is focus. Byte operations only, like the d2
// SVG pass: the input is loko's own output, not arbitrary XML.
func inlineSVG(svg []byte, captionID, describedBy, focus string) template.HTML {
	i := bytes.Index(svg, []byte("<svg"))
	if i < 0 {
		return ""
	}
	attrs := ` role="img" aria-labelledby="` + captionID + `"`
	if describedBy != "" {
		attrs += ` aria-describedby="` + describedBy + `"`
	}
	var out bytes.Buffer
	out.Grow(len(svg) - i + len(attrs) + len(" loko-focus"))
	out.WriteString("<svg" + attrs)
	rest := svg[i+len("<svg"):]
	if focus != "" {
		marker := []byte(`" data-loko-address="` + html.EscapeString(focus) + `"`)
		if j := bytes.Index(rest, marker); j >= 0 {
			out.Write(rest[:j])
			out.WriteString(" loko-focus")
			rest = rest[j:]
		}
	}
	out.Write(rest)
	return template.HTML(out.String()) //nolint:gosec // loko's own rendered SVG, see doc comment
}
