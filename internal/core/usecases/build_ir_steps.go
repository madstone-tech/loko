package usecases

import "github.com/madstone-tech/loko/internal/core/entities/arch"

// buildSteps compiles a dynamic view's steps in declaration order, numbered
// from 1: order is a flow's meaning, so steps are not sorted by address
// (feature 019, R5).
func buildSteps(view arch.Address, decls []arch.StepDecl, res *Resolved) []arch.Step {
	if len(decls) == 0 {
		return nil
	}
	steps := make([]arch.Step, 0, len(decls))
	for i, d := range decls {
		addr := arch.NewStepAddress(view, d.Name)
		steps = append(steps, arch.Step{
			Address:      addr,
			Name:         d.Name,
			Number:       i + 1,
			From:         res.StepFrom[addr],
			To:           res.StepTo[addr],
			Via:          res.StepVia[addr],
			Relationship: res.StepRelationship[addr],
			Reply:        d.Reply,
			Description:  d.Description,
			Range:        d.Range,
		})
	}
	return steps
}
