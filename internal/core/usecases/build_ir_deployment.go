package usecases

import (
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// buildEnvironments produces the sorted environment slice.
//
// Instances are FLATTENED onto Environment.Instances regardless of nesting
// depth, and placement is recorded as a cross-reference through
// Group.Contains and Instance.PlacedIn.
//
// That shape is what makes the clarified addressing rule expressible: an
// instance is reachable without walking the node tree, so moving it between
// groups changes only PlacedIn and leaves its address — its identity — intact
// (FR-012, FR-024). The nesting is still preserved in full for the renderer
// and the reconciler.
func buildEnvironments(model *arch.SourceModel, res *Resolved) []arch.Environment {
	envs := make([]arch.Environment, 0, len(model.Environments))

	for _, d := range model.Environments {
		envAddr := arch.NewEnvironmentAddress(d.Name)
		instances, placement := buildInstances(d, envAddr, res)

		envs = append(envs, arch.Environment{
			Address:   envAddr,
			Name:      d.Name,
			Provider:  d.Provider,
			Account:   d.Account,
			Region:    d.Region,
			Groups:    buildGroups(envAddr, nil, d.Groups, placement),
			Instances: instances,
		})
	}

	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Address.Compare(envs[j].Address) < 0
	})
	return envs
}

// buildInstances flattens every instance in the environment and returns them
// sorted, plus a map from group address to the instances placed in it.
func buildInstances(d arch.EnvironmentDecl, envAddr arch.Address,
	res *Resolved) ([]arch.Instance, map[arch.Address][]arch.Address) {

	placement := map[arch.Address][]arch.Address{}
	instances := make([]arch.Instance, 0)

	for _, placed := range d.AllInstances() {
		addr := arch.NewInstanceAddress(envAddr, placed.Instance.Name)

		var groupAddr arch.Address
		if len(placed.Path) > 0 {
			groupAddr = arch.NewGroupAddress(envAddr, placed.Path)
			placement[groupAddr] = append(placement[groupAddr], addr)
		}

		instances = append(instances, arch.Instance{
			Address:    addr,
			Name:       placed.Instance.Name,
			Of:         res.InstanceOf[addr],
			PlacedIn:   groupAddr,
			Attributes: buildAttributes(placed.Instance.Attributes),
			Claims:     buildClaims(placed.Instance.Claims),
		})
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Address.Compare(instances[j].Address) < 0
	})
	return instances, placement
}

// buildGroups rebuilds the placement tree with addresses assigned.
func buildGroups(envAddr arch.Address, prefix []string, decls []arch.GroupDecl,
	placement map[arch.Address][]arch.Address) []arch.Group {

	if len(decls) == 0 {
		return nil
	}

	groups := make([]arch.Group, 0, len(decls))
	for _, g := range decls {
		path := append(append([]string{}, prefix...), g.Name)
		addr := arch.NewGroupAddress(envAddr, path)

		groups = append(groups, arch.Group{
			Address:  addr,
			Name:     g.Name,
			Groups:   buildGroups(envAddr, path, g.Groups, placement),
			Contains: sortedAddresses(placement[addr]),
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Address.Compare(groups[j].Address) < 0
	})
	return groups
}

// buildAttributes returns instance attributes sorted by key.
func buildAttributes(in []arch.KeyValue) []arch.Attribute {
	if len(in) == 0 {
		return nil
	}
	out := make([]arch.Attribute, 0, len(in))
	for _, kv := range in {
		// KeyValue and Attribute are the same shape either side of the
		// SourceModel/IR boundary; the conversion keeps them separate types so
		// one can change without dragging the other with it.
		out = append(out, arch.Attribute(kv))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// buildClaims returns claims sorted by kind, then by address.
func buildClaims(in []arch.ClaimDecl) []arch.Claim {
	if len(in) == 0 {
		return nil
	}
	out := make([]arch.Claim, 0, len(in))
	for _, c := range in {
		out = append(out, arch.Claim{
			Kind:      c.Kind,
			Address:   c.Address,
			Addresses: sortedUnique(c.Addresses),
			Tags:      buildAttributes(c.Tags),
			Range:     c.Range,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Address < out[j].Address
	})
	return out
}
