package hclsource

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// SourceSuffix identifies an architecture source file. A plain ".hcl" file is
// not source: projects routinely hold unrelated HCL, and silently compiling it
// would be worse than ignoring it (FR-002).
const SourceSuffix = ".loko.hcl"

// skipDirs are never descended into. dist/ holds generated output, and reading
// it back would let a rendered artefact re-enter the model as authored source
// — exactly the two-sources-of-truth defect this release removes.
var skipDirs = map[string]bool{
	".git":         true,
	"dist":         true,
	"vendor":       true,
	"node_modules": true,
}

// SourceFile is one discovered input.
type SourceFile struct {
	// Rel is project-relative with forward slashes, on every platform.
	Rel string
	// Abs is the path to open.
	Abs string
}

// Discover walks root and returns every architecture source file beneath it,
// sorted by Rel (FR-001, FR-040).
//
// Finding nothing is not an error: it returns a no_source_found diagnostic so
// the caller can report it alongside anything else it knows (FR-005). The
// error return is reserved for a root that cannot be walked at all.
func Discover(root string) ([]SourceFile, arch.Diagnostics, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, nil, fmt.Errorf("reading project root %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("project root %s is not a directory", root)
	}

	var (
		files []SourceFile
		diags arch.Diagnostics
	)

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// An unreadable directory is a diagnostic against that path, not a
			// reason to abandon the run: the author should see every problem
			// from one invocation (FR-031).
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityError,
				Code:     arch.CodeSyntaxError,
				Summary:  "Cannot read path",
				Detail:   err.Error(),
				Range:    arch.SourceRange{File: arch.NormalizeFile(root, path)},
			})
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if path == root {
				return nil
			}
			name := d.Name()
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			return nil
		}

		if !isSourceName(d.Name()) {
			return nil
		}
		// Rel is computed against the walk root directly rather than through
		// NormalizeFile: both sides come from the same walk, so this is exact
		// whether the caller passed an absolute or a relative root.
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		files = append(files, SourceFile{
			Rel: filepath.ToSlash(filepath.Clean(rel)),
			Abs: path,
		})
		return nil
	})
	if walkErr != nil {
		return nil, diags, fmt.Errorf("walking project root %s: %w", root, walkErr)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Rel < files[j].Rel })

	if len(files) == 0 {
		diags = append(diags, arch.Diagnostic{
			Severity: arch.SeverityError,
			Code:     arch.CodeNoSourceFound,
			Summary:  "No architecture source found",
			Detail: fmt.Sprintf(
				"No *%s files were found under %s. An architecture is described in "+
					"*%s files; create one, for example arch%s, with a project block.",
				SourceSuffix, root, SourceSuffix, SourceSuffix),
		})
	}

	return files, diags, nil
}

// isSourceName reports whether a file name is architecture source. The name
// must have a segment before the suffix, so a bare "loko.hcl" is not source.
func isSourceName(name string) bool {
	return strings.HasSuffix(name, SourceSuffix) && len(name) > len(SourceSuffix)
}
