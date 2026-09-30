package usecases

import (
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// checkCollisions reports every pair of artifacts whose paths are equal, or
// equal ignoring case (FR-028). On a case-insensitive file system one would
// silently overwrite the other; checking planned paths rather than the disk
// catches it identically on every platform (research R9).
//
// artifacts must be sorted by path.
func checkCollisions(artifacts []viewmodel.Artifact, prov Provenance) arch.Diagnostics {
	groups := map[string][]int{}
	var keys []string
	for i, a := range artifacts {
		k := strings.ToLower(a.Path)
		if _, ok := groups[k]; !ok {
			keys = append(keys, k)
		}
		groups[k] = append(groups[k], i)
	}
	var diags arch.Diagnostics
	for _, k := range keys {
		idx := groups[k]
		for i := range idx {
			for j := i + 1; j < len(idx); j++ {
				diags = append(diags, collision(artifacts[idx[i]], artifacts[idx[j]], prov))
			}
		}
	}
	return diags
}

func collision(a, b viewmodel.Artifact, prov Provenance) arch.Diagnostic {
	name := func(x viewmodel.Artifact) string {
		if x.Owner == "" {
			return "the site index"
		}
		return x.Owner
	}
	d := arch.Diagnostic{
		Severity: arch.SeverityError,
		Code:     arch.CodeOutputPathCollision,
		Summary:  "Output files would collide",
		Detail: fmt.Sprintf("%s and %s would both be written as %s: the file names differ only by letter "+
			"case, so one would overwrite the other on a case-insensitive file system. Rename one of them.",
			name(a), name(b), b.Path),
		Address: arch.Address(b.Owner),
	}
	if r, ok := prov.RangeOf(arch.Address(b.Owner)); ok {
		d.Range = r
	}
	if r, ok := prov.RangeOf(arch.Address(a.Owner)); ok {
		d.Related = []arch.RelatedRange{{Message: "the other declaration", Range: r}}
	}
	return d
}
