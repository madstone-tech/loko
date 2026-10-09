package hclsource

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
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
	// overlay holds in-memory file contents, keyed by project-relative path,
	// that take precedence over disk (LoadOverlay).
	overlay map[string][]byte
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
		src, err := p.read(f)
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

// read returns a file's bytes, from the overlay when it holds the path.
func (p *parser) read(f SourceFile) ([]byte, error) {
	if src, ok := p.overlay[f.Rel]; ok {
		return src, nil
	}
	return os.ReadFile(f.Abs)
}

// applyOverlay installs the overlay and adds the overlay-only paths to the
// discovered file list, keeping it sorted by relative path.
func (p *parser) applyOverlay(root string, files []SourceFile, overlay []authoring.FileContent) []SourceFile {
	if len(overlay) == 0 {
		return files
	}
	p.overlay = make(map[string][]byte, len(overlay))
	for _, o := range overlay {
		p.overlay[o.Path] = o.New
		if !slices.ContainsFunc(files, func(f SourceFile) bool { return f.Rel == o.Path }) {
			files = append(files, SourceFile{Rel: o.Path, Abs: filepath.Join(root, filepath.FromSlash(o.Path))})
		}
	}
	slices.SortFunc(files, func(a, b SourceFile) int { return strings.Compare(a.Rel, b.Rel) })
	return files
}
