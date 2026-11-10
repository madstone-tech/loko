package usecases

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func changedPlan() authoring.Plan {
	return authoring.Plan{Files: []authoring.FileContent{{Path: "a.loko.hcl", Old: []byte("x\n"), New: []byte("y\n")}}}
}

type applyFixture struct {
	svc    *AuthoringService
	editor *fakeEditor
	source *fakeOverlaySource
	base   string
}

// newApplyFixture: a project that compiles, an editor returning a change, and
// a base revision the service has already handed out.
func newApplyFixture(m *arch.SourceModel) *applyFixture {
	ed := &fakeEditor{rev: testRevision("a.loko.hcl"), plan: changedPlan()}
	src := &fakeOverlaySource{stubSource: stubSource{model: m}}
	f := &applyFixture{editor: ed, source: src,
		svc: NewAuthoringService(AuthoringDeps{Root: "root", Source: src, Editor: ed})}
	f.svc.deps.Revisions.remember(ed.rev)
	f.base = ed.rev.Token()
	return f
}

func addSystem(name string) EditInput {
	return EditInput{Op: "add", Target: "element", Address: "system." + name}
}

func (f *applyFixture) apply(t *testing.T, preview bool, edits ...EditInput) *EditResult {
	t.Helper()
	res, err := f.svc.Apply(t.Context(), ApplyRequest{Edits: edits, BaseRevision: f.base, Preview: preview})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func refused(t *testing.T, res *EditResult, reason string, index int) {
	t.Helper()
	if res.OK || res.Refusal == nil || res.Refusal.Reason != reason || res.Refusal.Edit != index {
		t.Fatalf("result %+v (refusal %+v), want refusal %s at edit %d", res, res.Refusal, reason, index)
	}
}

func TestApplyInvalidEdit(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	res := f.apply(t, false, addSystem("a"), EditInput{Op: "upsert", Target: "element", Address: "system.b"})
	refused(t, res, authoring.ReasonInvalidEdit, 1)
	if len(f.editor.planned) != 0 || f.editor.commits.Load() != 0 {
		t.Error("an invalid edit reached the editor")
	}
	bad := f.apply(t, false, EditInput{Op: "add", Target: "element", Address: "system.c", Set: map[string]any{"tags": "x"}})
	refused(t, bad, authoring.ReasonInvalidEdit, 0)
	if res.Revision != f.base {
		t.Errorf("a refusal reports the current revision, got %q", res.Revision)
	}
}

func TestApplyUnknownRevisionIsStale(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	res, err := f.svc.Apply(t.Context(), ApplyRequest{Edits: []EditInput{addSystem("a")}, BaseRevision: "r1-0000000000000000"})
	if err != nil {
		t.Fatal(err)
	}
	refused(t, res, authoring.ReasonStaleRevision, 0)
}

func TestApplyCompileErrors(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	broken := describeModel()
	broken.Elements[1].Parent = ref("system.gone", 3)
	f.source.overlayFn = func([]authoring.FileContent) (*arch.SourceModel, arch.Diagnostics) { return broken, nil }
	res := f.apply(t, false, addSystem("a"))
	refused(t, res, authoring.ReasonCompileErrors, 0)
	if !res.Diags.HasErrors() || f.editor.commits.Load() != 0 {
		t.Errorf("compile errors: diags %v commits %d", res.Diags, f.editor.commits.Load())
	}
}

func TestApplyPreviewWritesNothing(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	res := f.apply(t, true, addSystem("a"))
	if !res.OK || !res.Preview || len(res.Files) != 1 || f.editor.commits.Load() != 0 {
		t.Errorf("preview: %+v commits %d", res, f.editor.commits.Load())
	}
}

func TestApplyNoOp(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	f.editor.plan = authoring.Plan{Files: []authoring.FileContent{{Path: "a.loko.hcl", Old: []byte("x"), New: []byte("x")}}}
	res := f.apply(t, false, addSystem("a"))
	if !res.OK || !res.NoOp || len(res.Files) != 0 || f.editor.commits.Load() != 0 {
		t.Errorf("no-op (FR-021): %+v", res)
	}
}

func TestApplyCommits(t *testing.T) {
	t.Parallel()
	m := describeModel()
	m.Elements = append(m.Elements, elem(arch.KindSystem, "lonely", 30))
	f := newApplyFixture(m)
	res := f.apply(t, false, addSystem("a"), addSystem("b"), addSystem("c"))
	if !res.OK || len(res.Files) != 1 || f.editor.commits.Load() != 1 {
		t.Fatalf("commit: %+v commits %d", res, f.editor.commits.Load())
	}
	if len(f.editor.planned) != 1 || len(f.editor.planned[0]) != 3 || f.editor.planned[0][2].Address != "system.c" {
		t.Errorf("a batch is planned once, in order: %+v", f.editor.planned)
	}
	if f.source.overlays.Load() != 1 {
		t.Errorf("a batch compiles once, got %d overlay compiles", f.source.overlays.Load())
	}
	if len(res.Diags) == 0 || res.Diags.HasErrors() {
		t.Errorf("a successful write reports warnings (FR-019): %v", res.Diags)
	}
	if res.Revision == "" {
		t.Error("a write returns the new revision")
	}
}

func TestApplyMapsEditorErrors(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	f.editor.planErr = &authoring.EditError{Index: 2, Reason: authoring.ReasonNotFound, Detail: "no container.x"}
	refused(t, f.apply(t, false, addSystem("a"), addSystem("b"), addSystem("c")), authoring.ReasonNotFound, 2)

	g := newApplyFixture(describeModel())
	g.editor.commitErr = &authoring.EditError{Reason: authoring.ReasonStaleRevision, Detail: "a.loko.hcl changed"}
	refused(t, g.apply(t, false, addSystem("a")), authoring.ReasonStaleRevision, 0)

	h := newApplyFixture(describeModel())
	h.editor.planErr = errors.New("disk on fire")
	if _, err := h.svc.Apply(t.Context(), ApplyRequest{Edits: []EditInput{addSystem("a")}, BaseRevision: h.base}); err == nil {
		t.Error("an unexpected editor failure is an error, not a refusal")
	}
}

func TestApplySerialises(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			_, _ = f.svc.Apply(t.Context(), ApplyRequest{Edits: []EditInput{addSystem("a")}, BaseRevision: f.base})
		})
	}
	wg.Wait()
	if f.editor.reentered.Load() {
		t.Error("concurrent writes interleaved (FR-017)")
	}
}

