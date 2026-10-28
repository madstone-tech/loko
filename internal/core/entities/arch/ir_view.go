package arch

// View is a compiled named subset of the architecture. Its references are
// resolved here and consumed by the renderer stage (FR-016).
type View struct {
	Address     Address   `json:"address" toon:"address"`
	Name        string    `json:"name" toon:"name"`
	Title       string    `json:"title,omitempty" toon:"title,omitempty"`
	Description string    `json:"description,omitempty" toon:"description,omitempty"`
	Include     []Address `json:"include,omitempty" toon:"include,omitempty"`
	Exclude     []Address `json:"exclude,omitempty" toon:"exclude,omitempty"`
	Tags        []string  `json:"tags,omitempty" toon:"tags,omitempty"`
	// Direction is the authored layout direction; the default is applied at projection.
	Direction string `json:"direction,omitempty" toon:"direction,omitempty"`
	// Layout is the authored layout engine; the project default is applied at projection.
	Layout string `json:"layout,omitempty" toon:"layout,omitempty"`
	// Style is the authored drawing style of a dynamic view (feature 019).
	Style string `json:"style,omitempty" toon:"style,omitempty"`
	// Steps is the flow of a dynamic view, in declaration order: order is its
	// meaning, so it is not sorted by address.
	Steps []Step `json:"steps,omitempty" toon:"steps,omitempty"`
}

// Step is one compiled interaction of a dynamic view.
type Step struct {
	Address Address `json:"address" toon:"address"`
	Name    string  `json:"name" toon:"name"`
	// Number is the step's position in its view, from 1.
	Number int     `json:"number" toon:"number"`
	From   Address `json:"from" toon:"from"`
	To     Address `json:"to" toon:"to"`
	// Via is the relationship the author named; Relationship is the one the
	// step uses: Via, or the first matching relationship, or empty when none
	// matches.
	Via          Address     `json:"via,omitempty" toon:"via,omitempty"`
	Relationship Address     `json:"relationship,omitempty" toon:"relationship,omitempty"`
	Reply        bool        `json:"reply,omitempty" toon:"reply,omitempty"`
	Description  string      `json:"description,omitempty" toon:"description,omitempty"`
	Range        SourceRange `json:"range" toon:"range"`
}
