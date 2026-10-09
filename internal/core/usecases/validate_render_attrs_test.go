package usecases

import (
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func renderModel() *arch.SourceModel {
	withAttr := func(e arch.ElementDecl, name string, line int) arch.ElementDecl {
		if e.AttrRanges == nil {
			e.AttrRanges = map[string]arch.SourceRange{}
		}
		e.AttrRanges[name] = at(line)
		return e
	}
	sys := withAttr(elem(arch.KindSystem, "s", 1), "title", 2)
	sys.Title = "Shop"
	db := withAttr(elem(arch.KindContainer, "db", 3), "shape", 4)
	db.Parent, db.Shape = ref("system.s", 3), "database"
	db.Relations = []arch.RelationDecl{{LocalName: "r", Target: ref("system.s", 5), Kind: "trigger", KindRange: at(5), Range: at(5)}}
	return &arch.SourceModel{
		Elements: []arch.ElementDecl{sys, db},
		Views:    []arch.ViewDecl{{Name: "v", Direction: "down", DirectionRange: at(9), Range: at(9)}},
	}
}

func TestValidateRenderAttributesAcceptsValid(t *testing.T) {
	t.Parallel()
	if d := ValidateRenderAttributes(renderModel()); len(d) != 0 {
		t.Errorf("valid attributes: %v", codes(d))
	}
}

func TestValidateRenderAttributesRejects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*arch.SourceModel)
		code   string
		line   int
		detail string
	}{
		{"unknown shape", func(m *arch.SourceModel) { m.Elements[1].Shape = "cylinder" }, arch.CodeInvalidAttributeValue, 4, "database, queue, topic, function, bucket"},
		{"shape on a system", func(m *arch.SourceModel) {
			m.Elements[0].Shape = "database"
			m.Elements[0].AttrRanges["shape"] = at(7)
		}, arch.CodeShapeNotAllowed, 7, "container"},
		{"empty title", func(m *arch.SourceModel) { m.Elements[0].Title = "" }, arch.CodeEmptyTitle, 2, ""},
		{"unknown kind", func(m *arch.SourceModel) { m.Elements[1].Relations[0].Kind = "event" }, arch.CodeInvalidAttributeValue, 5, "sync, async, trigger"},
		{"unknown direction", func(m *arch.SourceModel) { m.Views[0].Direction = "left" }, arch.CodeInvalidAttributeValue, 9, "down, right"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := renderModel()
			tt.mutate(m)
			d := ValidateRenderAttributes(m)
			if !slices.Equal(codes(d), []string{tt.code}) {
				t.Fatalf("codes = %v, want [%s]", codes(d), tt.code)
			}
			if d[0].Range.StartLine != tt.line || !strings.Contains(d[0].Detail, tt.detail) || d[0].Severity != arch.SeverityError {
				t.Errorf("diagnostic %+v: want line %d and detail containing %q", d[0], tt.line, tt.detail)
			}
		})
	}
}

func TestBuildIRCarriesRenderAttributes(t *testing.T) {
	t.Parallel()
	m := renderModel()
	m.Elements[1].Relations = append(m.Elements[1].Relations,
		arch.RelationDecl{LocalName: "s", Target: ref("system.s", 6), Kind: "sync", Tags: []string{"w", "r", "w"}, Range: at(6)})
	res, diags := ResolveModel(m)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	ir := BuildIR(m, res)
	db, _ := ir.Element("container.db")
	if db.Shape != "database" {
		t.Errorf("shape = %q", db.Shape)
	}
	if s, _ := ir.Element("system.s"); s.Title != "Shop" {
		t.Errorf("title = %q", s.Title)
	}
	trig, _ := ir.Relationship("container.db.uses.r")
	sync, _ := ir.Relationship("container.db.uses.s")
	if trig.Kind != "trigger" || sync.Kind != "" || strings.Join(sync.Tags, ",") != "r,w" {
		t.Errorf("relationships: %+v %+v", trig, sync)
	}
	if ir.Views[0].Direction != "down" {
		t.Errorf("view direction = %q", ir.Views[0].Direction)
	}
}
