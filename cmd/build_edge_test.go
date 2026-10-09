package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestBuildCaseCollision is FR-028: two outputs that differ only by case are
// an error naming both, and nothing is written.
func TestBuildCaseCollision(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "dist")
	r := runBuildIn(t, filepath.Join(fixtures, "case-collision"), out)
	if r.code != usecases.ExitErrors {
		t.Fatalf("exit = %d, want %d\n%s", r.code, usecases.ExitErrors, r.stderr)
	}
	if !strings.Contains(r.stderr, "container.Api") || !strings.Contains(r.stderr, "container.api") {
		t.Errorf("the collision must name both elements:\n%s", r.stderr)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Error("output was written despite the collision")
	}
}

// TestBuildEdgeCases covers the spec's edge-case list end to end.
func TestBuildEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nothing to draw", func(t *testing.T) {
		t.Parallel()
		out := filepath.Join(t.TempDir(), "dist")
		r := runBuildIn(t, filepath.Join(fixtures, "empty"), out)
		if r.code != usecases.ExitSuccess || !strings.Contains(r.stdout, "nothing to draw: the architecture declares no elements") {
			t.Errorf("exit=%d stdout=%q", r.code, r.stdout)
		}
		if _, err := os.Stat(filepath.Join(out, "diagrams")); !os.IsNotExist(err) {
			t.Error("an empty architecture produced diagrams")
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		out := t.TempDir()
		runBuildIn(t, filepath.Join(fixtures, "single-element"), out, "d2")
		if got := withPrefix(listFiles(t, out), "diagrams/"); strings.Join(got, ",") != "diagrams/landscape.d2" {
			t.Errorf("diagrams = %v, want only the landscape", got)
		}
	})

	t.Run("self-loop, cycle, unsafe name, hollow system", func(t *testing.T) {
		t.Parallel()
		out := t.TempDir()
		r := runBuildIn(t, filepath.Join(fixtures, "edge-cases"), out, "d2,html")
		if r.code == usecases.ExitErrors {
			t.Fatalf("build failed:\n%s", r.stderr)
		}
		land := readString(t, filepath.Join(out, "diagrams", "landscape.d2"))
		for _, edge := range []string{
			`"system___5fOdd_5fname-1" -> "system___5fOdd_5fname-1"`, // self-loop drawn, not dropped
			`"system__a" -> "system__b"`, `"system__b" -> "system__c"`, `"system__c" -> "system__a"`,
		} {
			if !strings.Contains(land, edge) {
				t.Errorf("landscape lacks %s\n%s", edge, land)
			}
		}
		if _, err := os.Stat(filepath.Join(out, "element", "system", "_Odd_name-1.html")); err != nil {
			t.Errorf("unsafe name not mapped to its escaped file name: %v", err)
		}
		if _, err := os.Stat(filepath.Join(out, "diagrams", "system-hollow.d2")); !os.IsNotExist(err) {
			t.Error("a system without containers got a view")
		}
	})

	t.Run("deep nesting", func(t *testing.T) {
		t.Parallel()
		out := t.TempDir()
		runBuildIn(t, filepath.Join(fixtures, "deployment-nested"), out, "d2")
		prod := readString(t, filepath.Join(out, "diagrams", "deployment-prod.d2"))
		if !strings.Contains(prod, `          "deployment__prod__node__l1__l2__l3__l4__l5": "l5" {`) {
			t.Errorf("five nested levels are not all shown:\n%s", prod)
		}
	})
}
