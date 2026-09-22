package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Resolved holds the outcome of resolution, keyed by the address of the thing
// that held the reference.
type Resolved struct {
	Table SymbolTable
	// Parent maps an element address to its containment parent.
	Parent map[arch.Address]arch.Address
	// Target maps a relationship address to its resolved endpoint.
	Target map[arch.Address]arch.Address
	// InstanceOf maps an instance address to the logical element it realises.
	InstanceOf map[arch.Address]arch.Address
	// ViewInclude and ViewExclude map a view address to its resolved members.
	ViewInclude map[arch.Address][]arch.Address
	ViewExclude map[arch.Address][]arch.Address
}

// logicalKinds is the set a reference to a logical element may resolve to.
var logicalKinds = []string{
	string(arch.KindPerson), string(arch.KindSystem), string(arch.KindContainer),
	string(arch.KindComponent), string(arch.KindExternal),
}

// ResolveModel builds the symbol table and resolves every reference in the
// model (FR-019 to FR-022).
//
// Resolution happens here, in core, rather than in the parser: references
// arrive as raw strings precisely so that this rule — and its diagnostics —
// can be unit-tested from a struct literal (research R2 and R7).
func ResolveModel(model *arch.SourceModel) (*Resolved, arch.Diagnostics) {
	st, diags := BuildSymbolTable(model)

	res := &Resolved{
		Table:       st,
		Parent:      map[arch.Address]arch.Address{},
		Target:      map[arch.Address]arch.Address{},
		InstanceOf:  map[arch.Address]arch.Address{},
		ViewInclude: map[arch.Address][]arch.Address{},
		ViewExclude: map[arch.Address][]arch.Address{},
	}

	for _, e := range model.Elements {
		self := e.Address()

		if !e.Parent.IsZero() {
			want := parentKinds(e.Kind)
			if addr, ok := st.resolve(e.Parent, want, &diags); ok {
				res.Parent[self] = addr
			}
		}

		for _, rel := range e.Relations {
			if rel.Target.IsZero() {
				continue
			}
			if addr, ok := st.resolve(rel.Target, logicalKinds, &diags); ok {
				res.Target[arch.NewRelationshipAddress(self, rel.LocalName)] = addr
			}
		}
	}

	for _, env := range model.Environments {
		envAddr := arch.NewEnvironmentAddress(env.Name)
		for _, placed := range env.AllInstances() {
			if placed.Instance.Of.IsZero() {
				continue
			}
			if addr, ok := st.resolve(placed.Instance.Of, logicalKinds, &diags); ok {
				res.InstanceOf[arch.NewInstanceAddress(envAddr, placed.Instance.Name)] = addr
			}
		}
	}

	// Views are resolved like anything else, so a broken reference in a view is
	// an error rather than a silently dropped member (FR-016).
	for _, v := range model.Views {
		addr := arch.NewViewAddress(v.Name)
		res.ViewInclude[addr] = st.resolveAll(v.Include, logicalKinds, &diags)
		res.ViewExclude[addr] = st.resolveAll(v.Exclude, logicalKinds, &diags)
	}

	return res, diags
}

// parentKinds returns the kinds a given element's parent may be (FR-010).
func parentKinds(kind arch.ElementKind) []string {
	switch kind {
	case arch.KindContainer:
		return []string{string(arch.KindSystem)}
	case arch.KindComponent:
		return []string{string(arch.KindContainer)}
	default:
		return nil
	}
}

func (st SymbolTable) resolveAll(refs []arch.Reference, want []string, diags *arch.Diagnostics) []arch.Address {
	var out []arch.Address
	for _, r := range refs {
		if addr, ok := st.resolve(r, want, diags); ok {
			out = append(out, addr)
		}
	}
	return out
}

// resolve looks one reference up, appending a diagnostic when it does not
// resolve or resolves to a kind this position does not accept.
func (st SymbolTable) resolve(r arch.Reference, want []string, diags *arch.Diagnostics) (arch.Address, bool) {
	addr := arch.Address(r.Raw)

	kind, ok := st.kinds[addr]
	if !ok {
		*diags = append(*diags, arch.Diagnostic{
			Severity: arch.SeverityError,
			Code:     arch.CodeUnresolvedReference,
			Summary:  "Unresolvable reference",
			Detail:   st.unresolvedDetail(r.Raw),
			Range:    r.Range,
		})
		return "", false
	}

	if len(want) > 0 && !containsStr(want, kind) {
		*diags = append(*diags, arch.Diagnostic{
			Severity: arch.SeverityError,
			Code:     arch.CodeWrongReferenceKind,
			Summary:  "Reference of the wrong kind",
			Detail: fmt.Sprintf("%s is a %s. This position requires %s.",
				addr, kind, orList(want)),
			Address: addr,
			Range:   r.Range,
			Related: []arch.RelatedRange{{Message: "declared here", Range: st.ranges[addr]}},
		})
		return "", false
	}

	return addr, true
}
