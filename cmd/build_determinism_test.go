package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// buildableFixtures are the fixture projects that build successfully.
func buildableFixtures(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "case-collision" {
			out = append(out, e.Name())
		}
	}
	return out
}

// TestBuildDeterministic is US3/AC1 and SC-002: two builds of an unchanged
// architecture are byte-identical, file for file. CI runs it on Linux and
// macOS.
func TestBuildDeterministic(t *testing.T) {
	t.Parallel()
	for _, name := range buildableFixtures(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(fixtures, name)
			a, b := t.TempDir(), t.TempDir()
			for _, out := range []string{a, b} {
				if r := runBuildIn(t, root, out); r.code == usecases.ExitErrors {
					t.Fatalf("build failed:\n%s", r.stderr)
				}
			}
			if ha, hb := hashTree(t, a), hashTree(t, b); !reflect.DeepEqual(ha, hb) {
				t.Errorf("two builds differ:\n%v\n%v", listFiles(t, a), listFiles(t, b))
			}
		})
	}
}
