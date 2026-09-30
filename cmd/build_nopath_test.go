package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestBuildWithEmptyPath is US6/AC1 and SC-005: every format builds with no
// executable search path at all. It runs in-process rather than starting the
// binary: no layer, tests included, may import os/exec (research R15), and a
// build that needed an external program would fail here just the same.
func TestBuildWithEmptyPath(t *testing.T) {
	t.Setenv("PATH", "")
	t.Setenv("HOME", t.TempDir())
	out := t.TempDir()
	r := runBuildIn(t, filepath.Join(fixtures, "two-systems"), out)
	if r.code != usecases.ExitSuccess {
		t.Fatalf("exit = %d with an empty PATH\n%s", r.code, r.stderr)
	}
	for _, id := range []string{"landscape", "system-shop", "container-api", "deployment-prod"} {
		if _, err := os.Stat(filepath.Join(out, "diagrams", id+".svg")); err != nil {
			t.Errorf("SVG not rendered without PATH: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "index.html")); err != nil {
		t.Error(err)
	}
}
