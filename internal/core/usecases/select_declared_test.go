package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func declaredIR(views ...arch.View) *arch.IR {
	ir := irOf(
		[]arch.Element{
			el(arch.KindPerson, "p", ""),
			el(arch.KindSystem, "shop", ""),
			el(arch.KindSystem, "pay", "", "pci"),
			el(arch.KindContainer, "api", "system.shop"),
			el(arch.KindContainer, "gw", "system.pay", "pci", "edge"),
			el(arch.KindContainer, "ledger", "system.pay"),
			el(arch.KindComponent, "k", "container.gw"),
		},
		[]arch.Relationship{
			rl("container.api", "gw", "container.gw", "charges", ""),
			rl("container.gw", "ledger", "container.ledger", "records", ""),
			rl("person.p", "api", "container.api", "uses", ""),
		},
	)
	return arch.NewIR(ir.Project, ir.Elements, ir.Relationships, nil, views, nil)
}

func addrs(as ...string) []arch.Address {
	out := make([]arch.Address, len(as))
	for i, a := range as {
		out[i] = arch.Address(a)
	}
	return out
}

func TestSelectDeclared(t *testing.T) {
	t.Parallel()
	ir := declaredIR()
	tests := []struct {
		name string
		view arch.View
		want []arch.Address
	}{
		{"include takes descendants", arch.View{Include: addrs("system.pay")},
			addrs("component.k", "container.gw", "container.ledger", "system.pay")},
		{"exclude removes a subtree", arch.View{Include: addrs("system.pay"), Exclude: addrs("container.gw")},
			addrs("container.ledger", "system.pay")},
		{"tags select elements carrying any listed tag", arch.View{Tags: []string{"edge", "nope"}},
			addrs("container.gw")},
		{"tags are OR-ed", arch.View{Tags: []string{"pci"}},
			addrs("container.gw", "system.pay")},
		{"include plus tags", arch.View{Include: addrs("person.p"), Tags: []string{"edge"}},
			addrs("container.gw", "person.p")},
		{"no include and no tags means everything", arch.View{Exclude: addrs("system.pay")},
			addrs("container.api", "person.p", "system.shop")},
		{"exclude everything", arch.View{Include: addrs("system.shop"), Exclude: addrs("system.shop")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := selectDeclared(ir, tt.view); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectDeclared = %v, want %v", got, tt.want)
			}
		})
	}
}

// declaredModel is the source model behind declaredIR's IR, with the view
// declarations carrying their own line numbers, so the compile-stage
// diagnostics can be tested end to end.
func declaredModel() *arch.SourceModel {
	tags := func(e arch.ElementDecl, names ...string) arch.ElementDecl {
		e.Tags = names
		return e
	}
	elem := func(kind arch.ElementKind, name, parent string, line int) arch.ElementDecl {
		e := arch.ElementDecl{Kind: kind, Name: name, Range: at(line)}
		if parent != "" {
			e.Parent = ref(parent, line)
		}
		return e
	}
	uses := func(local, target string, line int) arch.RelationDecl {
		return arch.RelationDecl{LocalName: local, Target: ref(target, line), Range: at(line)}
	}
	withRelations := func(e arch.ElementDecl, rels ...arch.RelationDecl) arch.ElementDecl {
		e.Relations = rels
		return e
	}
	return &arch.SourceModel{
		Project: arch.ProjectDecl{Name: "test", Declared: true, Range: at(1)},
		Files:   []string{"a.loko.hcl"},
		Elements: []arch.ElementDecl{
			withRelations(elem(arch.KindPerson, "p", "", 2), uses("api", "container.api", 11)),
			elem(arch.KindSystem, "shop", "", 3),
			tags(elem(arch.KindSystem, "pay", "", 4), "pci"),
			withRelations(elem(arch.KindContainer, "api", "system.shop", 5), uses("gw", "container.gw", 9)),
			withRelations(tags(elem(arch.KindContainer, "gw", "system.pay", 6), "pci", "edge"), uses("ledger", "container.ledger", 10)),
			elem(arch.KindContainer, "ledger", "system.pay", 7),
			elem(arch.KindComponent, "k", "container.gw", 8),
		},
		Views: []arch.ViewDecl{
			{Name: "landscape", Include: []arch.Reference{ref("person.p", 20)}, Range: at(20)},
			{Name: "nothing", Include: []arch.Reference{ref("system.shop", 25)},
				Exclude: []arch.Reference{ref("system.shop", 25)}, Range: at(25)},
			{Name: "pci", Tags: []string{"pci"}, Range: at(30)},
		},
	}
}

