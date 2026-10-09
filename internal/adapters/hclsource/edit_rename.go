package hclsource

import (
	"bytes"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// parentAttr is the parent reference each element kind takes.
var parentAttr = map[string]string{"container": "system", "component": "container"}

// rename moves an element to a new address (FR-022..FR-025): it rewrites the
// declaring block's type and label, rewrites every reference in every file
// (only the traversal tokens change), and records the move in a moved block
// appended to the declaring file. A kind change drops the parent attribute the
// new kind does not take; a batch sets the new parent.
func (w *workspace) rename(e authoring.Edit) error {
	if _, ok := w.locate(authoring.TargetElement, e.Address, authoring.BindingRef{}); !ok {
		return w.notFound(e.Address)
	}
	toKind, toName := splitElement(e.To)
	if taken, ok := w.findTop(toKind, toName); ok {
		return refuse(authoring.ReasonAddressInUse, "%s is already declared in %s", e.To, taken.path)
	}
	if err := w.dropMovesFrom(e.To); err != nil {
		return err
	}
	// Located after dropMovesFrom, which may have re-parsed the file.
	decl, _ := w.locate(authoring.TargetElement, e.Address, authoring.BindingRef{})
	fromKind, fromName := splitElement(e.Address)
	relabel(decl.block, toKind, toName)
	if old, now := parentAttr[fromKind], parentAttr[toKind]; old != "" && old != now {
		decl.block.Body().RemoveAttribute(old)
	}
	if err := w.rewriteReferences([]string{fromKind, fromName}, []string{toKind, toName}, decl.path); err != nil {
		return err
	}
	moved, err := newBlock(0, blockMoved, nil, []authoring.Attr{
		{Name: "from", Value: authoring.AttrValue{Kind: authoring.ValueRef, Ref: e.Address}},
		{Name: "to", Value: authoring.AttrValue{Kind: authoring.ValueRef, Ref: e.To}},
	})
	if err != nil {
		return err
	}
	body := decl.file.Body()
	if !endsWithNewline(body) {
		body.AppendNewline()
	}
	body.AppendNewline()
	body.AppendBlock(moved)
	w.store(decl.path, decl.file, false)
	return nil
}

// rewriteReferences renames every traversal starting with from, in every
// attribute of every block of every file. A file that cannot be parsed might
// hold a reference, so it refuses the rename rather than leave one behind.
func (w *workspace) rewriteReferences(from, to []string, declaring string) error {
	for _, p := range w.paths {
		f, err := w.file(p)
		if err != nil {
			return refuse(authoring.ReasonCompileErrors, "cannot rewrite references: %v", err)
		}
		before := tokensOf(f)
		renameIn(f.Body(), from, to)
		if p != declaring && !bytes.Equal(before, tokensOf(f)) {
			w.store(p, f, false)
		}
	}
	return nil
}

func renameIn(body *hclwrite.Body, from, to []string) {
	for _, a := range body.Attributes() {
		a.Expr().RenameVariablePrefix(from, to)
	}
	for _, b := range body.Blocks() {
		renameIn(b.Body(), from, to)
	}
}

// relabel changes a block's type and label, keeping one space between them:
// tokens hclwrite creates carry no layout of their own.
func relabel(b *hclwrite.Block, typ, label string) {
	b.SetType(typ)
	b.SetLabels([]string{label})
	toks := b.BuildTokens(nil)
	for i, t := range toks {
		if t.Type == hclsyntax.TokenIdent && i+1 < len(toks) {
			toks[i+1].SpacesBefore = 1
			return
		}
	}
}

// dropMovesFrom removes the moved block whose from is addr. Renaming onto an
// address the element once moved away from undoes that move; keeping the
// block would claim an address that is declared again.
func (w *workspace) dropMovesFrom(addr string) error {
	_, _, err := w.dropOneMove("from", addr)
	return err
}

// dropMovesTo removes the history of a removed element: every moved block
// whose to is addr, and, along chains, every block leading to one of those.
// A removed element has no address for its history to point at.
func (w *workspace) dropMovesTo(addr string) error {
	for gone := []string{addr}; len(gone) > 0; {
		from, found, err := w.dropOneMove("to", gone[0])
		if err != nil {
			return err
		}
		if !found {
			gone = gone[1:]
			continue
		}
		gone = append(gone, from)
	}
	return nil
}

// dropOneMove removes the first moved block whose operand equals addr,
// returning that block's from.
func (w *workspace) dropOneMove(operand, addr string) (string, bool, error) {
	for _, p := range w.paths {
		f, err := w.file(p)
		if err != nil {
			return "", false, refuse(authoring.ReasonCompileErrors, "cannot update moved blocks: %v", err)
		}
		for _, b := range f.Body().Blocks() {
			if b.Type() == blockMoved && movedOperand(b, operand) == addr {
				from := movedOperand(b, "from")
				f.Body().RemoveBlock(b)
				w.store(p, f, true)
				return from, true, nil
			}
		}
	}
	return "", false, nil
}

func movedOperand(b *hclwrite.Block, name string) string {
	a := b.Body().GetAttribute(name)
	if a == nil {
		return ""
	}
	return string(bytes.TrimSpace(a.Expr().BuildTokens(nil).Bytes()))
}
