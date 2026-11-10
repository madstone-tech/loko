package usecases

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// usesRel is one resolved relationship: its address, declaration and target.
type usesRel struct {
	addr   arch.Address
	decl   arch.RelationDecl
	target arch.Address
}

// ValidateRedundantRelationships reports a container relationship that one of
// the container's own components repeats with exactly the same target
// (contract Q-2). The system view draws both as one edge and shows neither
// description. Information only, so it never changes an exit code.
func ValidateRedundantRelationships(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	type key struct{ owner, target arch.Address }
	containerRels := map[key]usesRel{}
	repeats := map[arch.Address][]usesRel{}
	for _, e := range model.Elements {
		if e.Kind != arch.KindContainer {
			continue
		}
		for _, r := range relsOf(e, res) {
			containerRels[key{e.Address(), r.target}] = r
		}
	}
	for _, e := range model.Elements {
		if e.Kind != arch.KindComponent {
			continue
		}
		for _, r := range relsOf(e, res) {
			c, ok := containerRels[key{res.Parent[e.Address()], r.target}]
			if !ok {
				continue
			}
			repeats[c.addr] = append(repeats[c.addr], r)
		}
	}
	var diags arch.Diagnostics
	for _, k := range containerRels {
		if comps := repeats[k.addr]; comps != nil {
			diags = append(diags, redundantDiagnostic(k, comps))
		}
	}
	slices.SortFunc(diags, func(a, b arch.Diagnostic) int { return cmp.Compare(a.Address, b.Address) })
	return diags
}

// relsOf lists an element's relationships with a resolved target.
func relsOf(e arch.ElementDecl, res *Resolved) []usesRel {
	var out []usesRel
	for _, r := range e.Relations {
		addr := arch.NewRelationshipAddress(e.Address(), r.LocalName)
		if t, ok := res.Target[addr]; ok {
			out = append(out, usesRel{addr: addr, decl: r, target: t})
		}
	}
	return out
}

func redundantDiagnostic(c usesRel, comps []usesRel) arch.Diagnostic {
	slices.SortFunc(comps, func(a, b usesRel) int { return cmp.Compare(a.addr, b.addr) })
	first := comps[0]
	detail := fmt.Sprintf("%s and %s both use %s. The system view draws them as one edge and shows "+
		"neither description. Remove the container relationship, or give it a different target.",
		c.addr, first.addr, c.target)
	ck, fk := cmp.Or(c.decl.Kind, arch.RelSync), cmp.Or(first.decl.Kind, arch.RelSync)
	if ck != fk {
		detail += fmt.Sprintf(" The kinds differ: %s and %s.", ck, fk)
	}
	related := make([]arch.RelatedRange, len(comps))
	for i, r := range comps {
		related[i] = arch.RelatedRange{Message: string(r.addr), Range: r.decl.Range}
	}
	return arch.Diagnostic{
		Severity: arch.SeverityInfo, Code: arch.CodeRedundantRelationship,
		Summary: "Relationship repeats a component relationship", Detail: detail,
		Address: c.addr, Range: c.decl.Range, Related: related,
	}
}
