package usecases

import (
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Provenance records where every addressable declaration was authored. It
// feeds the generated-file notice (FR-020) and positions render-stage
// diagnostics.
//
// It is built from the SourceModel rather than the IR because environments,
// groups and views carry no range in the IR, and adding one would change the
// exported shape (research R7).
type Provenance struct {
	entries []provEntry // sorted by address
	files   []string
}

type provEntry struct {
	addr arch.Address
	rng  arch.SourceRange
}

// BuildProvenance indexes every element, relationship, environment, group,
// instance and view declaration by address.
func BuildProvenance(model *arch.SourceModel) Provenance {
	var p Provenance
	if model == nil {
		return p
	}
	add := func(a arch.Address, r arch.SourceRange) { p.entries = append(p.entries, provEntry{a, r}) }
	for _, e := range model.Elements {
		addr := e.Address()
		add(addr, e.Range)
		for _, r := range e.Relations {
			add(arch.NewRelationshipAddress(addr, r.LocalName), r.Range)
		}
	}
	for _, env := range model.Environments {
		envAddr := arch.NewEnvironmentAddress(env.Name)
		add(envAddr, env.Range)
		addPlacement(add, envAddr, nil, env.Groups, env.Instances)
	}
	for _, v := range model.Views {
		add(arch.NewViewAddress(v.Name), v.Range)
	}
	// Stable sort keeps the first declaration of a duplicate address; a
	// duplicate is a compile error, so it never reaches a render anyway.
	sort.SliceStable(p.entries, func(i, j int) bool { return p.entries[i].addr < p.entries[j].addr })

	p.files = append(p.files, model.Files...)
	sort.Strings(p.files)
	return p
}

func addPlacement(add func(arch.Address, arch.SourceRange), env arch.Address, path []string,
	groups []arch.GroupDecl, instances []arch.InstanceDecl) {

	for _, inst := range instances {
		add(arch.NewInstanceAddress(env, inst.Name), inst.Range)
	}
	for _, g := range groups {
		sub := append(append([]string(nil), path...), g.Name)
		add(arch.NewGroupAddress(env, sub), g.Range)
		addPlacement(add, env, sub, g.Groups, g.Instances)
	}
}

// RangeOf returns the declaring range of addr.
func (p Provenance) RangeOf(addr arch.Address) (arch.SourceRange, bool) {
	i := sort.Search(len(p.entries), func(i int) bool { return p.entries[i].addr >= addr })
	if i < len(p.entries) && p.entries[i].addr == addr {
		return p.entries[i].rng, true
	}
	return arch.SourceRange{}, false
}

// SourcesFor returns the sorted, de-duplicated files that declared addrs.
// Unknown addresses are skipped.
func (p Provenance) SourcesFor(addrs []arch.Address) []string {
	seen := map[string]bool{}
	var out []string
	for _, a := range addrs {
		r, ok := p.RangeOf(a)
		if !ok || r.File == "" || seen[r.File] {
			continue
		}
		seen[r.File] = true
		out = append(out, r.File)
	}
	sort.Strings(out)
	return out
}

// AllFiles returns every source file of the project, sorted. Site-wide
// artifacts name all of them.
func (p Provenance) AllFiles() []string { return p.files }
