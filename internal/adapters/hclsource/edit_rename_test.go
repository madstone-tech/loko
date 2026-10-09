package hclsource

import (
	"errors"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func rename(t *testing.T, from, to string) authoring.Edit {
	return mustEdit(t, authoring.Edit{Op: authoring.OpRename, Target: authoring.TargetElement, Address: from, To: to})
}

// TestRenameRewritesEveryReference: container.api is referenced as a parent
// (component.handler), a relationship target (container.web.uses.api), an
// instance's of (deploy.loko.hcl) and a view entry (views.loko.hcl).
func TestRenameRewritesEveryReference(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	p := planOne(t, root, rename(t, "container.api", "container.backend"))
	compiles(t, root, p)
	if len(p.Files) != 3 {
		t.Fatalf("rename touched %d files, want 3", len(p.Files))
	}
	for _, f := range p.Files {
		o, n := lines(f.Old), lines(f.New)
		if f.Path == "main.loko.hcl" {
			n = n[:len(o)] // the appended moved block is checked below
		}
		if len(o) != len(n) {
			t.Fatalf("%s: rename changed the line count", f.Path)
		}
		for i := range o {
			if o[i] != n[i] && strings.ReplaceAll(o[i], "container.api", "container.backend") != n[i] &&
				strings.Replace(o[i], `"api"`, `"backend"`, 1) != n[i] {
				t.Errorf("%s:%d changed more than the reference:\n- %s+ %s", f.Path, i+1, o[i], n[i])
			}
		}
	}
	main := string(p.Files[1].New)
	if !strings.HasSuffix(main, "\nmoved {\n  from = container.api\n  to   = container.backend\n}\n") {
		t.Errorf("moved block not appended canonically:\n%s", main[len(main)-120:])
	}
	if !strings.Contains(main, "container \"backend\" {\n  system      = system.shop") {
		t.Error("the declaring label was not rewritten")
	}
}

func TestRenameKindChangeInABatch(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	edits := []authoring.Edit{
		rename(t, "container.gateway", "component.gateway"),
		mustEdit(t, authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement, Address: "component.gateway",
			Set: []authoring.Attr{set("container", r("container.api"))}}),
	}
	p, err := NewEditor().Plan(t.Context(), root, edits)
	if err != nil {
		t.Fatal(err)
	}
	compiles(t, root, p)
	main := string(p.Files[1].New)
	if strings.Contains(main, "system = system.payments\n  uses") || !strings.Contains(main, "component \"gateway\" {") {
		t.Errorf("a kind change drops the parent attribute the new kind does not take:\n%s", main)
	}
}

func TestRenameRefusals(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	for _, tc := range []struct {
		from, to, reason, detail string
	}{
		{"container.api", "container.web", authoring.ReasonAddressInUse, "container.web"},
		{"container.nope", "container.x", authoring.ReasonNotFound, "container.nope"},
	} {
		_, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{rename(t, tc.from, tc.to)})
		var ee *authoring.EditError
		if !errors.As(err, &ee) || ee.Reason != tc.reason || !strings.Contains(ee.Detail, tc.detail) {
			t.Errorf("rename %s → %s: %v, want %s naming %s", tc.from, tc.to, err, tc.reason, tc.detail)
		}
	}
}

func TestRenameCollapsesChains(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	ed := NewEditor()
	p := planOne(t, root, rename(t, "external.bank", "external.acquirer"))
	commitStep(t, ed, root, p, false)
	p = planOne(t, root, rename(t, "external.acquirer", "external.psp"))
	compiles(t, root, p)
	main := string(p.Files[0].New)
	if !strings.Contains(main, "from = external.bank\n  to   = external.psp") ||
		!strings.Contains(main, "from = external.acquirer\n  to   = external.psp") {
		t.Errorf("an earlier moved.to follows the rename:\n%s", main[strings.Index(main, "moved"):])
	}
}

func TestRenameBackUndoesTheMove(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	ed := NewEditor()
	commitStep(t, ed, root, planOne(t, root, rename(t, "external.bank", "external.acquirer")), false)
	p := planOne(t, root, rename(t, "external.acquirer", "external.bank"))
	compiles(t, root, p)
	main := string(p.Files[0].New)
	if strings.Contains(main, "from = external.bank") || strings.Count(main, "moved {") != 1 ||
		!strings.Contains(main, `uses   "acquire" { target = external.bank }`) {
		t.Errorf("renaming back must drop the move it undoes:\n%s", main[strings.Index(main, "external \"bank\""):])
	}
}

func TestRemoveDropsTheElementsHistory(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	ed := NewEditor()
	commitStep(t, ed, root, planOne(t, root, rename(t, "external.bank", "external.acquirer")), false)
	commitStep(t, ed, root, planOne(t, root, rename(t, "external.acquirer", "external.psp")), false)
	edits := []authoring.Edit{
		mustEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetRelationship, Address: "container.gateway.uses.acquire"}),
		mustEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: "external.psp"}),
	}
	p, err := ed.Plan(t.Context(), root, edits)
	if err != nil {
		t.Fatal(err)
	}
	if main := string(p.Files[0].New); strings.Contains(main, "moved") {
		t.Errorf("moved blocks leading to a removed element must go with it:\n%s", main[strings.LastIndex(main, "}\n\n")-40:])
	}
	if strings.Contains(string(p.Files[0].New), "\n\n\n") && !strings.Contains(string(p.Files[0].Old), "\n\n\n") {
		t.Error("removing history left a doubled blank line")
	}
}
