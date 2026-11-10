package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
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
	runBuildIn(t, dagreFixture(t, "two-systems"), plain, "html")

	root := dagreFixture(t, "two-systems")
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
	branded := dagreFixture(t, "two-systems")
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
	r := runBuildIn(t, dagreFixture(t, "two-systems"), t.TempDir(), "html")
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

// TestBuildThemeAssets: theme assets are published under assets/theme/,
// pruned when removed, and an invalid asset fails the build without writing
// anything (contract T-4).
func TestBuildThemeAssets(t *testing.T) {
	t.Parallel()
	root := dagreFixture(t, "two-systems")
	mark := "<svg xmlns=\"http://www.w3.org/2000/svg\"><rect/></svg>\n"
	writeTheme(t, root, map[string]string{"assets/img/mark.svg": mark, "assets/style.css": "body{}"})
	out := t.TempDir()
	r := runBuildIn(t, root, out, "html")
	if r.code != usecases.ExitSuccess {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	if got := readString(t, filepath.Join(out, "assets", "theme", "img", "mark.svg")); got != mark {
		t.Errorf("published asset = %q, want the theme bytes %q", got, mark)
	}
	if got := readString(t, filepath.Join(out, "assets", "theme", "style.css")); got != "body{}" {
		t.Errorf("assets/theme/style.css = %q, want the theme bytes", got)
	}
	if css := readString(t, filepath.Join(out, "assets", "style.css")); !strings.Contains(css, "--font-ui") {
		t.Error("an asset named style.css replaced the built-in stylesheet")
	}
	if got, want := lineAfter(r.stdout, "built "), "theme: templates/ (assets/: 2 files)"; got != want {
		t.Errorf("theme line = %q, want %q", got, want)
	}

	if err := os.Remove(filepath.Join(root, "templates", "assets", "img", "mark.svg")); err != nil {
		t.Fatal(err)
	}
	res, err := usecases.Build(t.Context(), newBuildDeps(), usecases.BuildRequest{Root: root, Formats: []string{"html"}, OutDir: out})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(res.Report.Removed, "assets/theme/img/mark.svg") {
		t.Errorf("removed = %v, want assets/theme/img/mark.svg", res.Report.Removed)
	}
	if _, err := os.Stat(filepath.Join(out, "assets", "theme", "img", "mark.svg")); err == nil {
		t.Error("a removed asset is still published")
	}

	invalid := map[string]string{"templates/assets/link.svg": "link"}
	if err := os.Symlink(filepath.Join(root, "main.loko.hcl"), filepath.Join(root, "templates", "assets", "link.svg")); err != nil {
		t.Fatal(err)
	}
	if caseSensitive(t, root) {
		writeTheme(t, root, map[string]string{"assets/A.svg": "<svg/>", "assets/a.svg": "<svg/>"})
		invalid["templates/assets/a.svg"] = "case"
	}
	before := hashTree(t, out)
	r = runBuildIn(t, root, out, "d2")
	if r.code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", r.code, usecases.ExitErrors)
	}
	for file := range invalid {
		if !strings.Contains(r.stderr, file) || !strings.Contains(r.stderr, "Theme override is invalid") {
			t.Errorf("the error must name %s:\n%s", file, r.stderr)
		}
	}
	if !reflect.DeepEqual(before, hashTree(t, out)) {
		t.Error("an invalid theme asset changed the output directory")
	}
}

// caseSensitive reports whether the file system under dir tells A from a.
func caseSensitive(t *testing.T, dir string) bool {
	t.Helper()
	p := filepath.Join(dir, "CaseProbe")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(p)
	_, err := os.Stat(filepath.Join(dir, "caseprobe"))
	return err != nil
}
