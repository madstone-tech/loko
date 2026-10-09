package authoring

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

// Op is what an edit does to its target.
type Op string

// The four edit operations.
const (
	OpAdd    Op = "add"
	OpUpdate Op = "update"
	OpRemove Op = "remove"
	OpRename Op = "rename"
)

// TargetKind is the kind of declaration an edit addresses.
type TargetKind string

// The six editable declaration kinds.
const (
	TargetElement      TargetKind = "element"
	TargetRelationship TargetKind = "relationship"
	TargetEnvironment  TargetKind = "environment"
	TargetGroup        TargetKind = "group"
	TargetInstance     TargetKind = "instance"
	TargetBinding      TargetKind = "binding"
	TargetView         TargetKind = "view"

	// TargetViewEntry is internal: a cascading removal drops a removed
	// element from a view's include and exclude lists. NewEdit never accepts
	// it, so no caller can request one directly.
	TargetViewEntry TargetKind = "view_entry"
)

// BindingRef picks one binding block inside an instance: its kind, and its
// position among bindings of that kind (needed only when there is more than one).
type BindingRef struct {
	Kind  string
	Index int
}

// Edit is one requested change to one declaration. Construct it through
// NewEdit, which validates it; a zero or hand-built Edit is not trusted.
type Edit struct {
	Op      Op
	Target  TargetKind
	Address string
	Binding BindingRef
	Set     []Attr
	Clear   []string
	Cascade bool
	To      string
	File    string
	// Entry is the element address a view_entry edit removes from the view.
	Entry string
}

// FieldError is a validation failure that names the offending field, so a
// refusal can tell an assistant exactly what to fix.
type FieldError struct {
	Field string
	Msg   string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Msg }

func fieldError(field, format string, args ...any) error {
	return &FieldError{Field: field, Msg: fmt.Sprintf(format, args...)}
}

// NewEdit validates e and returns it with Set and Clear sorted by name.
func NewEdit(e Edit) (Edit, error) {
	checks := []func(Edit) error{checkOp, checkOptions, checkAddress, checkFile, checkAttrs}
	for _, check := range checks {
		if err := check(e); err != nil {
			return Edit{}, err
		}
	}
	e.Set = slices.Clone(e.Set)
	slices.SortFunc(e.Set, func(a, b Attr) int { return strings.Compare(a.Name, b.Name) })
	e.Clear = slices.Sorted(slices.Values(e.Clear))
	return e, nil
}

func checkOp(e Edit) error {
	switch e.Op {
	case OpAdd, OpUpdate, OpRemove:
	case OpRename:
		if e.Target != TargetElement {
			return fieldError("op", "rename applies only to elements, not %s", e.Target)
		}
	default:
		return fieldError("op", "%q is not one of add, update, remove, rename", e.Op)
	}
	if _, ok := legalAttrs[e.Target]; !ok {
		return fieldError("target", "%q is not one of element, relationship, environment, group, instance, binding, view", e.Target)
	}
	return nil
}

// checkOptions enforces which fields belong to which operation.
func checkOptions(e Edit) error {
	switch {
	case e.Cascade && e.Op != OpRemove:
		return fieldError("cascade", "only a remove can cascade")
	case e.To != "" && e.Op != OpRename:
		return fieldError("to", "only a rename takes a new address")
	case e.File != "" && e.Op != OpAdd:
		return fieldError("file", "only an add can choose a file")
	case len(e.Clear) > 0 && e.Op != OpUpdate:
		return fieldError("clear", "only an update can clear attributes")
	case len(e.Set) > 0 && e.Op != OpAdd && e.Op != OpUpdate:
		return fieldError("set", "only an add or update can set attributes")
	case e.Op == OpUpdate && len(e.Set) == 0 && len(e.Clear) == 0:
		return fieldError("set", "an update needs at least one attribute to set or clear")
	}
	if e.Op != OpRename {
		return nil
	}
	if e.To == e.Address {
		return fieldError("to", "%q is the current address", e.To)
	}
	if !isElementAddress(e.To) {
		return fieldError("to", "%q is not an element address such as component.api", e.To)
	}
	return nil
}

// checkFile accepts only a clean, relative path to a source file inside the
// project. Anything else could write outside the architecture (FR-020, FR-026).
func checkFile(e Edit) error {
	if e.File == "" {
		return nil
	}
	f := e.File
	switch {
	case !strings.HasSuffix(f, ".loko.hcl"):
		return fieldError("file", "%q must end in .loko.hcl", f)
	case path.IsAbs(f) || strings.HasPrefix(f, `\`) || strings.Contains(f, ":"):
		return fieldError("file", "%q must be relative to the project root", f)
	case path.Clean(f) != f || f == ".." || strings.HasPrefix(f, "../") || strings.Contains(f, `\`):
		return fieldError("file", "%q must be a clean path inside the project", f)
	}
	return nil
}

// ValidName reports whether s is a legal address segment: a letter or
// underscore, followed by letters, digits, underscores, or hyphens.
//
// It deliberately duplicates arch.ValidName, because entity packages may not
// import one another; naming_parity_test in usecases keeps the two in step.
func ValidName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
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