func mustResolve(m *arch.SourceModel) *Resolved {
	res, diags := ResolveModel(m)
	for _, d := range diags {
		if d.Severity == arch.SeverityError {
			panic("fixture must resolve cleanly: " + d.Code + " " + string(d.Address))
		}
	}
	return res
}

func TestResolveViewsDeclared(t *testing.T) {
	t.Parallel()
	m := declaredModel()

	// ResolveViews keeps the replacement and non-production behaviour: the
	// declared landscape replaces the derived one, view.nothing yields nothing.
	ir := BuildIR(m, mustResolve(m))
	views := ResolveViews(ir, Provenance{})
	var landscape []string
	for _, v := range views {
		if v.ID == "landscape" {
			landscape = append(landscape, string(v.Kind))
		}
		if v.ID == "nothing" {
			t.Error("an empty declared view must not be produced (FR-005)")
		}
		if v.ID == "pci" && (v.Kind != "declared" || v.Subject != "view.pci" || v.Selection == nil ||
			!reflect.DeepEqual(v.Selection.Tags, []string{"pci"})) {
			t.Errorf("declared view = %+v", v)
		}
	}
	if !reflect.DeepEqual(landscape, []string{"declared"}) {
		t.Errorf("landscape views = %v, want only the declared one (FR-004)", landscape)
	}

	// The reporting lives in the compile stage now, with the declaring line.
	byCode := map[string]arch.Diagnostic{}
	for _, d := range DeclaredViewDiagnostics(m, mustResolve(m)) {
		byCode[d.Code] = d
	}
	shadow, ok := byCode[arch.CodeViewShadowed]
	if !ok || shadow.Severity != arch.SeverityWarning || shadow.Range.StartLine != 20 || shadow.Address != "view.landscape" {
		t.Errorf("view_shadowed = %+v, %v", shadow, ok)
	}
	empty, ok := byCode[arch.CodeViewEmpty]
	if !ok || empty.Severity != arch.SeverityWarning || empty.Range.StartLine != 25 || empty.Address != "view.nothing" {
		t.Errorf("view_empty = %+v, %v", empty, ok)
	}
}

// TestDeclaredViewDiagnosticsSteps: a view with steps is a dynamic flow; it
// takes no selection (feature 019) and never reports view_empty, while the
// same model without the steps does.
func TestDeclaredViewDiagnosticsSteps(t *testing.T) {
	t.Parallel()
	m := declaredModel()
	m.Views[1].Steps = []arch.StepDecl{{Name: "go", From: ref("person.p", 26), To: ref("container.api", 26)}}
	res := mustResolve(m)
	for _, d := range DeclaredViewDiagnostics(m, res) {
		if d.Code == arch.CodeViewEmpty && d.Address == "view.nothing" {
			t.Errorf("a dynamic view reported view_empty: %+v", d)
		}
	}
}

func TestProjectDeclared(t *testing.T) {
	t.Parallel()
	ir := declaredIR(arch.View{Address: "view.path", Name: "path",
		Include: addrs("system.shop", "system.pay"), Exclude: addrs("container.ledger")})
	vm := projectOne(t, ir, "path")
	idx := nodeIndex(vm)

	if _, ok := idx["container__ledger"]; ok {
		t.Error("an excluded element appears")
	}
	if idx["container__gw"].Parent != "system__pay" || idx["component__k"].Parent != "container__gw" {
		t.Errorf("nesting under the nearest visible ancestor: gw=%+v k=%+v", idx["container__gw"], idx["component__k"])
	}
	for _, n := range vm.Nodes {
		if n.Role == "subject" {
			t.Errorf("a declared view has no subject boundary: %+v", n)
		}
	}
	// The relationship to the excluded ledger reaches the boundary rather
	// than lifting to the ledger's (visible) system (US2/AC2).
	want := []string{"container__api--container__gw", "container__gw--outside", "outside--container__api"}
	if got := edgeIDs(vm); !reflect.DeepEqual(got, want) {
		t.Errorf("edges = %v, want %v", got, want)
	}
}
