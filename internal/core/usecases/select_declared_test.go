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

func TestResolveViewsDeclared(t *testing.T) {
	t.Parallel()
	ir := declaredIR(
		arch.View{Address: "view.landscape", Name: "landscape", Include: addrs("person.p")},
		arch.View{Address: "view.nothing", Name: "nothing", Include: addrs("system.shop"), Exclude: addrs("system.shop")},
		arch.View{Address: "view.pci", Name: "pci", Tags: []string{"pci"}},
	)
	prov := Provenance{entries: []provEntry{
		{"view.landscape", file("views.loko.hcl", 1)},
		{"view.nothing", file("views.loko.hcl", 5)},
	}}
	views, diags := ResolveViews(ir, prov)

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

	byCode := map[string]arch.Diagnostic{}
	for _, d := range diags {
		byCode[d.Code] = d
	}
	shadow, ok := byCode[arch.CodeViewShadowed]
	if !ok || shadow.Severity != arch.SeverityWarning || shadow.Range.StartLine != 1 || shadow.Address != "view.landscape" {
		t.Errorf("view_shadowed = %+v, %v", shadow, ok)
	}
	empty, ok := byCode[arch.CodeViewEmpty]
	if !ok || empty.Severity != arch.SeverityWarning || empty.Range.StartLine != 5 || empty.Address != "view.nothing" {
		t.Errorf("view_empty = %+v, %v", empty, ok)
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
