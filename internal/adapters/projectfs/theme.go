package projectfs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
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

// LoadTheme returns every .gohtml, .css and .js file directly in
// <root>/templates/, sorted by name. A missing directory means the built-in
// theme is used, with no warning (US7/AC2). Unknown names are returned too,
// so the build can reject them rather than silently ignore a typo.
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
	return out, nil
}
