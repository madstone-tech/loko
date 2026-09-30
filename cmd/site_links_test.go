package cmd

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var linkAttr = regexp.MustCompile(`(?:href|src)="([^"]*)"`)

// TestSiteLinksResolve is FR-027 and SC-008: every internal link and image in
// every fixture site resolves to a file the build produced.
func TestSiteLinksResolve(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"two-systems", "declared-views", "deployment-nested", "edge-cases"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out := t.TempDir()
			if r := runBuildIn(t, filepath.Join(fixtures, name), out, "html"); r.code == usecases.ExitErrors {
				t.Fatalf("build failed:\n%s", r.stderr)
			}
			pages, checked := 0, 0
			for _, f := range listFiles(t, out) {
				if !strings.HasSuffix(f, ".html") {
					continue
				}
				pages++
				for _, m := range linkAttr.FindAllStringSubmatch(readString(t, filepath.Join(out, filepath.FromSlash(f))), -1) {
					href := m[1]
					if href == "" || strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "#") {
						continue
					}
					href, _, _ = strings.Cut(href, "#")
					target := path.Clean(path.Join(path.Dir(f), href))
					checked++
					if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(target))); err != nil {
						t.Errorf("%s: broken link %q (resolves to %s)", f, m[1], target)
					}
				}
			}
			if pages == 0 || checked == 0 {
				t.Fatalf("checked %d links on %d pages; the site is missing", checked, pages)
			}
		})
	}
}
