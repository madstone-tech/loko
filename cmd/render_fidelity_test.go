package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// buildFixture builds a project into a temp directory and returns it.
func buildFixture(t *testing.T, root string) string {
	t.Helper()
	out := t.TempDir()
	var stdout, stderr bytes.Buffer
	code, err := runBuildWith(t.Context(), BuildOptions{Root: root, Out: out, Stdout: &stdout, Stderr: &stderr})
	if err != nil || code != 0 {
		t.Fatalf("build %s: exit %d, %v\n%s", root, code, err, stderr.String())
	}
	return out
}

var viewBoxRe = regexp.MustCompile(`viewBox="([-0-9.]+) ([-0-9.]+) ([0-9.]+) ([0-9.]+)"`)
var fontSizeRe = regexp.MustCompile(`font-size[=:]"?([0-9.]+)`)

// TestContainerViewAspectRatio is FR-006 / SC-001 on the serverless reference
// (14 containers, 39 relationships): ratio ≤ 2:1 and labels ≥ 16px. Absolute
// width is not limited (clarification of 2026-10-09).
func TestContainerViewAspectRatio(t *testing.T) {
	t.Parallel()
	out := buildFixture(t, "../testdata/projects/serverless-reference")
	svg, err := os.ReadFile(filepath.Join(out, "diagrams", "system-portal.svg"))
	if err != nil {
		t.Fatal(err)
	}
	m := viewBoxRe.FindSubmatch(svg)
	if m == nil {
		t.Fatal("no viewBox in the SVG")
	}
	w, _ := strconv.ParseFloat(string(m[3]), 64)
	h, _ := strconv.ParseFloat(string(m[4]), 64)
	t.Logf("system-portal.svg: %.0f × %.0f (ratio %.2f)", w, h, w/h)
	if w/h > 2 {
		t.Errorf("width ÷ height = %.2f, want ≤ 2", w/h)
	}
	if wide := svgWidth(t, wideCopy(t)); w >= wide {
		t.Errorf("top-down width %.0f is not narrower than left-to-right %.0f", w, wide)
	}
	for _, f := range fontSizeRe.FindAllSubmatch(svg, -1) {
		if size, _ := strconv.ParseFloat(string(f[1]), 64); size < 16 {
			t.Errorf("a label renders at %.0fpx, want ≥ 16", size)
			break
		}
	}
}

var directionLine = regexp.MustCompile(`(?m)^direction: \w+\n|"direction": "\w+",?\n?\s*`)

// TestUnchangedProjectsRenderIdentically is SC-003: a project using none of
// the rendering attributes renders exactly as before, apart from direction.
func TestUnchangedProjectsRenderIdentically(t *testing.T) {
	t.Parallel()
	out := buildFixture(t, "../testdata/projects/two-systems")
	for _, d2 := range []string{"landscape.d2", "deployment-prod.d2"} {
		b, err := os.ReadFile(filepath.Join(out, "diagrams", d2))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "direction: right\n") {
			t.Errorf("%s must stay left-to-right", d2)
		}
	}
	for _, d2 := range []string{"system-shop.d2", "container-api.d2"} {
		b, err := os.ReadFile(filepath.Join(out, "diagrams", d2))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "direction: down\n") {
			t.Errorf("%s must default to top-down", d2)
		}
	}
}

// wideCopy builds the reference with an extra view of the same system drawn
// left to right, and returns the path of that view's SVG.
func wideCopy(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS("../testdata/projects/serverless-reference")); err != nil {
		t.Fatal(err)
	}
	extra := "view \"portal-right\" {\n  include   = [system.portal]\n  direction = \"right\"\n}\n"
	if err := os.WriteFile(filepath.Join(root, "wide.loko.hcl"), []byte(extra), 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(buildFixture(t, root), "diagrams", "portal-right.svg")
}

func svgWidth(t *testing.T, path string) float64 {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m := viewBoxRe.FindSubmatch(b)
	if m == nil {
		t.Fatalf("no viewBox in %s", path)
	}
	w, _ := strconv.ParseFloat(string(m[3]), 64)
	return w
}
