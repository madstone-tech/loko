package usecases

import (
	"fmt"
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ValidateStructure checks the containment rules: a container belongs to a
// system, a component belongs to a container, and containment forms a
// single-parent acyclic tree (FR-028).
//
// It does NOT look at relationships. Dependency cycles between elements are a
// normal architecture and must never be reported (FR-032) — this is the one
// place the Terraform mental model actively misleads, so the containment walk
// below is deliberately a separate function from anything that touches edges.
func ValidateStructure(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	var diags arch.Diagnostics
	diags = append(diags, checkParentPresence(model, res)...)
	diags = append(diags, checkContainmentAcyclic(model, res)...)
	return diags
}

// checkParentPresence reports a container or component with no parent. A
// parent that resolves to the wrong kind is already reported during
// resolution, which has the reference's own range and can be more precise.
func checkParentPresence(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	var diags arch.Diagnostics

	for _, e := range model.Elements {
		want := parentKinds(e.Kind)
		if len(want) == 0 {
			continue
		}
		if _, resolved := res.Parent[e.Address()]; resolved {
			continue
		}
		// An unresolvable or wrong-kind parent reference was already reported;
		// reporting it again here would double-count the same mistake.
		if !e.Parent.IsZero() {
			continue
		}
		diags = append(diags, arch.Diagnostic{
			Severity: arch.SeverityError,
			Code:     arch.CodeWrongParentKind,
			Summary:  fmt.Sprintf("%s has no %s", e.Kind, want[0]),
			Detail: fmt.Sprintf("Every %s must declare the %s that contains it, for example %s = %s.name.",
				e.Kind, want[0], want[0], want[0]),
			Address: e.Address(),
			Range:   e.Range,
		})
	}

	return diags
}

// checkContainmentAcyclic walks the Parent edge only.
//
// Iterative rather than recursive so a deep or cyclic chain cannot overflow the
// stack on hostile input, and so each cycle is reported once against its
// lowest-addressed member rather than once per entry point.
func checkContainmentAcyclic(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	const (
		unvisited = 0
		onStack   = 1
		done      = 2
	)

	state := make(map[arch.Address]int, len(model.Elements))
	reported := map[arch.Address]bool{}
	var diags arch.Diagnostics

	// Sorted start order keeps the diagnostic set deterministic (FR-033).
	starts := make([]arch.Address, 0, len(model.Elements))
	for _, e := range model.Elements {
		starts = append(starts, e.Address())
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i].Compare(starts[j]) < 0 })

	ranges := map[arch.Address]arch.SourceRange{}
	for _, e := range model.Elements {
		ranges[e.Address()] = e.Range
	}

	for _, start := range starts {
		if state[start] != unvisited {
			continue
		}

		var path []arch.Address
		node := start
		for state[node] != done {
			if state[node] == onStack {
				// Found a cycle; report it once, naming the whole ring so the
				// reader can see which link to break.
				if !reported[node] {
					reported[node] = true
					diags = append(diags, cycleDiagnostic(node, path, ranges))
				}
				break
			}
			state[node] = onStack
			path = append(path, node)

			parent, ok := res.Parent[node]
			if !ok {
				break
			}
			node = parent
		}

		for _, n := range path {
			state[n] = done
		}
	}

	return diags
}

func cycleDiagnostic(entry arch.Address, path []arch.Address,
	ranges map[arch.Address]arch.SourceRange) arch.Diagnostic {

	// Trim the path to the ring itself.
	ring := path
	for i, n := range path {
		if n == entry {
			ring = path[i:]
			break
		}
	}

	detail := "Containment must form a tree: "
	for i, n := range ring {
		if i > 0 {
			detail += " is inside "
		}
		detail += string(n)
	}
	detail += " is inside " + string(entry) + "."

	var related []arch.RelatedRange
	for _, n := range ring {
		if n == entry {
			continue
		}
		related = append(related, arch.RelatedRange{
			Message: "also in the cycle",
			Range:   ranges[n],
		})
	}

	return arch.Diagnostic{
		Severity: arch.SeverityError,
		Code:     arch.CodeContainmentCycle,
		Summary:  "Containment cycle",
		Detail:   detail,
		Address:  entry,
		Range:    ranges[entry],
		Related:  related,
	}
}

// ValidateProjectVersion checks the project's loko_version constraint against
// the running build (FR-028).
func ValidateProjectVersion(model *arch.SourceModel, buildVersion string) arch.Diagnostics {
	constraint := model.Project.Version
	if constraint == "" {
		return nil
	}

	ok, err := arch.ConstraintSatisfied(constraint, buildVersion)
	if err != nil {
		return arch.Diagnostics{{
			Severity: arch.SeverityError,
			Code:     arch.CodeVersionUnsatisfied,
			Summary:  "Invalid version constraint",
			Detail:   fmt.Sprintf("%q is not a valid constraint: %v", constraint, err),
			Range:    model.Project.VersionRange,
		}}
	}
	if ok {
		return nil
	}

	return arch.Diagnostics{{
		Severity: arch.SeverityError,
		Code:     arch.CodeVersionUnsatisfied,
		Summary:  "Unsupported loko version",
		Detail: fmt.Sprintf(
			"This project requires loko %s, but this build is %s.",
			constraint, buildVersion),
		Range: model.Project.VersionRange,
	}}
}
