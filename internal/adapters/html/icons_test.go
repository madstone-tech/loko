package html

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// TestIconCopiesMatch keeps the site's icons identical to the diagram's
// (feature 020 FR-011: one icon table). Adapters may not import each other,
// so the files are copied; this test catches a copy drifting.
func TestIconCopiesMatch(t *testing.T) {
	t.Parallel()
	d2Dir := filepath.Join("..", "d2", "icons")
	theirs, err := os.ReadDir(d2Dir)
	if err != nil {
		t.Fatal(err)
	}
	ours, err := fs.ReadDir(builtinTheme, "theme/icons")
	if err != nil {
		t.Fatal(err)
	}
	var a, b []string
	for _, e := range theirs {
		a = append(a, e.Name())
	}
	for _, e := range ours {
		b = append(b, e.Name())
	}
	if !slices.Equal(a, b) {
		t.Fatalf("icon files differ: d2 %v, html %v", a, b)
	}
	for _, name := range a {
		want, err := os.ReadFile(filepath.Join(d2Dir, name))
		if err != nil {
			t.Fatal(err)
		}
		got, err := fs.ReadFile(builtinTheme, "theme/icons/"+name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("theme/icons/%s differs from internal/adapters/d2/icons/%s", name, name)
		}
	}
}

// TestIconsCSSCoversTable: every icon the table can name has a rule.
func TestIconsCSSCoversTable(t *testing.T) {
	t.Parallel()
	css := renderSite(t, loadProjection(t), nil)["assets/icons.css"]
	names := []string{vm.FallbackIcon}
	for _, n := range vm.ShapeIcons {
		names = append(names, n)
	}
	for _, n := range vm.KindIcons {
		names = append(names, n)
	}
	for _, n := range names {
		if !strings.Contains(css, `.icon[data-icon="`+n+`"] { --icon: url("data:image/svg+xml;base64,`) {
			t.Errorf("icons.css has no rule for %q", n)
		}
	}
}
