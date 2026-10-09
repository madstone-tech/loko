package arch

// Logical-plane IR types. Every field here is resolved: parents and
// relationship endpoints are addresses that are known to exist.

// Element is a compiled logical-plane node.
type Element struct {
	Address Address     `json:"address" toon:"address"`
	Kind    ElementKind `json:"kind" toon:"kind"`
	Name    string      `json:"name" toon:"name"`
	// Parent is the containing element. Empty for person, system, and external.
	Parent      Address `json:"parent,omitempty" toon:"parent,omitempty"`
	Description string  `json:"description,omitempty" toon:"description,omitempty"`
	Owner       string  `json:"owner,omitempty" toon:"owner,omitempty"`
	Technology  string  `json:"technology,omitempty" toon:"technology,omitempty"`
	// Tags are sorted and de-duplicated at construction.
	Tags []string `json:"tags,omitempty" toon:"tags,omitempty"`
	Docs string   `json:"docs,omitempty" toon:"docs,omitempty"`
	// Title and Shape affect rendering only (feature 016); omitted when unset.
	Title string      `json:"title,omitempty" toon:"title,omitempty"`
	Shape string      `json:"shape,omitempty" toon:"shape,omitempty"`
	Range SourceRange `json:"range" toon:"range"`
}

// Relationship is a compiled directed edge.
//
// Address embeds the source element and the edge's local name
// ("container.api.uses.orders"), which is what lets the diff stage report a
// retargeted edge as a rewire instead of one edge vanishing and another
// appearing.
type Relationship struct {
	Address     Address `json:"address" toon:"address"`
	Source      Address `json:"source" toon:"source"`
	Target      Address `json:"target" toon:"target"`
	LocalName   string  `json:"localName" toon:"localName"`
	Description string  `json:"description,omitempty" toon:"description,omitempty"`
	Technology  string  `json:"technology,omitempty" toon:"technology,omitempty"`
	// Kind is async or trigger; sync is the default and never stored. Kind
	// and Tags affect rendering only (feature 016).
	Kind  string      `json:"kind,omitempty" toon:"kind,omitempty"`
	Tags  []string    `json:"tags,omitempty" toon:"tags,omitempty"`
	Range SourceRange `json:"range" toon:"range"`
}

// Project is the compiled project block.
type Project struct {
	Name        string `json:"name" toon:"name"`
	Description string `json:"description,omitempty" toon:"description,omitempty"`
	// LokoVersion is the raw constraint as authored, e.g. "~> 1.0".
	LokoVersion string `json:"lokoVersion,omitempty" toon:"lokoVersion,omitempty"`
	// Layout is the authored default layout engine for every view.
	Layout string `json:"layout,omitempty" toon:"layout,omitempty"`
}

// View is a compiled named subset of the architecture. Its references are
// resolved here and consumed by the renderer stage (FR-016).
type View struct {
	Address Address   `json:"address" toon:"address"`
	Name    string    `json:"name" toon:"name"`
	Include []Address `json:"include,omitempty" toon:"include,omitempty"`
	Exclude []Address `json:"exclude,omitempty" toon:"exclude,omitempty"`
	Tags    []string  `json:"tags,omitempty" toon:"tags,omitempty"`
	// Direction is the authored layout direction; the default is applied at projection.
	Direction string `json:"direction,omitempty" toon:"direction,omitempty"`
	// Layout is the authored layout engine; the project default is applied at projection.
	Layout string `json:"layout,omitempty" toon:"layout,omitempty"`
}

// Move is a compiled `moved` block.
type Move struct {
	From  Address     `json:"from" toon:"from"`
	To    Address     `json:"to" toon:"to"`
	Range SourceRange `json:"range" toon:"range"`
}
