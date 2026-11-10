package projectfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var _ usecases.ThemeSource = Theme{}

func TestLoadThemeMissingDirectory(t *testing.T) {
	t.Parallel()
	files, err := Theme{}.LoadTheme(context.Background(), t.TempDir())
	if files != nil || err != nil {
		t.Fatalf("no templates/ directory: got %v, %v; want nil, nil (US7/AC2)", files, err)
	}
}

func TestLoadTheme(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "templates")
	for name, body := range map[string]string{
		"style.css": "a{}", "layout.gohtml": "{{define \"header\"}}x{{end}}", "site.js": "//",
		"README.md": "ignored", "sytle.css": "typo, still returned so it can be rejected",
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	files, err := Theme{}.LoadTheme(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name+"@"+f.Origin)
	}
	want := []string{"layout.gohtml@templates/layout.gohtml", "site.js@templates/site.js",
		"style.css@templates/style.css", "sytle.css@templates/sytle.css"}
	if len(names) != len(want) {
		t.Fatalf("LoadTheme = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("LoadTheme[%d] = %s, want %s", i, names[i], want[i])
		}
	}
}

// TestLoadThemeAssets: every file below templates/assets/ is returned with
// its path and unchanged bytes, and a symbolic link is returned unread, so
// validation refuses it (contract T-4).
func TestLoadThemeAssets(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "templates")
	files := map[string]string{
		"custom.css": ".x{}", "notes.txt": "ignored",
		"assets/mark.svg": "<svg/>", "assets/img/deep/logo.png": "\x89PNG\x00raw", "assets/fonts/OFL.txt": "licence",
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, "custom.css"), filepath.Join(dir, "assets", "link.css")); err != nil {
		t.Fatal(err)
	}
	got, err := Theme{}.LoadTheme(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		name, rel, origin, bytes string
		link                     bool
	}
	var rows []row
	for _, f := range got {
		rows = append(rows, row{f.Name, f.Rel, f.Origin, string(f.Bytes), f.Link})
	}
	want := []row{
		{"custom.css", "", "templates/custom.css", ".x{}", false},
		{"OFL.txt", "assets/fonts/OFL.txt", "templates/assets/fonts/OFL.txt", "licence", false},
		{"logo.png", "assets/img/deep/logo.png", "templates/assets/img/deep/logo.png", "\x89PNG\x00raw", false},
		{"link.css", "assets/link.css", "templates/assets/link.css", "", true},
		{"mark.svg", "assets/mark.svg", "templates/assets/mark.svg", "<svg/>", false},
	}
	if len(rows) != len(want) {
		t.Fatalf("LoadTheme = %+v, want %+v", rows, want)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("LoadTheme[%d] = %+v, want %+v", i, rows[i], want[i])
		}
	}
}
