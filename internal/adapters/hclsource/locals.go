package hclsource

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// funcCall is one function call found in an expression.
type funcCall struct {
	name string
	rng  hcl.Range
}

// functionCalls walks an expression and returns every function call in it.
//
// Calls are rejected before evaluation so the message can name the five
// supported functions (FR-017a). Left to HCL, the author would get "call to
// unknown function" with no indication of what is available — and the set
// being small and fixed is exactly the sort of thing the error should teach.
func functionCalls(expr hcl.Expression) []funcCall {
	syn, ok := expr.(hclsyntax.Expression)
	if !ok {
		return nil
	}
	var out []funcCall
	_ = hclsyntax.Walk(syn, walker(func(node hclsyntax.Node) {
		if call, isCall := node.(*hclsyntax.FunctionCallExpr); isCall {
			out = append(out, funcCall{name: call.Name, rng: call.NameRange})
		}
	}))
	return out
}

// walker adapts a plain func to hclsyntax.Walker.
type walker func(hclsyntax.Node)

func (w walker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	w(node)
	return nil
}

func (w walker) Exit(hclsyntax.Node) hcl.Diagnostics { return nil }

// collectLocals gathers every locals block across every file into one map,
// before any other block is decoded.
//
// Collecting across all files first is what makes a local declared in one file
// usable from another, and makes declaration order irrelevant — the same
// property the two-pass reference resolution gives element references.
//
// Locals may reference other locals but never elements: a local is a value,
// and references are not values (research R2).
func (p *parser) collectLocals(files []parsedFile) (map[string]cty.Value, arch.Diagnostics) {
	type pending struct {
		name string
		expr hcl.Expression
		rng  hcl.Range
	}

	var (
		items []pending
		diags arch.Diagnostics
		seen  = map[string]hcl.Range{}
	)

	for _, f := range files {
		content, _, hclDiags := f.file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: blockLocals}},
		})
		diags = append(diags, p.conv.diags(hclDiags, arch.CodeUnknownBlock)...)

		for _, block := range content.Blocks {
			attrs, attrDiags := block.Body.JustAttributes()
			diags = append(diags, p.conv.diags(attrDiags, arch.CodeUnknownAttribute)...)
			for name, attr := range attrs {
				if prev, dup := seen[name]; dup {
					diags = append(diags, arch.Diagnostic{
						Severity: arch.SeverityError,
						Code:     arch.CodeDuplicateDeclaration,
						Summary:  "Duplicate local value",
						Detail:   fmt.Sprintf("A local named %q is already declared.", name),
						Range:    p.conv.rng(attr.NameRange.Ptr()),
						Related: []arch.RelatedRange{{
							Message: "first declared here",
							Range:   p.conv.rng(prev.Ptr()),
						}},
					})
					continue
				}
				seen[name] = attr.NameRange
				items = append(items, pending{name: name, expr: attr.Expr, rng: attr.NameRange})
			}
		}
	}

	// Resolve iteratively so a local may depend on another regardless of the
	// order they were written in. Each pass evaluates whatever is now
	// resolvable; progress stalling means the remainder is unresolvable.
	resolved := map[string]cty.Value{}
	remaining := items
	for len(remaining) > 0 {
		var stuck []pending
		progress := false

		for _, item := range remaining {
			ctx := &hcl.EvalContext{
				Variables: map[string]cty.Value{"local": cty.ObjectVal(resolved)},
				Functions: Functions,
			}
			if callDiags := p.checkFunctionCalls(item.expr); len(callDiags) > 0 {
				diags = append(diags, callDiags...)
				resolved[item.name] = cty.StringVal("")
				progress = true
				continue
			}
			v, evalDiags := item.expr.Value(ctx)
			if evalDiags.HasErrors() {
				stuck = append(stuck, item)
				continue
			}
			resolved[item.name] = v
			progress = true
		}

		if !progress {
			for _, item := range stuck {
				r := item.rng
				diags = append(diags, p.conv.errorf(&r, arch.CodeUnresolvedReference,
					"Unresolvable local value",
					fmt.Sprintf("Local %q cannot be evaluated. Locals may reference other "+
						"locals and the five string functions, but not elements: a reference "+
						"is not a value.", item.name)))
				resolved[item.name] = cty.StringVal("")
			}
			break
		}
		remaining = stuck
	}

	return resolved, diags
}
