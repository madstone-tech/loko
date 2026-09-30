package projectfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var _ usecases.ProseReader = Prose{}

// TestReadProse mirrors the containment rule of the docs_not_found warning
// (usecases.docsExist): a file inside the root is read; a missing file, a
// directory, or a path escaping the root is "not found", never an error.
func TestReadProse(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	root := filepath.Join(parent, "proj")
	if err := os.MkdirAll(filepath.Join(root, "docs", "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "a.md"), []byte("# A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "secret.md"), []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		docs  string
		text  string
		found bool
	}{
		{"./docs/a.md", "# A\n", true},
		{"docs/a.md", "# A\n", true},
		{"docs/missing.md", "", false},
		{"docs/dir", "", false},
		{"../secret.md", "", false},
		{"docs/../../secret.md", "", false},
	}
	for _, tt := range tests {
		text, found, err := Prose{}.ReadProse(context.Background(), root, tt.docs)
		if err != nil || text != tt.text || found != tt.found {
			t.Errorf("ReadProse(%q) = %q, %v, %v; want %q, %v, nil", tt.docs, text, found, err, tt.text, tt.found)
		}
	}
}
