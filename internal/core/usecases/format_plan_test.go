package usecases

import (
	"context"
	"sync"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// fakeFormatter returns canned canonical files.
type fakeFormatter struct {
	files []authoring.FileContent
	diags arch.Diagnostics
}

func (f *fakeFormatter) Format(context.Context, string) ([]string, arch.Diagnostics, error) {
	return nil, nil, nil
}
func (f *fakeFormatter) Check(context.Context, string) ([]string, arch.Diagnostics, error) {
	return nil, nil, nil
}
func (f *fakeFormatter) Canonical(context.Context, string) ([]authoring.FileContent, arch.Diagnostics, error) {
	return f.files, f.diags, nil
}

func formatFixture(files []authoring.FileContent, diags arch.Diagnostics) *applyFixture {
	f := newApplyFixture(describeModel())
	f.svc.deps.Formatter = &fakeFormatter{files: files, diags: diags}
	return f
}

func TestFormatPreviewAndCommit(t *testing.T) {
	t.Parallel()
	f := formatFixture(changedPlan().Files, nil)
	prev, err := f.svc.Format(t.Context(), f.base, true)
	if err != nil || !prev.OK || !prev.Preview || len(prev.Files) != 1 || f.editor.commits.Load() != 0 {
		t.Fatalf("preview: %+v %v, commits %d", prev, err, f.editor.commits.Load())
	}
	done, err := f.svc.Format(t.Context(), f.base, false)
	if err != nil || !done.OK || done.Preview || f.editor.commits.Load() != 1 {
		t.Errorf("commit: %+v %v, commits %d", done, err, f.editor.commits.Load())
	}
}

func TestFormatNothingToDo(t *testing.T) {
	t.Parallel()
	f := formatFixture(nil, nil)
	res, err := f.svc.Format(t.Context(), f.base, false)
	if err != nil || !res.OK || !res.NoOp || len(res.Files) != 0 || f.editor.commits.Load() != 0 {
		t.Errorf("nothing to format: %+v %v", res, err)
	}
}

func TestFormatRefusals(t *testing.T) {
	t.Parallel()
	f := formatFixture(changedPlan().Files, nil)
	stale, _ := f.svc.Format(t.Context(), "r1-0000000000000000", false)
	refused(t, stale, authoring.ReasonStaleRevision, 0)

	broken := formatFixture(changedPlan().Files, arch.Diagnostics{{Severity: arch.SeverityError, Code: arch.CodeSyntaxError}})
	res, _ := broken.svc.Format(t.Context(), broken.base, false)
	refused(t, res, authoring.ReasonCompileErrors, 0)
	if len(res.Diags) == 0 || broken.editor.commits.Load() != 0 {
		t.Errorf("syntax error: diagnostics %v, commits %d", res.Diags, broken.editor.commits.Load())
	}
}

// TestFormatQueuesWithEdits is FR-017 for fmt: a format and an edit on the
// same service never reach the editor at the same time.
func TestFormatQueuesWithEdits(t *testing.T) {
	t.Parallel()
	f := formatFixture(changedPlan().Files, nil)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() { _, _ = f.svc.Format(t.Context(), f.base, false) })
		wg.Go(func() {
			_, _ = f.svc.Apply(t.Context(), ApplyRequest{Edits: []EditInput{addSystem("a")}, BaseRevision: f.base})
		})
	}
	wg.Wait()
	if f.editor.reentered.Load() {
		t.Error("fmt and apply_edit overlapped in the editor")
	}
}
