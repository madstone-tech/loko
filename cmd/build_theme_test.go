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
}
