package viewmodel

// PageContext is an element page's place in the model: what it is part of,
// its siblings, its neighbours and the diagram that shows it in context
// (feature 021, data-model.md). Lists are sorted by address.
type PageContext struct {
	// Ancestors run from the top level down to the parent; empty at top level.
	Ancestors []LinkRef `json:"ancestors,omitempty"`
	// Siblings are the parent's children, or the top-level elements of the
	// same kind, including this element with Current set. Empty when alone.
	Siblings   []SiblingRow   `json:"siblings,omitempty"`
	Neighbours []NeighbourRow `json:"neighbours,omitempty"`
	// Prev and Next are the siblings either side; nil at the ends.
	Prev *LinkRef `json:"prev,omitempty"`
	Next *LinkRef `json:"next,omitempty"`
	// ContextDiagram is the view that shows this element among its siblings:
	// the parent's own view, else the first view containing the element, else
	// the landscape for a top-level element. Always a produced view, or "".
	ContextDiagram ViewID `json:"contextDiagram,omitempty"`
	// ProseInherited is set when the element's docs is its parent's file; the
	// page then links to the parent instead of repeating it (FR-009).
	ProseInherited bool      `json:"proseInherited,omitempty"`
	Source         SourceRef `json:"source"`
	// SiteSources feed the HTML page's generated-file notice: Sources plus the
	// files of everything the site page also depicts (parent, siblings,
	// neighbours, the In context view; research R9). Sources itself is left
	// as it was, so the markdown output does not change (FR-043).
	SiteSources []string `json:"siteSources,omitempty"`
}

// SiblingRow is one entry of a page's Siblings list.
type SiblingRow struct {
	Ref     LinkRef `json:"ref"`
	Current bool    `json:"current,omitempty"`
	// Calls: the page's element has a relationship to this sibling.
	Calls bool `json:"calls,omitempty"`
	// CalledBy: this sibling has a relationship to the page's element.
	CalledBy bool `json:"calledBy,omitempty"`
}

// Neighbour directions, seen from the page's element.
const (
	DirectionOut  = "out"
	DirectionIn   = "in"
	DirectionBoth = "both"
)

// NeighbourRow is one other element linked to the page's element, however
// many relationships link them. Its Other, Description, Technology and Tags
// fields match RelationRow's, so a theme's relation-table keeps working.
type NeighbourRow struct {
	Other     LinkRef `json:"other"`
	Direction string  `json:"direction"`
	// Technology joins the relationships' technologies, sorted and unique.
	Technology string `json:"technology,omitempty"`
	// Description is the first non-empty one in relationship-address order.
	Description   string   `json:"description,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Relationships int      `json:"relationships"`
	Sibling       bool     `json:"sibling,omitempty"`
}

// SourceRef is where an element is declared.
type SourceRef struct {
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"` // 1-based; 0 when unknown
}

// Overview holds the index page's counts.
type Overview struct {
	Systems       int `json:"systems"`
	Externals     int `json:"externals"`
	Containers    int `json:"containers"`
	DataStores    int `json:"dataStores"` // database, queue or topic shape
	Relationships int `json:"relationships"`
	Async         int `json:"async"`
	Warnings      int `json:"warnings"`
	Infos         int `json:"infos"`
}
