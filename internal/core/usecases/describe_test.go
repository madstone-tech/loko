package usecases

import (
	"fmt"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// describeModel: system shop { container web (React), container api (Go) {
// component h } }, system pay { container gw }, person u, and an environment
// prod placing api inside node vpc.
func describeModel() *arch.SourceModel {
	child := func(kind arch.ElementKind, name, parent, tech string, line int) arch.ElementDecl {
		e := elem(kind, name, line)
		e.Parent, e.Technology = ref(parent, line), tech
		return e
	}
	shop := elem(arch.KindSystem, "shop", 2)
	shop.Description, shop.Owner = "Storefront", "commerce"
	web := child(arch.KindContainer, "web", "system.shop", "React", 3)
	web.Relations = []arch.RelationDecl{{LocalName: "calls", Target: ref("container.api", 4), Range: at(4)}}
	api := child(arch.KindContainer, "api", "system.shop", "Go", 5)
	api.Relations = []arch.RelationDecl{{LocalName: "charge", Target: ref("container.gw", 6), Range: at(6)}}
	u := elem(arch.KindPerson, "u", 9)
	u.Relations = []arch.RelationDecl{{LocalName: "browse", Target: ref("container.web", 10), Range: at(10)}}
	return &arch.SourceModel{
		Project: arch.ProjectDecl{Name: "demo", Description: "Demo", Declared: true, Range: at(1)},
		Files:   []string{"a.loko.hcl"},
		Elements: []arch.ElementDecl{shop, web, api,
			child(arch.KindComponent, "h", "container.api", "", 7),
			elem(arch.KindSystem, "pay", 8),
			child(arch.KindContainer, "gw", "system.pay", "", 8), u},
		Environments: []arch.EnvironmentDecl{envWithInstance("prod", []string{"vpc"}, "api", "container.api")},
	}
}

func authDeps(m *arch.SourceModel, diags ...arch.Diagnostic) AuthoringDeps {
	return AuthoringDeps{
		Root:   "root",
		Source: &fakeOverlaySource{stubSource: stubSource{model: m, diags: diags}},
		Editor: &fakeEditor{rev: testRevision("a.loko.hcl")},
	}
}

func viewAddrs(views []ElementView) string {
	var out []string
	for _, v := range views {
		out = append(out, v.Address)
	}
	return strings.Join(out, " ")
}

func describe(t *testing.T, req DescribeRequest) *DescribeResult {
	t.Helper()
	got, err := Describe(t.Context(), authDeps(describeModel()), req)
	if err != nil {
		t.Fatal(err)
	}
	if !got.OK || got.Revision != testRevision("a.loko.hcl").Token() {
		t.Fatalf("describe %+v: ok=%v revision=%q", req, got.OK, got.Revision)
	}
	return got
}

func TestDescribeSummary(t *testing.T) {
	t.Parallel()
	got := describe(t, DescribeRequest{Level: "summary"})
	if got.Project.Name != "demo" || fmt.Sprint(got.Counts) != "[{component 1} {container 3} {person 1} {system 2}]" {
		t.Errorf("project/counts: %+v %v", got.Project, got.Counts)
	}
	if viewAddrs(got.Elements) != "person.u system.pay system.shop" {
		t.Errorf("summary lists top-level elements only, got %q", viewAddrs(got.Elements))
	}
	if len(got.Relationships) != 0 || len(got.Environments) != 1 || got.Environments[0].Name != "prod" ||
		len(got.Environments[0].Instances) != 0 {
		t.Errorf("summary: relationships %v environments %+v", got.Relationships, got.Environments)
	}
}

func TestDescribeStructure(t *testing.T) {
	t.Parallel()
	got := describe(t, DescribeRequest{Level: "structure"})
	want := "container.api container.gw container.web person.u system.pay system.shop"
	if viewAddrs(got.Elements) != want {
		t.Errorf("structure stops at containers: got %q", viewAddrs(got.Elements))
	}
	for _, e := range got.Elements {
		if e.Address == "container.api" && (e.Technology != "Go" || e.Components != 1 || e.Parent != "system.shop") {
			t.Errorf("container.api view: %+v", e)
		}
		if e.Description != "" {
			t.Errorf("structure omits descriptions: %+v", e)
		}
	}
}

func TestDescribeFull(t *testing.T) {
	t.Parallel()
	got := describe(t, DescribeRequest{Level: "full"})
	if len(got.Elements) != 7 || len(got.Relationships) != 3 {
		t.Fatalf("full: %d elements, %d relationships", len(got.Elements), len(got.Relationships))
	}
	if got.Elements[len(got.Elements)-1].Owner != "commerce" {
		t.Errorf("full carries every attribute: %+v", got.Elements[len(got.Elements)-1])
	}
	inst := got.Environments[0].Instances
	if len(inst) != 1 || inst[0].Of != "container.api" || inst[0].PlacedIn != "deployment.prod.node.vpc" {
		t.Errorf("full placement: %+v", inst)
	}
}

func TestDescribeScoped(t *testing.T) {
	t.Parallel()
	got := describe(t, DescribeRequest{Level: "structure", Address: "container.api"})
	if viewAddrs(got.Elements) != "component.h container.api" {
		t.Errorf("scope = element and children, got %q", viewAddrs(got.Elements))
	}
	var rels []string
	for _, r := range got.Relationships {
		rels = append(rels, r.Source+">"+r.Target)
	}
	if strings.Join(rels, " ") != "container.api>container.gw container.web>container.api" {
		t.Errorf("scope relationships: %v", rels)
	}
	missing, _ := Describe(t.Context(), authDeps(describeModel()), DescribeRequest{Address: "container.apu"})
	if missing.OK || missing.Error == nil || missing.Error.Reason != "not_found" || missing.Error.Suggestions[0] != "container.api" {
		t.Errorf("unknown scope: %+v", missing.Error)
	}
}

func TestDescribeCompileErrors(t *testing.T) {
	t.Parallel()
	m := describeModel()
	m.Elements[1].Parent = ref("system.missing", 3)
	got, err := Describe(t.Context(), authDeps(m), DescribeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got.OK || !got.Diags.HasErrors() || len(got.Elements) != 0 || got.Revision == "" {
		t.Errorf("compile errors must return diagnostics and no data (FR-007): %+v", got)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	m := describeModel()
	m.Elements[1].Parent = ref("system.missing", 3)
	got, err := Validate(t.Context(), authDeps(m))
	if err != nil {
		t.Fatal(err)
	}
	if got.OK || got.Errors == 0 || got.Diags[0].Range.File == "" || got.Revision == "" {
		t.Errorf("validate: %+v", got)
	}
	clean, _ := Validate(t.Context(), authDeps(describeModel()))
	if !clean.OK || clean.Errors != 0 {
		t.Errorf("clean project: %+v", clean)
	}
}
