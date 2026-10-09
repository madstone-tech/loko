package hclsource

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// Edits are made through the hclwrite tree, but a file is always serialised
// from its raw token stream (tokensOf), never with File.Bytes(): that runs the
// formatter over the whole file and would realign lines nobody asked to change
// (FR-013). Tokens an edit creates carry no layout of their own, so new code is
// first written and formatted on its own, at its nesting depth, then parsed
// back: the parsed tokens carry canonical spacing (FR-014).

// tokensOf serialises f exactly as its tokens stand.
func tokensOf(f *hclwrite.File) []byte { return f.BuildTokens(nil).Bytes() }

// parseLossless parses src for editing, refusing any file whose token stream
// would not reproduce it byte for byte (research R1).
func parseLossless(src []byte, rel string) (*hclwrite.File, error) {
	f, diags := hclwrite.ParseConfig(src, rel, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("%s does not parse: %s", rel, diags.Error())
	}
	if !bytes.Equal(tokensOf(f), src) {
		return nil, fmt.Errorf("%s cannot be edited without changing bytes it does not own", rel)
	}
	return f, nil
}

// canonical formats text nested depth blocks deep and returns the innermost
// body of the parsed result, so whatever it holds has canonical layout for
// that depth.
func canonical(depth int, inner string) (*hclwrite.Body, error) {
	text := strings.Repeat("w {\n", depth) + inner + strings.Repeat("}\n", depth)
	f, diags := hclwrite.ParseConfig(hclwrite.Format([]byte(text)), "edit", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("generated source does not parse: %s", diags.Error())
	}
	body := f.Body()
	for range depth {
		body = body.Blocks()[0].Body()
	}
	return body, nil
}

// valueTokens renders an attribute value. A reference is a bare traversal,
// never a quoted string.
func valueTokens(v authoring.AttrValue) hclwrite.Tokens {
	switch v.Kind {
	case authoring.ValueRef:
		kind, name, _ := strings.Cut(v.Ref, ".")
		return hclwrite.TokensForTraversal(hcl.Traversal{hcl.TraverseRoot{Name: kind}, hcl.TraverseAttr{Name: name}})
	case authoring.ValueList:
		elems := make([]hclwrite.Tokens, 0, len(v.List))
		for _, s := range v.List {
			elems = append(elems, hclwrite.TokensForValue(cty.StringVal(s)))
		}
		return hclwrite.TokensForTuple(elems)
	case authoring.ValueMap:
		attrs := make([]hclwrite.ObjectAttrTokens, 0, len(v.Map))
		for _, m := range v.Map {
			attrs = append(attrs, hclwrite.ObjectAttrTokens{Name: hclwrite.TokensForIdentifier(m.Key), Value: valueTokens(m.Value)})
		}
		return hclwrite.TokensForObject(attrs)
	case authoring.ValueNumber:
		return hclwrite.TokensForValue(cty.NumberFloatVal(v.Num))
	case authoring.ValueBool:
		return hclwrite.TokensForValue(cty.BoolVal(v.Bool))
	}
	return hclwrite.TokensForValue(cty.StringVal(v.Str))
}

// exprTokens returns v's tokens laid out canonically for an attribute at depth.
func exprTokens(depth int, v authoring.AttrValue) (hclwrite.Tokens, error) {
	body, err := canonical(depth, "a = "+string(valueTokens(v).Bytes())+"\n")
	if err != nil {
		return nil, err
	}
	return body.GetAttribute("a").Expr().BuildTokens(nil), nil
}

// newBlock builds a canonical block for nesting at depth. Reference
// attributes come first, then the rest in name order.
func newBlock(depth int, typ string, labels []string, attrs []authoring.Attr) (*hclwrite.Block, error) {
	scratch := hclwrite.NewEmptyFile()
	b := scratch.Body().AppendNewBlock(typ, labels)
	for _, refsFirst := range []bool{true, false} {
		for _, a := range attrs {
			if (a.Value.Kind == authoring.ValueRef) == refsFirst {
				b.Body().SetAttributeRaw(a.Name, valueTokens(a.Value))
			}
		}
	}
	body, err := canonical(depth, string(tokensOf(scratch)))
	if err != nil {
		return nil, err
	}
	return body.Blocks()[0], nil
}

// blockIndent is the indentation of a block's header line.
func blockIndent(b *hclwrite.Block) int {
	for _, t := range b.BuildTokens(nil) {
		if t.Type == hclsyntax.TokenIdent {
			return t.SpacesBefore
		}
	}
	return 0
}

