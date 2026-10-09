package usecases

// ProjectInfo is the project block.
type ProjectInfo struct {
	Name        string `json:"name" toon:"name"`
	Description string `json:"description,omitempty" toon:"description,omitempty"`
	Layout      string `json:"layout,omitempty" toon:"layout,omitempty"`
}

// KindCount is how many elements of one kind exist.
type KindCount struct {
	Kind  string `json:"kind" toon:"kind"`
	Count int    `json:"count" toon:"count"`
}

// ElementView is an element at the requested level of detail.
type ElementView struct {
	Address string `json:"address" toon:"address"`
	Title   string `json:"title,omitempty" toon:"title,omitempty"`
	// Kind and Name repeat the address; only summary and full carry them,
	// which keeps the structure level within its token budget (R6).
	Kind        string   `json:"kind,omitempty" toon:"kind,omitempty"`
	Name        string   `json:"name,omitempty" toon:"name,omitempty"`
	Parent      string   `json:"parent,omitempty" toon:"parent,omitempty"`
	Technology  string   `json:"technology,omitempty" toon:"technology,omitempty"`
	Shape       string   `json:"shape,omitempty" toon:"shape,omitempty"`
	Components  int      `json:"components,omitempty" toon:"components,omitempty"`
	Description string   `json:"description,omitempty" toon:"description,omitempty"`
	Owner       string   `json:"owner,omitempty" toon:"owner,omitempty"`
	Tags        []string `json:"tags,omitempty" toon:"tags,omitempty"`
	Docs        string   `json:"docs,omitempty" toon:"docs,omitempty"`
}

// RelationshipView is one relationship.
type RelationshipView struct {
	Address     string   `json:"address" toon:"address"`
	Source      string   `json:"source" toon:"source"`
	Target      string   `json:"target" toon:"target"`
	Description string   `json:"description,omitempty" toon:"description,omitempty"`
	Technology  string   `json:"technology,omitempty" toon:"technology,omitempty"`
	Kind        string   `json:"kind,omitempty" toon:"kind,omitempty"`
	Tags        []string `json:"tags,omitempty" toon:"tags,omitempty"`
}

// EnvironmentView is a deployment environment; instances only at full.
type EnvironmentView struct {
	Address   string         `json:"address" toon:"address"`
	Name      string         `json:"name" toon:"name"`
	Instances []InstanceView `json:"instances,omitempty" toon:"instances,omitempty"`
}

// ViewInfo is a declared view; include and exclude only at full.
type ViewInfo struct {
	Address   string   `json:"address" toon:"address"`
	Tags      []string `json:"tags,omitempty" toon:"tags,omitempty"`
	Direction string   `json:"direction,omitempty" toon:"direction,omitempty"`
	Layout    string   `json:"layout,omitempty" toon:"layout,omitempty"`
	Include   []string `json:"include,omitempty" toon:"include,omitempty"`
	Exclude   []string `json:"exclude,omitempty" toon:"exclude,omitempty"`
}

// InstanceView is one placed instance.
type InstanceView struct {
	Address  string `json:"address" toon:"address"`
	Of       string `json:"of" toon:"of"`
	PlacedIn string `json:"placedIn,omitempty" toon:"placedIn,omitempty"`
}
