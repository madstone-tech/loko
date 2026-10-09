package usecases

import (
	"context"
	"errors"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// applyBatch runs a validated batch: expand removals, plan in memory, compile
// the planned source, then report, preview, or commit.
func (s *AuthoringService) applyBatch(ctx context.Context, b authoring.Batch) (*EditResult, error) {
	ir, _, err := compileIR(ctx, s.deps)
	if err != nil {
		return nil, err
	}
	exp, refusal := expandRemovals(ir, b.Edits)
	if refusal != nil {
		return s.refuse(ctx, refusal)
	}
	plan, err := s.deps.Editor.Plan(ctx, s.deps.Root, exp.edits)
	if r := editRefusal(err, exp.origin); r != nil {
		return s.refuse(ctx, r)
	} else if err != nil {
		return nil, err
	}
	res, err := compileOverlay(ctx, s.deps, plan.Files)
	if err != nil {
		return nil, err
	}
	if r := danglingRefusal(ir, b.Edits, res.Model); r != nil {
		return s.refuse(ctx, r)
	}
	if res.HasErrors() {
		out, err := s.refuse(ctx, &Refusal{Reason: authoring.ReasonCompileErrors,
			Detail: "the edited architecture does not compile; nothing was written"})
		if out != nil {
			out.Diags = res.Diags.SortedForOutput()
		}
		return out, err
	}
	return s.finish(ctx, b, plan, exp.removed, res.Diags.SortedForOutput())
}

// finish reports a no-op, returns a preview, or commits.
func (s *AuthoringService) finish(ctx context.Context, b authoring.Batch, plan authoring.Plan, removed []string, warnings arch.Diagnostics) (*EditResult, error) {
	out := &EditResult{OK: true, Removed: removed, Diags: warnings}
	switch {
	case !plan.Changed():
		out.NoOp = true
	case b.Preview:
		out.Preview, out.Files = true, s.deps.Editor.Diff(plan)
	default:
		out.Files = s.deps.Editor.Diff(plan)
		err := s.deps.Editor.Commit(ctx, s.deps.Root, plan, b.BaseRevision)
		if r := editRefusal(err, nil); r != nil {
			return s.refuse(ctx, r)
		} else if err != nil {
			return nil, err
		}
	}
	rev, err := currentRevision(ctx, s.deps)
	if err != nil {
		return nil, err
	}
	out.Revision = rev
	return out, nil
}

// editRefusal turns an editor's *authoring.EditError into a refusal, mapping
// an expanded edit's index back to the batch edit it came from.
func editRefusal(err error, origin []int) *Refusal {
	var ee *authoring.EditError
	if !errors.As(err, &ee) {
		return nil
	}
	idx := ee.Index
	if idx >= 0 && idx < len(origin) {
		idx = origin[idx]
	}
	return &Refusal{Reason: ee.Reason, Edit: idx, Detail: ee.Detail, Dependents: ee.Dependents}
}
