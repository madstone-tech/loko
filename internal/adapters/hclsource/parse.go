package hclsource

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// parsedFile pairs a successfully parsed file with its project-relative path.
type parsedFile struct {
	rel  string
	file *hcl.File
}

// parser holds the HCL parser and its file cache for one run.
//
// The cache is retained deliberately: rendering a diagnostic with the offending
// line and a caret needs the original bytes, and that is most of what makes a
// compiler's errors feel usable. It is discarded when the run ends.
type parser struct {
	hcl  *hclparse.Parser
	conv converter
	// evalCtx carries the locals and the five functions. It is populated once,
	// after locals are gathered across every file, so a local declared in one
	// file is usable from another.
	evalCtx *hcl.EvalContext
}

func newParser(root string) *parser {
	return &parser{
		hcl:  hclparse.NewParser(),
		conv: converter{root: root},
	}
}

// parseAll reads and parses every discovered file.
//
// A file that fails to parse yields a syntax_error and is dropped from the
// returned set; the remaining files are still parsed, so one run reports every
// problem it can find rather than stopping at the first bad file (FR-031,
// FR-028a).
func (p *parser) parseAll(files []SourceFile) ([]parsedFile, arch.Diagnostics) {
	var (
		out   []parsedFile
		diags arch.Diagnostics
	)

	for _, f := range files {
		src, err := os.ReadFile(f.Abs)
		if err != nil {
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityError,
				Code:     arch.CodeSyntaxError,
				Summary:  "Cannot read source file",
				Detail:   err.Error(),
				Range:    arch.SourceRange{File: f.Rel},
			})
			continue
		}

		// Parse under the project-relative name so every range HCL produces is
		// already relative; no post-hoc rewriting, and nothing machine-specific
		// can leak into an export (FR-036c).
		file, hclDiags := p.hcl.ParseHCL(src, f.Rel)
		if hclDiags.HasErrors() {
			diags = append(diags, p.conv.diags(hclDiags, arch.CodeSyntaxError)...)
			continue
		}
		// Non-fatal parse diagnostics are still worth surfacing.
		for _, d := range hclDiags {
			if d.Severity != hcl.DiagError {
				diags = append(diags, p.conv.diag(d, arch.CodeSyntaxError))
			}
		}
		if file == nil {
			continue
		}
		out = append(out, parsedFile{rel: f.Rel, file: file})
	}

	return out, diags
}
