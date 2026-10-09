package authoring

import (
	"strings"
	"testing"
)

func TestRenderAttributeEdits(t *testing.T) {
	t.Parallel()
	ok := []Edit{
		{Op: OpUpdate, Target: TargetElement, Address: "system.s", Set: []Attr{{"title", str("Shop")}}},
		{Op: OpUpdate, Target: TargetElement, Address: "container.db", Set: []Attr{{"shape", str("database")}}},
		{Op: OpUpdate, Target: TargetElement, Address: "external.psp", Set: []Attr{{"shape", str("queue")}}},
		{Op: OpUpdate, Target: TargetRelationship, Address: "container.a.uses.b", Set: []Attr{{"kind", str("trigger")}, {"tags", list("read")}}},
		{Op: OpUpdate, Target: TargetView, Address: "view.v", Set: []Attr{{"direction", str("right")}}},
		{Op: OpUpdate, Target: TargetView, Address: "view.v", Set: []Attr{{"layout", str("elk")}}},
		{Op: OpUpdate, Target: TargetRelationship, Address: "container.a.uses.b", Clear: []string{"kind"}},
	}
	for _, e := range ok {
		if _, err := NewEdit(e); err != nil {
			t.Errorf("%s %s: %v", e.Target, e.Address, err)
		}
	}
	bad := []struct {
		edit Edit
		want string
	}{
		{Edit{Op: OpUpdate, Target: TargetElement, Address: "system.s", Set: []Attr{{"shape", str("database")}}}, "set.shape"},
		{Edit{Op: OpUpdate, Target: TargetElement, Address: "container.db", Set: []Attr{{"shape", str("cylinder")}}}, "database, queue, topic, function, bucket"},
		{Edit{Op: OpUpdate, Target: TargetRelationship, Address: "container.a.uses.b", Set: []Attr{{"kind", str("event")}}}, "sync, async, trigger"},
		{Edit{Op: OpUpdate, Target: TargetView, Address: "view.v", Set: []Attr{{"direction", str("left")}}}, "down, right"},
		{Edit{Op: OpUpdate, Target: TargetView, Address: "view.v", Set: []Attr{{"layout", str("neato")}}}, "dagre, elk"},
		{Edit{Op: OpUpdate, Target: TargetElement, Address: "system.s", Set: []Attr{{"title", str("")}}}, "set.title"},
	}
	for _, tt := range bad {
		_, err := NewEdit(tt.edit)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%+v: err = %v, want one containing %q", tt.edit.Set, err, tt.want)
		}
	}
}
