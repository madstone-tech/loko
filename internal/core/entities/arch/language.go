package arch

import (
	"slices"
	"strings"
)

// LanguageBlock describes one block of the architecture language: what it is
// called, how it is labelled, where it may appear and what it accepts. The
// table of them (schemas.go) is the one definition the decoder, the edit
// schema's parity test and the schema MCP tool all read (feature 017, R1).
type LanguageBlock struct {
	Name string
	// Labels names the block's labels in order; empty for unlabelled blocks.
	Labels []string
	// LabelValues is the closed set the label must come from, when there is one.
	LabelValues []string
	// In lists where the block may appear: InFile for the top level, otherwise
	// the names of the blocks it nests in.
	In    []string
	Attrs []LanguageAttr
	// Example is the smallest valid use of the block, canonically formatted.
	// A nested block's example shows it inside its parent.
	Example string
}

// LanguageAttr describes one attribute of a block.
type LanguageAttr struct {
	Name     string
	Type     string
	Required bool
	// Values is the closed set the value must come from, when there is one.
	Values []string
	// Kinds narrows an element attribute to these element kinds; empty means
	// every kind that has the attribute.
	Kinds []string
	Note  string
}

// InFile marks a block that appears at the top level of a source file.
const InFile = "file"

// Attribute value types, as the schema tool reports them.
const (
	AttrString     = "string"
	AttrStringList = "list(string)"
	AttrRef        = "ref"
	AttrRefList    = "list(ref)"
	AttrMap        = "map"
	AttrBool       = "bool"
)

// LanguageBlockNamed returns the block with the given name.
func LanguageBlockNamed(name string) (LanguageBlock, bool) {
	for _, b := range Language {
		if b.Name == name {
			return b, true
		}
	}
	return LanguageBlock{}, false
}

// AttrNames returns a block's attribute names, sorted as the table is.
func (b LanguageBlock) AttrNames() []string {
	names := make([]string, len(b.Attrs))
	for i, a := range b.Attrs {
		names[i] = a.Name
	}
	return names
}

// elementAttrs returns the common element attributes, plus the parent
// reference a container or component takes, sorted by name.
func elementAttrs(parent ...LanguageAttr) []LanguageAttr {
	out := append(slices.Clone(commonElementAttrs), parent...)
	slices.SortFunc(out, func(a, b LanguageAttr) int { return strings.Compare(a.Name, b.Name) })
	return out
}

func claimKindNames() []string {
	out := make([]string, len(AllClaimKinds))
	for i, k := range AllClaimKinds {
		out[i] = string(k)
	}
	return out
}
