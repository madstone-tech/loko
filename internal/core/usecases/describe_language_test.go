package usecases

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func TestDescribeLanguageAll(t *testing.T) {
	t.Parallel()
	res := DescribeLanguage("")
	if !res.OK || res.Error != nil {
		t.Fatalf("not ok: %+v", res.Error)
	}
	if len(res.Blocks) != len(arch.Language) {
		t.Errorf("%d blocks, want %d", len(res.Blocks), len(arch.Language))
	}
	if len(res.Addresses) == 0 {
		t.Error("no address forms")
	}
	if !strings.Contains(res.Docs, "relative to the project root") {
		t.Errorf("docs rule %q", res.Docs)
	}
	i := slices.IndexFunc(res.Blocks, func(b LanguageBlockView) bool { return b.Name == "binding" })
	if got := res.Blocks[i].LabelValues; !slices.Equal(got, []string{"terraform", "cloudformation"}) {
		t.Errorf("binding label values %v", got)
	}
}

func TestDescribeLanguageOneBlock(t *testing.T) {
	t.Parallel()
	res := DescribeLanguage("view")
	if !res.OK || len(res.Blocks) != 1 || res.Blocks[0].Name != "view" {
		t.Fatalf("got %+v", res)
	}
	if res.Addresses != nil || res.Docs != "" {
		t.Error("a one-block answer carries the address forms or the docs rule")
	}
	var layout LanguageAttrView
	for _, a := range res.Blocks[0].Attributes {
		if a.Name == "layout" {
			layout = a
		}
	}
	if !slices.Equal(layout.Values, []string{"dagre", "elk"}) {
		t.Errorf("view.layout values %v", layout.Values)
	}
}

func TestDescribeLanguageUnknownBlock(t *testing.T) {
	t.Parallel()
	res := DescribeLanguage("sequence")
	if res.OK || res.Error == nil || res.Error.Reason != "unknown_block" {
		t.Fatalf("got %+v", res)
	}
	if !slices.Contains(res.Error.Suggestions, "view") || len(res.Error.Suggestions) != len(arch.Language) {
		t.Errorf("suggestions %v", res.Error.Suggestions)
	}
}

// TestDescribeLanguageIsStable is FR-006: the same answer, byte for byte.
func TestDescribeLanguageIsStable(t *testing.T) {
	t.Parallel()
	first, _ := json.Marshal(DescribeLanguage(""))
	for range 20 {
		next, _ := json.Marshal(DescribeLanguage(""))
		if !bytes.Equal(first, next) {
			t.Fatal("the language description differs between calls")
		}
	}
}
