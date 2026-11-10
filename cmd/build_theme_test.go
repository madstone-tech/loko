package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestBuildTheme is US7 end to end: an override from a directory outside the
// tool is used (SC-009), diagrams are untouched, and a malformed override
// fails the build naming the file without writing anything (FR-035).
func TestBuildTheme(t *testing.T) {
	t.Parallel()
	plain := t.TempDir()
	runBuildIn(t, filepath.Join(fixtures, "two-systems"), plain, "html")

	root := copyFixture(t, "two-systems")
	if err := os.MkdirAll(filepath.Join(root, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	layout := filepath.Join(root, "templates", "layout.gohtml")
	if err := os.WriteFile(layout, []byte(`{{define "header"}}<header class="acme">ACME</header>{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	themed := t.TempDir()
	if r := runBuildIn(t, root, themed, "html"); r.code != usecases.ExitSuccess {
		t.Fatalf("themed build exit = %d\n%s", r.code, r.stderr)
	}
	for _, f := range withPrefix(listFiles(t, themed), "element/") {
		if !strings.Contains(readString(t, filepath.Join(themed, f)), `class="acme"`) {
			t.Errorf("%s does not use the overridden header", f)
		}
	}
	diagrams := func(dir string) map[string][32]byte {
		all := hashTree(t, dir)
		for k := range all {
			if !strings.HasPrefix(k, "diagrams/") {
				delete(all, k)
			}
		}
		return all
	}
	if !reflect.DeepEqual(diagrams(plain), diagrams(themed)) {
		t.Error("a theme override changed the diagrams")
	}

	if err := os.WriteFile(filepath.Join(root, "templates", "partials.gohtml"), []byte(`{{define "nope"}}{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	before := hashTree(t, themed)
	r := runBuildIn(t, root, themed, "html")
	if r.code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", r.code, usecases.ExitErrors)
	}
	if !strings.Contains(r.stderr, "templates/partials.gohtml") || !strings.Contains(r.stderr, `"nope"`) {
		t.Errorf("the error must name the file and the block:\n%s", r.stderr)
	}
	if !reflect.DeepEqual(before, hashTree(t, themed)) {
		t.Error("a malformed theme changed the output directory")
	}

	// The build names the theme it applied (contract T-6).
	branded := copyFixture(t, "two-systems")
	writeTheme(t, branded, map[string]string{
		"custom.css":    ":root { --font-ui: serif; }",
		"layout.gohtml": `{{define "footer"}}<footer>F</footer>{{end}}{{define "brand"}}<div class="brand">B</div>{{end}}`,
	})
	r = runBuildIn(t, branded, t.TempDir(), "html")
	if r.code != usecases.ExitSuccess {
		t.Fatalf("branded build exit = %d\n%s", r.code, r.stderr)
	}
	if got, want := lineAfter(r.stdout, "built "), "theme: templates/ (custom.css; layout.gohtml: brand, footer)"; got != want {
		t.Errorf("line after built = %q, want %q\n%s", got, want, r.stdout)
	}
}

// TestBuildNoThemeLine: a project without templates/ prints no theme line.
func TestBuildNoThemeLine(t *testing.T) {
	t.Parallel()
	r := runBuildIn(t, copyFixture(t, "two-systems"), t.TempDir(), "html")
	if r.code != usecases.ExitSuccess {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	for line := range strings.SplitSeq(r.stdout, "\n") {
		if strings.HasPrefix(line, "theme:") {
			t.Errorf("unthemed build printed %q", line)
		}
	}
}

// writeTheme writes files below root/templates, creating directories.
func writeTheme(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(root, "templates", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// lineAfter returns the line that follows the first line starting with
// prefix, or "" when there is none.
func lineAfter(out, prefix string) string {
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, prefix) && i+1 < len(lines) {
			return lines[i+1]
		}
	}
	return ""
}
