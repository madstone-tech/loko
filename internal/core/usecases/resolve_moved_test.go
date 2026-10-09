package usecases

import (
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func moved(from, to string, line int) arch.MovedDecl {
	return arch.MovedDecl{From: ref(from, line), To: ref(to, line), Range: at(line)}
}

func movedModel(moves ...arch.MovedDecl) *arch.SourceModel {
	return &arch.SourceModel{
		Elements: []arch.ElementDecl{elem(arch.KindSystem, "s", 1), elem(arch.KindComponent, "api", 2), elem(arch.KindSystem, "c", 3)},
		Moved:    moves,
	}
}

func TestValidateMoved(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		moves []arch.MovedDecl
		want  []string
	}{
		{"a kind change", []arch.MovedDecl{moved("container.api", "component.api", 10)}, nil},
		{"a chain: only the end is declared", []arch.MovedDecl{moved("system.a", "system.b", 10), moved("system.b", "system.c", 11)}, nil},
		{"from still declared", []arch.MovedDecl{moved("system.s", "system.c", 10)}, []string{arch.CodeMovedFromDeclared}},
		{"to not declared", []arch.MovedDecl{moved("system.old", "system.nowhere", 10)}, []string{arch.CodeMovedToUnresolved}},
		{"a chain that never lands", []arch.MovedDecl{moved("system.a", "system.b", 10), moved("system.b", "system.x", 11)},
			[]string{arch.CodeMovedToUnresolved, arch.CodeMovedToUnresolved}},
		{"a cycle", []arch.MovedDecl{moved("system.a", "system.b", 10), moved("system.b", "system.a", 11)},
			[]string{arch.CodeMovedToUnresolved, arch.CodeMovedToUnresolved}},
		{"duplicate from", []arch.MovedDecl{moved("system.old", "system.s", 10), moved("system.old", "system.c", 12)},
			[]string{arch.CodeMovedDuplicateFrom}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateMoved(movedModel(tt.moves...))
			if !slices.Equal(codes(got), tt.want) {
				t.Errorf("codes = %v, want %v", codes(got), tt.want)
			}
			for _, d := range got {
				if d.Range.StartLine == 0 {
					t.Errorf("%s has no range", d.Code)
				}
				if d.Code == arch.CodeMovedDuplicateFrom && (len(d.Related) != 1 || d.Related[0].Range.StartLine != 10) {
					t.Errorf("duplicate_from points at the first block: %+v", d.Related)
				}
			}
		})
	}
}

func TestBuildIRCarriesMoves(t *testing.T) {
	t.Parallel()
	m := movedModel(moved("system.z", "system.s", 10), moved("container.api", "component.api", 11))
	res, diags := ResolveModel(m)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	ir := BuildIR(m, res)
	if len(ir.Moves) != 2 || ir.Moves[0].From != "container.api" || ir.Moves[1].To != "system.s" {
		t.Errorf("moves sorted by from: %+v", ir.Moves)
	}
	if empty := BuildIR(movedModel(), res); empty.Moves != nil {
		t.Error("no moved blocks means no moves field (omitempty keeps old exports byte-identical)")
	}
}
