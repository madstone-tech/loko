package hclsource

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// Commit writes every changed file in the plan, or none (FR-012a). It refuses
// with stale_revision if a file the plan changes differs on disk from base,
// and with path_refused for anything but *.loko.hcl inside root.
//
// Each file is first written to a temporary sibling, so a failure while
// writing leaves the originals untouched; only then are the temporaries
// renamed into place. If a rename fails, the files already renamed are
// restored from the plan's original bytes.
func (e *Editor) Commit(ctx context.Context, root string, plan authoring.Plan, base authoring.Revision) error {
	changed := slices.DeleteFunc(slices.Clone(plan.Files), func(f authoring.FileContent) bool {
		return f.Old != nil && bytes.Equal(f.Old, f.New)
	})
	for _, f := range changed {
		if err := checkPath(root, f.Path); err != nil {
			return err
		}
	}
	if err := checkStale(root, changed, base); err != nil {
		return err
	}
	staged, err := stage(root, changed)
	defer removeAll(staged)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Again, just before the renames: keep the window for a racing save small.
	if err := checkStale(root, changed, base); err != nil {
		return err
	}
	return swapIn(root, changed, staged)
}

// checkStale compares each file the plan changes against the revision the
// caller read. A file that has appeared since is stale too.
func checkStale(root string, files []authoring.FileContent, base authoring.Revision) error {
	var stale []string
	for _, f := range files {
		want, existed := base.Hash(f.Path)
		got, err := hashFile(abs(root, f.Path))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if existed {
				stale = append(stale, f.Path)
			}
		case err != nil:
			return err
		case !existed || got != want || (f.Old != nil && got != hashBytes(f.Old)):
			stale = append(stale, f.Path)
		}
	}
	if len(stale) > 0 {
		return refuse(authoring.ReasonStaleRevision, "%s changed since your last read; read again and retry", strings.Join(stale, ", "))
	}
	return nil
}

func abs(root, rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }

// stage writes each file's new content to a temporary file beside it, with
// the original's permissions. It returns the temporaries written so far.
func stage(root string, files []authoring.FileContent) ([]string, error) {
	var staged []string
	for _, f := range files {
		dst := abs(root, f.Path)
		mode := fs.FileMode(0o644)
		if fi, err := os.Stat(dst); err == nil {
			mode = fi.Mode().Perm()
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return staged, err
		}
		tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".tmp-*")
		if err != nil {
			return staged, err
		}
		staged = append(staged, tmp.Name())
		_, werr := tmp.Write(f.New)
		serr := tmp.Sync()
		cerr := tmp.Close()
		if err := errors.Join(werr, serr, cerr, os.Chmod(tmp.Name(), mode)); err != nil {
			return staged, err
		}
	}
	return staged, nil
}

// swapIn renames the temporaries into place, restoring the originals if any
// rename fails.
func swapIn(root string, files []authoring.FileContent, staged []string) error {
	for i, f := range files {
		if err := os.Rename(staged[i], abs(root, f.Path)); err != nil {
			return errors.Join(err, restore(root, files[:i]))
		}
		staged[i] = ""
	}
	return nil
}

// restore puts back the original content of files already committed.
func restore(root string, files []authoring.FileContent) error {
	var errs []error
	for _, f := range files {
		if f.Old == nil {
			errs = append(errs, os.Remove(abs(root, f.Path)))
			continue
		}
		errs = append(errs, os.WriteFile(abs(root, f.Path), f.Old, 0o644))
	}
	return errors.Join(errs...)
}

func removeAll(paths []string) {
	for _, p := range paths {
		if p != "" {
			_ = os.Remove(p)
		}
	}
}
