package usecases

import (
	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// projectDeployment draws one environment: its placement-group tree exactly
// as authored, however deep (edge case: deeply nested placement), each
// instance inside its group, and connections derived from the logical
// relationships between the elements the instances realise.
//
// An endpoint lifts to its nearest ancestor-or-self that is instantiated in
// this environment. A relationship whose other end has no instance here
// becomes a boundary connection (FR-009).
func projectDeployment(ir *arch.IR, envAddr arch.Address) ([]viewmodel.Node, []viewmodel.Edge, bool) {
	env, _ := ir.Environment(envAddr)
	envID := viewmodel.NodeIDFor(string(env.Address))
	nodes := []viewmodel.Node{{
		ID:      envID,
		Address: string(env.Address),
		Role:    viewmodel.RoleSubject,
		Label:   env.Name,
		Style:   viewmodel.StyleFor(viewmodel.RoleSubject, "", nil),
	}}
	nodes = appendGroups(nodes, env.Groups, envID)

	instancesOf := map[arch.Address][]string{}
	for _, inst := range env.Instances {
		parent := envID
		if inst.PlacedIn != "" {
			parent = viewmodel.NodeIDFor(string(inst.PlacedIn))
		}
		n := instanceNode(ir, inst, parent)
		nodes = append(nodes, n)
		instancesOf[inst.Of] = append(instancesOf[inst.Of], n.ID)
	}

	resolve := func(addr arch.Address) []string {
		for cur := addr; cur != ""; {
			if ids, ok := instancesOf[cur]; ok {
				return ids
			}
			e, ok := ir.Element(cur)
			if !ok {
				return nil
			}
			cur = e.Parent
		}
		return nil
	}
	edges, outside := liftEdges(ir.Relationships, resolve)
	return nodes, edges, outside
}

func appendGroups(nodes []viewmodel.Node, groups []arch.Group, parent string) []viewmodel.Node {
	for _, g := range groups {
		id := viewmodel.NodeIDFor(string(g.Address))
		nodes = append(nodes, viewmodel.Node{
			ID:      id,
			Address: string(g.Address),
			Role:    viewmodel.RoleGroup,
			Label:   g.Name,
			Parent:  parent,
			Style:   viewmodel.StyleFor(viewmodel.RoleGroup, "", nil),
		})
		nodes = appendGroups(nodes, g.Groups, id)
	}
	return nodes
}

func instanceNode(ir *arch.IR, inst arch.Instance, parent string) viewmodel.Node {
	of, _ := ir.Element(inst.Of)
	return viewmodel.Node{
		ID:          viewmodel.NodeIDFor(string(inst.Address)),
		Address:     string(inst.Address),
		Role:        viewmodel.RoleInstance,
		Kind:        string(of.Kind),
		Label:       inst.Name,
		Technology:  of.Technology,
		Description: of.Description,
		Parent:      parent,
		Style:       viewmodel.WithShape(viewmodel.StyleFor(viewmodel.RoleInstance, string(of.Kind), of.Tags), of.Shape),
		Link:        viewmodel.ElementPath(string(inst.Of), "html"),
	}
}
