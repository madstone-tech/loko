package hclsource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// Plan applies edits, in order, to in-memory copies of the source files and
// returns the new content of every file it touched. Nothing is written.
func (e *Editor) Plan(ctx context.Context, root string, edits []authoring.Edit) (authoring.Plan, error) {
	w, err := newWorkspace(root)
	if err != nil {
		return authoring.Plan{}, err
	}
	for i, ed := range edits {
		if err := ctx.Err(); err != nil {
			return authoring.Plan{}, err
		}
		if err := w.apply(ed); err != nil {
			var ee *authoring.EditError
			if errors.As(err, &ee) {
				ee.Index = i
				return authoring.Plan{}, ee
			}
			return authoring.Plan{}, fmt.Errorf("edit %d: %w", i, err)
		}
	}
	return w.plan(), nil
}

func (w *workspace) apply(e authoring.Edit) error {
	switch e.Op {
	case authoring.OpAdd:
		return w.add(e)
	case authoring.OpUpdate:
		return w.update(e)
	case authoring.OpRemove:
		return w.remove(e)
	case authoring.OpRename:
		return w.rename(e)
	}
	return refuse(authoring.ReasonInvalidEdit, "unknown op %q", e.Op)
}

// workspace is the set of source files one plan reads and changes. Files are
// read lazily and parsed once until an edit changes them.
type workspace struct {
	root    string
	paths   []string                  // every source file, sorted
	orig    map[string][]byte         // content before the plan; nil for a created file
	cur     map[string][]byte         // content now
	parsed  map[string]*hclwrite.File // parse cache for cur
	broken  map[string]error          // files that cannot be edited
	touched map[string]bool
}

func newWorkspace(root string) (*workspace, error) {
	files, _, err := Discover(root)
	if err != nil {
		return nil, err
	}
	w := &workspace{root: root, orig: map[string][]byte{}, cur: map[string][]byte{},
		parsed: map[string]*hclwrite.File{}, broken: map[string]error{}, touched: map[string]bool{}}
	for _, f := range files {
		w.paths = append(w.paths, f.Rel)
	}
	return w, nil
}

func (w *workspace) content(path string) ([]byte, error) {
	if b, ok := w.cur[path]; ok {
		return b, nil
	}
	b, err := os.ReadFile(filepath.Join(w.root, filepath.FromSlash(path)))
	if err != nil {
		return nil, err
	}
	w.orig[path], w.cur[path] = b, b
	return b, nil
}

// file returns the parsed, editable form of a file.
func (w *workspace) file(path string) (*hclwrite.File, error) {
	if f, ok := w.parsed[path]; ok {
		return f, nil
	}
	src, err := w.content(path)
	if err != nil {
		return nil, err
	}
	f, err := parseLossless(src, path)
	if err != nil {
		w.broken[path] = err
		return nil, err
	}
	w.parsed[path] = f
	return f, nil
}

// store records an edited file's new content. A removal also closes up the
// blank lines either side of what it removed.
func (w *workspace) store(path string, f *hclwrite.File, removal bool) {
	b := tokensOf(f)
	if removal {
		b = closeJunction(w.cur[path], b)
	}
	w.cur[path] = b
	w.touched[path] = true
	delete(w.parsed, path) // reparse from the new bytes, so the next edit sees clean tokens
}

// create starts a new, empty source file.
func (w *workspace) create(path string) *hclwrite.File {
	w.orig[path], w.cur[path] = nil, []byte{}
	w.paths = append(w.paths, path)
	slices.Sort(w.paths)
	f := hclwrite.NewEmptyFile()
	w.parsed[path] = f
	return f
}

func (w *workspace) exists(path string) bool { return slices.Contains(w.paths, path) }

func (w *workspace) plan() authoring.Plan {
	var p authoring.Plan
	for _, path := range w.paths {
		if w.touched[path] {
			p.Files = append(p.Files, authoring.FileContent{Path: path, Old: w.orig[path], New: w.cur[path]})
		}
	}
	return p
}
