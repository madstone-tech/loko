package viewmodel

// StepRow is one step of a dynamic view, in order (feature 019): a row of the
// page's Steps table, and an arrow of the sequence diagram, which needs the
// order the sorted edges do not keep.
type StepRow struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	// From and To link the step's elements' pages; FromNode and ToNode are
	// their node IDs in this view.
	From       LinkRef `json:"from"`
	To         LinkRef `json:"to"`
	FromNode   string  `json:"fromNode"`
	ToNode     string  `json:"toNode"`
	Text       string  `json:"text,omitempty"`
	Technology string  `json:"technology,omitempty"`
	Reply      bool    `json:"reply,omitempty"`
}

// StyleSequence is View.Style for a dynamic view drawn as a sequence diagram.
const StyleSequence = "sequence"
