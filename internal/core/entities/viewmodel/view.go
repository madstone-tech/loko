package viewmodel

import "sort"

// ViewKind classifies how a view was obtained (research R3).
type ViewKind string

const (
	KindLandscapeView  ViewKind = "landscape"
	KindSystemView     ViewKind = "system"
	KindContainerView  ViewKind = "container"
	KindDeploymentView ViewKind = "deployment"
	KindDeclaredView   ViewKind = "declared"
)

// rank is the kind's position in the view sort order.
func (k ViewKind) rank() int {
	switch k {
	case KindLandscapeView:
		return 0
	case KindSystemView:
		return 1
	case KindContainerView:
		return 2
	case KindDeploymentView:
		return 3
	default:
		return 4
	}
}

// ViewID is a view's stable identity. It is a function of what the view
// depicts, never of position or discovery order (FR-006), and it names the
// files the view produces.
type ViewID string

// Selection is a declared view's include / exclude / tag selection, carried
// as addresses. viewmodel does not import the IR package: the entity layer may
// import only the standard library.
type Selection struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// View is a named slice of the architecture at one level of detail.
type View struct {
	ID    ViewID   `json:"id"`
	Kind  ViewKind `json:"kind"`
	Title string   `json:"title"`
	// TitleDeclared is set when the author wrote the title, which is then
	// drawn on the diagram; otherwise Title is the name and nothing is drawn
	// (feature 018). Derived, so not encoded.
	TitleDeclared bool `json:"-"`
	// Description is a declared view's description, shown under the page
	// heading and never drawn (feature 018).
	Description string `json:"description,omitempty"`
	// Subject is the address of the system, container, environment or
	// view.<label> the view depicts. Empty for the landscape.
	Subject   string     `json:"subject,omitempty"`
	Selection *Selection `json:"selection,omitempty"`
	// Direction is the layout direction, "down" or "right"; the projection
	// always sets it (feature 016, research R5).
	Direction string `json:"direction"`
	// Layout is the layout engine; "" means dagre, the default.
	Layout string `json:"layout,omitempty"`
	// Style is "sequence" for a dynamic view drawn as a sequence diagram;
	// empty for numbered edges, the default (feature 019).
	Style string `json:"style,omitempty"`
}

// SortViews orders views by kind (landscape, system, container, deployment,
// declared), then by ID.
func SortViews(vs []View) {
	sort.SliceStable(vs, func(i, j int) bool {
		if ri, rj := vs[i].Kind.rank(), vs[j].Kind.rank(); ri != rj {
			return ri < rj
		}
		return vs[i].ID < vs[j].ID
	})
}

// Mode is how a view came to be, for its page label and index chip (feature
// 021, data-model "View mode"): sequence, declared, deployment or automatic.
func (v View) Mode() string {
	switch {
	case v.Style == "sequence":
		return "sequence"
	case v.Kind == KindDeclaredView:
		return "declared"
	case v.Kind == KindDeploymentView:
		return "deployment"
	default:
		return "automatic"
	}
}
