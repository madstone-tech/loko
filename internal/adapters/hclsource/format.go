package hclsource

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Formatter implements the usecases.SourceFormatter port.
type Formatter struct{}

// NewFormatter returns the HCL source formatter.
func NewFormatter() *Formatter { return &Formatter{} }

// Format rewrites every non-canonical file in place and returns the paths it
// changed, sorted (FR-035).
func (f *Formatter) Format(ctx context.Context, root string) ([]string, arch.Diagnostics, error) {
	return f.run(ctx, root, true)
}

// Check reports which files are not canonically formatted without writing
// anything (FR-035a).
func (f *Formatter) Check(ctx context.Context, root string) ([]string, arch.Diagnostics, error) {
	return f.run(ctx, root, false)
}

// run does the work for both modes. Sharing one implementation is what
// guarantees --check cannot disagree with what a real run would do; two code
// paths would eventually drift and a CI gate that disagrees with the formatter
// is worse than no gate.
func (f *Formatter) run(_ context.Context, root string, write bool) ([]string, arch.Diagnostics, error) {
	files, diags, err := Discover(root)
	if err != nil {
		return nil, diags, err
	}

	conv := converter{root: root}
	var changed []string

	for _, file := range files {
		src, readErr := os.ReadFile(file.Abs)
		if readErr != nil {
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityError,
				Code:     arch.CodeSyntaxError,
				Summary:  "Cannot read source file",
				Detail:   readErr.Error(),
				Range:    arch.SourceRange{File: file.Rel},
			})
			continue
		}

		formatted, fileDiags := formatSource(conv, file.Rel, src)
		diags = append(diags, fileDiags...)
		if formatted == nil || bytes.Equal(src, formatted) {
			continue
		}

		changed = append(changed, file.Rel)
		if !write {
			continue
		}
		if writeErr := writeFilePreservingMode(file.Abs, formatted); writeErr != nil {
			return changed, diags, fmt.Errorf("writing %s: %w", file.Rel, writeErr)
		}
	}

	// Discover already returns files sorted, so changed follows that order.
	return changed, diags, nil
}

// formatSource returns the canonical form of src, or nil when the file does
// not parse.
//
// The parse check exists because hclwrite.Format is purely lexical: it will
// happily "format" a file with an unclosed block, producing something that
// still does not parse and has been silently rewritten. Refusing to touch a
// file we cannot parse is what makes the command safe to run over a whole
// project (FR-035).
func formatSource(conv converter, rel string, src []byte) ([]byte, arch.Diagnostics) {
	_, hclDiags := hclwrite.ParseConfig(src, rel, hcl.InitialPos)
	if hclDiags.HasErrors() {
		return nil, conv.diags(hclDiags, arch.CodeSyntaxError)
	}
	return hclwrite.Format(src), nil
}

// writeFilePreservingMode rewrites a file without changing its permissions.
//
// The write is direct rather than write-to-temp-and-rename: a rename would
// break hard links and, on some systems, reset ownership — surprising
// behaviour for a formatter, which should be the least invasive thing in the
// toolchain.
func writeFilePreservingMode(path string, data []byte) error {
	info, err := os.Stat(path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, data, mode)
}
