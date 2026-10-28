package usecases

import (
	"cmp"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// FlowHit is one step an element takes part in (feature 019): the dynamic
// view, the step's number and label, and whether the element is its from or
// its to.
type FlowHit struct {
	View   string `json:"view" toon:"view"`
	Number int    `json:"number" toon:"number"`
	Step   string `json:"step" toon:"step"`
	Role   string `json:"role" toon:"role"`
}

// flows lists the steps addr takes part in, directly or through an element it
// contains, sorted by view then number. A step from and to the element is
// listed once, as from.
func flows(ir *arch.IR, addr arch.Address) []FlowHit {
	var out []FlowHit
	for _, v := range ir.Views {
		for _, s := range v.Steps {
			role := ""
			switch {
			case within(ir, s.From, addr):
				role = "from"
			case within(ir, s.To, addr):
				role = "to"
			default:
				continue
			}
			out = append(out, FlowHit{View: string(v.Address), Number: s.Number, Step: s.Name, Role: role})
		}
	}
	slices.SortFunc(out, func(a, b FlowHit) int { return cmp.Or(cmp.Compare(a.View, b.View), cmp.Compare(a.Number, b.Number)) })
	return out
}

// stepInfos describes a dynamic view's steps, in order; nil for a view
// without steps.
func stepInfos(steps []arch.Step) []StepInfo {
	if len(steps) == 0 {
		return nil
	}
	out := make([]StepInfo, len(steps))
	for i, s := range steps {
		out[i] = StepInfo{Number: s.Number, Name: s.Name, From: string(s.From), To: string(s.To), Via: string(s.Via),
			Relationship: string(s.Relationship), Reply: s.Reply, Description: s.Description}
	}
	return out
}
