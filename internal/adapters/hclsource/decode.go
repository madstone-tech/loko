package hclsource

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Source implements the usecases.ArchitectureSource port.
type Source struct{}

// New returns the HCL architecture source.
func New() *Source { return &Source{} }

// Load discovers, parses, and decodes every architecture source file beneath
// root into one unresolved SourceModel.
//
// It reports syntax-level problems only — unreadable files, invalid syntax,
// unknown blocks, unknown attributes, unknown functions. References come back
// as raw strings; resolving them and every semantic rule belong to core
// (research R7).
//
// A SourceModel is returned even when diagnostics contain errors, so the
// compiler can validate whatever parsed and report everything in one run.
func (s *Source) Load(_ context.Context, root string) (*arch.SourceModel, arch.Diagnostics, error) {
	files, diags, err := Discover(root)
	if err != nil {
		return nil, diags, err
	}

	p := newParser(root)
	parsed, parseDiags := p.parseAll(files)
	diags = append(diags, parseDiags...)

	model := &arch.SourceModel{}
	for _, f := range parsed {
		model.Files = append(model.Files, f.rel)
	}
	sort.Strings(model.Files)

	// Locals are collected across every file first, so a local declared in one
	// file is usable from another and declaration order does not matter.
	locals, localDiags := p.collectLocals(parsed)
	diags = append(diags, localDiags...)
	p.evalCtx = &hcl.EvalContext{
		Variables: map[string]cty.Value{"local": cty.ObjectVal(locals)},
		Functions: Functions,
	}

	for _, f := range parsed {
		diags = append(diags, p.decodeFile(f, model)...)
	}

	return model, diags, nil
}

// decodeFile walks one file's top-level blocks.
func (p *parser) decodeFile(f parsedFile, model *arch.SourceModel) arch.Diagnostics {
	content, remain, hclDiags := f.file.Body.PartialContent(fileSchema())
	diags := p.conv.diags(hclDiags, arch.CodeUnknownBlock)
	diags = append(diags, p.reportExtraneous(remain, "a loko source file", nil)...)

	for _, block := range content.Blocks {
		switch block.Type {
		case blockProject:
			diags = append(diags, p.decodeProject(block, model)...)
		case blockView:
			diags = append(diags, p.decodeView(block, model)...)
		case blockReconcile:
			diags = append(diags, p.decodeReconcile(block, model)...)
		case blockLocals:
			// Already gathered in the pre-pass.
		case blockDeployment:
			diags = append(diags, p.decodeDeployment(block, model)...)
		default:
			if arch.ValidElementKind(block.Type) {
				diags = append(diags, p.decodeElement(block, model)...)
				continue
			}
			diags = append(diags, p.unknownBlockDiag(block.Type, block.DefRange, "a loko source file"))
		}
	}
	return diags
}

// label returns a block's label and validates it as an address segment. An
// invalid label is reported here because an address built from it would be
// unusable downstream.
func (p *parser) label(block *hcl.Block, index int) (string, arch.Diagnostics) {
	if index >= len(block.Labels) {
		return "", arch.Diagnostics{p.conv.errorf(&block.DefRange, arch.CodeUnknownBlock,
			"Missing block name",
			fmt.Sprintf("A %q block requires a name label, for example %s \"api\" { }.",
				block.Type, block.Type))}
	}
	name := block.Labels[index]
	if !arch.ValidName(name) {
		r := block.DefRange
		if index < len(block.LabelRanges) {
			r = block.LabelRanges[index]
		}
		return name, arch.Diagnostics{p.conv.errorf(&r, arch.CodeUnknownBlock,
			"Invalid name",
			fmt.Sprintf("%q is not a valid name. Names start with a letter or underscore "+
				"and continue with letters, digits, underscores, or hyphens.", name))}
	}
	return name, nil
}

// attrs decodes a body against a schema and returns the matched attributes
// plus diagnostics for anything unexpected.
func (p *parser) attrs(body hcl.Body, legal []string, context string,
	blocks ...hcl.BlockHeaderSchema) (*hcl.BodyContent, arch.Diagnostics) {

	content, remain, hclDiags := body.PartialContent(bodySchema(legal, blocks...))
	diags := p.conv.diags(hclDiags, arch.CodeUnknownAttribute)
	diags = append(diags, p.reportExtraneous(remain, context, legal)...)
	return content, diags
}

