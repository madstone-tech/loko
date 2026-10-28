package d2

import (
	"bytes"
	"cmp"
	"slices"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// NodePlan is what the SVG pass adds to one element or instance card: its
// icon, drawn in the card's text colour, and the role of each label line.
type NodePlan struct {
	Icon   string
	Colour string
	Lines  []LineRole
	// Address is the node's model address, written on its group as
	// data-loko-address (feature 021, contracts/svg-address.md).
	Address string
}

// LabelPlan is produced by the emitter alongside the D2 source, for the SVG
// pass only (data-model "Label plan"). Keys are d2's absolute IDs: a node's
// path, and an edge's "(<src> -> <dst>)[<n>]" under the common parent.
type LabelPlan struct {
	Nodes map[string]NodePlan
	Edges map[string][]LineRole
}

func newLabelPlan() LabelPlan {
	return LabelPlan{Nodes: map[string]NodePlan{}, Edges: map[string][]LineRole{}}
}

// hasDescription reports whether any card has a description line, which is
// drawn in d2's regular face (research R3).
func (p LabelPlan) hasDescription() bool {
	for _, n := range p.Nodes {
		if slices.Contains(n.Lines, roleDescription) {
			return true
		}
	}
	return false
}

// Key serialises the plan deterministically for the render-cache key
// (research R10): roles cannot be recovered from the D2 source text.
func (p LabelPlan) Key() []byte {
	var b bytes.Buffer
	for _, k := range sortedKeys(p.Nodes) {
		n := p.Nodes[k]
		b.WriteString("n\x00" + k + "\x00" + n.Icon + "\x00" + n.Colour + "\x00" + n.Address + "\x00")
		for _, r := range n.Lines {
			b.WriteByte(byte(r))
		}
		b.WriteByte('\n')
	}
	for _, k := range sortedKeys(p.Edges) {
		b.WriteString("e\x00" + k + "\x00")
		for _, r := range p.Edges[k] {
			b.WriteByte(byte(r))
		}
		b.WriteByte('\n')
	}
	return b.Bytes()
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// labelLine is one line of a label and its role.
type labelLine struct {
	text string
	role LineRole
}

// joinLines is the label text d2 draws, one tspan per line, and the role of
// each line in the same order.
func joinLines(lines []labelLine) (string, []LineRole) {
	texts := make([]string, len(lines))
	roles := make([]LineRole, len(lines))
	for i, l := range lines {
		texts[i], roles[i] = l.text, l.role
	}
	return strings.Join(texts, "\n"), roles
}

// kindTitles name a kind on a card's kind line. The emitter upper-cases it,
// not CSS, so viewers that ignore CSS and the .d2 output agree (research R8).
var kindTitles = map[string]string{
	vm.KindPerson: "PERSON", vm.KindSystem: "SYSTEM", vm.KindContainer: "CONTAINER",
	vm.KindComponent: "COMPONENT", vm.KindExternal: "EXTERNAL",
}

// isCard reports whether a node is drawn as an element card with an icon.
// Boundaries and the outside node keep their own treatment (FR-013).
func isCard(n vm.Node) bool { return n.Role == vm.RoleElement || n.Role == vm.RoleInstance }

// nodeLines is a card's label: the title or name in bold, the kind line
// (CONTAINER · Go, ending with the name when titled), then the description
// (FR-005). Boundaries and the outside node keep the label text they had.
func nodeLines(n vm.Node) []labelLine {
	switch {
	case n.Role == vm.RoleGroup:
		return plainLines(groupLabel(n))
	case n.Role == vm.RoleSubject:
		return plainLines(subjectLabel(n))
	case !isCard(n):
		return plainLines(n.Label)
	}
	var lines []labelLine
	if n.Title != "" {
		lines = appendLines(lines, wrapTitle(n.Title), roleTitle)
	} else {
		lines = append(lines, labelLine{n.Label, roleTitle})
	}
	var kind []string
	if k := kindTitles[n.Kind]; k != "" {
		kind = append(kind, k)
	}
	if n.Technology != "" {
		kind = append(kind, n.Technology)
	}
	if n.Title != "" {
		if len(kind) == 0 {
			kind = append(kind, "ELEMENT")
		}
		kind = append(kind, n.Label)
	}
	if len(kind) > 0 {
		lines = append(lines, labelLine{strings.Join(kind, " · "), roleKind})
	}
	if n.Description != "" {
		lines = append(lines, labelLine{"", roleBlank})
		lines = appendLines(lines, wrapDescription(n.Description), roleDescription)
	}
	return lines
}

// subjectLabel is a view subject's boundary label, unchanged by feature 020:
// its name or title over [Kind: Technology]. Its description is on the page.
func subjectLabel(n vm.Node) string {
	if n.Kind == "" {
		return n.Label
	}
	tag := kindTitles[n.Kind]
	if tag != "" {
		tag = tag[:1] + strings.ToLower(tag[1:])
	}
	if n.Technology != "" {
		if tag != "" {
			tag += ": "
		}
		tag += n.Technology
	}
	switch {
	case n.Title != "":
		return wrapTitle(n.Title) + "\n[" + cmp.Or(tag, "Element") + "] · " + n.Label
	case tag != "":
		return n.Label + "\n[" + tag + "]"
	}
	return n.Label
}

// edgeLines are an edge's description, its bare technology and its tags,
// each on its own line (FR-006, research R8).
func edgeLines(e vm.Edge) []labelLine {
	lines := edgeTextLines(e.Label, e.Technology)
	if len(e.Tags) > 0 {
		lines = append(lines, labelLine{"#" + strings.Join(e.Tags, " #"), roleEdgeTags})
	}
	return lines
}

func edgeTextLines(label, technology string) []labelLine {
	var lines []labelLine
	if label != "" {
		lines = appendLines(lines, label, roleEdgeLabel)
	}
	if technology != "" {
		lines = append(lines, labelLine{technology, roleEdgeTech})
	}
	return lines
}

// styledEdge reports whether the SVG pass has anything to restyle on an
// edge: lines that all keep d2's drawing stay out of the plan.
func styledEdge(roles []LineRole) bool {
	return slices.ContainsFunc(roles, func(r LineRole) bool { return r != roleEdgeLabel })
}

func appendLines(lines []labelLine, text string, role LineRole) []labelLine {
	for _, l := range strings.Split(text, "\n") {
		lines = append(lines, labelLine{l, role})
	}
	return lines
}

func plainLines(text string) []labelLine { return appendLines(nil, text, roleTitle) }

// titleWidth is the longest line a title is drawn with: D2 does not wrap plain
// labels, so a long title would otherwise stretch its box across the diagram.
const titleWidth = 40

// descWidth is the longest line a card's description is drawn with, for the
// same reason: the description went through appendLines un-wrapped, so one
// long sentence stretched its card to thousands of pixels and the whole
// diagram shrank into an unreadable ribbon (real-model report: a landscape
// canvas of 6,366 × 669 px, descriptions up to 282 characters on one line).
// Titles are names and stay wider; prose is cheaper to wrap.
const descWidth = 50

// wrapText breaks each paragraph at word boundaries into lines of at most
// width characters, never truncating. Existing line breaks are kept: an
// author's manual break is formatting, not something to re-flow.
func wrapText(text string, width int) string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		var lines []string
		line := ""
		for _, w := range strings.Fields(para) {
			switch {
			case line == "":
				line = w
			case len([]rune(line))+1+len([]rune(w)) <= width:
				line += " " + w
			default:
				lines = append(lines, line)
				line = w
			}
		}
		out = append(out, append(lines, line)...)
	}
	return strings.Join(out, "\n")
}

// wrapTitle breaks a title at word boundaries into lines of at most
// titleWidth characters, never truncating it.
func wrapTitle(title string) string {
	return wrapText(title, titleWidth)
}

// wrapDescription breaks a description at word boundaries into lines of at
// most descWidth characters, never truncating it (ADR-0021). The SVG pass
// styles every produced line as roleDescription, so line breaks are invisible
// to it.
func wrapDescription(text string) string {
	return wrapText(text, descWidth)
}

// groupLabel labels a deployment node like an element box: its title over
// [technology] and its name, or its name over [technology]. The description
// and tags are listed on the deployment page, never drawn (feature 018, R3).
// A node with none of them keeps its 1.1.0 label, the bare name.
func groupLabel(n vm.Node) string {
	switch {
	case n.Title != "" && n.Technology != "":
		return wrapTitle(n.Title) + "\n[" + n.Technology + "] · " + n.Label
	case n.Title != "":
		return wrapTitle(n.Title) + "\n" + n.Label
	case n.Technology != "":
		return n.Label + "\n[" + n.Technology + "]"
	}
	return n.Label
}
