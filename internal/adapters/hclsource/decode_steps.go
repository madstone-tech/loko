package hclsource

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// decodeSteps reads a view's step blocks in declaration order: a step's
// position is its number (feature 019). References come back raw; core
// resolves and matches them.
func (p *parser) decodeSteps(blocks hcl.Blocks, decl *arch.ViewDecl) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, nested := range blocks {
		if nested.Type != blockStep {
			diags = append(diags, p.unknownBlockDiag(nested.Type, nested.DefRange, "a view block"))
			continue
		}
		s, d := p.decodeStep(nested)
		diags = append(diags, d...)
		decl.Steps = append(decl.Steps, s)
	}
	return diags
}

func (p *parser) decodeStep(block *hcl.Block) (arch.StepDecl, arch.Diagnostics) {
	name, diags := p.label(block, 0)
	s := arch.StepDecl{Name: name, Range: p.conv.rng(&block.DefRange)}
	if diags.HasErrors() {
		return s, diags
	}
	content, attrDiags := p.attrs(block.Body, stepAttrs, "a step block")
	diags = append(diags, attrDiags...)
	for _, r := range []struct {
		name string
		into *arch.Reference
	}{{"from", &s.From}, {"to", &s.To}, {"via", &s.Via}} {
		if attr, ok := content.Attributes[r.name]; ok {
			ref, d := p.reference(attr)
			diags, *r.into = append(diags, d...), ref
		}
	}
	if attr, ok := content.Attributes["description"]; ok {
		v, d := p.evalString(attr)
		diags, s.Description = append(diags, d...), v
	}
	if attr, ok := content.Attributes["reply"]; ok {
		b, d := p.evalBool(attr)
		diags, s.Reply = append(diags, d...), b
	}
	return s, diags
}