// TestApplyAfterRestart: a new server has remembered no revisions, but a
// token that matches the files as they are now is still current.
func TestApplyAfterRestart(t *testing.T) {
	t.Parallel()
	ed := &fakeEditor{rev: testRevision("a.loko.hcl"), plan: changedPlan()}
	svc := NewAuthoringService(AuthoringDeps{Root: "root", Editor: ed,
		Source: &fakeOverlaySource{stubSource: stubSource{model: describeModel()}}})
	res, err := svc.Apply(t.Context(), ApplyRequest{Edits: []EditInput{addSystem("a")}, BaseRevision: ed.rev.Token()})
	if err != nil || !res.OK || ed.commits.Load() != 1 {
		t.Fatalf("a current token after a restart: %+v %v", res.Refusal, err)
	}
}

func TestApplyViewInput(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	res := f.apply(t, true, EditInput{Op: "add", Target: "view", Address: "view.flow",
		Set: map[string]any{"include": []any{"container.api", "system.shop"}, "tags": []any{"pci"}}})
	if !res.OK {
		t.Fatalf("view add: %+v", res.Refusal)
	}
	e := f.editor.planned[0][0]
	if e.Set[0].Name != "include" || e.Set[0].Value.Kind != authoring.ValueRefList || e.Set[1].Value.Kind != authoring.ValueList {
		t.Errorf("include must become a reference list, tags a string list: %+v", e.Set)
	}
}

// lonelyModel is describeModel with one more system, which has no
// description and so adds a diagnostic of its own.
func lonelyModel() *arch.SourceModel {
	m := describeModel()
	m.Elements = append(m.Elements, elem(arch.KindSystem, "lonely", 30))
	return m
}

// planned returns an overlay function that compiles before for the source on
// disk and after for the planned source.
func planned(before, after func() *arch.SourceModel) func([]authoring.FileContent) (*arch.SourceModel, arch.Diagnostics) {
	return func(o []authoring.FileContent) (*arch.SourceModel, arch.Diagnostics) {
		if len(o) == 0 {
			return before(), nil
		}
		return after(), nil
	}
}

