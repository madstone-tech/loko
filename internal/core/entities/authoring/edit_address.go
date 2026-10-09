package authoring

import (
	"slices"
	"strings"
)

// ElementKinds are the five element block types, sorted.
var ElementKinds = []string{"component", "container", "external", "person", "system"}

// isElementAddress reports whether a is "kind.name" for an element kind.
func isElementAddress(a string) bool {
	kind, name, ok := strings.Cut(a, ".")
	return ok && slices.Contains(ElementKinds, kind) && ValidName(name)
}

// AddressParts is an edit address split into what the editor needs to find or
// place the declaration.
type AddressParts struct {
	// Element is the element address: the element itself, or a relationship's source.
	Element string
	// Local is a relationship's or instance's own name.
	Local string
	// Env is the environment name for deployment targets.
	Env string
	// Groups is the node path: the group itself, or where to place a new instance.
	Groups []string
}

// SplitAddress parses a for the given target kind. The forms are:
//
//	element       kind.name
//	relationship  kind.name.uses.local
//	environment   deployment.env
//	group         deployment.env.node.g1[.g2…]
//	instance      deployment.env[.node.g1[.g2…]].instance.name
//	binding       the instance's address; the binding is picked by BindingRef
//	view          view.name
//
// An instance's identity is env-scoped (deployment.env.instance.name). The
// optional node path places a new instance inside a group; it is ignored when
// locating an existing one.
func SplitAddress(t TargetKind, a string) (AddressParts, bool) {
	seg := strings.Split(a, ".")
	if slices.ContainsFunc(seg, func(s string) bool { return !ValidName(s) }) {
		return AddressParts{}, false
	}
	switch t {
	case TargetElement:
		return AddressParts{Element: a}, len(seg) == 2 && isElementAddress(a)
	case TargetRelationship:
		el := strings.Join(seg[:min(2, len(seg))], ".")
		return AddressParts{Element: el, Local: seg[len(seg)-1]},
			len(seg) == 4 && seg[2] == "uses" && isElementAddress(el)
	case TargetEnvironment:
		return AddressParts{Env: seg[len(seg)-1]}, len(seg) == 2 && seg[0] == "deployment"
	case TargetGroup:
		ok := len(seg) >= 4 && seg[0] == "deployment" && seg[2] == "node" &&
			!slices.Contains(seg[3:], "instance")
		return AddressParts{Env: seg[1], Groups: tail(seg, 3)}, ok
	case TargetInstance, TargetBinding:
		return splitInstance(seg)
	case TargetView:
		return AddressParts{Local: seg[len(seg)-1]}, len(seg) == 2 && seg[0] == "view"
	}
	return AddressParts{}, false
}

func splitInstance(seg []string) (AddressParts, bool) {
	n := len(seg)
	if n < 4 || seg[0] != "deployment" || seg[n-2] != "instance" {
		return AddressParts{}, false
	}
	p := AddressParts{Env: seg[1], Local: seg[n-1]}
	switch {
	case n == 4:
		return p, true
	case n >= 6 && seg[2] == "node" && !slices.Contains(seg[3:n-2], "instance"):
		p.Groups = slices.Clone(seg[3 : n-2])
		return p, true
	}
	return AddressParts{}, false
}

func tail(seg []string, from int) []string {
	if len(seg) <= from {
		return nil
	}
	return slices.Clone(seg[from:])
}

// CanonicalAddress is the address the compiled IR uses for the declaration:
// an instance or binding loses its placement path.
func CanonicalAddress(t TargetKind, a string) string {
	p, ok := SplitAddress(t, a)
	if !ok || (t != TargetInstance && t != TargetBinding) {
		return a
	}
	return "deployment." + p.Env + ".instance." + p.Local
}

func checkAddress(e Edit) error {
	if _, ok := SplitAddress(e.Target, e.Address); !ok {
		return fieldError("address", "%q is not a valid %s address", e.Address, e.Target)
	}
	if e.Op != OpAdd && e.Target == TargetInstance && strings.Contains(e.Address, ".node.") {
		return fieldError("address", "address an existing instance as deployment.<env>.instance.<name>")
	}
	if e.Target != TargetBinding {
		return nil
	}
	if e.Binding.Kind != "terraform" && e.Binding.Kind != "cloudformation" {
		return fieldError("binding.kind", "%q is not terraform or cloudformation", e.Binding.Kind)
	}
	if e.Binding.Index < 0 {
		return fieldError("binding.index", "must not be negative")
	}
	return nil
}
