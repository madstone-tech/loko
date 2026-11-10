package usecases

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// redundantModel has system s with containers api and db, and a container
// api that uses target. Each extra component of api uses its own target.
func redundantModel(apiKind string, comps ...arch.ElementDecl) *arch.SourceModel {
	m := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "s", 1),
		elem(arch.KindSystem, "ext", 2),
		{Kind: arch.KindContainer, Name: "ledger", Range: at(3), Parent: ref("system.ext", 3)},
		{Kind: arch.KindComponent, Name: "book", Range: at(4), Parent: ref("container.ledger", 4)},
		{Kind: arch.KindContainer, Name: "db", Range: at(5), Parent: ref("system.s", 5)},
		{Kind: arch.KindContainer, Name: "api", Range: at(6), Parent: ref("system.s", 6),
			Relations: []arch.RelationDecl{{LocalName: "rec", Target: ref("container.ledger", 7), Kind: apiKind, Range: at(7)}}},
	}}
	m.Elements = append(m.Elements, comps...)
	return m
}

func redundantComp(name, parent, target, kind string, line int) arch.ElementDecl {
	return arch.ElementDecl{Kind: arch.KindComponent, Name: name, Range: at(line), Parent: ref(parent, line),
		Relations: []arch.RelationDecl{{LocalName: "rec", Target: ref(target, line+1), Kind: kind, Range: at(line + 1)}}}
}

func TestValidateRedundantRelationships(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		model   *arch.SourceModel
		related []int
		detail  []string
		absent  string
	}{
		{"one component repeats it", redundantModel("", redundantComp("h", "container.api", "container.ledger", "", 10)),
			[]int{11}, []string{"container.api.uses.rec and component.h.uses.rec both use container.ledger.",
				"Remove the container relationship, or give it a different target."}, "The kinds differ"},
		{"two components sorted by address", redundantModel("",
			redundantComp("z", "container.api", "container.ledger", "", 20), redundantComp("a", "container.api", "container.ledger", "", 10)),
			[]int{11, 21}, []string{"component.a.uses.rec"}, ""},
		{"component of another container", redundantModel("", redundantComp("h", "container.db", "container.ledger", "", 10)),
			nil, nil, ""},
		{"component targets a child of the target", redundantModel("", redundantComp("h", "container.api", "component.book", "", 10)),
			nil, nil, ""},
		{"kinds differ", redundantModel("", redundantComp("h", "container.api", "container.ledger", "async", 10)),
			[]int{11}, []string{"The kinds differ: sync and async."}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, rd := ResolveModel(tt.model)
			if len(rd) != 0 {
				t.Fatalf("fixture: %v", rd)
			}
			d := ValidateRedundantRelationships(tt.model, res)
			if tt.related == nil {
				if len(d) != 0 {
					t.Fatalf("got %v, want none", d)
				}
				return
			}
			if len(d) != 1 {
				t.Fatalf("got %d diagnostics, want 1: %v", len(d), d)
			}
			g := d[0]
			if g.Code != arch.CodeRedundantRelationship || g.Severity != arch.SeverityInfo ||
				g.Summary != "Relationship repeats a component relationship" ||
				g.Address != "container.api.uses.rec" || g.Range != at(7) {
				t.Errorf("diagnostic %+v", g)
			}
			if len(g.Related) != len(tt.related) {
				t.Fatalf("related %v, want lines %v", g.Related, tt.related)
			}
			for i, line := range tt.related {
				if g.Related[i].Range != at(line) {
					t.Errorf("related[%d] %v, want line %d", i, g.Related[i], line)
				}
			}
			for _, want := range tt.detail {
				if !strings.Contains(g.Detail, want) {
					t.Errorf("detail %q lacks %q", g.Detail, want)
				}
			}
			if tt.absent != "" && strings.Contains(g.Detail, tt.absent) {
				t.Errorf("detail %q has %q", g.Detail, tt.absent)
			}
			if strings.Contains(g.Detail, ";") {
				t.Errorf("detail has a semicolon: %q", g.Detail)
			}
		})
	}
}
