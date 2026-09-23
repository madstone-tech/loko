package usecases

import (
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// SymbolTable records every declared address and what kind of thing it names.
//
// It is built in one pass over the SourceModel before anything is resolved.
// That is the whole of FR-020's two-pass requirement: with the table complete,
// resolution is a map lookup, so declaration order and file order cannot
// affect the outcome.
type SymbolTable struct {
	// kinds maps address to a symbol kind: an element kind, or the pseudo
	// kinds below for things that are not logical elements.
	kinds map[arch.Address]string
	// ranges records where each address was declared, for related-location
	// reporting.
	ranges map[arch.Address]arch.SourceRange
}

// Pseudo kinds for addressable things that are not logical elements.
const (
	symEnvironment = "deployment"
	symInstance    = "instance"
	symView        = "view"
	symGroup       = "node"
)

// Has reports whether addr is declared.
func (st SymbolTable) Has(addr arch.Address) bool {
	_, ok := st.kinds[addr]
	return ok
}

// Kind returns the symbol kind at addr.
func (st SymbolTable) Kind(addr arch.Address) (string, bool) {
	k, ok := st.kinds[addr]
	return k, ok
}

// Range returns where addr was declared.
func (st SymbolTable) Range(addr arch.Address) arch.SourceRange { return st.ranges[addr] }

// BuildSymbolTable collects every declared address, reporting a duplicate as
// an error that names both sites (FR-004).
func BuildSymbolTable(model *arch.SourceModel) (SymbolTable, arch.Diagnostics) {
	st := SymbolTable{
		kinds:  map[arch.Address]string{},
		ranges: map[arch.Address]arch.SourceRange{},
	}
	var diags arch.Diagnostics

	declare := func(addr arch.Address, kind, noun string, r arch.SourceRange) {
		if prev, exists := st.ranges[addr]; exists {
			// An instance collision gets its own code and its own explanation.
			// It is the one duplicate a reader is likely to find surprising,
			// because the two declarations can sit in different node blocks and
			// still collide — instance addresses omit the placement path by
			// design, so names are unique per environment (FR-012a).
			code, summary := arch.CodeDuplicateDeclaration, "Duplicate declaration"
			detail := fmt.Sprintf("%s %s is already declared.", noun, addr)
			if kind == symInstance {
				code, summary = arch.CodeDuplicateInstanceName, "Duplicate instance name"
				detail = fmt.Sprintf(
					"An instance named %q is already declared in this environment. "+
						"Instance names are unique per deployment, not per node: an instance's "+
						"address omits its placement path so that moving it between nodes "+
						"preserves its identity.", instanceName(addr))
			}
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityError,
				Code:     code,
				Summary:  summary,
				Detail:   detail,
				Address:  addr,
				Range:    r,
				Related:  []arch.RelatedRange{{Message: "first declared here", Range: prev}},
			})
			return
		}
		st.kinds[addr] = kind
		st.ranges[addr] = r
	}

	for _, e := range model.Elements {
		declare(e.Address(), string(e.Kind), "Element", e.Range)
		for _, rel := range e.Relations {
			declare(arch.NewRelationshipAddress(e.Address(), rel.LocalName),
				"uses", "Relationship", rel.Range)
		}
	}

	for _, env := range model.Environments {
		envAddr := arch.NewEnvironmentAddress(env.Name)
		declare(envAddr, symEnvironment, "Environment", env.Range)
		declareGroups(&st, &diags, declare, envAddr, env.Groups)
		// Instances are declared against the environment regardless of how
		// deeply they are nested, which is what makes a name collision across
		// two different nodes an error (FR-012a).
		for _, placed := range env.AllInstances() {
			declare(arch.NewInstanceAddress(envAddr, placed.Instance.Name),
				symInstance, "Instance", placed.Instance.Range)
		}
	}

	for _, v := range model.Views {
		declare(arch.NewViewAddress(v.Name), symView, "View", v.Range)
	}

	return st, diags
}

func declareGroups(st *SymbolTable, diags *arch.Diagnostics,
	declare func(arch.Address, string, string, arch.SourceRange), env arch.Address, groups []arch.GroupDecl) {

	var walk func(prefix []string, gs []arch.GroupDecl)
	walk = func(prefix []string, gs []arch.GroupDecl) {
		for _, g := range gs {
			path := append(append([]string{}, prefix...), g.Name)
			declare(arch.NewGroupAddress(env, path), symGroup, "Placement group", g.Range)
			walk(path, g.Groups)
		}
	}
	walk(nil, groups)
}

// instanceName returns the trailing segment of an instance address.
func instanceName(addr arch.Address) string {
	s := string(addr)
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		return s[i+1:]
	}
	return s
}
