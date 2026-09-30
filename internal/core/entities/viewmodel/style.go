package viewmodel

import "sort"

// Element kinds, mirrored from the IR as plain strings: this package may
// import only the standard library.
const (
	KindPerson    = "person"
	KindSystem    = "system"
	KindContainer = "container"
	KindComponent = "component"
	KindExternal  = "external"
)

// Shape is a node's drawn outline.
type Shape string

const (
	ShapePerson    Shape = "person"
	ShapeRectangle Shape = "rectangle"
	ShapeBoundary  Shape = "boundary"
	ShapeOval      Shape = "oval"
)

// Style is the visual treatment the projection assigns to a node. Backends
// translate it into their own syntax and decide nothing themselves (FR-008).
type Style struct {
	Shape     Shape    `json:"shape"`
	Fill      string   `json:"fill,omitempty"`
	Stroke    string   `json:"stroke,omitempty"`
	FontColor string   `json:"fontColor,omitempty"`
	Dashed    bool     `json:"dashed,omitempty"`
	Classes   []string `json:"classes,omitempty"`
}

// EdgeStyle is the visual treatment of an edge.
type EdgeStyle struct {
	Dashed bool `json:"dashed,omitempty"`
}

type palette struct {
	shape              Shape
	fill, stroke, font string
	dashed             bool
}

// elementPalette is the C4 convention table of research R6. Element kind
// determines shape consistently across every view and backend (FR-016); an
// external is visibly distinguished by its dashed outline (FR-017).
var elementPalette = map[string]palette{
	KindPerson:    {ShapePerson, "#08427b", "#073b6f", "#ffffff", false},
	KindSystem:    {ShapeRectangle, "#1168bd", "#0b4884", "#ffffff", false},
	KindContainer: {ShapeRectangle, "#438dd5", "#3c7fc0", "#ffffff", false},
	KindComponent: {ShapeRectangle, "#85bbf0", "#78a8d8", "#000000", false},
	KindExternal:  {ShapeRectangle, "#999999", "#8a8a8a", "#ffffff", true},
}

// StyleFor returns the style for a node. It is a pure function of its inputs
// (FR-019).
func StyleFor(role NodeRole, kind string, tags []string) Style {
	var s Style
	switch role {
	case RoleSubject, RoleGroup:
		s = Style{Shape: ShapeBoundary, Stroke: "#444444", FontColor: "#444444", Dashed: true}
	case RoleOutside:
		s = Style{Shape: ShapeOval, Stroke: "#999999", FontColor: "#666666", Dashed: true}
	default:
		p, ok := elementPalette[kind]
		if !ok {
			p = elementPalette[KindSystem]
		}
		s = Style{Shape: p.shape, Fill: p.fill, Stroke: p.stroke, FontColor: p.font, Dashed: p.dashed}
	}
	s.Classes = classesFor(kind, tags)
	return s
}

// classesFor carries the kind and every tag into the output as CSS classes, so
// a user's own styling can target them (FR-018).
func classesFor(kind string, tags []string) []string {
	var cs []string
	if kind != "" {
		cs = append(cs, "kind-"+Segment(kind))
		if kind == KindExternal {
			cs = append(cs, "external")
		}
	}
	for _, t := range tags {
		cs = append(cs, "tag-"+Segment(t))
	}
	sort.Strings(cs)
	return dedupe(cs)
}

func dedupe(sorted []string) []string {
	if len(sorted) < 2 {
		return sorted
	}
	out := sorted[:1]
	for _, s := range sorted[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}
