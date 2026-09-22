package hclsource

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// decodeProject handles the project block, which replaced the previous
// configuration file (FR-003).
//
// A second project block is an error rather than a merge: the project has one
// name and one version constraint, and silently picking a winner would make
// the compiled output depend on file ordering.
func (p *parser) decodeProject(block *hcl.Block, model *arch.SourceModel) arch.Diagnostics {
	name, diags := p.label(block, 0)
	if diags.HasErrors() {
		return diags
	}

	if model.Project.Declared {
		return append(diags, arch.Diagnostic{
			Severity: arch.SeverityError,
			Code:     arch.CodeDuplicateDeclaration,
			Summary:  "Duplicate project block",
			Detail: fmt.Sprintf("A project block is already declared as %q. "+
				"Exactly one project block may exist across all source files.", model.Project.Name),
			Range: p.conv.rng(&block.DefRange),
			Related: []arch.RelatedRange{{
				Message: "first declared here",
				Range:   model.Project.Range,
			}},
		})
	}

	content, attrDiags := p.attrs(block.Body, projectAttrs, "a project block")
	diags = append(diags, attrDiags...)

	decl := arch.ProjectDecl{
		Name:     name,
		Range:    p.conv.rng(&block.DefRange),
		Declared: true,
	}
	if attr, ok := content.Attributes["description"]; ok {
		s, d := p.evalString(attr)
		diags, decl.Description = append(diags, d...), s
	}
	if attr, ok := content.Attributes["loko_version"]; ok {
		s, d := p.evalString(attr)
		diags, decl.Version = append(diags, d...), s
		decl.VersionRange = p.conv.rng(attr.Expr.Range().Ptr())
	}

	model.Project = decl
	return diags
}

// decodeView handles a view block. Views are validated in this release and
// rendered in a later one (FR-016), so their references are extracted here and
// resolved by core like any other.
func (p *parser) decodeView(block *hcl.Block, model *arch.SourceModel) arch.Diagnostics {
	name, diags := p.label(block, 0)
	if diags.HasErrors() {
		return diags
	}

	content, attrDiags := p.attrs(block.Body, viewAttrs, "a view block")
	diags = append(diags, attrDiags...)

	decl := arch.ViewDecl{Name: name, Range: p.conv.rng(&block.DefRange)}

	if attr, ok := content.Attributes["include"]; ok {
		refs, d := p.referenceList(attr)
		diags, decl.Include = append(diags, d...), refs
	}
	if attr, ok := content.Attributes["exclude"]; ok {
		refs, d := p.referenceList(attr)
		diags, decl.Exclude = append(diags, d...), refs
	}
	if attr, ok := content.Attributes["tags"]; ok {
		tags, d := p.evalStringList(attr)
		diags, decl.Tags = append(diags, d...), tags
	}

	model.Views = append(model.Views, decl)
	return diags
}

// decodeReconcile handles the reconcile block. Its patterns are carried
// through untouched for the reconciliation stage (FR-015) and have no effect
// on this release's behaviour.
func (p *parser) decodeReconcile(block *hcl.Block, model *arch.SourceModel) arch.Diagnostics {
	content, diags := p.attrs(block.Body, reconcileAttrs, "a reconcile block")

	attr, ok := content.Attributes["ignore"]
	if !ok {
		return diags
	}
	patterns, d := p.evalStringList(attr)
	diags = append(diags, d...)
	r := p.conv.rng(attr.Expr.Range().Ptr())
	for _, pattern := range patterns {
		model.Ignores = append(model.Ignores, arch.IgnorePattern{Pattern: pattern, Range: r})
	}
	return diags
}
