package arch

// View declarations, as authored. Split from source_model.go to keep both
// files inside the entity budget.

// ViewDecl is a named, filtered subset of the architecture. A view with steps
// is a dynamic view: an ordered flow (feature 019), which takes no include,
// exclude or tags.
type ViewDecl struct {
	Name string
	// Title and Description say what the view shows (feature 018); the name
	// stays its identifier. TitleRange is set when a title is written.
	Title       string
	TitleRange  SourceRange
	Description string
	Include     []Reference
	Exclude     []Reference
	Tags        []string
	// IncludeRange, ExcludeRange and TagsRange locate those attributes, so a
	// dynamic view that sets one is reported there (feature 019).
	IncludeRange SourceRange
	ExcludeRange SourceRange
	TagsRange    SourceRange
	// Direction is the layout direction; "" means the view's default.
	Direction      string
	DirectionRange SourceRange
	// Layout is the layout engine; "" means the project's.
	Layout      string
	LayoutRange SourceRange
	// Style draws a dynamic view with numbered edges ("edges", the default)
	// or as a sequence diagram ("sequence").
	Style      string
	StyleRange SourceRange
	// Steps are the flow, in declaration order: a step's number is its
	// position.
	Steps []StepDecl
	Range SourceRange
}

// StepDecl is a `step` block: one interaction of a dynamic view. The label is
// its stable identity; its number comes from its position.
type StepDecl struct {
	Name string
	From Reference
	To   Reference
	// Via names the relationship the step uses; zero when unset.
	Via         Reference
	Description string
	// Reply marks a step along a relationship in the opposite direction.
	Reply bool
	Range SourceRange
}
