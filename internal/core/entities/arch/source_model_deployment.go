package arch

// Deployment-plane declarations, as authored. Split from source_model.go to
// keep both files inside the 300 effective-line entity budget.

// EnvironmentDecl is a `deployment` block: a named environment that
// instantiates the logical architecture (FR-011).
type EnvironmentDecl struct {
	Name     string
	Provider string
	Account  string
	Region   string
	// Groups are `node` blocks nested directly under the environment.
	Groups []GroupDecl
	// Instances are declared directly under the environment, outside any node.
	Instances []InstanceDecl
	Range     SourceRange
}

// GroupDecl is a `node` block. Nests to arbitrary depth (FR-012).
//
// A group carries its own address but contributes no segment to the addresses
// of the instances inside it, so re-parenting an instance preserves its
// identity (FR-024). Instance names are therefore unique per environment,
// not per group (FR-012a).
type GroupDecl struct {
	Name      string
	Groups    []GroupDecl
	Instances []InstanceDecl
	Range     SourceRange
}

// InstanceDecl realises one logical element inside one environment (FR-013).
type InstanceDecl struct {
	Name string
	// Of references the logical element this instance realises. Required.
	Of Reference
	// Attributes are environment-specific values such as sizing or limits,
	// held as an ordered slice so the export stays byte-stable (FR-040).
	Attributes []KeyValue
	Claims     []ClaimDecl
	Range      SourceRange
}

// ClaimKind identifies the infrastructure format a claim selects against.
// Only these two ship in v1.0 (FR-014).
type ClaimKind string

const (
	ClaimTerraform      ClaimKind = "terraform"
	ClaimCloudFormation ClaimKind = "cloudformation"
)

// AllClaimKinds lists every binding kind this release accepts.
var AllClaimKinds = []ClaimKind{ClaimTerraform, ClaimCloudFormation}

// ValidClaimKind reports whether s names a supported binding kind.
func ValidClaimKind(s string) bool {
	for _, k := range AllClaimKinds {
		if string(k) == s {
			return true
		}
	}
	return false
}

// ClaimDecl is a `binding` block: an assertion that an instance corresponds to
// one or more physical resources (FR-014).
//
// Exactly one of Address, Addresses, or Tags selects the resources. Which one
// was used is preserved rather than normalised, because the reconciliation
// stage treats an exact address, a glob, and a tag match differently.
type ClaimDecl struct {
	Kind ClaimKind
	// Address is an exact identifier or a glob such as "module.api.*".
	Address   string
	Addresses []string
	Tags      []KeyValue
	Range     SourceRange
}

// Identifiers returns every physical identifier this claim selects by exact
// address or glob, so duplicate-claim detection can compare them without
// caring which selector form was authored. A tag selector matches at
// reconcile time against observed resources, so it yields nothing here.
func (c ClaimDecl) Identifiers() []string {
	if c.Address != "" {
		return []string{c.Address}
	}
	return c.Addresses
}

// Selectors reports how many of the three selector forms were supplied. Zero
// or more than one is a validation error, raised in core rather than here.
func (c ClaimDecl) Selectors() int {
	n := 0
	if c.Address != "" {
		n++
	}
	if len(c.Addresses) > 0 {
		n++
	}
	if len(c.Tags) > 0 {
		n++
	}
	return n
}

// AllInstances walks the environment and returns every instance regardless of
// nesting depth, paired with the placement path that holds it.
//
// Core uses this to flatten instances onto Environment.Instances while
// recording placement as a cross-reference — the shape that makes an
// instance's identity independent of where it sits.
func (e EnvironmentDecl) AllInstances() []PlacedInstance {
	var out []PlacedInstance
	for _, inst := range e.Instances {
		out = append(out, PlacedInstance{Instance: inst})
	}
	for _, g := range e.Groups {
		out = append(out, collectInstances(g, nil)...)
	}
	return out
}

// PlacedInstance pairs an instance declaration with the node path holding it.
// A nil or empty Path means the instance sits directly under the environment.
type PlacedInstance struct {
	Instance InstanceDecl
	Path     []string
}

func collectInstances(g GroupDecl, prefix []string) []PlacedInstance {
	path := make([]string, len(prefix), len(prefix)+1)
	copy(path, prefix)
	path = append(path, g.Name)

	var out []PlacedInstance
	for _, inst := range g.Instances {
		// Copy the path per instance: callers retain these slices, and a
		// shared backing array would alias as the walk continues.
		p := make([]string, len(path))
		copy(p, path)
		out = append(out, PlacedInstance{Instance: inst, Path: p})
	}
	for _, child := range g.Groups {
		out = append(out, collectInstances(child, path)...)
	}
	return out
}

// String implements fmt.Stringer, for the same reason ElementKind does.
func (k ClaimKind) String() string { return string(k) }
