package authoring

import (
	"strings"
	"testing"
)

func refs(a ...string) AttrValue { return AttrValue{Kind: ValueRefList, List: a} }

func TestViewEdits(t *testing.T) {
	t.Parallel()
	ok := []Edit{
		{Op: OpAdd, Target: TargetView, Address: "view.event-flow",
			Set: []Attr{{"include", refs("container.orders", "system.s")}, {"tags", list("sns")}}},
		{Op: OpUpdate, Target: TargetView, Address: "view.event-flow", Clear: []string{"exclude"}},
		{Op: OpRemove, Target: TargetView, Address: "view.event-flow"},
		{Op: OpAdd, Target: TargetView, Address: "view.all"},
	}
	for _, e := range ok {
		if _, err := NewEdit(e); err != nil {
			t.Errorf("NewEdit(%s %s) = %v", e.Op, e.Address, err)
		}
	}
	bad := []struct {
		edit  Edit
		field string
	}{
		{Edit{Op: OpAdd, Target: TargetView, Address: "views.x"}, "address"},
		{Edit{Op: OpAdd, Target: TargetView, Address: "view.x.y"}, "address"},
		{Edit{Op: OpAdd, Target: TargetView, Address: "view.x", Set: []Attr{{"include", list("container.a")}}}, "set.include"},
		{Edit{Op: OpAdd, Target: TargetView, Address: "view.x", Set: []Attr{{"include", refs("container.a", "deployment.p")}}}, "set.include"},
		{Edit{Op: OpAdd, Target: TargetView, Address: "view.x", Set: []Attr{{"region", str("x")}}}, "set.region"},
		{Edit{Op: OpRename, Target: TargetView, Address: "view.x", To: "view.y"}, "op"},
	}
	for _, tt := range bad {
		_, err := NewEdit(tt.edit)
		if err == nil || !strings.Contains(err.Error(), tt.field) {
			t.Errorf("%+v: err = %v, want one naming %q", tt.edit, err, tt.field)
		}
	}
	if p, ok := SplitAddress(TargetView, "view.event-flow"); !ok || p.Local != "event-flow" {
		t.Errorf("SplitAddress(view) = %+v %v", p, ok)
	}
}
