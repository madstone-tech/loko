package usecases

import (
	"maps"
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// attrTypeOf maps the edit schema's value shapes onto the language table's
// attribute types.
var attrTypeOf = map[authoring.ValueKind]string{
	authoring.ValueString:  arch.AttrString,
	authoring.ValueList:    arch.AttrStringList,
	authoring.ValueRef:     arch.AttrRef,
	authoring.ValueRefList: arch.AttrRefList,
	authoring.ValueMap:     arch.AttrMap,
	authoring.ValueBool:    arch.AttrBool,
}

// TestLanguageParity keeps the edit schema (authoring), a copy that exists
// because entity packages may not import one another, in step with the
// language table (arch.Language): the same attributes, value types and
// allowed values for every edit target. Blocks with no edit target
// (reconcile, moved, locals) are not compared.
func TestLanguageParity(t *testing.T) {
	t.Parallel()
	type pair struct {
		target  authoring.TargetKind
		address string // an address of the target, for element kinds
		block   string
	}
	pairs := []pair{
		{authoring.TargetRelationship, "", "uses"},
		{authoring.TargetEnvironment, "", "deployment"},
		{authoring.TargetGroup, "", "node"},
		{authoring.TargetInstance, "", "instance"},
		{authoring.TargetBinding, "", "binding"},
		{authoring.TargetView, "", "view"},
	}
	for _, kind := range authoring.ElementKinds {
		pairs = append(pairs, pair{authoring.TargetElement, kind + ".x", kind})
	}
	for _, p := range pairs {
		b, ok := arch.LanguageBlockNamed(p.block)
		if !ok {
			t.Fatalf("no block %q", p.block)
		}
		want := map[string]string{}
		for _, a := range b.Attrs {
			if p.target == authoring.TargetElement && len(a.Kinds) > 0 && !slices.Contains(a.Kinds, p.block) {
				continue
			}
			want[a.Name] = a.Type
			if got := authoring.AllowedValues(p.target, a.Name); !slices.Equal(got, a.Values) {
				t.Errorf("%s %s.%s: edit schema allows %v, language table %v", p.target, p.block, a.Name, got, a.Values)
			}
		}
		got := map[string]string{}
		for name, vk := range authoring.LegalAttrs(p.target, p.address) {
			got[name] = attrTypeOf[vk]
		}
		if !maps.Equal(got, want) {
			t.Errorf("%s (%s): edit schema %v, language table %v", p.target, p.block, got, want)
		}
	}
}
