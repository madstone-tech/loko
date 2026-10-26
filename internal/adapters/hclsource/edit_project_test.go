package hclsource

import (
	"errors"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func addProject(t *testing.T, file string) authoring.Edit {
	t.Helper()
	return mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetProject, Address: "project",
		File: file, Set: []authoring.Attr{set("name", s("shop")), set("description", s("Online shop"))}})
}

func planErr(t *testing.T, root string, e authoring.Edit) *authoring.EditError {
	t.Helper()
	_, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{e})
	var ee *authoring.EditError
	if !errors.As(err, &ee) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	return ee
}

func TestAddProjectInEmptyRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	plan, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{addProject(t, "architecture/workspace.loko.hcl")})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].Path != "architecture/workspace.loko.hcl" {
		t.Fatalf("plan %+v", plan.Files)
	}
	if got, want := string(plan.Files[0].New), "project \"shop\" {\n  description = \"Online shop\"\n}\n"; got != want {
		t.Errorf("new file:\n%s\nwant:\n%s", got, want)
	}
}

func TestProjectEditRefusals(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{"main.loko.hcl": "project \"p\" {}\n"})
	if ee := planErr(t, root, addProject(t, "other.loko.hcl")); ee.Reason != authoring.ReasonProjectExists || !strings.Contains(ee.Detail, "main.loko.hcl") {
		t.Errorf("second project: %v", ee)
	}
	empty := t.TempDir()
	add := mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: "person.user"})
	if ee := planErr(t, empty, add); ee.Reason != authoring.ReasonNoProject || !strings.Contains(ee.Detail, `"target":"project"`) {
		t.Errorf("element before project: %v", ee)
	}
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{"main.loko.hcl": "project \"p\" {\n  description = \"d\"\n}\n"})
	e := mustEdit(t, authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetProject, Address: "project",
		Set: []authoring.Attr{set("ignore_warnings", l("missing_docs"))}})
	plan, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{e})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(plan.Files[0].New); !strings.Contains(got, `ignore_warnings = ["missing_docs"]`) {
		t.Errorf("updated project:\n%s", got)
	}
}
