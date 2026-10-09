package hclsource

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Block type names. These are the language's permanent surface: once a name
// ships, users' files depend on it and there is no migration command.
const (
	blockProject    = "project"
	blockLocals     = "locals"
	blockView       = "view"
	blockReconcile  = "reconcile"
	blockDeployment = "deployment"
	blockNode       = "node"
	blockInstance   = "instance"
	blockBinding    = "binding"
	blockUses       = "uses"
	blockMoved      = "moved"
)

// deliberatelyExcluded names constructs a reader might reasonably expect,
// which this release does not have (FR-018). Rejecting them by name lets the
// error teach rather than merely refuse — "unknown block" alone would leave
// the author wondering whether they had mistyped something.
var deliberatelyExcluded = map[string]string{
	"for_each": "iteration is deliberately excluded from v1.0",
	"dynamic":  "dynamic block generation is deliberately excluded from v1.0",
	"variable": "input variables are deliberately excluded from v1.0; use locals",
	"module":   "modules are deliberately excluded from v1.0",
	"resource": "loko describes architecture, not resources; see binding blocks to claim them",
	"output":   "outputs are deliberately excluded from v1.0; use loko export",
	"provider": "provider configuration belongs to your infrastructure tool, not to loko",
}

// fileSchema lists every block that may appear at the top level of a source
// file. PartialContent against this schema turns anything else into a
// diagnostic rather than a hard failure, so one run reports every unknown
// construct at once (FR-031).
func fileSchema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: blockProject, LabelNames: []string{"name"}},
			{Type: string(arch.KindPerson), LabelNames: []string{"name"}},
			{Type: string(arch.KindSystem), LabelNames: []string{"name"}},
			{Type: string(arch.KindContainer), LabelNames: []string{"name"}},
			{Type: string(arch.KindComponent), LabelNames: []string{"name"}},
			{Type: string(arch.KindExternal), LabelNames: []string{"name"}},
			{Type: blockDeployment, LabelNames: []string{"name"}},
			{Type: blockView, LabelNames: []string{"name"}},
			{Type: blockReconcile},
			{Type: blockMoved},
			{Type: blockLocals},
		},
	}
}

// Attribute sets per block. Held as sorted slices so the "did you mean" list in
// an unknown-attribute diagnostic is deterministic.
var (
	commonElementAttrs = []string{"description", "docs", "owner", "shape", "tags", "technology", "title"}
	usesAttrs          = []string{"description", "kind", "tags", "target", "technology"}
	projectAttrs       = []string{"description", "loko_version"}
	deploymentAttrs    = []string{"account", "provider", "region"}
	instanceAttrs      = []string{"attributes", "of"}
	bindingAttrs       = []string{"address", "addresses", "tags"}
	viewAttrs          = []string{"direction", "exclude", "include", "tags"}
	reconcileAttrs     = []string{"ignore"}
)

// elementAttrs returns the attributes legal on a given element kind. Containers
// and components additionally take their parent reference.
func elementAttrs(kind arch.ElementKind) []string {
	switch kind {
	case arch.KindContainer:
		return withAttr(commonElementAttrs, string(arch.KindSystem))
	case arch.KindComponent:
		return withAttr(commonElementAttrs, string(arch.KindContainer))
	default:
		return commonElementAttrs
	}
}

func withAttr(base []string, extra string) []string {
	out := make([]string, 0, len(base)+1)
	out = append(out, base...)
	out = append(out, extra)
	sort.Strings(out)
	return out
}

// bodySchema builds a schema from an attribute name list plus block headers.
// Every attribute is optional here: requiredness is a validation rule, and
// validation belongs to core, not to the parser.
func bodySchema(attrs []string, blocks ...hcl.BlockHeaderSchema) *hcl.BodySchema {
	s := &hcl.BodySchema{Blocks: blocks}
	for _, a := range attrs {
		s.Attributes = append(s.Attributes, hcl.AttributeSchema{Name: a})
	}
	return s
}