// evalString evaluates an attribute to a string.
func (p *parser) evalString(attr *hcl.Attribute) (string, arch.Diagnostics) {
	v, diags := p.evalValue(attr)
	if diags.HasErrors() {
		return "", diags
	}
	if v.Kind != arch.ValueString {
		return "", append(diags, p.conv.errorf(attr.Expr.Range().Ptr(),
			arch.CodeUnknownAttribute, "Expected a string",
			fmt.Sprintf("%q must be a string.", attr.Name)))
	}
	return v.Str, diags
}

// evalStringList evaluates an attribute to a list of strings.
func (p *parser) evalStringList(attr *hcl.Attribute) ([]string, arch.Diagnostics) {
	v, diags := p.evalValue(attr)
	if diags.HasErrors() {
		return nil, diags
	}
	if v.Kind != arch.ValueList {
		return nil, append(diags, p.conv.errorf(attr.Expr.Range().Ptr(),
			arch.CodeUnknownAttribute, "Expected a list",
			fmt.Sprintf("%q must be a list of strings.", attr.Name)))
	}
	out := make([]string, 0, len(v.List))
	for _, item := range v.List {
		if item.Kind != arch.ValueString {
			diags = append(diags, p.conv.errorf(attr.Expr.Range().Ptr(),
				arch.CodeUnknownAttribute, "Expected a list of strings",
				fmt.Sprintf("Every entry of %q must be a string.", attr.Name)))
			continue
		}
		out = append(out, item.Str)
	}
	return out, diags
}

// evalValue evaluates any attribute into the core Value type, keeping cty on
// this side of the FR-044 boundary.
func (p *parser) evalValue(attr *hcl.Attribute) (arch.Value, arch.Diagnostics) {
	if diags := p.checkFunctionCalls(attr.Expr); len(diags) > 0 {
		return arch.Value{}, diags
	}
	v, hclDiags := attr.Expr.Value(p.evalCtx)
	if hclDiags.HasErrors() {
		return arch.Value{}, p.conv.diags(hclDiags, arch.CodeUnknownAttribute)
	}
	return p.ctyToValue(v, attr.Expr.Range().Ptr()), nil
}

// checkFunctionCalls rejects calls outside the five supported functions before
// evaluation, so the message can name the available set (FR-017a) rather than
// leaving HCL to say "call to unknown function".
func (p *parser) checkFunctionCalls(expr hcl.Expression) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, call := range functionCalls(expr) {
		if _, known := Functions[call.name]; known {
			continue
		}
		r := call.rng
		diags = append(diags, p.conv.errorf(&r,
			arch.CodeUnknownFunction, "Unknown function", unknownFunctionDetail(call.name)))
	}
	return diags
}

// ctyToValue converts an evaluated cty value into the core Value type.
func (p *parser) ctyToValue(v cty.Value, r *hcl.Range) arch.Value {
	src := p.conv.rng(r)
	if v.IsNull() || !v.IsKnown() {
		return arch.Value{Kind: arch.ValueNull, Source: src}
	}

	t := v.Type()
	switch {
	case t == cty.String:
		return arch.Value{Kind: arch.ValueString, Str: v.AsString(), Source: src}
	case t == cty.Bool:
		return arch.Value{Kind: arch.ValueBool, Bool: v.True(), Source: src}
	case t == cty.Number:
		f, _ := v.AsBigFloat().Float64()
		return arch.Value{Kind: arch.ValueNumber, Num: f, Source: src}
	case t.IsTupleType() || t.IsListType() || t.IsSetType():
		var items []arch.Value
		for it := v.ElementIterator(); it.Next(); {
			_, ev := it.Element()
			items = append(items, p.ctyToValue(ev, r))
		}
		return arch.Value{Kind: arch.ValueList, List: items, Source: src}
	case t.IsObjectType() || t.IsMapType():
		var kvs []arch.KeyValue
		for it := v.ElementIterator(); it.Next(); {
			k, ev := it.Element()
			kvs = append(kvs, arch.KeyValue{Key: k.AsString(), Value: p.ctyToValue(ev, r)})
		}
		// Sorted at the boundary so attribute maps are already deterministic
		// before they reach the IR builder (FR-040).
		sort.Slice(kvs, func(i, j int) bool { return kvs[i].Key < kvs[j].Key })
		return arch.Value{Kind: arch.ValueMap, MapKV: kvs, Source: src}
	default:
		return arch.Value{Kind: arch.ValueString, Str: strings.TrimSpace(v.GoString()), Source: src}
	}
}
