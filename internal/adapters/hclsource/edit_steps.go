package hclsource

import (
	"slices"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// addStep adds a step to its view (feature 019, R9): appended, or inserted
// before the step e.Before names. Order is a flow's meaning, so the insertion
// point matters; every other byte of the file is kept.
func (w *workspace) addStep(e authoring.Edit, p authoring.AddressParts) error {
	view, ok := w.findTop(blockView, p.View)
	if e.Before == "" {
		return w.addNested(e, view, ok, "view."+p.View, blockStep, p.Local)
	}
	if !ok {
		return w.notFound("view." + p.View)
	}
	next, ok := view.child(blockStep, e.Before)
	if !ok {
		return w.notFound("view." + p.View + ".step." + e.Before)
	}
	blk, err := newBlock(next.depth, blockStep, []string{p.Local}, e.Set)
	if err != nil {
		return err
	}
	toks := next.file.BuildTokens(nil)
	at := declStart(toks, slices.Index(toks, next.block.BuildTokens(nil)[0]))
	head := hclwrite.Tokens(toks[:at]).Bytes()
	src := slices.Concat(head, blk.BuildTokens(nil).Bytes(), []byte("\n"), tokensOf(next.file)[len(head):])
	w.cur[next.path], w.touched[next.path] = src, true
	delete(w.parsed, next.path)
	return nil
}

// declStart is the index of the first token of the declaration whose block
// starts at toks[i]: the block, or the comment lines directly above it.
func declStart(toks hclwrite.Tokens, i int) int {
	for i > 0 && toks[i-1].Type == hclsyntax.TokenComment && lineStart(toks, i-1) {
		i--
	}
	return i
}
