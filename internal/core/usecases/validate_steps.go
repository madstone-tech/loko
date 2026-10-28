package usecases

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// stepRel is a relationship as a step sees it: drawn from From to To. A
// trigger is declared on the invoked element but drawn from its trigger, so a
// flow is written queue → function (016 R4).
type stepRel struct {
	addr     arch.Address
	from, to arch.Address
}

// ValidateSteps checks dynamic views (feature 019, research R2, R3): a view
// with steps takes no selection, a style needs steps, and each step must
// match a declared relationship. The relationship each step uses is recorded
// in res.StepRelationship for the IR.
func ValidateSteps(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	var diags arch.Diagnostics
	rels := stepRelationships(model, res)
	for _, v := range model.Views {
		diags = append(diags, selectionConflicts(v)...)
		view := arch.NewViewAddress(v.Name)
		for _, s := range v.Steps {
			if d, ok := matchStep(res, rels, arch.NewStepAddress(view, s.Name), s); ok {
				diags = append(diags, d)
			}
		}
	}
	return diags
}

// stepRelationships lists every resolved relationship in drawn direction,
// sorted by address so the first match is deterministic.
func stepRelationships(model *arch.SourceModel, res *Resolved) []stepRel {
	var out []stepRel
	for _, e := range model.Elements {
		for _, r := range e.Relations {
			addr := arch.NewRelationshipAddress(e.Address(), r.LocalName)
			target, ok := res.Target[addr]
			if !ok {
				continue
			}
			sr := stepRel{addr: addr, from: e.Address(), to: target}
			if r.Kind == arch.RelTrigger {
				sr.from, sr.to = sr.to, sr.from
			}
			out = append(out, sr)
		}
	}
	slices.SortFunc(out, func(a, b stepRel) int { return cmp.Compare(a.addr, b.addr) })
	return out
}

func selectionConflicts(v arch.ViewDecl) arch.Diagnostics {
	view := arch.NewViewAddress(v.Name)
	conflict := func(attr string, r arch.SourceRange, detail string) arch.Diagnostic {
		return arch.Diagnostic{Severity: arch.SeverityError, Code: arch.CodeDynamicViewSelection,
			Summary: "Dynamic view setting conflicts", Address: view, Range: r,
			Detail: fmt.Sprintf("%s %s %s", view, detail, attr)}
	}
	if len(v.Steps) == 0 {
		if v.Style != "" {
			return arch.Diagnostics{conflict("style", v.StyleRange,
				"has no steps, so there is no flow to draw with")}
		}
		return nil
	}
	var diags arch.Diagnostics
	for _, a := range []struct {
		name string
		set  bool
		r    arch.SourceRange
	}{{"include", len(v.Include) > 0, v.IncludeRange}, {"exclude", len(v.Exclude) > 0, v.ExcludeRange},
		{"tags", len(v.Tags) > 0, v.TagsRange}} {
		if a.set {
			diags = append(diags, conflict(a.name, a.r,
				"has steps, so its elements are the steps' elements; remove"))
		}
	}
	return diags
}

// matchStep finds the relationship a step uses and reports a step that
// names one that does not connect its elements, or matches none.
func matchStep(res *Resolved, rels []stepRel, addr arch.Address, s arch.StepDecl) (arch.Diagnostic, bool) {
	from, okF := res.StepFrom[addr]
	to, okT := res.StepTo[addr]
	if !okF || !okT {
		return arch.Diagnostic{}, false // already reported as unresolved
	}
	if s.Reply {
		from, to = to, from
	}
	fits := func(r stepRel) bool { return related(res, from, r.from) && related(res, to, r.to) }
	if via, named := res.StepVia[addr]; named {
		i := slices.IndexFunc(rels, func(r stepRel) bool { return r.addr == via })
		if i >= 0 && fits(rels[i]) {
			res.StepRelationship[addr] = via
			return arch.Diagnostic{}, false
		}
		var ends string
		if i >= 0 {
			ends = fmt.Sprintf(" connects %s to %s, but", rels[i].from, rels[i].to)
		}
		return arch.Diagnostic{Severity: arch.SeverityError, Code: arch.CodeStepMismatch,
			Summary: "Step does not match its relationship", Address: addr, Range: s.Range,
			Detail: fmt.Sprintf("%s%s the step goes from %s to %s. Point via at a relationship between "+
				"them, or fix from and to.", via, ends, s.From.Raw, s.To.Raw)}, true
	}
	if i := slices.IndexFunc(rels, fits); i >= 0 {
		res.StepRelationship[addr] = rels[i].addr
		return arch.Diagnostic{}, false
	}
	return arch.Diagnostic{Severity: arch.SeverityWarning, Code: arch.CodeStepUnmatched,
		Summary: "Step matches no relationship", Address: addr, Range: s.Range,
		Detail: fmt.Sprintf("No relationship connects %s to %s, or their parents or children, in that "+
			"direction. The step is drawn; declare the relationship, or set reply = true for a reply.",
			s.From.Raw, s.To.Raw)}, true
}

// related reports whether a and b are the same element or one contains the
// other.
func related(res *Resolved, a, b arch.Address) bool {
	return encloses(res, a, b) || encloses(res, b, a)
}

func encloses(res *Resolved, outer, inner arch.Address) bool {
	seen := map[arch.Address]bool{}
	for cur := inner; cur != "" && !seen[cur]; cur = res.Parent[cur] {
		if cur == outer {
			return true
		}
		seen[cur] = true
	}
	return false
}
