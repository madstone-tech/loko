package usecases

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// Format canonicalises the project's source as `loko fmt` does, but as a
// plan: previewed as diffs, or written through the editor with the same
// revision check as any edit, so it cannot overwrite a change another client
// just made (feature 017, FR-019). It queues with edits on the service.
func (s *AuthoringService) Format(ctx context.Context, base string, preview bool) (*EditResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rev, ok := s.baseRevision(ctx, base)
	if !ok {
		return s.refuse(ctx, &Refusal{Reason: authoring.ReasonStaleRevision,
			Detail: "base_revision is unknown or expired; read again (describe or validate) and retry"})
	}
	files, diags, err := s.deps.Formatter.Canonical(ctx, s.deps.Root)
	if err != nil {
		return nil, err
	}
	if diags.HasErrors() {
		out, err := s.refuse(ctx, &Refusal{Reason: authoring.ReasonCompileErrors,
			Detail: "a source file does not parse; nothing was formatted"})
		if out != nil {
			out.Diags = diags.SortedForOutput()
		}
		return out, err
	}
	return s.finish(ctx, preview, rev, authoring.Plan{Files: files}, nil, nil)
}
