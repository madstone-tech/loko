package hclsource

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// movedSchema: both attributes are required. HCL reports a missing one as
// "Missing required argument", which the converter files under
// unknown_attribute like any other attribute error.
var movedSchema = &hcl.BodySchema{Attributes: []hcl.AttributeSchema{
	{Name: "from", Required: true},
	{Name: "to", Required: true},
}}

// decodeMoved reads a `moved { from = …  to = … }` block. Both operands are
// static traversals: from names an address that no longer exists, which is
// why references are extracted rather than evaluated (traversal.go). Whether
// they resolve is checked in core (ValidateMoved).
func (p *parser) decodeMoved(block *hcl.Block, model *arch.SourceModel) arch.Diagnostics {
	content, remain, hclDiags := block.Body.PartialContent(movedSchema)
	diags := p.conv.diags(hclDiags, arch.CodeUnknownAttribute)
	diags = append(diags, p.reportExtraneous(remain, "a moved block", []string{"from", "to"})...)

	decl := arch.MovedDecl{Range: p.conv.rng(&block.DefRange)}
	for name, dst := range map[string]*arch.Reference{"from": &decl.From, "to": &decl.To} {
		if attr, ok := content.Attributes[name]; ok {
			ref, d := p.reference(attr)
			diags, *dst = append(diags, d...), ref
		}
	}
	if decl.From.IsZero() || decl.To.IsZero() {
		return diags
	}
	model.Moved = append(model.Moved, decl)
	return diags
}
