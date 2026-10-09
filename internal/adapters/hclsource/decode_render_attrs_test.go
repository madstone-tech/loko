package hclsource

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func TestDecodeRenderAttributes(t *testing.T) {
	t.Parallel()
	m, diags, err := New().Load(t.Context(), "testdata/golden/render_attrs_valid/input")
	if err != nil || diags.HasErrors() {
		t.Fatal(err, diags)
	}
	byName := map[string]arch.ElementDecl{}
	for _, e := range m.Elements {
		byName[e.Name] = e
	}
	if db := byName["db"]; db.Title != "Orders database" || db.Shape != "database" || db.AttrRanges["shape"].StartLine == 0 {
		t.Errorf("db: %+v", db)
	}
	rel := byName["worker"].Relations[0]
	if rel.Kind != "trigger" || rel.KindRange.StartLine != 21 || len(rel.Tags) != 1 || rel.Tags[0] != "read" {
		t.Errorf("relation: %+v", rel)
	}
	if v := m.Views[0]; v.Direction != "right" || v.DirectionRange.StartLine == 0 {
		t.Errorf("view: %+v", v)
	}
}
