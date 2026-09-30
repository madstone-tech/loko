package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestBuildDeclaredViews is US2 (FR-003..FR-005) end to end.
func TestBuildDeclaredViews(t *testing.T) {
	t.Parallel()
	root := filepath.Join(fixtures, "declared-views")
	out := t.TempDir()
	r := runBuildIn(t, root, out, "d2,svg")
	if r.code != usecases.ExitSuccess {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}

	path := readString(t, filepath.Join(out, "diagrams", "payment-path.d2"))
	if strings.Contains(path, `"container__ledger"`) {
		t.Error("payment-path draws the excluded ledger (AC2)")
	}
	if !strings.Contains(path, `-> "outside"`) {
		t.Error("payment-path's connections to the ledger do not reach the boundary (AC2)")
	}
	if _, err := os.Stat(filepath.Join(out, "diagrams", "payment-path.svg")); err != nil {
		t.Error(err)
	}

	pci := readString(t, filepath.Join(out, "diagrams", "pci-only.d2"))
	for _, want := range []string{`"system__payments"`, `"container__gateway"`} {
		if !strings.Contains(pci, want) {
			t.Errorf("pci-only lacks tagged element %s (AC3)", want)
		}
	}

	if _, err := os.Stat(filepath.Join(out, "diagrams", "nothing.d2")); !os.IsNotExist(err) {
		t.Error("an empty declared view produced a file (AC4)")
	}
	if !strings.Contains(r.stderr, "View selects nothing") || !strings.Contains(r.stderr, "view.nothing") {
		t.Errorf("no warning names the empty view (AC4):\n%s", r.stderr)
	}

	land := readString(t, filepath.Join(out, "diagrams", "landscape.d2"))
	if !strings.Contains(land, `"person__customer"`) || strings.Contains(land, `"system__shop"`) {
		t.Errorf("landscape should be the declared view with only the customer (AC5):\n%s", land)
	}
	if !strings.Contains(r.stderr, "View replaces an automatic view") {
		t.Errorf("no warning names the shadowed automatic view (AC5):\n%s", r.stderr)
	}

	if strict := runBuildStrict(t, root, t.TempDir()); strict != usecases.ExitWarnings {
		t.Errorf("--strict exit = %d, want %d", strict, usecases.ExitWarnings)
	}
}
