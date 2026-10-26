package outputdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfine(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir()) // macOS: /var is a symlink
	if err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	s := New()
	for _, out := range []string{"dist", "out/site", ""} {
		got, err := s.Confine(t.Context(), root, out)
		if err != nil {
			t.Errorf("%q: %v", out, err)
			continue
		}
		rel, _ := filepath.Rel(root, got)
		if strings.HasPrefix(rel, "..") {
			t.Errorf("%q resolved outside the root: %s", out, got)
		}
		if out == "" && filepath.Base(got) != "dist" {
			t.Errorf("default is %s, want dist", got)
		}
	}
	for _, out := range []string{"../x", outside, "escape/site"} {
		if _, err := s.Confine(t.Context(), root, out); err == nil || !strings.Contains(err.Error(), "outside the project") {
			t.Errorf("%q: %v, want a refusal", out, err)
		}
	}
}

// TestConfineDotRoot: the CLI's default root is ".", the case 1.0.2 fixed in
// hclsource.checkPath. Not parallel: t.Chdir changes the working directory.
func TestConfineDotRoot(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	got, err := New().Confine(t.Context(), ".", "dist")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("Confine returned %q, want an absolute path", got)
	}
}
