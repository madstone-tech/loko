package usecases

import (
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// expansion is a batch with cascades expanded. origin maps each edit back to
// the batch edit it came from; removed lists every declaration removed.
type expansion struct {
	edits   []authoring.Edit
	origin  []int
	removed []string
}

func (x *expansion) push(e authoring.Edit, from int) {
	key := func(e authoring.Edit) string { return string(e.Target) + " " + e.Address + " " + e.Entry }
	if slices.ContainsFunc(x.edits, func(o authoring.Edit) bool { return o.Op == e.Op && key(o) == key(e) }) {
		return
	}
	x.edits, x.origin = append(x.edits, e), append(x.origin, from)
}

// expandRemovals applies FR-018a: a cascading removal also removes everything
// that depends on its target. A group or environment that still holds
// instances is refused without cascade, since removing it removes them.
func expandRemovals(ir *arch.IR, edits []authoring.Edit) (*expansion, *Refusal) {
	x := &expansion{}
	for i, e := range edits {
		x.push(e, i)
		if e.Op != authoring.OpRemove {
			continue
		}
		if e.Cascade && ir == nil {
			return nil, &Refusal{Reason: authoring.ReasonInvalidEdit, Edit: i,
				Detail: "a cascading removal needs a project that compiles; fix the errors first, or remove declarations one at a time"}
		}
		if ir == nil {
			x.removed = append(x.removed, authoring.CanonicalAddress(e.Target, e.Address))
			continue
		}
		switch e.Target {
		case authoring.TargetElement:
			x.removed = append(x.removed, cascadeElement(ir, x, e, i)...)
		case authoring.TargetGroup, authoring.TargetEnvironment:
			// Instances an earlier edit in the batch removed no longer hold it.
			held := slices.DeleteFunc(instancesWithin(ir, e.Address), func(a string) bool {
				return slices.Contains(x.removed, a)
			})
			if len(held) > 0 && !e.Cascade {
				return nil, &Refusal{Reason: authoring.ReasonDanglingReferences, Edit: i, Dependents: held,
					Detail: e.Address + " still holds instances; remove them first, or remove with cascade"}
			}
			x.removed = append(append(x.removed, e.Address), held...)
		case authoring.TargetRelationship, authoring.TargetInstance:
			x.removed = append(x.removed, authoring.CanonicalAddress(e.Target, e.Address))
		}
	}
	slices.Sort(x.removed)
	x.removed = slices.Compact(x.removed)
	return x, nil
}

// cascadeElement returns everything removing e removes, and, for a cascade,
// queues the extra edits: descendants, relationships into the subtree,
// instances of it, and its view entries. Relationships out of the subtree are
// nested in its blocks and go with them.
func cascadeElement(ir *arch.IR, x *expansion, e authoring.Edit, from int) []string {
	subtree := descendants(ir, arch.Address(e.Address))
	removed := addrStrings(subtree)
	in := func(a arch.Address) bool { return slices.Contains(subtree, a) }
	for _, d := range subtree[1:] {
		x.queue(e, from, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: string(d)})
	}
	for _, r := range ir.Relationships {
		if in(r.Source) || in(r.Target) {
			removed = append(removed, string(r.Address))
		}
		if in(r.Target) && !in(r.Source) {
			x.queue(e, from, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetRelationship, Address: string(r.Address)})
		}
	}
	for _, env := range ir.Environments {
		for _, inst := range env.Instances {
			if in(inst.Of) {
				removed = append(removed, string(inst.Address))
				x.queue(e, from, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetInstance, Address: string(inst.Address)})
			}
		}
	}
	for _, v := range ir.Views {
		for _, a := range slices.Concat(v.Include, v.Exclude) {
			if in(a) {
				x.queue(e, from, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetViewEntry, Address: string(v.Address), Entry: string(a)})
			}
		}
	}
	return removed
}

// queue adds a cascade's extra edit, only when the edit asked for cascade.
func (x *expansion) queue(parent authoring.Edit, from int, e authoring.Edit) {
	if parent.Cascade {
		x.push(e, from)
	}
}

// descendants returns root and everything it contains, root first, the rest
// sorted.
func descendants(ir *arch.IR, root arch.Address) []arch.Address {
	var out []arch.Address
	for _, c := range ir.Children(root) {
		out = append(out, descendants(ir, c.Address)...)
	}
	slices.Sort(out)
	return append([]arch.Address{root}, out...)
}

func instancesWithin(ir *arch.IR, group string) []string {
	env := strings.Join(strings.SplitN(group, ".", 3)[:2], ".")
	var out []string
	for _, e := range ir.Environments {
		for _, in := range e.Instances {
			p := string(in.PlacedIn)
			if string(e.Address) == env && (group == env || p == group || strings.HasPrefix(p, group+".")) {
				out = append(out, string(in.Address))
			}
		}
	}
	slices.Sort(out)
	return out
}

func addrStrings(as []arch.Address) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = string(a)
	}
	return out
}
