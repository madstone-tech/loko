package arch

import "fmt"

// SchemaVersion is the shape of the exported IR (FR-036a).
//
// It is incremented only when the export's shape changes incompatibly, and it
// is deliberately independent of the loko release version: tying it to the
// release would change the exported bytes on every release and break the
// byte-identical guarantee of SC-003.
const SchemaVersion = 1

// IR is the immutable, fully-resolved compiled architecture. Every command and
// every later stage consumes this; none of them parse source themselves
// (FR-026).
//
// Every exported field is an ORDERED SLICE, never a map. Go randomises map
// iteration, so a map reaching an encoder is a latent nondeterminism bug that
// passes most of the time. Making the public shape map-free removes the
// failure mode structurally instead of relying on each encoder to sort
// (FR-040, research R6). Unexported lookup indexes exist for O(1) access and
// never reach an encoder.
//
// IR has no mutating methods (FR-025). Build one with NewIR.
type IR struct {
	SchemaVersion int            `json:"schemaVersion" toon:"schemaVersion"`
	Project       Project        `json:"project" toon:"project"`
	Elements      []Element      `json:"elements" toon:"elements"`
	Relationships []Relationship `json:"relationships" toon:"relationships"`
	Environments  []Environment  `json:"environments" toon:"environments"`
	Views         []View         `json:"views" toon:"views"`
	Ignores       []string       `json:"ignores" toon:"ignores"`

	byElement      map[Address]int
	byRelationship map[Address]int
	byEnvironment  map[Address]int
	byInstance     map[Address]instanceRef
	outgoing       map[Address][]int
	incoming       map[Address][]int
}

type instanceRef struct{ env, inst int }

// NewIR builds the lookup indexes over already-sorted slices and returns the
// finished value. Callers sort before calling; NewIR does not reorder, so the
// ordering rule lives in exactly one place (the IR builder use case).
func NewIR(project Project, elements []Element, rels []Relationship,
	envs []Environment, views []View, ignores []string) *IR {

	ir := &IR{
		SchemaVersion:  SchemaVersion,
		Project:        project,
		Elements:       elements,
		Relationships:  rels,
		Environments:   envs,
		Views:          views,
		Ignores:        ignores,
		byElement:      make(map[Address]int, len(elements)),
		byRelationship: make(map[Address]int, len(rels)),
		byEnvironment:  make(map[Address]int, len(envs)),
		byInstance:     make(map[Address]instanceRef),
		outgoing:       make(map[Address][]int),
		incoming:       make(map[Address][]int),
	}

	for i, e := range elements {
		ir.byElement[e.Address] = i
	}
	for i, r := range rels {
		ir.byRelationship[r.Address] = i
		ir.outgoing[r.Source] = append(ir.outgoing[r.Source], i)
		ir.incoming[r.Target] = append(ir.incoming[r.Target], i)
	}
	for ei, env := range envs {
		ir.byEnvironment[env.Address] = ei
		for ii, inst := range env.Instances {
			ir.byInstance[inst.Address] = instanceRef{env: ei, inst: ii}
		}
	}
	return ir
}

// Element returns the element at addr.
func (ir *IR) Element(addr Address) (Element, bool) {
	i, ok := ir.byElement[addr]
	if !ok {
		return Element{}, false
	}
	return ir.Elements[i], true
}

// Relationship returns the relationship at addr.
func (ir *IR) Relationship(addr Address) (Relationship, bool) {
	i, ok := ir.byRelationship[addr]
	if !ok {
		return Relationship{}, false
	}
	return ir.Relationships[i], true
}

// Environment returns the environment at addr.
func (ir *IR) Environment(addr Address) (Environment, bool) {
	i, ok := ir.byEnvironment[addr]
	if !ok {
		return Environment{}, false
	}
	return ir.Environments[i], true
}

// Instance returns the deployment instance at addr, from any environment.
func (ir *IR) Instance(addr Address) (Instance, bool) {
	ref, ok := ir.byInstance[addr]
	if !ok {
		return Instance{}, false
	}
	return ir.Environments[ref.env].Instances[ref.inst], true
}

// Has reports whether addr names anything in the compiled architecture.
func (ir *IR) Has(addr Address) bool {
	if _, ok := ir.byElement[addr]; ok {
		return true
	}
	if _, ok := ir.byRelationship[addr]; ok {
		return true
	}
	if _, ok := ir.byEnvironment[addr]; ok {
		return true
	}
	_, ok := ir.byInstance[addr]
	return ok
}

// ElementsByKind returns every element of the given kind, in IR order.
func (ir *IR) ElementsByKind(kind ElementKind) []Element {
	var out []Element
	for _, e := range ir.Elements {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// OutgoingFrom returns the relationships whose source is addr, in IR order.
func (ir *IR) OutgoingFrom(addr Address) []Relationship {
	return ir.relsAt(ir.outgoing[addr])
}

// IncomingTo returns the relationships whose target is addr, in IR order.
func (ir *IR) IncomingTo(addr Address) []Relationship {
	return ir.relsAt(ir.incoming[addr])
}

func (ir *IR) relsAt(idx []int) []Relationship {
	if len(idx) == 0 {
		return nil
	}
	out := make([]Relationship, 0, len(idx))
	for _, i := range idx {
		out = append(out, ir.Relationships[i])
	}
	return out
}

// Children returns the elements whose Parent is addr, in IR order.
func (ir *IR) Children(addr Address) []Element {
	var out []Element
	for _, e := range ir.Elements {
		if e.Parent == addr {
			out = append(out, e)
		}
	}
	return out
}

// CheckSchemaVersion validates a schemaVersion read from a previously exported
// artefact (FR-036b). An unrecognised version is refused outright, naming both
// what was found and what is supported, rather than read on a best-effort
// basis: a partial read of an incompatible shape is worse than a clear stop.
func CheckSchemaVersion(found int) error {
	if found == SchemaVersion {
		return nil
	}
	return fmt.Errorf(
		"unsupported export schema version %d: this build of loko reads version %d only",
		found, SchemaVersion)
}
