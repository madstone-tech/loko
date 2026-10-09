package authoring

import (
	"slices"
	"strings"
)

// ValueKind is the shape of an attribute value.
type ValueKind string

// Attribute value shapes. Number and Bool appear only inside a Map.
const (
	ValueString  ValueKind = "string"
	ValueList    ValueKind = "list"
	ValueRef     ValueKind = "ref"
	ValueRefList ValueKind = "ref_list" // a list of element references, in List
	ValueMap     ValueKind = "map"
	ValueNumber  ValueKind = "number"
	ValueBool    ValueKind = "bool"
)

// AttrValue is one attribute's value. Only the field matching Kind is used.
// A Ref is written as a bare traversal (system.shop), never a quoted string.
type AttrValue struct {
	Kind ValueKind
	Str  string
	List []string
	Ref  string
	Map  []MapEntry
	Num  float64
	Bool bool
}

// MapEntry is one key of a flat map value (instance attributes, binding tags).
type MapEntry struct {
	Key   string
	Value AttrValue
}

// Attr is one attribute to set.
type Attr struct {
	Name  string
	Value AttrValue
}

// legalAttrs mirrors the language schema: the attributes each target may set,
// and the value shape each takes. Element parents are added per kind below.
var legalAttrs = map[TargetKind]map[string]ValueKind{
	TargetElement: {"description": ValueString, "docs": ValueString, "owner": ValueString,
		"tags": ValueList, "technology": ValueString, "title": ValueString},
	TargetRelationship: {"description": ValueString, "kind": ValueString, "tags": ValueList,
		"target": ValueRef, "technology": ValueString},
	TargetEnvironment: {"account": ValueString, "provider": ValueString, "region": ValueString},
	TargetGroup:       {},
	TargetInstance:    {"attributes": ValueMap, "of": ValueRef},
	TargetBinding:     {"address": ValueString, "addresses": ValueList, "tags": ValueMap},
	TargetView:        {"direction": ValueString, "exclude": ValueRefList, "include": ValueRefList, "tags": ValueList},
}

// elementParent is the parent reference attribute each element kind takes.
var elementParent = map[string]string{"container": "system", "component": "container"}

// requiredOnAdd lists attributes an add must set.
var requiredOnAdd = map[TargetKind][]string{
	TargetRelationship: {"target"},
	TargetInstance:     {"of"},
}

// LegalAttrs returns the attributes an edit of target t at address a may set,
// with their value shapes.
func LegalAttrs(t TargetKind, a string) map[string]ValueKind {
	out := map[string]ValueKind{}
	for k, v := range legalAttrs[t] {
		out[k] = v
	}
	if t == TargetElement {
		kind, _, _ := strings.Cut(a, ".")
		if parent, ok := elementParent[kind]; ok {
			out[parent] = ValueRef
		}
		if slices.Contains(shapeKinds, kind) {
			out["shape"] = ValueString
		}
	}
	return out
}

func checkAttrs(e Edit) error {
	legal := LegalAttrs(e.Target, e.Address)
	seen := map[string]bool{}
	for _, a := range e.Set {
		want, ok := legal[a.Name]
		if !ok {
			return fieldError("set."+a.Name, "not an attribute of %s; legal: %s", e.Target, names(legal))
		}
		if seen[a.Name] {
			return fieldError("set."+a.Name, "set twice")
		}
		seen[a.Name] = true
		if err := checkValue("set."+a.Name, want, a.Value); err != nil {
			return err
		}
		if err := checkEnum(e.Target, a); err != nil {
			return err
		}
	}
	for _, c := range e.Clear {
		if _, ok := legal[c]; !ok {
			return fieldError("clear."+c, "not an attribute of %s; legal: %s", e.Target, names(legal))
		}
	}
	return checkRequired(e, seen)
}

func checkRequired(e Edit, seen map[string]bool) error {
	if e.Op != OpAdd {
		return nil
	}
	required := requiredOnAdd[e.Target]
	if e.Target == TargetElement {
		kind, _, _ := strings.Cut(e.Address, ".")
		if parent, ok := elementParent[kind]; ok {
			required = []string{parent}
		}
	}
	for _, r := range required {
		if !seen[r] {
			return fieldError("set."+r, "required when adding a %s", e.Target)
		}
	}
	if e.Target == TargetBinding {
		n := 0
		for _, s := range []string{"address", "addresses", "tags"} {
			if seen[s] {
				n++
			}
		}
		if n != 1 {
			return fieldError("set", "a binding sets exactly one of address, addresses, tags")
		}
	}
	return nil
}

func checkValue(field string, want ValueKind, v AttrValue) error {
	if v.Kind != want {
		return fieldError(field, "must be a %s, got %s", want, v.Kind)
	}
	switch want {
	case ValueRef:
		if !isElementAddress(v.Ref) {
			return fieldError(field, "%q is not an element address such as system.shop", v.Ref)
		}
	case ValueRefList:
		for _, r := range v.List {
			if !isElementAddress(r) {
				return fieldError(field, "%q is not an element address such as system.shop", r)
			}
		}
	case ValueMap:
		for _, m := range v.Map {
			if !ValidName(m.Key) {
				return fieldError(field+"."+m.Key, "is not a valid key")
			}
			if !slices.Contains([]ValueKind{ValueString, ValueNumber, ValueBool}, m.Value.Kind) {
				return fieldError(field+"."+m.Key, "must be a string, number or bool")
			}
		}
	}
	return nil
}

func names(m map[string]ValueKind) string {
	if len(m) == 0 {
		return "(none)"
	}
	return strings.Join(slices.Sorted(func(yield func(string) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}), ", ")
}
