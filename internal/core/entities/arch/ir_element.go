package arch

// Logical-plane IR types. Every field here is resolved: parents and
// relationship endpoints are addresses that are known to exist.

// Element is a compiled logical-plane node.
type Element struct {
	Address Address     `json:"address"`
	Kind    ElementKind `json:"kind"`
	Name    string      `json:"name"`
	// Parent is the containing element. Empty for person, system, and external.
	Parent      Address `json:"parent,omitempty"`
	Description string  `json:"description,omitempty"`
	Owner       string  `json:"owner,omitempty"`
	Technology  string  `json:"technology,omitempty"`
	// Tags are sorted and de-duplicated at construction.
	Tags  []string    `json:"tags,omitempty"`
	Docs  string      `json:"docs,omitempty"`
	Range SourceRange `json:"range"`
}

// Relationship is a compiled directed edge.
//
// Address embeds the source element and the edge's local name
// ("container.api.uses.orders"), which is what lets the diff stage report a
// retargeted edge as a rewire instead of one edge vanishing and another
// appearing.
type Relationship struct {
	Address     Address     `json:"address"`
	Source      Address     `json:"source"`
	Target      Address     `json:"target"`
	LocalName   string      `json:"localName"`
	Description string      `json:"description,omitempty"`
	Technology  string      `json:"technology,omitempty"`
	Range       SourceRange `json:"range"`
}

// Project is the compiled project block.
type Project struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// LokoVersion is the raw constraint as authored, e.g. "~> 1.0".
	LokoVersion string `json:"lokoVersion,omitempty"`
}

// View is a compiled named subset of the architecture. Its references are
// resolved here and consumed by the renderer stage (FR-016).
type View struct {
	Address Address   `json:"address"`
	Name    string    `json:"name"`
	Include []Address `json:"include,omitempty"`
	Exclude []Address `json:"exclude,omitempty"`
	Tags    []string  `json:"tags,omitempty"`
}
