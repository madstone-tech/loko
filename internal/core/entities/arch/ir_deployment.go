package arch

// Deployment-plane IR types.

// Environment is a compiled deployment target.
//
// Instances is FLAT and COMPLETE: every instance in the environment appears
// there regardless of nesting depth, sorted by address. Groups preserves the
// placement tree, and placement is recorded as a cross-reference via
// Group.Contains and Instance.PlacedIn.
//
// That shape is what makes the clarified addressing rule expressible. An
// instance is reachable without walking the tree, so its identity does not
// depend on where it sits, while the nesting is still fully preserved for the
// renderer and the reconciler.
type Environment struct {
	Address   Address    `json:"address" toon:"address"`
	Name      string     `json:"name" toon:"name"`
	Provider  string     `json:"provider,omitempty" toon:"provider,omitempty"`
	Account   string     `json:"account,omitempty" toon:"account,omitempty"`
	Region    string     `json:"region,omitempty" toon:"region,omitempty"`
	Groups    []Group    `json:"groups,omitempty" toon:"groups,omitempty"`
	Instances []Instance `json:"instances" toon:"instances"`
}

// Group is a compiled placement group. It has its own address reflecting its
// position in the nesting but contributes no segment to the instances it
// holds.
type Group struct {
	Address Address `json:"address" toon:"address"`
	Name    string  `json:"name" toon:"name"`
	Groups  []Group `json:"groups,omitempty" toon:"groups,omitempty"`
	// Contains lists the addresses of instances placed directly in this group,
	// sorted.
	Contains []Address `json:"contains,omitempty" toon:"contains,omitempty"`
}

// Instance is a compiled realisation of one logical element in one
// environment.
type Instance struct {
	// Address carries no placement-group segment (FR-012, FR-024), so moving
	// an instance between groups preserves its identity.
	Address Address `json:"address" toon:"address"`
	Name    string  `json:"name" toon:"name"`
	// Of is the resolved logical element this instance realises.
	Of Address `json:"of" toon:"of"`
	// PlacedIn is the group holding it; empty when directly under the
	// environment.
	PlacedIn Address `json:"placedIn,omitempty" toon:"placedIn,omitempty"`
	// Attributes are sorted by key. An ordered slice rather than a map, so
	// ordering is explicit rather than left to an encoder (FR-040).
	Attributes []Attribute `json:"attributes,omitempty" toon:"attributes,omitempty"`
	// Claims are sorted by kind, then by address.
	Claims []Claim `json:"claims,omitempty" toon:"claims,omitempty"`
}

// Attribute is one environment-specific key/value on an instance.
type Attribute struct {
	Key   string `json:"key" toon:"key"`
	Value Value  `json:"value" toon:"value"`
}

// Claim is a compiled binding: an assertion that an instance corresponds to
// one or more physical resources.
type Claim struct {
	Kind      ClaimKind   `json:"kind" toon:"kind"`
	Address   string      `json:"address,omitempty" toon:"address,omitempty"`
	Addresses []string    `json:"addresses,omitempty" toon:"addresses,omitempty"`
	Tags      []Attribute `json:"tags,omitempty" toon:"tags,omitempty"`
	Range     SourceRange `json:"range" toon:"range"`
}

// Identifiers returns every physical identifier this claim selects by exact
// address or glob, so duplicate-claim detection can compare them without
// caring which selector form was authored.
func (c Claim) Identifiers() []string {
	if c.Address != "" {
		return []string{c.Address}
	}
	return c.Addresses
}
