package usecases

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ViewRow is one resolved view in a views listing (feature 023, contract Q-1).
type ViewRow struct {
	Name  string `json:"name" toon:"name"`
	Kind  string `json:"kind" toon:"kind"`
	Title string `json:"title" toon:"title"`
}

// RelationshipRow is one relationship in a relationships listing. Kind is
// sync when the source declares none.
type RelationshipRow struct {
	Address     string `json:"address" toon:"address"`
	Source      string `json:"source" toon:"source"`
	Target      string `json:"target" toon:"target"`
	Kind        string `json:"kind" toon:"kind"`
	Description string `json:"description,omitempty" toon:"description,omitempty"`
}

// listingElementKinds are the values the elements listing filters by.
var listingElementKinds = []arch.ElementKind{
	arch.KindPerson, arch.KindSystem, arch.KindContainer, arch.KindComponent, arch.KindExternal,
}

// checkElementKind refuses an element kind filter that names no logical kind.
func checkElementKind(kind string) *ReadError {
	if kind == "" || slices.Contains(listingElementKinds, arch.ElementKind(kind)) {
		return nil
	}
	names := make([]string, len(listingElementKinds))
	for i, k := range listingElementKinds {
		names[i] = string(k)
	}
	return &ReadError{Reason: "invalid_query", Detail: fmt.Sprintf(
		"element kind %q is not one of %s", kind, strings.Join(names, ", "))}
}

// listViews lists every view the build produces, generated and declared,
// sorted by name. A declared view replaces the generated view it shadows.
func listViews(ir *arch.IR) []ViewRow {
	views := ResolveViews(ir, Provenance{})
	out := make([]ViewRow, 0, len(views))
	for _, v := range views {
		out = append(out, ViewRow{Name: string(v.ID), Kind: string(v.Kind), Title: v.Title})
	}
	slices.SortFunc(out, func(a, b ViewRow) int { return cmp.Compare(a.Name, b.Name) })
	return out
}

// listElements lists every element of the given kind, or of every kind when
// kind is empty, sorted by address.
func listElements(ir *arch.IR, kind string) []ElementHit {
	out := []ElementHit{}
	for _, e := range ir.Elements {
		if kind == "" || string(e.Kind) == kind {
			out = append(out, ElementHit{Address: string(e.Address), Kind: string(e.Kind), Name: e.Name, Title: e.Title})
		}
	}
	slices.SortFunc(out, func(a, b ElementHit) int { return cmp.Compare(a.Address, b.Address) })
	return out
}

// listRelationships lists every relationship, sorted by address.
func listRelationships(ir *arch.IR) []RelationshipRow {
	out := make([]RelationshipRow, 0, len(ir.Relationships))
	for _, r := range ir.Relationships {
		out = append(out, RelationshipRow{
			Address: string(r.Address), Source: string(r.Source), Target: string(r.Target),
			Kind: cmp.Or(r.Kind, arch.RelSync), Description: r.Description,
		})
	}
	slices.SortFunc(out, func(a, b RelationshipRow) int { return cmp.Compare(a.Address, b.Address) })
	return out
}
