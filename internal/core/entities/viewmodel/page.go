package viewmodel

// LinkRef is a resolved link to an element's page.
type LinkRef struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	// Title is the element's display title, when it has one (feature 016).
	Title    string `json:"title,omitempty"`
	Kind     string `json:"kind"`
	PagePath string `json:"pagePath"`
}

// RelationRow is one row of a page's uses / used-by table. It deliberately
// omits the other element's description, so editing one element's description
// does not change every neighbour's page (FR-022, SC-003).
type RelationRow struct {
	Relationship string  `json:"relationship"`
	Other        LinkRef `json:"other"`
	Description  string  `json:"description,omitempty"`
	Technology   string  `json:"technology,omitempty"`
	// Tags are the relationship's tags (feature 016, FR-010).
	Tags []string `json:"tags,omitempty"`
}

// ElementPage is the input to the prose outputs for one logical element
// (FR-025).
type ElementPage struct {
	Address     string   `json:"address"`
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Title       string   `json:"title,omitempty"`
	Shape       string   `json:"shape,omitempty"`
	Description string   `json:"description,omitempty"`
	Technology  string   `json:"technology,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Classes     []string `json:"classes,omitempty"`
	// Prose is the markdown text of the element's docs file.
	Prose     string `json:"prose,omitempty"`
	ProsePath string `json:"prosePath,omitempty"`
	// ProseMissing is set when docs was declared but the file does not exist;
	// the page is still produced (FR-026).
	ProseMissing bool          `json:"proseMissing,omitempty"`
	Parent       *LinkRef      `json:"parent,omitempty"`
	Children     []LinkRef     `json:"children,omitempty"`
	Uses         []RelationRow `json:"uses,omitempty"`
	UsedBy       []RelationRow `json:"usedBy,omitempty"`
	// Diagram is the view this page embeds; always one that was produced.
	Diagram  ViewID `json:"diagram,omitempty"`
	PagePath string `json:"pagePath"`
	// Sources feed the generated-file notice of this page.
	Sources []string `json:"sources"`
}

// ProjectHeader is the project block, as the outputs show it.
type ProjectHeader struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Projection is the complete intermediate value handed to every backend. It is
// the only thing a backend reads (FR-011).
type Projection struct {
	Project      ProjectHeader `json:"project"`
	Views        []ViewModel   `json:"views"`
	Pages        []ElementPage `json:"pages"`
	Environments []LinkRef     `json:"environments,omitempty"`
	// Sources is every source file in the project; it feeds the notice of
	// site-wide files such as index pages and assets.
	Sources []string `json:"sources"`
}

// View returns the view model with the given ID.
func (p *Projection) View(id ViewID) (ViewModel, bool) {
	for _, v := range p.Views {
		if v.View.ID == id {
			return v, true
		}
	}
	return ViewModel{}, false
}
