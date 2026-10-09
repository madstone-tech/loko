package arch

// SourceModel is what the parser adapter produces: a flat set of declarations
// exactly as authored, with references still unresolved.
//
// Keeping this separate from IR is the central structural decision of this
// feature (research R7). The adapter owns syntax; core owns the symbol table,
// reference resolution, and every validation rule. Because a SourceModel is a
// plain struct, the whole rule suite is testable from literals with no files
// on disk, which keeps the golden fixtures focused on parsing.
//
// An unresolvable reference is representable here on purpose — detecting it is
// core's job, not the parser's (FR-021).
type SourceModel struct {
	Project      ProjectDecl
	Elements     []ElementDecl
	Environments []EnvironmentDecl
	Views        []ViewDecl
	Ignores      []IgnorePattern
	// Moved records renames: each block says an element once declared at
	// From is now declared at To (feature 015).
	Moved []MovedDecl
	// Files lists every discovered source file, sorted, so the compiler can
	// report which inputs it actually read.
	Files []string
}

// Reference is a cross-element pointer before resolution: the traversal
// exactly as written, plus where it was written.
//
// References are extracted as static traversals and never evaluated
// (research R2). That is what keeps resolution in core, makes declaration
// order irrelevant, and lets the diff stage's `moved` block — whose operands
// deliberately name addresses that no longer exist — use the ordinary
// mechanism instead of an exception to the evaluator.
type Reference struct {
	Raw   string
	Range SourceRange
}

// IsZero reports whether the reference was omitted.
func (r Reference) IsZero() bool { return r.Raw == "" }

// ValueKind tags the variants of Value.
type ValueKind int

const (
	ValueNull ValueKind = iota
	ValueString
	ValueNumber
	ValueBool
	ValueList
	ValueMap
)

// Value is a closed sum over the scalar and collection shapes an attribute map
// can hold (FR-013). It exists so that evaluated attributes can cross the
// adapter boundary without core ever seeing the evaluator's own value type
// (FR-044).
type Value struct {
	Kind   ValueKind
	Str    string
	Num    float64
	Bool   bool
	List   []Value
	MapKV  []KeyValue
	Source SourceRange
}

// KeyValue is one entry of a Value map. Map entries are held as an ordered
// slice, not a Go map, so that ordering is explicit and the export stays
// byte-identical across runs (FR-040).
type KeyValue struct {
	Key   string
	Value Value
}

// StringValue builds a string Value.
func StringValue(s string) Value { return Value{Kind: ValueString, Str: s} }

// ProjectDecl carries the settings that replaced the previous configuration
// file (FR-003).
type ProjectDecl struct {
	Name        string
	Description string
	// Version is the raw constraint as authored, e.g. "~> 1.0". It is parsed
	// and evaluated during validation, not here.
	Version      string
	VersionRange SourceRange
	Range        SourceRange
	// Declared is false when no project block was found, which lets the
	// compiler distinguish an absent block from an empty one.
	Declared bool
}

// ElementDecl is one logical-plane declaration as authored.
type ElementDecl struct {
	Kind        ElementKind
	Name        string
	Description string
	Owner       string
	Technology  string
	Tags        []string
	Docs        string
	// Title and Shape affect rendering only (feature 016).
	Title string
	Shape string
	// Parent is the `system` reference on a container or the `container`
	// reference on a component. Zero for person, system, and external.
	Parent    Reference
	Relations []RelationDecl
	Range     SourceRange
	// AttrRanges locates each attribute by name so a diagnostic can point at
	// the offending value rather than at the whole block.
	AttrRanges map[string]SourceRange
}

// Address returns the declaration's address, e.g. "container.api".
func (e ElementDecl) Address() Address { return NewElementAddress(e.Kind, e.Name) }

// RelationDecl is a `uses` block, declared inside the element it originates
// from (FR-008). The block label is its local name, which is what gives the
// edge a stable address and lets a retarget read as a rewire.
type RelationDecl struct {
	LocalName   string
	Target      Reference
	Description string
	Technology  string
	// Kind ("" means sync) and Tags affect rendering only (feature 016).
	Kind      string
	KindRange SourceRange
	Tags      []string
	Range     SourceRange
}

// ViewDecl is a named, filtered subset of the architecture. Validated in this
// release and rendered in a later one (FR-016).
type ViewDecl struct {
	Name    string
	Include []Reference
	Exclude []Reference
	Tags    []string
	// Direction is the layout direction; "" means the view's default.
	Direction      string
	DirectionRange SourceRange
	Range          SourceRange
}

// IgnorePattern is a physical-resource pattern the project declares it does
// not model. Carried through untouched for the reconciliation stage (FR-015).
type IgnorePattern struct {
	Pattern string
	Range   SourceRange
}

// MovedDecl is a `moved` block: a rename recorded in the source so history
// survives it. From names an address that no longer exists; To names the
// element it became, possibly of another kind.
type MovedDecl struct {
	From  Reference
	To    Reference
	Range SourceRange
}
