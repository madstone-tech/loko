package usecases

import (
	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// LanguageResult describes the architecture language for a client that is
// about to edit it (feature 017, FR-001..FR-003). Addresses and Docs are set
// only when every block is asked for.
type LanguageResult struct {
	OK        bool                `json:"ok" toon:"ok"`
	Blocks    []LanguageBlockView `json:"blocks,omitempty" toon:"blocks,omitempty"`
	Addresses []AddressFormView   `json:"addresses,omitempty" toon:"addresses,omitempty"`
	Docs      string              `json:"docs,omitempty" toon:"docs,omitempty"`
	Error     *ReadError          `json:"error,omitempty" toon:"error,omitempty"`
}

// LanguageBlockView is one block of the language.
type LanguageBlockView struct {
	Name        string             `json:"name" toon:"name"`
	Labels      []string           `json:"labels,omitempty" toon:"labels,omitempty"`
	LabelValues []string           `json:"label_values,omitempty" toon:"label_values,omitempty"`
	In          []string           `json:"in" toon:"in"`
	Attributes  []LanguageAttrView `json:"attributes,omitempty" toon:"attributes,omitempty"`
	Example     string             `json:"example" toon:"example"`
}

// LanguageAttrView is one attribute of a block.
type LanguageAttrView struct {
	Name     string   `json:"name" toon:"name"`
	Type     string   `json:"type" toon:"type"`
	Required bool     `json:"required,omitempty" toon:"required,omitempty"`
	Values   []string `json:"values,omitempty" toon:"values,omitempty"`
	Kinds    []string `json:"kinds,omitempty" toon:"kinds,omitempty"`
	Note     string   `json:"note,omitempty" toon:"note,omitempty"`
}

// AddressFormView is how one apply_edit target is addressed.
type AddressFormView struct {
	Target  string `json:"target" toon:"target"`
	Pattern string `json:"pattern" toon:"pattern"`
	Example string `json:"example" toon:"example"`
	Note    string `json:"note,omitempty" toon:"note,omitempty"`
}

// docsRule is the one path rule a reader of a nested source file does not
// expect, so the description states it outright.
const docsRule = "docs paths are relative to the project root, not to the file that declares them"

// DescribeLanguage returns every block, or only the named one. An unknown
// name is refused with the names that exist.
func DescribeLanguage(block string) *LanguageResult {
	if block == "" {
		out := &LanguageResult{OK: true, Docs: docsRule}
		for _, b := range arch.Language {
			out.Blocks = append(out.Blocks, blockView(b))
		}
		for _, f := range authoring.AddressForms {
			out.Addresses = append(out.Addresses,
				AddressFormView{Target: string(f.Target), Pattern: f.Pattern, Example: f.Example, Note: f.Note})
		}
		return out
	}
	b, ok := arch.LanguageBlockNamed(block)
	if !ok {
		names := make([]string, len(arch.Language))
		for i, b := range arch.Language {
			names[i] = b.Name
		}
		return &LanguageResult{Error: &ReadError{Reason: "unknown_block",
			Detail: "no block named " + block, Suggestions: names}}
	}
	return &LanguageResult{OK: true, Blocks: []LanguageBlockView{blockView(b)}}
}

func blockView(b arch.LanguageBlock) LanguageBlockView {
	v := LanguageBlockView{Name: b.Name, Labels: b.Labels, LabelValues: b.LabelValues, In: b.In, Example: b.Example}
	for _, a := range b.Attrs {
		v.Attributes = append(v.Attributes, LanguageAttrView{Name: a.Name, Type: a.Type, Required: a.Required,
			Values: a.Values, Kinds: a.Kinds, Note: a.Note})
	}
	return v
}
