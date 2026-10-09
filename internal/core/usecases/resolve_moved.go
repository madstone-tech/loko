package usecases

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ValidateMoved checks the moved blocks (contracts/language-moved.md): from
// must not be declared, to must resolve to a declared element (directly or
// along a chain of moves), and no address may be moved twice. A kind change
// is allowed; parent rules are enforced on the renamed element itself.
func ValidateMoved(model *arch.SourceModel) arch.Diagnostics {
	declared := map[arch.Address]bool{}
	for _, e := range model.Elements {
		declared[e.Address()] = true
	}
	first := map[arch.Address]arch.MovedDecl{}
	next := map[arch.Address]arch.Address{}
	var diags arch.Diagnostics
	for _, m := range model.Moved {
		from := arch.Address(m.From.Raw)
		if prev, dup := first[from]; dup {
			diags = append(diags, movedDiag(m.From.Range, arch.CodeMovedDuplicateFrom, "Address moved twice",
				fmt.Sprintf("%s is already the source of a moved block; each address can be moved once.", from),
				arch.RelatedRange{Message: "first moved here", Range: prev.Range}))
			continue
		}
		first[from], next[from] = m, arch.Address(m.To.Raw)
		if declared[from] {
			diags = append(diags, movedDiag(m.From.Range, arch.CodeMovedFromDeclared, "Moved from an address that is still declared",
				fmt.Sprintf("%s is still declared, so it cannot have moved. Remove the moved block or the declaration.", from)))
		}
	}
	for _, m := range model.Moved {
		if first[arch.Address(m.From.Raw)].Range != m.Range {
			continue // a duplicate, already reported
		}
		if !lands(arch.Address(m.To.Raw), declared, next) {
			diags = append(diags, movedDiag(m.To.Range, arch.CodeMovedToUnresolved, "Moved to an address that is not declared",
				fmt.Sprintf("%s is not declared, and no chain of moved blocks leads from it to a declared element.", m.To.Raw)))
		}
	}
	return diags
}

// lands follows a chain of moves until it reaches a declared element, or a
// dead end or a cycle.
func lands(a arch.Address, declared map[arch.Address]bool, next map[arch.Address]arch.Address) bool {
	seen := map[arch.Address]bool{}
	for !declared[a] {
		to, ok := next[a]
		if !ok || seen[a] {
			return false
		}
		seen[a], a = true, to
	}
	return true
}

func movedDiag(r arch.SourceRange, code, summary, detail string, related ...arch.RelatedRange) arch.Diagnostic {
	return arch.Diagnostic{Severity: arch.SeverityError, Code: code, Summary: summary, Detail: detail, Range: r, Related: related}
}

// buildMoves carries the moved blocks into the IR, sorted by From, or nil.
func buildMoves(model *arch.SourceModel) []arch.Move {
	if len(model.Moved) == 0 {
		return nil
	}
	out := make([]arch.Move, 0, len(model.Moved))
	for _, m := range model.Moved {
		out = append(out, arch.Move{From: arch.Address(m.From.Raw), To: arch.Address(m.To.Raw), Range: m.Range})
	}
	slices.SortFunc(out, func(a, b arch.Move) int { return cmp.Compare(a.From, b.From) })
	return out
}
