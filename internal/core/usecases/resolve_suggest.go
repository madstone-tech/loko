package usecases

import (
	"fmt"
	"sort"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// unresolvedDetail names the closest declared address when there is a
// plausible one. A typo is the common case, and naming the likely intended
// element is most of the diagnostic's value.
func (st SymbolTable) unresolvedDetail(raw string) string {
	base := fmt.Sprintf("No element %q is declared.", raw)
	if best, ok := st.nearest(raw); ok {
		return base + fmt.Sprintf(" Did you mean %s?", best)
	}
	return base
}

// nearest returns the declared address closest to raw by edit distance, when
// one is close enough to be worth suggesting.
func (st SymbolTable) nearest(raw string) (arch.Address, bool) {
	// A suggestion is only useful if it is a near miss; beyond a third of the
	// string being wrong it is noise.
	budget := len(raw) / 3
	if budget < 1 {
		budget = 1
	}

	candidates := make([]string, 0, len(st.kinds))
	for addr := range st.kinds {
		candidates = append(candidates, string(addr))
	}
	// Sorted so the suggestion is deterministic when two are equally close.
	sort.Strings(candidates)

	best, bestDist := "", budget+1
	for _, c := range candidates {
		if d := editDistance(raw, c); d < bestDist {
			best, bestDist = c, d
		}
	}
	if best == "" {
		return "", false
	}
	return arch.Address(best), true
}

// editDistance is Levenshtein distance over two rows.
func editDistance(a, b string) int {
	if a == b {
		return 0
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = minInt(minInt(curr[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// orList renders a kind list as "a system" or "a container or a component".
func orList(kinds []string) string {
	switch len(kinds) {
	case 0:
		return "a declared element"
	case 1:
		return "a " + kinds[0]
	}
	parts := make([]string, 0, len(kinds))
	for _, k := range kinds {
		parts = append(parts, "a "+k)
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " or " + parts[len(parts)-1]
}
