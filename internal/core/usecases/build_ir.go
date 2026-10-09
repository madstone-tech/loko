package usecases

import (
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// BuildIR turns a resolved SourceModel into the immutable compiled value.
//
// Ordering is assigned HERE, once, and every slice is sorted byte-wise on its
// address before the IR is constructed. Encoders then serialise in slice order
// and sort nothing themselves. That is what makes determinism structural
// rather than a discipline each backend has to remember (FR-040, research R6).
//
// Elements whose references did not resolve are still included, with the
// unresolved field left empty. The diagnostics already say what is broken;
// dropping the element too would produce a second, confusing wave of "not
// found" errors from anything consuming the IR.
func BuildIR(model *arch.SourceModel, res *Resolved) *arch.IR {
	elements, relationships := buildLogical(model, res)
	environments := buildEnvironments(model, res)
	views := buildViews(model, res)

	ir := arch.NewIR(
		arch.Project{
			Name:        model.Project.Name,
			Description: model.Project.Description,
			LokoVersion: model.Project.Version,
		},
		elements,
		relationships,
		environments,
		views,
		buildIgnores(model),
	)
	ir.Moves = buildMoves(model)
	return ir
}

// buildLogical produces the sorted element and relationship slices.
func buildLogical(model *arch.SourceModel, res *Resolved) ([]arch.Element, []arch.Relationship) {
	elements := make([]arch.Element, 0, len(model.Elements))
	var relationships []arch.Relationship

	for _, d := range model.Elements {
		addr := d.Address()
		elements = append(elements, arch.Element{
			Address:     addr,
			Kind:        d.Kind,
			Name:        d.Name,
			Parent:      res.Parent[addr],
			Description: d.Description,
			Owner:       d.Owner,
			Technology:  d.Technology,
			Tags:        sortedUnique(d.Tags),
			Docs:        d.Docs,
			Title:       d.Title,
			Shape:       d.Shape,
			Range:       d.Range,
		})

		for _, rel := range d.Relations {
			relAddr := arch.NewRelationshipAddress(addr, rel.LocalName)
			relationships = append(relationships, arch.Relationship{
				Address:     relAddr,
				Source:      addr,
				Target:      res.Target[relAddr],
				LocalName:   rel.LocalName,
				Description: rel.Description,
				Technology:  rel.Technology,
				Kind:        storedKind(rel.Kind),
				Tags:        sortedUnique(rel.Tags),
				Range:       rel.Range,
			})
		}
	}

	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Address.Compare(elements[j].Address) < 0
	})
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].Address.Compare(relationships[j].Address) < 0
	})
	return elements, relationships
}

// buildViews produces the sorted view slice with resolved members.
func buildViews(model *arch.SourceModel, res *Resolved) []arch.View {
	views := make([]arch.View, 0, len(model.Views))
	for _, v := range model.Views {
		addr := arch.NewViewAddress(v.Name)
		views = append(views, arch.View{
			Address:   addr,
			Name:      v.Name,
			Include:   sortedAddresses(res.ViewInclude[addr]),
			Exclude:   sortedAddresses(res.ViewExclude[addr]),
			Tags:      sortedUnique(v.Tags),
			Direction: v.Direction,
		})
	}
	sort.Slice(views, func(i, j int) bool {
		return views[i].Address.Compare(views[j].Address) < 0
	})
	return views
}

// buildIgnores returns the reconcile patterns, sorted and de-duplicated. They
// are carried through untouched for the reconciliation stage (FR-015).
func buildIgnores(model *arch.SourceModel) []string {
	patterns := make([]string, 0, len(model.Ignores))
	for _, ig := range model.Ignores {
		patterns = append(patterns, ig.Pattern)
	}
	return sortedUnique(patterns)
}

// sortedUnique returns a sorted, de-duplicated copy. It returns nil for an
// empty input so the field is omitted from the export rather than emitted as
// an empty array — two encoders disagreeing on [] versus absent would be a
// byte-stability bug.
func sortedUnique(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func sortedAddresses(in []arch.Address) []arch.Address {
	if len(in) == 0 {
		return nil
	}
	out := make([]arch.Address, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool { return out[i].Compare(out[j]) < 0 })
	return out
}

// storedKind normalises a relationship kind for the IR: sync is the default,
// so it is never stored and exports omit it.
func storedKind(k string) string {
	if k == arch.RelSync {
		return ""
	}
	return k
}
