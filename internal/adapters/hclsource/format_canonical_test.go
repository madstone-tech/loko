package hclsource

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCanonical: each non-canonical file with its current and canonical
// bytes, sorted by path; canonical and unparsable files absent; nothing
// written (feature 017, R9).
func TestCanonical(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"b.loko.hcl":      "system \"b\" {\ndescription=\"x\"\n}\n",
		"a.loko.hcl":      "system \"a\" {\n  description   = \"x\"\n}\n",
		"ok.loko.hcl":     "system \"ok\" {\n  description = \"x\"\n}\n",
		"broken.loko.hcl": "system \"broken\" {\n",
	})
	before, _ := os.ReadFile(filepath.Join(root, "b.loko.hcl"))
	files, diags, err := NewFormatter().Canonical(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Path != "a.loko.hcl" || files[1].Path != "b.loko.hcl" {
		t.Fatalf("files %+v", files)
	}
	if string(files[1].Old) != string(before) || string(files[1].New) != "system \"b\" {\n  description = \"x\"\n}\n" {
		t.Errorf("b: old %q new %q", files[1].Old, files[1].New)
	}
	if !diags.HasErrors() {
		t.Error("the unparsable file is not reported")
	}
	after, _ := os.ReadFile(filepath.Join(root, "b.loko.hcl"))
	if string(after) != string(before) {
		t.Error("Canonical wrote a file")
	}
}
