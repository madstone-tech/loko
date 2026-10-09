package usecases

import (
	"context"
	"sync/atomic"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// fakeOverlaySource is a concrete OverlaySource. It returns a canned model for
// Load, and for LoadOverlay the model produced by overlayFn from the overlay
// it was given, so a test can make a planned edit compile or fail.
type fakeOverlaySource struct {
	stubSource
	overlayFn func([]authoring.FileContent) (*arch.SourceModel, arch.Diagnostics)
	overlays  atomic.Int32
}

func (f *fakeOverlaySource) LoadOverlay(_ context.Context, _ string, overlay []authoring.FileContent) (*arch.SourceModel, arch.Diagnostics, error) {
	if len(overlay) > 0 {
		f.overlays.Add(1)
	}
	if f.overlayFn == nil {
		return f.model, f.diags, f.err
	}
	m, d := f.overlayFn(overlay)
	return m, d, nil
}

// fakeEditor is a concrete SourceEditor that records calls, returns canned
// plans and errors, and detects reentrant use: two calls overlapping means the
// service failed to serialise writes (FR-017).
type fakeEditor struct {
	rev       authoring.Revision
	plan      authoring.Plan
	planErr   error
	commitErr error

	planned   [][]authoring.Edit
	commits   atomic.Int32
	active    atomic.Int32
	reentered atomic.Bool
}

func (f *fakeEditor) enter() func() {
	if f.active.Add(1) > 1 {
		f.reentered.Store(true)
	}
	return func() { f.active.Add(-1) }
}

func (f *fakeEditor) Plan(_ context.Context, _ string, edits []authoring.Edit) (authoring.Plan, error) {
	defer f.enter()()
	f.planned = append(f.planned, edits)
	return f.plan, f.planErr
}

func (f *fakeEditor) Diff(p authoring.Plan) []authoring.FileChange {
	out := make([]authoring.FileChange, 0, len(p.Files))
	for _, file := range p.Files {
		out = append(out, authoring.FileChange{Path: file.Path, Diff: "--- a/" + file.Path + "\n"})
	}
	return out
}

func (f *fakeEditor) Revision(context.Context, string) (authoring.Revision, error) {
	return f.rev, nil
}

func (f *fakeEditor) Commit(context.Context, string, authoring.Plan, authoring.Revision) error {
	defer f.enter()()
	if f.commitErr != nil {
		return f.commitErr
	}
	f.commits.Add(1)
	return nil
}

func testRevision(paths ...string) authoring.Revision {
	files := make([]authoring.FileHash, 0, len(paths))
	for _, p := range paths {
		files = append(files, authoring.FileHash{Path: p, SHA256: "0000000000000000000000000000000000000000000000000000000000000000"})
	}
	return authoring.NewRevision(files)
}