// bodyIndent is the indentation for a new line in b's body: that of an
// existing attribute when there is one, otherwise one level in.
func bodyIndent(b *hclwrite.Block) int {
	for _, a := range b.Body().Attributes() {
		for _, t := range a.BuildTokens(nil) {
			if t.Type == hclsyntax.TokenIdent {
				return t.SpacesBefore
			}
		}
	}
	return blockIndent(b) + 2
}

// expand turns a one-line block (`uses "x" { target = y }`) into the
// multi-line form, so a line can be added to it. Only b's own lines change.
func expand(b *hclwrite.Block) {
	toks := b.BuildTokens(nil)
	open, close := -1, -1
	for i, t := range toks {
		switch {
		case t.Type == hclsyntax.TokenOBrace && open < 0:
			open = i
		case t.Type == hclsyntax.TokenCBrace:
			close = i
		case t.Type == hclsyntax.TokenNewline && open >= 0 && close < 0:
			return // already multi-line
		}
	}
	if open < 0 || close < 0 {
		return
	}
	indent := blockIndent(b)
	toks[open].Bytes = []byte("{\n")
	if open+1 < close {
		toks[open+1].SpacesBefore = indent + 2
		b.Body().AppendNewline()
	}
	toks[close].SpacesBefore = indent
}

// endsWithNewline reports whether a body's last token ends a line, so an item
// appended after it starts on its own line.
func endsWithNewline(b *hclwrite.Body) bool {
	toks := b.BuildTokens(nil)
	if len(toks) == 0 {
		return true
	}
	last := toks[len(toks)-1]
	return last.Type == hclsyntax.TokenNewline || bytes.HasSuffix(last.Bytes, []byte("\n"))
}

// hasContent reports whether a body holds anything but blank lines.
func hasContent(b *hclwrite.Body) bool {
	for _, t := range b.BuildTokens(nil) {
		if t.Type != hclsyntax.TokenNewline {
			return true
		}
	}
	return false
}

// closeJunction removes one blank line where a removal left two in a row (or
// one at the very start of the file), so a removed declaration takes one
// adjacent blank line with it (research R1 § Declaration span).
func closeJunction(old, new []byte) []byte {
	i := 0
	for i < len(new) && i < len(old) && old[i] == new[i] {
		i++
	}
	before := bytes.HasSuffix(new[:i], []byte("\n\n")) || i == 0
	after := bytes.HasPrefix(new[i:], []byte("\n"))
	if before && after {
		return append(new[:i:i], new[i+1:]...)
	}
	if i == len(new) && bytes.HasSuffix(new, []byte("\n\n")) {
		return new[:len(new)-1]
	}
	return new
}

// dropAttachedComments clears the comments directly above a block, with no
// blank line between, that hclwrite does not attach to it: it attaches # and
// // comments only, never /* */ ones. Together with the block they form its
// declaration span (research R1), so they go when the block goes.
func dropAttachedComments(parent *hclwrite.Body, b *hclwrite.Block) {
	toks := parent.BuildTokens(nil)
	first := b.BuildTokens(nil)[0]
	i := slices.Index(toks, first)
	for i > 0 {
		prev := toks[i-1]
		switch {
		case prev.Type == hclsyntax.TokenComment && bytes.HasSuffix(prev.Bytes, []byte("\n")) && lineStart(toks, i-1):
			i--
		case prev.Type == hclsyntax.TokenNewline && i > 1 && toks[i-2].Type == hclsyntax.TokenComment &&
			bytes.HasPrefix(toks[i-2].Bytes, []byte("/*")) && lineStart(toks, i-2):
			i -= 2
		default:
			return
		}
		for _, t := range toks[i : i+1+boolInt(prev.Type == hclsyntax.TokenNewline)] {
			t.Bytes, t.SpacesBefore = nil, 0
		}
	}
}

// lineStart reports whether toks[i] begins a line.
func lineStart(toks hclwrite.Tokens, i int) bool {
	if i == 0 {
		return true
	}
	p := toks[i-1]
	return p.Type == hclsyntax.TokenNewline || bytes.HasSuffix(p.Bytes, []byte("\n"))
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// sameTokens reports whether two expressions are the same apart from layout:
// equal tokens, ignoring spacing and line breaks.
func sameTokens(a, b hclwrite.Tokens) bool {
	significant := func(ts hclwrite.Tokens) []string {
		var out []string
		for _, t := range ts {
			if t.Type != hclsyntax.TokenNewline {
				out = append(out, string(t.Type)+string(t.Bytes))
			}
		}
		return out
	}
	return slices.Equal(significant(a), significant(b))
}