// unknownBlockDiag reports a block the language does not define, naming a
// reason when the block is one that was deliberately left out.
func (p *parser) unknownBlockDiag(blockType string, r hcl.Range, context string) arch.Diagnostic {
	detail := fmt.Sprintf("%q is not a block in %s.", blockType, context)
	if reason, excluded := deliberatelyExcluded[blockType]; excluded {
		detail = fmt.Sprintf("%q is not supported: %s.", blockType, reason)
	}
	return p.conv.errorf(&r, arch.CodeUnknownBlock, "Unsupported block type", detail)
}

// unknownAttrDiag reports an attribute the language does not define, listing
// what is legal there. Listing them is what turns the error into documentation.
func (p *parser) unknownAttrDiag(name string, r hcl.Range, context string, legal []string) arch.Diagnostic {
	return p.conv.errorf(&r, arch.CodeUnknownAttribute, "Unsupported attribute",
		fmt.Sprintf("%q is not an attribute of %s. Supported attributes are: %s.",
			name, context, strings.Join(legal, ", ")))
}

// reportExtraneous converts the leftovers PartialContent hands back into
// unknown-block and unknown-attribute diagnostics.
//
// HCL returns the unmatched remainder rather than failing, which is precisely
// what lets one run report every unknown construct instead of stopping at the
// first.
func (p *parser) reportExtraneous(remain hcl.Body, context string, legalAttrs []string) arch.Diagnostics {
	if remain == nil {
		return nil
	}
	// An empty schema matches nothing, so everything left over shows up as a
	// diagnostic from HCL itself, carrying exact ranges.
	_, hclDiags := remain.Content(&hcl.BodySchema{})

	var out arch.Diagnostics
	for _, d := range hclDiags {
		if d.Subject == nil {
			continue
		}
		if name, ok := blockTypeFromDiag(d); ok {
			out = append(out, p.unknownBlockDiag(name, *d.Subject, context))
			continue
		}
		if name, ok := attrNameFromDiag(d); ok {
			out = append(out, p.unknownAttrDiag(name, *d.Subject, context, legalAttrs))
			continue
		}
		// Shape HCL did not describe in a way we recognise: surface it rather
		// than swallow it, tagged as an unknown block.
		out = append(out, p.conv.errorf(d.Subject, arch.CodeUnknownBlock,
			"Unsupported construct", d.Detail))
	}
	return out
}

// HCL's summaries for these two cases are "Unsupported block type" and
// "Unsupported argument"; the offending name appears quoted in the detail
// ("Blocks of type %q are not expected here.", "An argument named %q is not
// expected here."). Matching is case-insensitive on the summary and tolerant
// of either field carrying the name, so a rewording upstream degrades to the
// generic branch rather than mislabelling a block as an attribute.
func blockTypeFromDiag(d *hcl.Diagnostic) (string, bool) {
	hay := strings.ToLower(d.Summary + " " + d.Detail)
	if !strings.Contains(hay, "block") {
		return "", false
	}
	if name := quotedName(d.Detail); name != "" {
		return name, true
	}
	return quotedName(d.Summary), true
}

func attrNameFromDiag(d *hcl.Diagnostic) (string, bool) {
	hay := strings.ToLower(d.Summary + " " + d.Detail)
	if !strings.Contains(hay, "argument") && !strings.Contains(hay, "attribute") {
		return "", false
	}
	if name := quotedName(d.Detail); name != "" {
		return name, true
	}
	return quotedName(d.Summary), true
}

// quotedName pulls the first double-quoted identifier out of HCL's message.
// Returns "" when there is none, which the callers render as a generic name.
func quotedName(s string) string {
	_, after, ok := strings.Cut(s, `"`)
	if !ok {
		return ""
	}
	rest := after
	before0, _, ok0 := strings.Cut(rest, `"`)
	if !ok0 {
		return ""
	}
	return before0
}
