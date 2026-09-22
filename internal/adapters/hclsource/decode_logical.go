package hclsource

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// decodeElement handles person, system, container, component, and external
// blocks, along with the uses blocks nested inside them (FR-006, FR-008).
func (p *parser) decodeElement(block *hcl.Block, model *arch.SourceModel) arch.Diagnostics {
	kind := arch.ElementKind(block.Type)

	name, diags := p.label(block, 0)
	if diags.HasErrors() {
		return diags
	}

	legal := elementAttrs(kind)
	content, attrDiags := p.attrs(block.Body, legal, "a "+block.Type+" block",
		hcl.BlockHeaderSchema{Type: blockUses, LabelNames: []string{"name"}})
	diags = append(diags, attrDiags...)

	decl := arch.ElementDecl{
		Kind:       kind,
		Name:       name,
		Range:      p.conv.rng(&block.DefRange),
		AttrRanges: map[string]arch.SourceRange{},
	}

	for attrName, attr := range content.Attributes {
		decl.AttrRanges[attrName] = p.conv.rng(attr.Expr.Range().Ptr())
	}

	// The parent reference is named after the parent's kind: a container
	// declares `system = system.payments`, a component declares
	// `container = container.api` (FR-010).
	switch kind {
	case arch.KindContainer:
		diags = append(diags, p.parentRef(content, string(arch.KindSystem), &decl)...)
	case arch.KindComponent:
		diags = append(diags, p.parentRef(content, string(arch.KindContainer), &decl)...)
	}

	if attr, ok := content.Attributes["description"]; ok {
		s, d := p.evalString(attr)
		diags, decl.Description = append(diags, d...), s
	}
	if attr, ok := content.Attributes["owner"]; ok {
		s, d := p.evalString(attr)
		diags, decl.Owner = append(diags, d...), s
	}
	if attr, ok := content.Attributes["technology"]; ok {
		s, d := p.evalString(attr)
		diags, decl.Technology = append(diags, d...), s
	}
	if attr, ok := content.Attributes["docs"]; ok {
		s, d := p.evalString(attr)
		diags, decl.Docs = append(diags, d...), s
	}
	if attr, ok := content.Attributes["tags"]; ok {
		tags, d := p.evalStringList(attr)
		diags, decl.Tags = append(diags, d...), tags
	}

	for _, nested := range content.Blocks {
		if nested.Type != blockUses {
			diags = append(diags, p.unknownBlockDiag(nested.Type, nested.DefRange, "a "+block.Type+" block"))
			continue
		}
		rel, d := p.decodeUses(nested)
		diags = append(diags, d...)
		decl.Relations = append(decl.Relations, rel)
	}

	model.Elements = append(model.Elements, decl)
	return diags
}

// parentRef extracts the containment reference for a container or component.
// Absence is not reported here: whether the parent is required is a validation
// rule, and validation belongs to core.
func (p *parser) parentRef(content *hcl.BodyContent, attrName string, decl *arch.ElementDecl) arch.Diagnostics {
	attr, ok := content.Attributes[attrName]
	if !ok {
		return nil
	}
	ref, diags := p.reference(attr)
	decl.Parent = ref
	return diags
}

// decodeUses handles a relationship block. The label is its local name, which
// is what gives the edge a stable address and lets the diff stage report a
// retarget as a rewire rather than one edge vanishing and another appearing.
func (p *parser) decodeUses(block *hcl.Block) (arch.RelationDecl, arch.Diagnostics) {
	name, diags := p.label(block, 0)
	rel := arch.RelationDecl{LocalName: name, Range: p.conv.rng(&block.DefRange)}
	if diags.HasErrors() {
		return rel, diags
	}

	content, attrDiags := p.attrs(block.Body, usesAttrs, "a uses block")
	diags = append(diags, attrDiags...)

	if attr, ok := content.Attributes["target"]; ok {
		ref, d := p.reference(attr)
		diags, rel.Target = append(diags, d...), ref
	}
	if attr, ok := content.Attributes["description"]; ok {
		s, d := p.evalString(attr)
		diags, rel.Description = append(diags, d...), s
	}
	if attr, ok := content.Attributes["technology"]; ok {
		s, d := p.evalString(attr)
		diags, rel.Technology = append(diags, d...), s
	}

	return rel, diags
}
