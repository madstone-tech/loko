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
