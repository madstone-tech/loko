package hclsource

import (
	"os"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestResolveDocs covers each row of the docs resolution rule (feature 018,
// contracts/language.md): from the declaring file first, the project root as
// a fallback, never outside the root.
func TestResolveDocs(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"architecture/docs/near.md":   "near",
		"architecture/docs/both.md":   "near copy",
		"docs/both.md":                "root copy",
		"architecture/docs/rooted.md": "",
		"top.md":                      "top",
	})
	r, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	tests := []struct {
		name, file, docs string
		path             string
		via              arch.DocsVia
		cands            [2]string
	}{
		{"file", "architecture/a.loko.hcl", "docs/near.md", "architecture/docs/near.md", arch.DocsFile,
			[2]string{"architecture/docs/near.md", "docs/near.md"}},
		{"root", "architecture/a.loko.hcl", "architecture/docs/rooted.md", "architecture/docs/rooted.md", arch.DocsRoot,
			[2]string{"architecture/architecture/docs/rooted.md", "architecture/docs/rooted.md"}},
		{"both", "architecture/a.loko.hcl", "docs/both.md", "architecture/docs/both.md", arch.DocsBoth,
			[2]string{"architecture/docs/both.md", "docs/both.md"}},
		{"missing", "architecture/a.loko.hcl", "docs/none.md", "", arch.DocsMissing,
			[2]string{"architecture/docs/none.md", "docs/none.md"}},
		{"leaves the root", "a.loko.hcl", "../outside.md", "", arch.DocsMissing,
			[2]string{"../outside.md", "../outside.md"}},
		{"parent within the root", "architecture/a.loko.hcl", "../top.md", "top.md", arch.DocsFile,
			[2]string{"top.md", "../top.md"}},
		{"declared at the root", "main.loko.hcl", "architecture/docs/near.md", "architecture/docs/near.md", arch.DocsFile,
			[2]string{"architecture/docs/near.md", "architecture/docs/near.md"}},
		{"declared at the root, as written", "main.loko.hcl", "./top.md", "./top.md", arch.DocsFile,
			[2]string{"top.md", "top.md"}},
	}
	for _, tt := range tests {
		path, via, cands := resolveDocs(r, tt.file, tt.docs)
		if path != tt.path || via != tt.via || cands != tt.cands {
			t.Errorf("%s: resolveDocs(%q, %q) = %q, %s, %q; want %q, %s, %q",
				tt.name, tt.file, tt.docs, path, via, cands, tt.path, tt.via, tt.cands)
		}
	}
}

// TestLoadResolvesDocs: Load records the outcome on every element with docs.
func TestLoadResolvesDocs(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"main.loko.hcl":              "project \"p\" {}\n",
		"payments/payments.loko.hcl": "system \"payments\" {\n  docs = \"docs/overview.md\"\n}\n",
		"payments/docs/overview.md":  "# Payments\n",
	})
	m, diags, err := New().Load(t.Context(), root)
	if err != nil || len(diags) != 0 {
		t.Fatal(err, diags)
	}
	e := m.Elements[0]
	if e.DocsPath != "payments/docs/overview.md" || e.DocsVia != arch.DocsFile || e.Docs != "docs/overview.md" {
		t.Errorf("element docs: %q %s (declared %q)", e.DocsPath, e.DocsVia, e.Docs)
	}
}
