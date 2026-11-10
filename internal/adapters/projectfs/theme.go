package projectfs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// ThemeDir is where theme overrides live, relative to the project root.
const ThemeDir = "templates"

// Theme is the ThemeSource over the local file system.
type Theme struct{}

// themeExts are the override kinds the site uses; other files in templates/
// are ignored.
var themeExts = map[string]bool{".gohtml": true, ".css": true, ".js": true}

// assetsDir holds theme assets, below ThemeDir (feature 023).
const assetsDir = "assets"

// LoadTheme returns every .gohtml, .css and .js file directly in
// <root>/templates/, sorted by name, then every file below
// <root>/templates/assets/, sorted by path. A missing directory means the
// built-in theme is used, with no warning (US7/AC2). Unknown names are
// returned too, so the build can reject them rather than silently ignore a
// typo.
func (Theme) LoadTheme(_ context.Context, root string) ([]vm.ThemeFile, error) {
	dir := filepath.Join(root, ThemeDir)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading theme overrides in %s: %w", dir, err)
	}
	var out []vm.ThemeFile
	for _, e := range entries {
		if e.IsDir() || !themeExts[filepath.Ext(e.Name())] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading theme override %s: %w", e.Name(), err)
		}
		out = append(out, vm.ThemeFile{Name: e.Name(), Bytes: data, Origin: ThemeDir + "/" + e.Name()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	assets, err := loadAssets(filepath.Join(dir, assetsDir))
	if err != nil {
		return nil, err
	}
	return append(out, assets...), nil
}

// loadAssets returns every regular file below dir at any depth, with
// unchanged bytes. A symbolic link is not followed: it is returned with Link
// set and no bytes, so the build refuses it (contract T-4).
func loadAssets(dir string) ([]vm.ThemeFile, error) {
	var out []vm.ThemeFile
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && p == dir {
			return filepath.SkipAll
		}
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		rel = path.Join(assetsDir, filepath.ToSlash(rel))
		f := vm.ThemeFile{Name: d.Name(), Rel: rel, Origin: ThemeDir + "/" + rel}
		if d.Type()&fs.ModeSymlink != 0 {
			f.Link = true
		} else if !d.Type().IsRegular() {
			return nil
		} else if f.Bytes, err = os.ReadFile(p); err != nil {
			return fmt.Errorf("reading theme asset %s: %w", f.Origin, err)
		}
		out = append(out, f)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading theme assets in %s: %w", dir, err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Rel < out[j].Rel })
	return out, nil
}
