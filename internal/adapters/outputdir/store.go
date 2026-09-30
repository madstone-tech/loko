package outputdir

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Store is the file-system ArtifactStore.
type Store struct{}

// New returns a Store.
func New() *Store { return &Store{} }

// Commit writes artifacts into outDir (research R8):
//
//  1. a staging directory inside outDir proves it is writable; if it cannot
//     be created the error names outDir and nothing is written;
//  2. viewmodel.PlanCommit decides what to write, what is already identical
//     on disk, and which previously owned paths to remove;
//  3. changed files are staged, then renamed into place; owned orphans are
//     removed; the manifest is written last.
//
// Files not named in the previous manifest are never touched.
func (s *Store) Commit(ctx context.Context, outDir string, artifacts []vm.Artifact, sources []string) (usecases.CommitReport, error) {
	var rep usecases.CommitReport
	for _, a := range artifacts {
		if !vm.SafePath(a.Path) || a.Path == vm.ManifestFile {
			return rep, fmt.Errorf("refusing to write unsafe artifact path %q", a.Path)
		}
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return rep, fmt.Errorf("creating output directory %s: %w", outDir, err)
	}
	staging, err := os.MkdirTemp(outDir, ".loko-staging-*")
	if err != nil {
		return rep, fmt.Errorf("output directory %s is not writable: %w", outDir, err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	plan := vm.PlanCommit(readManifest(outDir), artifacts, onDisk(outDir))
	byPath := make(map[string][]byte, len(artifacts))
	for _, a := range artifacts {
		byPath[a.Path] = a.Bytes
	}
	for _, p := range plan.Write {
		if err := ctx.Err(); err != nil {
			return rep, err
		}
		if err := writeFile(staging, p, byPath[p]); err != nil {
			return rep, fmt.Errorf("staging %s in %s: %w", p, outDir, err)
		}
	}
	for _, p := range plan.Write {
		if err := moveIntoPlace(staging, outDir, p); err != nil {
			return rep, err
		}
	}
	rep.Written, rep.Unchanged = plan.Write, plan.Unchanged
	rep.Removed = prune(outDir, plan.Remove)

	manifest := vm.NewManifest(artifacts).Encode(sources)
	if err := writeFile(staging, vm.ManifestFile, manifest); err != nil {
		return rep, fmt.Errorf("writing manifest in %s: %w", outDir, err)
	}
	if err := moveIntoPlace(staging, outDir, vm.ManifestFile); err != nil {
		return rep, err
	}
	return rep, nil
}

func readManifest(outDir string) vm.Manifest {
	data, err := os.ReadFile(filepath.Join(outDir, vm.ManifestFile))
	if err != nil {
		return vm.Manifest{}
	}
	return vm.DecodeManifest(data)
}

func onDisk(outDir string) func(string) ([]byte, bool) {
	return func(p string) ([]byte, bool) {
		b, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(p)))
		return b, err == nil
	}
}

func writeFile(root, p string, data []byte) error {
	full := filepath.Join(root, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func moveIntoPlace(staging, outDir, p string) error {
	dst := filepath.Join(outDir, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(dst), err)
	}
	if err := os.Rename(filepath.Join(staging, filepath.FromSlash(p)), dst); err != nil {
		return fmt.Errorf("writing %s: %w", dst, err)
	}
	return nil
}

// prune removes owned orphans and any directory that pruning left empty. It
// reports only what it actually removed; a path already gone is not an error.
func prune(outDir string, paths []string) []string {
	var removed []string
	for _, p := range paths {
		full := filepath.Join(outDir, filepath.FromSlash(p))
		if err := os.Remove(full); err != nil {
			continue // already gone, or unremovable: never fail a build over an orphan
		}
		removed = append(removed, p)
		for dir := path.Dir(p); dir != "." && dir != "/"; dir = path.Dir(dir) {
			if os.Remove(filepath.Join(outDir, filepath.FromSlash(dir))) != nil {
				break // not empty, or not ours to judge
			}
		}
	}
	return removed
}
