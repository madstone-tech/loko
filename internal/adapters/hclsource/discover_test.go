package hclsource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// writeTree materialises a map of relative path -> contents under a temp dir.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
	return root
}

// TestDiscoverMergesNestedDirectories covers FR-001: every *.loko.hcl beneath
// the root, at any depth, is discovered and merged into one architecture.
func TestDiscoverMergesNestedDirectories(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"arch.loko.hcl":                  "",
		"systems/payments.loko.hcl":      "",
		"systems/deep/nested/x.loko.hcl": "",
	})

	got, diags, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if diags.HasErrors() {
		t.Fatalf("unexpected errors: %v", diags)
	}

	want := []string{
		"arch.loko.hcl",
		"systems/deep/nested/x.loko.hcl",
		"systems/payments.loko.hcl",
	}
	assertPaths(t, got, want)
}

// TestDiscoverIgnoresOtherFiles covers FR-002.
func TestDiscoverIgnoresOtherFiles(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"arch.loko.hcl":      "",
		"docs/payments.md":   "# prose",
		"main.tf":            "",
		"notes.hcl":          "", // .hcl but not *.loko.hcl
		"loko.hcl":           "", // no leading name segment
		"dist/out.d2":        "",
		".hidden/x.loko.hcl": "",
	})

	got, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	assertPaths(t, got, []string{"arch.loko.hcl"})
}

// TestDiscoverSkipsGeneratedAndVCSDirs keeps a generated dist/ or a .git/
// checkout from being read back in as authored source.
func TestDiscoverSkipsGeneratedAndVCSDirs(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"arch.loko.hcl":              "",
		"dist/generated.loko.hcl":    "",
		".git/hooks/sneaky.loko.hcl": "",
		"vendor/dep/v.loko.hcl":      "",
		"node_modules/n.loko.hcl":    "",
	})

	got, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	assertPaths(t, got, []string{"arch.loko.hcl"})
}

// TestDiscoverNoSourceFound covers FR-005: an empty project gets a clear,
// distinct message rather than silently succeeding on an empty model.
func TestDiscoverNoSourceFound(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{"docs/readme.md": "nothing here"})

	got, diags, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want no files", got)
	}
	if len(diags) != 1 || diags[0].Code != arch.CodeNoSourceFound {
		t.Fatalf("want one %s diagnostic, got %v", arch.CodeNoSourceFound, diags)
	}
	if diags[0].Severity != arch.SeverityError {
		t.Errorf("no_source_found severity = %v, want error", diags[0].Severity)
	}
}

// TestDiscoverIsSorted underpins FR-040: discovery order must not vary with
// the file system, or two machines would produce different exports.
func TestDiscoverIsSorted(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"zeta.loko.hcl":     "",
		"alpha.loko.hcl":    "",
		"m/mid.loko.hcl":    "",
		"beta.loko.hcl":     "",
		"a/a/deep.loko.hcl": "",
	})

	got, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Rel >= got[i].Rel {
			t.Fatalf("not sorted at %d: %q then %q", i, got[i-1].Rel, got[i].Rel)
		}
	}
}

// TestDiscoverMissingRoot reserves the error return for failures that make the
// whole run impossible, per the ArchitectureSource contract.
func TestDiscoverMissingRoot(t *testing.T) {
	t.Parallel()

	if _, _, err := Discover(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("Discover on a missing root returned nil error, want an error")
	}
}

func assertPaths(t *testing.T, got []SourceFile, want []string) {
	t.Helper()
	if len(got) != len(want) {
		var rels []string
		for _, f := range got {
			rels = append(rels, f.Rel)
		}
		t.Fatalf("got %d files %v, want %d %v", len(got), rels, len(want), want)
	}
	for i := range want {
		if got[i].Rel != want[i] {
			t.Errorf("file %d: got %q, want %q", i, got[i].Rel, want[i])
		}
	}
}
