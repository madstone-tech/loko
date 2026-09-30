package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildLocalChange is US3/AC4, FR-022 and SC-003: changing one element's
// description leaves unrelated files byte-unchanged.
func TestBuildLocalChange(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "two-systems")
	out := t.TempDir()
	runBuildIn(t, root, out)
	before := hashTree(t, out)

	f := filepath.Join(root, "main.loko.hcl")
	src := readString(t, f)
	edited := strings.Replace(src, `description = "HTTP entry point"`, `description = "HTTP entry point, now with retries"`, 1)
	if edited == src {
		t.Fatal("fixture changed: the component description to edit was not found")
	}
	if err := os.WriteFile(f, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	runBuildIn(t, root, out)
	after := hashTree(t, out)

	// What may change when component.handler's description changes
	// (contracts/output-layout.md § Determinism): its own pages, its parent's
	// pages, the views that draw it, and the search index.
	permitted := map[string]bool{
		"diagrams/container-api.d2": true, "diagrams/container-api.svg": true,
		"view/container-api.html": true, "md/view/container-api.md": true,
		"element/component/handler.html": true, "md/element/component/handler.md": true,
		"element/container/api.html": true, "md/element/container/api.md": true,
		"assets/search-index.js": true,
	}
	changed := 0
	for path, h := range after {
		if before[path] == h {
			continue
		}
		changed++
		if !permitted[path] {
			t.Errorf("%s changed, but it does not depict component.handler", path)
		}
	}
	if unchanged := float64(len(after)-changed) / float64(len(after)); unchanged < 0.9 {
		t.Errorf("%.0f%% of files unchanged (%d of %d changed), want at least 90%% (SC-003)",
			unchanged*100, changed, len(after))
	}
}

// TestBuildPrune is US3/AC5 and FR-023: a deleted element's files go; files
// the tool did not write stay.
func TestBuildPrune(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "two-systems")
	out := t.TempDir()
	runBuildIn(t, root, out)
	if _, err := os.Stat(filepath.Join(out, "diagrams", "container-gateway.svg")); err != nil {
		t.Fatalf("precondition: %v", err)
	}
	notes := filepath.Join(out, "NOTES.txt")
	if err := os.WriteFile(notes, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Delete gateway's two components, the last blocks in the file, so the
	// container loses its container view.
	f := filepath.Join(root, "main.loko.hcl")
	src := readString(t, f)
	cut := strings.Index(src, `component "authorizer"`)
	if cut < 0 {
		t.Fatal("fixture changed: component authorizer not found")
	}
	if err := os.WriteFile(f, []byte(src[:cut]), 0o644); err != nil {
		t.Fatal(err)
	}
	runBuildIn(t, root, out)

	for _, gone := range []string{"diagrams/container-gateway.svg", "diagrams/container-gateway.d2"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(gone))); !os.IsNotExist(err) {
			t.Errorf("%s survived its element's deletion", gone)
		}
	}
	if readString(t, notes) != "mine" {
		t.Error("a file the tool did not write was touched")
	}
}
