package usecases

import (
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// dependentsModel: container api (in system s) holds component h; web uses
// api; h uses web; prod places instance api inside node vpc; view v
// includes api.
func dependentsModel() *arch.SourceModel {
	child := func(kind arch.ElementKind, name, parent string, line int) arch.ElementDecl {
		e := elem(kind, name, line)
		e.Parent = ref(parent, line)
		return e
	}
	web := child(arch.KindContainer, "web", "system.s", 3)
	web.Relations = []arch.RelationDecl{{LocalName: "calls", Target: ref("container.api", 4), Range: at(4)}}
	h := child(arch.KindComponent, "h", "container.api", 5)
	h.Relations = []arch.RelationDecl{{LocalName: "out", Target: ref("container.web", 6), Range: at(6)}}
	return &arch.SourceModel{
		Project:      arch.ProjectDecl{Name: "d", Declared: true, Range: at(1)},
		Files:        []string{"a.loko.hcl"},
		Elements:     []arch.ElementDecl{elem(arch.KindSystem, "s", 2), child(arch.KindContainer, "api", "system.s", 2), web, h},
		Environments: []arch.EnvironmentDecl{envWithInstance("prod", []string{"vpc"}, "api", "container.api")},
		Views:        []arch.ViewDecl{{Name: "v", Include: []arch.Reference{ref("container.api", 9)}, Range: at(9)}},
	}
}

// without returns m minus the named elements: what the planned source would
// compile to after the removal, with every referrer still in place.
func without(m *arch.SourceModel, names ...string) *arch.SourceModel {
	out := *m
	out.Elements = slices.DeleteFunc(slices.Clone(m.Elements), func(e arch.ElementDecl) bool { return slices.Contains(names, e.Name) })
	return &out
}

func TestRemoveRefusesDangling(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(dependentsModel())
	f.source.overlayFn = func([]authoring.FileContent) (*arch.SourceModel, arch.Diagnostics) {
		return without(dependentsModel(), "api"), nil
	}
	res := f.apply(t, false, EditInput{Op: "remove", Target: "element", Address: "container.api"})
	refused(t, res, authoring.ReasonDanglingReferences, 0)
	want := "component.h container.web.uses.calls deployment.prod.instance.api view.v"
	if got := strings.Join(res.Refusal.Dependents, " "); got != want {
		t.Errorf("dependents = %q, want %q", got, want)
	}
	if f.editor.commits.Load() != 0 {
		t.Error("a dangling removal was committed")
	}
}

func TestRemoveWithinBatchIsNotDangling(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(dependentsModel())
	// The planned source no longer refers to api: the batch removed or
	// re-pointed every referrer.
	clean := without(dependentsModel(), "api", "h")
	clean.Elements[1].Relations = nil
	clean.Environments, clean.Views = nil, nil
	f.source.overlayFn = func([]authoring.FileContent) (*arch.SourceModel, arch.Diagnostics) { return clean, nil }
	res := f.apply(t, false, EditInput{Op: "remove", Target: "element", Address: "container.api"})
	if !res.OK {
		t.Errorf("removal with no remaining referrers: %+v", res.Refusal)
	}
}

func TestRemoveCascade(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(dependentsModel())
	res := f.apply(t, true, EditInput{Op: "remove", Target: "element", Address: "container.api", Cascade: true})
	if !res.OK {
		t.Fatalf("cascade: %+v", res.Refusal)
	}
	wantRemoved := "component.h component.h.uses.out container.api container.web.uses.calls deployment.prod.instance.api"
	if got := strings.Join(res.Removed, " "); got != wantRemoved {
		t.Errorf("removed = %q, want %q", got, wantRemoved)
	}
	var planned []string
	for _, e := range f.editor.planned[0] {
		planned = append(planned, string(e.Op)+" "+string(e.Target)+" "+e.Address+" "+e.Entry)
	}
	want := []string{
		"remove element container.api ",
		"remove element component.h ",
		"remove relationship container.web.uses.calls ",
		"remove instance deployment.prod.instance.api ",
		"remove view_entry view.v container.api",
	}
	if !slices.Equal(planned, want) {
		t.Errorf("expanded edits:\n%s\nwant:\n%s", strings.Join(planned, "\n"), strings.Join(want, "\n"))
	}
}

func TestRemoveGroupHoldingInstances(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(dependentsModel())
	res := f.apply(t, false, EditInput{Op: "remove", Target: "group", Address: "deployment.prod.node.vpc"})
	refused(t, res, authoring.ReasonDanglingReferences, 0)
	if strings.Join(res.Refusal.Dependents, " ") != "deployment.prod.instance.api" {
		t.Errorf("dependents = %v", res.Refusal.Dependents)
	}
	ok := f.apply(t, true, EditInput{Op: "remove", Target: "group", Address: "deployment.prod.node.vpc", Cascade: true})
	if !ok.OK || strings.Join(ok.Removed, " ") != "deployment.prod.instance.api deployment.prod.node.vpc" {
		t.Errorf("group cascade: %+v %v", ok.Refusal, ok.Removed)
	}
}

// TestRemoveGroupAfterItsInstances: instances removed earlier in the same
// batch no longer hold the group, so the group's removal is accepted.
func TestRemoveGroupAfterItsInstances(t *testing.T) {
	t.Parallel()
	f := newApplyFixture(dependentsModel())
	res := f.apply(t, true,
		EditInput{Op: "remove", Target: "instance", Address: "deployment.prod.instance.api"},
		EditInput{Op: "remove", Target: "group", Address: "deployment.prod.node.vpc"})
	if !res.OK || strings.Join(res.Removed, " ") != "deployment.prod.instance.api deployment.prod.node.vpc" {
		t.Errorf("group after its instance in one batch: %+v %v", res.Refusal, res.Removed)
	}
	env := f.apply(t, true,
		EditInput{Op: "remove", Target: "element", Address: "container.api", Cascade: true},
		EditInput{Op: "remove", Target: "environment", Address: "deployment.prod"})
	if !env.OK {
		t.Errorf("environment after a cascade removed its instances: %+v", env.Refusal)
	}
}
