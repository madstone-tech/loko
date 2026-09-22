package hclsource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// reference extracts a cross-element reference from an attribute expression.
//
// The expression is read as a STATIC TRAVERSAL and never evaluated
// (research R2). Three things follow, and they are the reason the whole
// compiler is shaped this way:
//
//  1. Resolution can live in core. The adapter emits a dotted string and a
//     range; the symbol table, the unresolvable-reference error, and the
//     wrong-kind error all belong to the use-case layer where the constitution
//     wants validation to sit.
//  2. Declaration order stops mattering for free. There is no partially
//     populated evaluation context to sequence — pass one collects addresses,
//     pass two does map lookups.
//  3. The diff stage's `moved` block, whose operands deliberately name
//     addresses that no longer exist, becomes an ordinary use of this function
//     rather than a carve-out in the evaluator.
//
// A quoted string in a reference position is rejected: "container.api" is a
// description, not a pointer, and accepting it would let a typo read as valid.
func (p *parser) reference(attr *hcl.Attribute) (arch.Reference, arch.Diagnostics) {
	traversal, hclDiags := hcl.AbsTraversalForExpr(attr.Expr)
	if hclDiags.HasErrors() {
		return arch.Reference{}, arch.Diagnostics{p.conv.errorf(
			attr.Expr.Range().Ptr(),
			arch.CodeWrongReferenceKind,
			"Expected a reference, not a value",
			fmt.Sprintf(
				"%q must be a reference such as container.api, written without quotes. "+
					"A quoted string is a value, not a pointer to an element.",
				attr.Name),
		)}
	}

	raw, ok := traversalString(traversal)
	if !ok {
		return arch.Reference{}, arch.Diagnostics{p.conv.errorf(
			attr.Expr.Range().Ptr(),
			arch.CodeWrongReferenceKind,
			"Unsupported reference form",
			fmt.Sprintf(
				"%q must be a plain dotted reference such as container.api. "+
					"Index and splat expressions are not part of this language.",
				attr.Name),
		)}
	}

	return arch.Reference{Raw: raw, Range: p.conv.rng(attr.Expr.Range().Ptr())}, nil
}

// referenceList extracts a list of references, as used by a view's include and
// exclude attributes.
func (p *parser) referenceList(attr *hcl.Attribute) ([]arch.Reference, arch.Diagnostics) {
	exprs, hclDiags := hcl.ExprList(attr.Expr)
	if hclDiags.HasErrors() {
		return nil, arch.Diagnostics{p.conv.errorf(
			attr.Expr.Range().Ptr(),
			arch.CodeWrongReferenceKind,
			"Expected a list of references",
			fmt.Sprintf("%q must be a list, for example [system.payments, container.api].", attr.Name),
		)}
	}

	var (
		refs  []arch.Reference
		diags arch.Diagnostics
	)
	for _, expr := range exprs {
		traversal, tDiags := hcl.AbsTraversalForExpr(expr)
		if tDiags.HasErrors() {
			diags = append(diags, p.conv.errorf(
				expr.Range().Ptr(),
				arch.CodeWrongReferenceKind,
				"Expected a reference, not a value",
				fmt.Sprintf("Each entry of %q must be a reference such as container.api, written without quotes.", attr.Name),
			))
			continue
		}
		raw, ok := traversalString(traversal)
		if !ok {
			diags = append(diags, p.conv.errorf(
				expr.Range().Ptr(),
				arch.CodeWrongReferenceKind,
				"Unsupported reference form",
				fmt.Sprintf("Each entry of %q must be a plain dotted reference such as container.api.", attr.Name),
			))
			continue
		}
		refs = append(refs, arch.Reference{Raw: raw, Range: p.conv.rng(expr.Range().Ptr())})
	}
	return refs, diags
}

// traversalString renders a traversal as its dotted source form. It reports
// false for anything that is not a plain chain of names, such as an index or a
// splat, which this language does not have.
func traversalString(t hcl.Traversal) (string, bool) {
	if len(t) == 0 {
		return "", false
	}
	var b strings.Builder
	for i, step := range t {
		switch s := step.(type) {
		case hcl.TraverseRoot:
			if i != 0 {
				return "", false
			}
			b.WriteString(s.Name)
		case hcl.TraverseAttr:
			b.WriteByte('.')
			b.WriteString(s.Name)
		default:
			return "", false
		}
	}
	return b.String(), true
}
