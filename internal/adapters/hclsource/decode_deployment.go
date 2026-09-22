package hclsource

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// decodeDeployment handles a deployment block and the node/instance/binding
// tree inside it (FR-011 to FR-014).
func (p *parser) decodeDeployment(block *hcl.Block, model *arch.SourceModel) arch.Diagnostics {
	name, diags := p.label(block, 0)
	if diags.HasErrors() {
		return diags
	}

	content, attrDiags := p.attrs(block.Body, deploymentAttrs, "a deployment block",
		hcl.BlockHeaderSchema{Type: blockNode, LabelNames: []string{"name"}},
		hcl.BlockHeaderSchema{Type: blockInstance, LabelNames: []string{"name"}})
	diags = append(diags, attrDiags...)

	decl := arch.EnvironmentDecl{Name: name, Range: p.conv.rng(&block.DefRange)}

	for _, attrName := range deploymentAttrs {
		attr, ok := content.Attributes[attrName]
		if !ok {
			continue
		}
		s, d := p.evalString(attr)
		diags = append(diags, d...)
		switch attrName {
		case "provider":
			decl.Provider = s
		case "account":
			decl.Account = s
		case "region":
			decl.Region = s
		}
	}

	for _, nested := range content.Blocks {
		switch nested.Type {
		case blockNode:
			g, d := p.decodeNode(nested)
			diags = append(diags, d...)
			decl.Groups = append(decl.Groups, g)
		case blockInstance:
			inst, d := p.decodeInstance(nested)
			diags = append(diags, d...)
			decl.Instances = append(decl.Instances, inst)
		default:
			diags = append(diags, p.unknownBlockDiag(nested.Type, nested.DefRange, "a deployment block"))
		}
	}

	model.Environments = append(model.Environments, decl)
	return diags
}

// decodeNode handles a placement group, which nests to arbitrary depth
// (FR-012).
//
// A node carries its own address but contributes no segment to the instances
// inside it, so re-parenting an instance preserves its identity (FR-024). That
// is a property of address construction in core; here the tree is simply
// recorded as authored.
func (p *parser) decodeNode(block *hcl.Block) (arch.GroupDecl, arch.Diagnostics) {
	name, diags := p.label(block, 0)
	g := arch.GroupDecl{Name: name, Range: p.conv.rng(&block.DefRange)}
	if diags.HasErrors() {
		return g, diags
	}

	content, attrDiags := p.attrs(block.Body, nil, "a node block",
		hcl.BlockHeaderSchema{Type: blockNode, LabelNames: []string{"name"}},
		hcl.BlockHeaderSchema{Type: blockInstance, LabelNames: []string{"name"}})
	diags = append(diags, attrDiags...)

	for _, nested := range content.Blocks {
		switch nested.Type {
		case blockNode:
			child, d := p.decodeNode(nested)
			diags = append(diags, d...)
			g.Groups = append(g.Groups, child)
		case blockInstance:
			inst, d := p.decodeInstance(nested)
			diags = append(diags, d...)
			g.Instances = append(g.Instances, inst)
		default:
			diags = append(diags, p.unknownBlockDiag(nested.Type, nested.DefRange, "a node block"))
		}
	}
	return g, diags
}

// decodeInstance handles an instance block (FR-013).
func (p *parser) decodeInstance(block *hcl.Block) (arch.InstanceDecl, arch.Diagnostics) {
	name, diags := p.label(block, 0)
	inst := arch.InstanceDecl{Name: name, Range: p.conv.rng(&block.DefRange)}
	if diags.HasErrors() {
		return inst, diags
	}

	content, attrDiags := p.attrs(block.Body, instanceAttrs, "an instance block",
		hcl.BlockHeaderSchema{Type: blockBinding, LabelNames: []string{"kind"}})
	diags = append(diags, attrDiags...)

	if attr, ok := content.Attributes["of"]; ok {
		ref, d := p.reference(attr)
		diags, inst.Of = append(diags, d...), ref
	}
	if attr, ok := content.Attributes["attributes"]; ok {
		v, d := p.evalValue(attr)
		diags = append(diags, d...)
		if v.Kind == arch.ValueMap {
			inst.Attributes = v.MapKV
		} else if !d.HasErrors() {
			diags = append(diags, p.conv.errorf(attr.Expr.Range().Ptr(),
				arch.CodeUnknownAttribute, "Expected a map",
				`"attributes" must be a map, for example { memory = 1024, timeout = 30 }.`))
		}
	}

	for _, nested := range content.Blocks {
		if nested.Type != blockBinding {
			diags = append(diags, p.unknownBlockDiag(nested.Type, nested.DefRange, "an instance block"))
			continue
		}
		claim, d := p.decodeBinding(nested)
		diags = append(diags, d...)
		inst.Claims = append(inst.Claims, claim)
	}

	return inst, diags
}

// decodeBinding handles a binding block, whose label is the claim kind
// (FR-014). Only terraform and cloudformation ship in v1.0.
func (p *parser) decodeBinding(block *hcl.Block) (arch.ClaimDecl, arch.Diagnostics) {
	claim := arch.ClaimDecl{Range: p.conv.rng(&block.DefRange)}

	var diags arch.Diagnostics
	if len(block.Labels) == 0 {
		return claim, arch.Diagnostics{p.conv.errorf(&block.DefRange, arch.CodeUnknownBlock,
			"Missing binding kind",
			`A binding block requires a kind label, for example binding "terraform" { }.`)}
	}

	kind := block.Labels[0]
	if !arch.ValidClaimKind(kind) {
		r := block.LabelRanges[0]
		diags = append(diags, p.conv.errorf(&r, arch.CodeUnknownBlock,
			"Unsupported binding kind",
			fmt.Sprintf("%q is not a binding kind. v1.0 supports terraform and cloudformation.", kind)))
	}
	claim.Kind = arch.ClaimKind(kind)

	content, attrDiags := p.attrs(block.Body, bindingAttrs, "a binding block")
	diags = append(diags, attrDiags...)

	if attr, ok := content.Attributes["address"]; ok {
		s, d := p.evalString(attr)
		diags, claim.Address = append(diags, d...), s
	}
	if attr, ok := content.Attributes["addresses"]; ok {
		list, d := p.evalStringList(attr)
		diags, claim.Addresses = append(diags, d...), list
	}
	if attr, ok := content.Attributes["tags"]; ok {
		v, d := p.evalValue(attr)
		diags = append(diags, d...)
		if v.Kind == arch.ValueMap {
			claim.Tags = v.MapKV
		} else if !d.HasErrors() {
			diags = append(diags, p.conv.errorf(attr.Expr.Range().Ptr(),
				arch.CodeUnknownAttribute, "Expected a map",
				`"tags" must be a map of strings, for example { Application = "payments" }.`))
		}
	}

	return claim, diags
}
