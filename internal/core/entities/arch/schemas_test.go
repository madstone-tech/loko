package arch

import (
	"slices"
	"strings"
	"testing"
)

func TestLanguageTableIsWellFormed(t *testing.T) {
	t.Parallel()
	types := []string{AttrString, AttrStringList, AttrRef, AttrRefList, AttrMap, AttrBool}
	if !slices.IsSortedFunc(Language, func(a, b LanguageBlock) int { return strings.Compare(a.Name, b.Name) }) {
		t.Error("Language is not sorted by block name")
	}
	names := map[string]bool{}
	for _, b := range Language {
		if names[b.Name] {
			t.Errorf("block %q appears twice", b.Name)
		}
		names[b.Name] = true
	}
	for _, b := range Language {
		for _, in := range b.In {
			if in != InFile && !names[in] {
				t.Errorf("block %q nests in unknown block %q", b.Name, in)
			}
		}
		if b.Example == "" {
			t.Errorf("block %q has no example", b.Name)
		}
		if !slices.IsSortedFunc(b.Attrs, func(x, y LanguageAttr) int { return strings.Compare(x.Name, y.Name) }) {
			t.Errorf("block %q attributes are not sorted", b.Name)
		}
		seen := map[string]bool{}
		for _, a := range b.Attrs {
			if seen[a.Name] {
				t.Errorf("%s.%s appears twice", b.Name, a.Name)
			}
			seen[a.Name] = true
			if !slices.Contains(types, a.Type) {
				t.Errorf("%s.%s has unknown type %q", b.Name, a.Name, a.Type)
			}
		}
	}
}

func TestLanguageValueSets(t *testing.T) {
	t.Parallel()
	attr := func(block, name string) LanguageAttr {
		t.Helper()
		b, ok := LanguageBlockNamed(block)
		if !ok {
			t.Fatalf("no block %q", block)
		}
		i := slices.IndexFunc(b.Attrs, func(a LanguageAttr) bool { return a.Name == name })
		if i < 0 {
			t.Fatalf("no attribute %s.%s", block, name)
		}
		return b.Attrs[i]
	}
	for _, c := range []struct {
		block, name string
		want        []string
	}{
		{"project", "layout", Layouts},
		{"view", "layout", Layouts},
		{"view", "direction", Directions},
		{"uses", "kind", RelationshipKinds},
		{"container", "shape", Shapes},
		{"person", "shape", Shapes},
	} {
		if got := attr(c.block, c.name).Values; !slices.Equal(got, c.want) {
			t.Errorf("%s.%s values = %v, want %v", c.block, c.name, got, c.want)
		}
	}
	if got := attr("system", "shape").Kinds; !slices.Equal(got, []string{"container", "external"}) {
		t.Errorf("shape kinds = %v, want container, external", got)
	}
	for _, req := range [][2]string{{"uses", "target"}, {"instance", "of"}, {"container", "system"},
		{"component", "container"}, {"moved", "from"}, {"moved", "to"}} {
		if !attr(req[0], req[1]).Required {
			t.Errorf("%s.%s is not required", req[0], req[1])
		}
	}
	if note := attr("system", "docs").Note; !strings.Contains(note, "relative to the project root") {
		t.Errorf("docs note %q does not state the path rule", note)
	}
}

func TestLanguageLabelValues(t *testing.T) {
	t.Parallel()
	for _, b := range Language {
		if b.Name == "binding" {
			want := make([]string, len(AllClaimKinds))
			for i, k := range AllClaimKinds {
				want[i] = string(k)
			}
			if !slices.Equal(b.LabelValues, want) {
				t.Errorf("binding label values = %v, want %v", b.LabelValues, want)
			}
			continue
		}
		if len(b.LabelValues) != 0 {
			t.Errorf("block %q has label values %v", b.Name, b.LabelValues)
		}
	}
}

func TestLanguageBlockNamed(t *testing.T) {
	t.Parallel()
	if b, ok := LanguageBlockNamed("view"); !ok || b.Name != "view" {
		t.Error("view not found")
	}
	if _, ok := LanguageBlockNamed("sequence"); ok {
		t.Error("sequence found")
	}
}
