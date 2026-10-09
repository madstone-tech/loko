package arch

import "testing"

func TestRenderAttributeSets(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"database", "queue", "topic", "function", "bucket"} {
		if !ValidShape(s) {
			t.Errorf("ValidShape(%q) = false", s)
		}
	}
	for _, s := range []string{"", "cylinder", "Database"} {
		if ValidShape(s) {
			t.Errorf("ValidShape(%q) = true", s)
		}
	}
	for _, k := range []string{"sync", "async", "trigger"} {
		if !ValidRelationshipKind(k) {
			t.Errorf("ValidRelationshipKind(%q) = false", k)
		}
	}
	if ValidRelationshipKind("event") || ValidRelationshipKind("") {
		t.Error("only sync, async and trigger are relationship kinds")
	}
	for _, d := range []string{"down", "right"} {
		if !ValidDirection(d) {
			t.Errorf("ValidDirection(%q) = false", d)
		}
	}
	if ValidDirection("left") || ValidDirection("") {
		t.Error("only down and right are directions")
	}
	if !ShapeAllowedOn(KindContainer) || !ShapeAllowedOn(KindExternal) || ShapeAllowedOn(KindSystem) || ShapeAllowedOn(KindPerson) || ShapeAllowedOn(KindComponent) {
		t.Error("shape is allowed on containers and externals only")
	}
	if len(Shapes) != 5 || len(RelationshipKinds) != 3 || len(Directions) != 2 {
		t.Error("value sets changed size")
	}
}
