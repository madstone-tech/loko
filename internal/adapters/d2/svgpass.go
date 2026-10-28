package d2

import (
	"bytes"
	"encoding/base64"
	"slices"
	"strconv"

	"oss.terrastruct.com/d2/lib/svg"
)

// groupPlan is what the pass does to one d2 group: an icon for a card, and
// a style per label line.
type groupPlan struct {
	icon, colour string
	lines        []LineRole
}

var (
	groupOpen = []byte(`<g class="`)
	textOpen  = []byte("<text")
	textClose = []byte("</text>")
	tspanOpen = []byte("<tspan")
)

// applyPlan adds the cards' icons and the label line styles to a rendered
// diagram (feature 020, research R1, R2). It is one linear scan over the
// bytes: d2 draws every node and edge as a flat group whose first class is
// the base64 of its ID, with the shape first and the label <text> last.
// fontHash is the prefix of the diagram's d2 font families (research R3).
func applyPlan(doc []byte, plan LabelPlan, fontHash string) []byte {
	if len(plan.Nodes) == 0 && len(plan.Edges) == 0 {
		return doc
	}
	groups := make(map[string]groupPlan, len(plan.Nodes)+len(plan.Edges))
	for id, n := range plan.Nodes {
		groups[groupClass(id)] = groupPlan{icon: n.Icon, colour: n.Colour, lines: n.Lines}
	}
	for id, lines := range plan.Edges {
		groups[groupClass(id)] = groupPlan{lines: lines}
	}

	var out bytes.Buffer
	out.Grow(len(doc) + len(doc)/8)
	rest := doc
	if icons := planIcons(plan); len(icons) > 0 {
		if i := bytes.Index(rest, []byte("<svg")); i >= 0 {
			if j := bytes.IndexByte(rest[i:], '>'); j >= 0 {
				out.Write(rest[:i+j+1])
				writeDefs(&out, icons)
				rest = rest[i+j+1:]
			}
		}
	}
	for {
		i := bytes.Index(rest, groupOpen)
		if i < 0 {
			break
		}
		start := i + len(groupOpen)
		n := bytes.IndexAny(rest[start:], `" `)
		if n < 0 {
			break
		}
		out.Write(rest[:start+n])
		g, ok := groups[string(rest[start:start+n])]
		rest = rest[start+n:]
		if ok {
			rest = applyGroup(&out, rest, g, fontHash)
		}
	}
	out.Write(rest)
	return out.Bytes()
}

// groupClass is the class d2 gives the group of an object or edge ID.
func groupClass(id string) string {
	return base64.URLEncoding.EncodeToString([]byte(svg.EscapeText(id)))
}

// planIcons are the icons the plan's cards use, sorted, once each.
func planIcons(plan LabelPlan) []string {
	var icons []string
	for _, n := range plan.Nodes {
		if n.Icon != "" && !slices.Contains(icons, n.Icon) {
			icons = append(icons, n.Icon)
		}
	}
	slices.Sort(icons)
	return icons
}

func writeDefs(out *bytes.Buffer, icons []string) {
	out.WriteString("<defs>")
	for _, name := range icons {
		out.WriteString(iconSymbol(name))
	}
	out.WriteString("</defs>")
}

// applyGroup writes one planned group from just after its first class up to
// the end of its label, and returns what follows. The icon goes between the
// card's shape and its label, so the label is drawn over it.
func applyGroup(out *bytes.Buffer, rest []byte, g groupPlan, fontHash string) []byte {
	t := bytes.Index(rest, textOpen)
	if t < 0 {
		return rest
	}
	out.Write(rest[:t])
	if g.icon != "" {
		if x, y, ok := rectOrigin(rest[:t]); ok {
			writeIcon(out, g, x, y)
		}
	}
	rest = rest[t:]
	end := bytes.Index(rest, textClose)
	if end < 0 {
		return rest
	}
	text := rest[:end]
	for line := 0; ; line++ {
		k := bytes.Index(text, tspanOpen)
		if k < 0 {
			break
		}
		k += len(tspanOpen)
		out.Write(text[:k])
		if line < len(g.lines) {
			if s := g.lines[line].inlineStyle(fontHash); s != "" {
				out.WriteString(` style="` + s + `"`)
			}
		}
		text = text[k:]
	}
	out.Write(text)
	return rest[end:]
}

// rectOrigin is the x and y of the first <rect> in a card's shape.
func rectOrigin(shape []byte) (x, y float64, ok bool) {
	i := bytes.Index(shape, []byte("<rect "))
	if i < 0 {
		return 0, 0, false
	}
	tag := shape[i:]
	if j := bytes.IndexByte(tag, '>'); j >= 0 {
		tag = tag[:j]
	}
	x, okX := floatAttr(tag, " x=\"")
	y, okY := floatAttr(tag, " y=\"")
	return x, y, okX && okY
}

func floatAttr(tag []byte, prefix string) (float64, bool) {
	i := bytes.Index(tag, []byte(prefix))
	if i < 0 {
		return 0, false
	}
	v := tag[i+len(prefix):]
	j := bytes.IndexByte(v, '"')
	if j < 0 {
		return 0, false
	}
	f, err := strconv.ParseFloat(string(v[:j]), 64)
	return f, err == nil
}

// iconInset and iconSize place the icon in the card's top-left corner, inside
// the 20px padding d2 leaves around a label (research R1).
const iconInset, iconSize = 4, "16"

func writeIcon(out *bytes.Buffer, g groupPlan, x, y float64) {
	out.WriteString(`<use class="loko-icon" href="#loko-icon-` + g.icon + `" x="` + formatFloat(x+iconInset) +
		`" y="` + formatFloat(y+iconInset) + `" width="` + iconSize + `" height="` + iconSize +
		`" color="` + g.colour + `"/>`)
}

func formatFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
