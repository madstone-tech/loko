package d2

// LineRole is what one line of a node or edge label is. The emitter records
// it per line, and the SVG pass styles each line by its role (feature 020,
// research R2), because roles cannot be read back from the text.
type LineRole uint8

const (
	roleTitle       LineRole = iota // a node's name or title line, in d2's bold
	roleKind                        // CONTAINER · Go
	roleBlank                       // the spacer before a description
	roleDescription                 // a node's description line
	roleEdgeLabel                   // an edge's description line
	roleEdgeTech                    // an edge's technology line
	roleEdgeTags                    // an edge's #tag line
)

// lineStyle is how the SVG pass draws one role.
type lineStyle struct {
	css string
	// regular switches the line to d2's regular font family. d2 embeds each
	// weight as its own single-face family, so font-weight cannot select a
	// lighter face (research R3).
	regular bool
	opacity float64 // the fill-opacity in css, for the contrast test
	size    float64 // the font-size in css, in em of d2's 16px
}

// lineStyles is the only place label text styles are defined (FR-011). No
// role sets font-weight (research R3, R4).
var lineStyles = [...]lineStyle{
	roleTitle:       {opacity: 1, size: 1},
	roleKind:        {css: "font-size:0.7em;letter-spacing:0.06em;fill-opacity:0.85", opacity: 0.85, size: 0.7},
	roleBlank:       {opacity: 1, size: 1},
	roleDescription: {css: "font-size:0.9em;fill-opacity:0.9", regular: true, opacity: 0.9, size: 0.9},
	roleEdgeLabel:   {opacity: 1, size: 1},
	roleEdgeTech:    {css: "font-size:0.75em;letter-spacing:0.04em;fill-opacity:0.8", opacity: 0.8, size: 0.75},
	roleEdgeTags:    {css: "font-size:0.75em;fill-opacity:0.72", opacity: 0.72, size: 0.75},
}

// inlineStyle is the style attribute value for a line of this role in a
// diagram whose d2 font families are prefixed with fontHash; empty when the
// line stays as d2 drew it.
func (r LineRole) inlineStyle(fontHash string) string {
	s := lineStyles[r]
	if s.regular {
		return "font-family:" + fontHash + "-font-regular;" + s.css
	}
	return s.css
}
