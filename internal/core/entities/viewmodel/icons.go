package viewmodel

// ShapeIcons and KindIcons are the icon table (feature 020 FR-002, FR-011):
// the Lucide icon that says what an element is. A declared shape wins over
// the kind. The table lives in the projection's package so that the diagram
// and the site draw the same icon for an element (feature 021); each backend
// only translates the name into its own glyph.
var ShapeIcons = map[Shape]string{
	ShapePerson:   "user",
	ShapeDatabase: "database",
	ShapeQueue:    "list-ordered",
	ShapeTopic:    "megaphone",
	ShapeFunction: "square-function",
	ShapeBucket:   "archive",
}

var KindIcons = map[string]string{
	KindPerson:    "user",
	KindSystem:    "boxes",
	KindContainer: "container",
	KindComponent: "blocks",
	KindExternal:  "globe",
}

// FallbackIcon is drawn for a kind without an icon; the d2 package's
// TestIconTableComplete fails when that happens (feature 020 FR-017).
const FallbackIcon = "boxes"

// IconFor is the icon name for an element of the given shape and kind.
func IconFor(shape Shape, kind string) string {
	if name, ok := ShapeIcons[shape]; ok {
		return name
	}
	if name, ok := KindIcons[kind]; ok {
		return name
	}
	return FallbackIcon
}
