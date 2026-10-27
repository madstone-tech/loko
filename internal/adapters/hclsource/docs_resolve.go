package hclsource

import (
	"os"
	"path"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// resolveAllDocs resolves every element's docs path against the files on
// disk, confined to the project root (feature 018, R9). Overlaid source files
// are resolved the same way: docs files are never overlaid. A root that
// cannot be opened leaves the paths unresolved, so no docs diagnostic is
// given rather than a wrong one.
func resolveAllDocs(root string, model *arch.SourceModel) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return
	}
	defer func() { _ = r.Close() }() // read-only; a close error loses nothing
	for i := range model.Elements {
		e := &model.Elements[i]
		if e.Docs != "" {
			e.DocsPath, e.DocsVia, e.DocsCandidates = resolveDocs(r, e.Range.File, e.Docs)
		}
	}
}

// resolveDocs resolves one docs path declared in file (relative to the root,
// forward slashes): from the file's directory first, then from the root. It
// stats at most two paths. A path from a file at the root, or found only from
// the root, is returned as written, so 1.1.0 models export unchanged.
func resolveDocs(r *os.Root, file, docs string) (string, arch.DocsVia, [2]string) {
	dir := path.Dir(file)
	cands := [2]string{path.Clean(path.Join(dir, docs)), path.Clean(docs)}
	near, inFile := stat(r, cands[0])
	far, inRoot := stat(r, cands[1])
	switch {
	case inFile && (dir == "." || cands[0] == cands[1]):
		return docs, arch.DocsFile, cands
	case inFile && inRoot && !os.SameFile(near, far):
		return cands[0], arch.DocsBoth, cands
	case inFile:
		return cands[0], arch.DocsFile, cands
	case inRoot:
		return docs, arch.DocsRoot, cands
	}
	return "", arch.DocsMissing, cands
}

// stat reports whether rel names a regular file inside the root. A path that
// leaves the root is not found, whatever is there.
func stat(r *os.Root, rel string) (os.FileInfo, bool) {
	name := filepath.FromSlash(rel)
	if !filepath.IsLocal(name) {
		return nil, false
	}
	info, err := r.Stat(name)
	if err != nil || info.IsDir() {
		return nil, false
	}
	return info, true
}
