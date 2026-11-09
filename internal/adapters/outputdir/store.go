package outputdir

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
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
//     removed; the manifest is written last, into the state directory beside
//     outDir; a manifest left in outDir by 1.5.0 is then removed.
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

	if err := writeManifest(outDir, vm.NewManifest(artifacts).Encode(sources)); err != nil {
		return rep, err
	}
	// Retire the manifest a 1.5.0 build left inside the output directory. This
	// runs only once the state manifest is in place, so a failed commit keeps
	// it for the retry.
	if err := os.Remove(filepath.Join(outDir, vm.ManifestFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return rep, fmt.Errorf("removing the superseded manifest in %s: %w", outDir, err)
	}
	return rep, nil
}

// stateDir is the hidden directory beside outDir that holds the manifest:
// .<base>.loko, for example .dist.loko. It is never inside outDir, so copying
// or publishing the output directory carries no tool state (feature 022).
func stateDir(outDir string) string {
	return filepath.Join(filepath.Dir(outDir), "."+filepath.Base(outDir)+".loko")
}

// stateManifest is the manifest path inside stateDir.
func stateManifest(outDir string) string { return filepath.Join(stateDir(outDir), "manifest") }

// readManifest reads the state manifest. Only when it does not exist does it
// read the manifest a 1.5.0 build kept inside outDir. Any other failure, or a
// corrupt file, gives an empty manifest, which prunes nothing.
func readManifest(outDir string) vm.Manifest {
	data, err := os.ReadFile(stateManifest(outDir))
	if errors.Is(err, fs.ErrNotExist) {
		data, err = os.ReadFile(filepath.Join(outDir, vm.ManifestFile))
	}
	if err != nil {
		return vm.Manifest{}
	}
	return vm.DecodeManifest(data)
}

// writeManifest replaces the state manifest atomically: a temporary file in
// the state directory, then a rename.
func writeManifest(outDir string, data []byte) error {
	dir := stateDir(outDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating the manifest directory for %s: %w", outDir, err)
	}
	tmp, err := os.CreateTemp(dir, ".manifest-*")
	if err != nil {
		return fmt.Errorf("writing the manifest for %s: %w", outDir, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, werr := tmp.Write(data)
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Chmod(tmp.Name(), 0o644)
	}
	if werr == nil {
		werr = os.Rename(tmp.Name(), stateManifest(outDir))
	}
	if werr != nil {
		return fmt.Errorf("writing the manifest for %s: %w", outDir, werr)
	}
	return nil
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
