package usecases

import (
	"fmt"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// danglingRefusal applies FR-018 to the planned source: a removal is refused
// while anything still refers to the removed element or to anything inside
// it. Checking after planning means a batch that also removes or re-points
// every referrer is not refused.
func danglingRefusal(ir *arch.IR, edits []authoring.Edit, planned *arch.SourceModel) *Refusal {
	if planned == nil {
		return nil
	}
	for i, e := range edits {
		if e.Op != authoring.OpRemove || e.Target != authoring.TargetElement || e.Cascade {
			continue
		}
		gone := []arch.Address{arch.Address(e.Address)}
		if ir != nil {
			gone = descendants(ir, arch.Address(e.Address))
		}
		if deps := referrers(planned, gone); len(deps) > 0 {
			return &Refusal{Reason: authoring.ReasonDanglingReferences, Edit: i, Dependents: deps,
				Detail: fmt.Sprintf("%s is still referred to by %s; remove or change those first, or remove with cascade",
					e.Address, strings.Join(deps, ", "))}
		}
	}
	return nil
}

// referrers lists, sorted, every declaration whose references name one of gone.
func referrers(m *arch.SourceModel, gone []arch.Address) []string {
	hit := func(r arch.Reference) bool { return slices.Contains(gone, arch.Address(r.Raw)) }
	var out []string
	for _, e := range m.Elements {
		if hit(e.Parent) {
			out = append(out, string(e.Address()))
		}
		for _, rd := range e.Relations {
			if hit(rd.Target) {
				out = append(out, string(arch.NewRelationshipAddress(e.Address(), rd.LocalName)))
			}
		}
	}
	for _, env := range m.Environments {
		forEachInstance(env.Instances, env.Groups, func(in arch.InstanceDecl) {
			if hit(in.Of) {
				out = append(out, "deployment."+env.Name+".instance."+in.Name)
			}
		})
	}
	for _, v := range m.Views {
		if slices.ContainsFunc(slices.Concat(v.Include, v.Exclude), hit) {
			out = append(out, string(arch.NewViewAddress(v.Name)))
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

func forEachInstance(direct []arch.InstanceDecl, groups []arch.GroupDecl, fn func(arch.InstanceDecl)) {
	for _, in := range direct {
		fn(in)
	}
	for _, g := range groups {
		forEachInstance(g.Instances, g.Groups, fn)
	}
}
