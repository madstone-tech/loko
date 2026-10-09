package hclsource

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// add creates a declaration: top-level ones in the file research R8 picks,
// nested ones at the end of their parent's body.
func (w *workspace) add(e authoring.Edit) error {
	if _, taken := w.locate(e.Target, authoring.CanonicalAddress(e.Target, e.Address), e.Binding); taken && e.Target != authoring.TargetBinding {
		return refuse(authoring.ReasonAddressInUse, "%s is already declared", authoring.CanonicalAddress(e.Target, e.Address))
	}
	p, _ := authoring.SplitAddress(e.Target, e.Address)
	switch e.Target {
	case authoring.TargetElement:
		kind, name := splitElement(e.Address)
		return w.addTop(e, kind, name)
	case authoring.TargetEnvironment:
		return w.addTop(e, blockDeployment, p.Env)
	case authoring.TargetRelationship:
		kind, name := splitElement(p.Element)
		parent, ok := w.findTop(kind, name)
		return w.addNested(e, parent, ok, p.Element, blockUses, p.Local)
	case authoring.TargetGroup:
		parent, ok := w.group(p.Env, p.Groups[:len(p.Groups)-1])
		return w.addNested(e, parent, ok, groupAddress(p.Env, p.Groups[:len(p.Groups)-1]), blockNode, p.Groups[len(p.Groups)-1])
	case authoring.TargetInstance:
		parent, ok := w.group(p.Env, p.Groups)
		return w.addNested(e, parent, ok, groupAddress(p.Env, p.Groups), blockInstance, p.Local)
	case authoring.TargetBinding:
		parent, ok := w.instance(p.Env, p.Local)
		return w.addNested(e, parent, ok, e.Address, blockBinding, e.Binding.Kind)
	}
	return refuse(authoring.ReasonInvalidEdit, "cannot add a %s", e.Target)
}

func groupAddress(env string, path []string) string {
	return strings.Join(append([]string{"deployment", env}, nodePath(path)...), ".")
}

func nodePath(path []string) []string {
	if len(path) == 0 {
		return nil
	}
	return append([]string{"node"}, path...)
}

func (w *workspace) addTop(e authoring.Edit, typ, label string) error {
	path, err := w.placement(e)
	if err != nil {
		return err
	}
	var f *hclwrite.File
	if w.exists(path) {
		if f, err = w.file(path); err != nil {
			return refuse(authoring.ReasonCompileErrors, "%v", err)
		}
	} else {
		f = w.create(path)
	}
	blk, err := newBlock(0, typ, []string{label}, e.Set)
	if err != nil {
		return err
	}
	body := f.Body()
	if hasContent(body) {
		if !endsWithNewline(body) {
			body.AppendNewline()
		}
		body.AppendNewline()
	}
	body.AppendBlock(blk)
	w.store(path, f, false)
	return nil
}

// placement picks the file for a new top-level declaration (research R8): an
// explicit file; else beside the parent a container or component names; else
// the file declaring the project.
func (w *workspace) placement(e authoring.Edit) (string, error) {
	if e.File != "" {
		if err := checkPath(w.root, e.File); err != nil {
			return "", err
		}
		return e.File, nil
	}
	for _, a := range e.Set {
		if a.Name == "system" || a.Name == "container" {
			kind, name := splitElement(a.Value.Ref)
			if parent, ok := w.findTop(kind, name); ok {
				return parent.path, nil
			}
		}
	}
	if proj, ok := w.findTop(blockProject, ""); ok {
		return proj.path, nil
	}
	if len(w.paths) > 0 {
		return w.paths[0], nil
	}
	return "main.loko.hcl", nil
}

func (w *workspace) addNested(e authoring.Edit, parent located, found bool, parentAddr, typ, label string) error {
	if !found {
		return w.notFound(parentAddr)
	}
	expand(parent.block)
	blk, err := newBlock(parent.depth+1, typ, []string{label}, e.Set)
	if err != nil {
		return err
	}
	body := parent.block.Body()
	if !endsWithNewline(body) {
		body.AppendNewline()
	}
	if hasContent(body) {
		body.AppendNewline()
	}
	body.AppendBlock(blk)
	w.store(parent.path, parent.file, false)
	return nil
}

// update sets and clears attributes of one declaration, changing only the
// attribute lines involved.
func (w *workspace) update(e authoring.Edit) error {
	at, ok := w.locate(e.Target, e.Address, e.Binding)
	if !ok {
		return w.notFound(e.Address)
	}
	for _, a := range e.Set {
		if err := setAttr(at, a); err != nil {
			return err
		}
	}
	for _, c := range e.Clear {
		at.block.Body().RemoveAttribute(c)
	}
	w.store(at.path, at.file, false)
	return nil
}

func setAttr(at located, a authoring.Attr) error {
	body := at.block.Body()
	toks, err := exprTokens(at.depth+1, a.Value)
	if err != nil {
		return err
	}
	if existing := body.GetAttribute(a.Name); existing != nil {
		if sameTokens(existing.Expr().BuildTokens(nil), toks) {
			return nil // the same value, however it is spaced: a no-op (FR-021)
		}
		// Keep the author's spacing after "=" as well as before it.
		if old := existing.Expr().BuildTokens(nil); len(old) > 0 && len(toks) > 0 {
			toks[0].SpacesBefore = old[0].SpacesBefore
		}
		body.SetAttributeRaw(a.Name, toks)
		return nil
	}
	expand(at.block)
	indent := bodyIndent(at.block)
	if !endsWithNewline(body) {
		body.AppendNewline()
	}
	attr := body.SetAttributeRaw(a.Name, toks)
	if line := attr.BuildTokens(nil); len(line) > 1 {
		line[0].SpacesBefore, line[1].SpacesBefore = indent, 1
	}
	return nil
}

// remove deletes one declaration with its attached comments.
func (w *workspace) remove(e authoring.Edit) error {
	if e.Target == authoring.TargetViewEntry {
		return w.dropViewEntry(e)
	}
	at, ok := w.locate(e.Target, e.Address, e.Binding)
	if !ok {
		return w.notFound(e.Address)
	}
	dropAttachedComments(at.parent, at.block)
	at.parent.RemoveBlock(at.block)
	w.store(at.path, at.file, true)
	if e.Target == authoring.TargetElement {
		return w.dropMovesTo(e.Address)
	}
	return nil
}

// dropViewEntry removes one element from a view's include and exclude lists;
// a cascading removal uses it so the view does not dangle.
func (w *workspace) dropViewEntry(e authoring.Edit) error {
	at, ok := w.locate(e.Target, e.Address, e.Binding)
	if !ok {
		return w.notFound(e.Address)
	}
	for _, list := range []string{"include", "exclude"} {
		attr := at.block.Body().GetAttribute(list)
		if attr == nil {
			continue
		}
		var keep []hclwrite.Tokens
		for _, tr := range attr.Expr().Variables() {
			if t := tr.BuildTokens(nil); strings.TrimSpace(string(t.Bytes())) != e.Entry {
				keep = append(keep, t)
			}
		}
		toks, err := tupleTokens(at.depth+1, keep)
		if err != nil {
			return err
		}
		at.block.Body().SetAttributeRaw(list, toks)
	}
	w.store(at.path, at.file, false)
	return nil
}

func tupleTokens(depth int, elems []hclwrite.Tokens) (hclwrite.Tokens, error) {
	body, err := canonical(depth, "a = "+string(hclwrite.TokensForTuple(elems).Bytes())+"\n")
	if err != nil {
		return nil, err
	}
	return body.GetAttribute("a").Expr().BuildTokens(nil), nil
}