func hasAddress(d arch.Diagnostics, addr arch.Address) bool {
	return slices.ContainsFunc(d, func(x arch.Diagnostic) bool { return x.Address == addr })
}

func (c DiagnosticCounts) total() int { return c.Errors + c.Warnings + c.Info }

// TestApplyDiagnosticDelta is contract W-2: a successful write reports what
// it added and removed, and counts the rest.
func TestApplyDiagnosticDelta(t *testing.T) {
	t.Parallel()
	add := newApplyFixture(describeModel())
	add.source.overlayFn = planned(describeModel, lonelyModel)
	res := add.apply(t, true, addSystem("lonely"))
	if !res.OK || !hasAddress(res.DiagsAdded, "system.lonely") || len(res.DiagsRemoved) != 0 {
		t.Fatalf("added: %+v removed: %+v", res.DiagsAdded, res.DiagsRemoved)
	}
	if res.DiagsUnchanged.total() == 0 || res.DiagsUnchanged.total()+len(res.DiagsAdded) != len(res.Diags) {
		t.Errorf("unchanged %+v + added %d must equal the full set %d", res.DiagsUnchanged, len(res.DiagsAdded), len(res.Diags))
	}

	rm := newApplyFixture(lonelyModel())
	rm.source.overlayFn = planned(lonelyModel, describeModel)
	res = rm.apply(t, false, addSystem("x"))
	if !res.OK || len(res.DiagsAdded) != 0 || !hasAddress(res.DiagsRemoved, "system.lonely") {
		t.Fatalf("added: %+v removed: %+v", res.DiagsAdded, res.DiagsRemoved)
	}
	if i := slices.IndexFunc(res.DiagsRemoved, func(d arch.Diagnostic) bool { return d.Address == "system.lonely" }); res.DiagsRemoved[i].Range.StartLine != 30 {
		t.Errorf("a removed diagnostic keeps its old position: %+v", res.DiagsRemoved[i].Range)
	}

	mv := newApplyFixture(describeModel())
	mv.source.overlayFn = planned(describeModel, lonelyModel)
	res, err := mv.svc.Move(t.Context(), MoveRequest{From: "system.pay", To: "system.pay2", BaseRevision: mv.base, Preview: true})
	if err != nil || !res.OK || !hasAddress(res.DiagsAdded, "system.lonely") {
		t.Fatalf("move: %+v %v", res, err)
	}
	if e := mv.editor.planned[0][0]; e.Op != authoring.OpRename || e.Address != "system.pay" || e.To != "system.pay2" {
		t.Errorf("move plans one rename: %+v", e)
	}
}

// TestApplyRefusalKeepsFullDiagnostics is contract W-3.
func TestApplyRefusalKeepsFullDiagnostics(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	broken := lonelyModel()
	broken.Elements[1].Parent = ref("system.gone", 3)
	f.source.overlayFn = planned(describeModel, func() *arch.SourceModel { return broken })
	res := f.apply(t, false, addSystem("a"))
	refused(t, res, authoring.ReasonCompileErrors, 0)
	if !res.Diags.HasErrors() || !hasAddress(res.Diags, "system.lonely") || res.DiagsAdded != nil || res.DiagsUnchanged.total() != 0 {
		t.Errorf("a refusal keeps the full set and no delta: %+v", res)
	}
	res.ShapeEdit(false, false)
	if len(res.Diagnostics) != len(res.Diags) || res.DiagnosticsAdded != nil || res.DiagnosticsUnchanged != nil {
		t.Errorf("a shaped refusal lists every diagnostic: %+v", res)
	}
}

func TestShapeEdit(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(describeModel())
	f.source.overlayFn = planned(describeModel, lonelyModel)
	res := f.apply(t, true, addSystem("lonely"))
	res.ShapeEdit(false, false)
	if res.Diagnostics != nil || len(res.DiagnosticsAdded) != len(res.DiagsAdded) || res.DiagnosticsRemoved != nil ||
		res.DiagnosticsUnchanged == nil || *res.DiagnosticsUnchanged != res.DiagsUnchanged {
		t.Errorf("changed: %+v", res)
	}
	res.ShapeEdit(true, true)
	if len(res.Diagnostics) != len(res.Diags) || res.Diagnostics[0].Range == nil ||
		res.DiagnosticsAdded != nil || res.DiagnosticsUnchanged != nil {
		t.Errorf("all: %+v", res)
	}
}
