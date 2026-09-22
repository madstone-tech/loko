package arch

import "strings"

// Address is the stable identity of every element, relationship, environment,
// placement group, and instance in a compiled architecture. It is the key the
// IR is organised by and the key the diff stage compares on.
//
// The canonical forms are frozen for v1.x (data-model.md §1):
//
//	container.api                            element
//	container.api.uses.orders                relationship
//	deployment.prod                          environment
//	deployment.prod.node.vpc-main.subnet-a   placement group
//	deployment.prod.instance.api             instance
//	view.payment-path                        view
//
// An instance address deliberately carries no placement-group segment, so
// moving an instance between groups preserves its identity (FR-012, FR-024).
//
// Addresses are built by the constructors below and never by string
// concatenation at the call site, so the forms cannot drift.
type Address string

// Segment separators. These are not configurable: the diff stage depends on
// address stability across releases.
const (
	sepDot        = "."
	segUses       = "uses"
	segNode       = "node"
	segInstance   = "instance"
	prefixDeploy  = "deployment"
	prefixViewKey = "view"
)

// ElementKind enumerates the logical-plane element types (FR-006).
type ElementKind string

const (
	KindPerson    ElementKind = "person"
	KindSystem    ElementKind = "system"
	KindContainer ElementKind = "container"
	KindComponent ElementKind = "component"
	KindExternal  ElementKind = "external"
)

// AllElementKinds lists every logical element kind, in declaration order.
var AllElementKinds = []ElementKind{
	KindPerson, KindSystem, KindContainer, KindComponent, KindExternal,
}

// ValidElementKind reports whether s names a logical element kind.
func ValidElementKind(s string) bool {
	for _, k := range AllElementKinds {
		if string(k) == s {
			return true
		}
	}
	return false
}

// NewElementAddress returns the address of a logical element, e.g.
// "container.api".
func NewElementAddress(kind ElementKind, name string) Address {
	return Address(string(kind) + sepDot + name)
}

// NewRelationshipAddress returns the address of a relationship nested inside
// source, e.g. "container.api.uses.orders". The local name is what makes a
// retargeted edge readable as a rewire rather than a delete plus an add.
func NewRelationshipAddress(source Address, localName string) Address {
	return Address(string(source) + sepDot + segUses + sepDot + localName)
}

// NewEnvironmentAddress returns the address of a deployment environment, e.g.
// "deployment.prod".
func NewEnvironmentAddress(name string) Address {
	return Address(prefixDeploy + sepDot + name)
}

// NewGroupAddress returns the address of a placement group identified by its
// path from the environment root, e.g.
// "deployment.prod.node.vpc-main.subnet-a". An empty path returns env itself.
func NewGroupAddress(env Address, path []string) Address {
	if len(path) == 0 {
		return env
	}
	var b strings.Builder
	b.WriteString(string(env))
	b.WriteString(sepDot)
	b.WriteString(segNode)
	for _, p := range path {
		b.WriteString(sepDot)
		b.WriteString(p)
	}
	return Address(b.String())
}

// NewInstanceAddress returns the address of a deployment instance, e.g.
// "deployment.prod.instance.api".
//
// The placement group holding the instance is deliberately absent: instance
// names are unique per environment (FR-012a) precisely so that identity can be
// independent of placement.
func NewInstanceAddress(env Address, name string) Address {
	return Address(string(env) + sepDot + segInstance + sepDot + name)
}

// NewViewAddress returns the address of a view, e.g. "view.payment-path".
func NewViewAddress(name string) Address {
	return Address(prefixViewKey + sepDot + name)
}

// Prefix returns the leading segment of the address — the element kind for a
// logical address, or "deployment"/"view" otherwise. It returns "" for the
// empty address.
func (a Address) Prefix() string {
	s := string(a)
	if i := strings.Index(s, sepDot); i >= 0 {
		return s[:i]
	}
	return s
}

// Compare orders addresses byte-wise. Ordering must not depend on locale
// collation: identical source has to export byte-identically on every machine
// (FR-040), which is the precondition for golden-file testing.
func (a Address) Compare(b Address) int {
	return strings.Compare(string(a), string(b))
}

// String implements fmt.Stringer.
func (a Address) String() string { return string(a) }

// ValidName reports whether s is a legal address segment: a letter or
// underscore, followed by letters, digits, underscores, or hyphens.
//
// Written as an explicit scan rather than a regexp so that this package keeps
// no dependency beyond strings and stays cheap on large projects, where it is
// called once per declaration.
func ValidName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
			// Always legal.
		case c >= '0' && c <= '9', c == '-':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// String implements fmt.Stringer. Encoders that discover the interface — the
// TOON backend among them — render the kind as its lowercase name rather than
// rejecting a named string type.
func (k ElementKind) String() string { return string(k) }
