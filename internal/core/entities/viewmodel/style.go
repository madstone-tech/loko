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

	// Element shapes chosen by the `shape` attribute (feature 016).
	ShapeDatabase Shape = "database"
	ShapeQueue    Shape = "queue"
	ShapeTopic    Shape = "topic"
	ShapeFunction Shape = "function"
	ShapeBucket   Shape = "bucket"
)

// Style is the visual treatment the projection assigns to a node. Backends
// translate it into their own syntax and decide nothing themselves (FR-008).
// Mapping Shape and the kind to a backend's own glyph, as the SVG backend's
// icons do, is such a translation (feature 020, research R6).
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
	// Dashed marks an edge crossing the view's boundary.
	Dashed bool `json:"dashed,omitempty"`
	// Async marks an edge whose relationships are all async (feature 016);
	// it is drawn with a different dash than a crossing edge.
	Async bool `json:"async,omitempty"`
}

type palette struct {
	shape              Shape
	fill, stroke, font string
	dashed             bool
}

// elementPalette is the C4 convention of a darker fill for a higher level, in
// the Madstone Signal Blue ramp (feature 020, research R5). Every text line
// passes WCAG AA on its fill. Element kind determines shape consistently
// across every view and backend (FR-016); an external is visibly distinguished
// by its dark neutral fill and dashed outline (FR-017).
var elementPalette = map[string]palette{
	KindPerson:    {ShapePerson, "#1A4275", "#14365F", "#FFFFFF", false},
	KindSystem:    {ShapeRectangle, "#1E4F8F", "#183F73", "#FFFFFF", false},
	KindContainer: {ShapeRectangle, "#2D66AE", "#245390", "#FFFFFF", false},
	KindComponent: {ShapeRectangle, "#93C5FD", "#60A5FA", "#0F1117", false},
	KindExternal:  {ShapeRectangle, "#525252", "#3D3D3D", "#FFFFFF", true},
}

// EdgeColours is the stroke and label colour of every edge.
type EdgeColours struct{ Stroke, Font string }

// EdgePalette draws edges in brand neutrals, not the renderer's default blue
// (feature 020, FR-009).
var EdgePalette = EdgeColours{Stroke: "#737373", Font: "#3D3D3D"}

// StyleFor returns the style for a node. It is a pure function of its inputs
// (FR-019).
func StyleFor(role NodeRole, kind string, tags []string) Style {
	var s Style
	switch role {
	case RoleSubject, RoleGroup:
		s = Style{Shape: ShapeBoundary, Stroke: "#737373", FontColor: "#3D3D3D", Dashed: true}
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
		cs = append(cs, "kind-"+Ident(kind))
		if kind == KindExternal {
			cs = append(cs, "external")
		}
	}
	for _, t := range tags {
		cs = append(cs, "tag-"+Ident(t))
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

// WithShape applies an element's `shape` attribute to its style. Colours stay
// the kind's, so a shaped container is still recognisably a container; an
// empty shape keeps the kind's default.
func WithShape(s Style, shape string) Style {
	if shape != "" {
		s.Shape = Shape(shape)
	}
	return s
}
