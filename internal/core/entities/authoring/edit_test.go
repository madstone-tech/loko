package authoring

import (
	"strings"
	"testing"
)

func str(s string) AttrValue     { return AttrValue{Kind: ValueString, Str: s} }
func ref(a string) AttrValue     { return AttrValue{Kind: ValueRef, Ref: a} }
func list(s ...string) AttrValue { return AttrValue{Kind: ValueList, List: s} }

func TestNewEditAccepts(t *testing.T) {
	t.Parallel()
	ok := []Edit{
		{Op: OpAdd, Target: TargetElement, Address: "system.shop", Set: []Attr{{"description", str("Shop")}}},
		{Op: OpAdd, Target: TargetElement, Address: "container.api", Set: []Attr{{"system", ref("system.shop")}, {"tags", list("edge")}}},
		{Op: OpAdd, Target: TargetElement, Address: "component.h", Set: []Attr{{"container", ref("container.api")}}},
		{Op: OpAdd, Target: TargetRelationship, Address: "container.api.uses.orders", Set: []Attr{{"target", ref("container.db")}}},
		{Op: OpAdd, Target: TargetEnvironment, Address: "deployment.prod", Set: []Attr{{"region", str("us-east-1")}}},
		{Op: OpAdd, Target: TargetGroup, Address: "deployment.prod.node.vpc.subnet-a"},
		{Op: OpAdd, Target: TargetInstance, Address: "deployment.prod.instance.api", Set: []Attr{{"of", ref("container.api")}}},
		{Op: OpAdd, Target: TargetBinding, Address: "deployment.prod.instance.api", Binding: BindingRef{Kind: "terraform"},
			Set: []Attr{{"address", str("module.api.this")}}},
		{Op: OpUpdate, Target: TargetElement, Address: "container.api", Clear: []string{"owner"}},
		{Op: OpRemove, Target: TargetElement, Address: "container.api", Cascade: true},
		{Op: OpRename, Target: TargetElement, Address: "container.api", To: "component.api"},
		{Op: OpAdd, Target: TargetElement, Address: "system.x", File: "sub/x.loko.hcl"},
	}
	for _, e := range ok {
		if _, err := NewEdit(e); err != nil {
			t.Errorf("NewEdit(%s %s %s) = %v, want nil", e.Op, e.Target, e.Address, err)
		}
	}
}

func TestNewEditRejects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		edit  Edit
		field string
	}{
		{"rename a relationship", Edit{Op: OpRename, Target: TargetRelationship, Address: "container.a.uses.b", To: "container.a.uses.c"}, "op"},
		{"cascade on update", Edit{Op: OpUpdate, Target: TargetElement, Address: "system.s", Cascade: true, Clear: []string{"owner"}}, "cascade"},
		{"to on add", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", To: "system.t"}, "to"},
		{"file on update", Edit{Op: OpUpdate, Target: TargetElement, Address: "system.s", File: "x.loko.hcl", Clear: []string{"owner"}}, "file"},
		{"bad element kind", Edit{Op: OpAdd, Target: TargetElement, Address: "service.s"}, "address"},
		{"bad name", Edit{Op: OpAdd, Target: TargetElement, Address: "system.1abc"}, "address"},
		{"relationship form", Edit{Op: OpAdd, Target: TargetRelationship, Address: "container.a.calls.b", Set: []Attr{{"target", ref("container.b")}}}, "address"},
		{"group form", Edit{Op: OpAdd, Target: TargetGroup, Address: "deployment.prod.instance.x"}, "address"},
		{"illegal attribute", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", Set: []Attr{{"region", str("x")}}}, "set.region"},
		{"system on a system", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", Set: []Attr{{"system", ref("system.t")}}}, "set.system"},
		{"reference as string", Edit{Op: OpAdd, Target: TargetElement, Address: "container.c", Set: []Attr{{"system", str("system.s")}}}, "set.system"},
		{"tags not a list", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", Set: []Attr{{"tags", str("x")}}}, "set.tags"},
		{"relationship needs target", Edit{Op: OpAdd, Target: TargetRelationship, Address: "container.a.uses.b"}, "set.target"},
		{"instance needs of", Edit{Op: OpAdd, Target: TargetInstance, Address: "deployment.p.instance.a"}, "set.of"},
		{"container needs system", Edit{Op: OpAdd, Target: TargetElement, Address: "container.c"}, "set.system"},
		{"binding needs exactly one selector", Edit{Op: OpAdd, Target: TargetBinding, Address: "deployment.p.instance.a",
			Binding: BindingRef{Kind: "terraform"}, Set: []Attr{{"address", str("a")}, {"addresses", list("b")}}}, "set"},
		{"binding kind", Edit{Op: OpAdd, Target: TargetBinding, Address: "deployment.p.instance.a",
			Binding: BindingRef{Kind: "pulumi"}, Set: []Attr{{"address", str("a")}}}, "binding.kind"},
		{"file extension", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", File: "x.hcl"}, "file"},
		{"file absolute", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", File: "/x.loko.hcl"}, "file"},
		{"file escapes root", Edit{Op: OpAdd, Target: TargetElement, Address: "system.s", File: "../x.loko.hcl"}, "file"},
		{"empty update", Edit{Op: OpUpdate, Target: TargetElement, Address: "system.s"}, "set"},
		{"rename to itself", Edit{Op: OpRename, Target: TargetElement, Address: "system.s", To: "system.s"}, "to"},
		{"rename to non-element", Edit{Op: OpRename, Target: TargetElement, Address: "system.s", To: "deployment.s"}, "to"},
		{"unknown op", Edit{Op: "upsert", Target: TargetElement, Address: "system.s"}, "op"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewEdit(tt.edit)
			if err == nil {
				t.Fatal("NewEdit accepted an invalid edit")
			}
			if !strings.Contains(err.Error(), tt.field) {
				t.Errorf("error %q does not name field %q", err, tt.field)
			}
		})
	}
}

func TestNewEditNormalises(t *testing.T) {
	t.Parallel()
	e, err := NewEdit(Edit{Op: OpUpdate, Target: TargetElement, Address: "system.s",
		Set: []Attr{{"owner", str("o")}, {"description", str("d")}}, Clear: []string{"tags", "docs"}})
	if err != nil {
		t.Fatal(err)
	}
	if e.Set[0].Name != "description" || e.Clear[0] != "docs" {
		t.Errorf("Set and Clear must be sorted: %+v", e)
	}
}

func TestValidName(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]bool{"a": true, "_x": true, "a-b_1": true, "1a": false, "-a": false, "a.b": false, "": false} {
		if got := ValidName(name); got != want {
			t.Errorf("ValidName(%q) = %v, want %v", name, got, want)
		}
	}
}
