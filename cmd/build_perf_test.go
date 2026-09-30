package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// writeLargeProject generates systems×containers×components elements across
// one file per system, each container calling the next.
func writeLargeProject(t *testing.T, systems, containers, components int) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "project.loko.hcl"), []byte(`project "large" {}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for s := range systems {
		var b strings.Builder
		fmt.Fprintf(&b, "system \"s%d\" {\n  description = \"System %d\"\n}\n", s, s)
		for c := range containers {
			fmt.Fprintf(&b, "container \"s%dc%d\" {\n  system = system.s%d\n  description = \"Container %d of system %d\"\n", s, c, s, c, s)
			if c+1 < containers {
				fmt.Fprintf(&b, "  uses \"next\" { target = container.s%dc%d }\n", s, c+1)
			} else if s+1 < systems {
				fmt.Fprintf(&b, "  uses \"next\" { target = container.s%dc0 }\n", s+1)
			}
			b.WriteString("}\n")
			for k := range components {
				fmt.Fprintf(&b, "component \"s%dc%dk%d\" {\n  container = container.s%dc%d\n", s, c, k, s, c)
				if k > 0 {
					fmt.Fprintf(&b, "  uses \"prev\" { target = component.s%dc%dk%d }\n", s, c, k-1)
				}
				b.WriteString("}\n")
			}
		}
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("s%d.loko.hcl", s)), []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestPerformanceBuild is SC-006: 1,000 elements build in under 10 s, and a
// single-description edit rebuilt through the serve path — the same backend
// instances, so the SVG cache is warm — is reflected in under 2 s.
func TestPerformanceBuild(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("performance test: skipped under -short and -race, whose overhead invalidates wall-clock budgets")
	}
	root := writeLargeProject(t, 20, 10, 4) // 20 + 200 + 800 = 1,020 elements
	deps := newBuildDeps()
	req := usecases.BuildRequest{Root: root, BuildVersion: buildVersion(), Formats: usecases.SupportedFormats(deps.Backends)}

	start := time.Now()
	res, err := usecases.BuildArtifacts(context.Background(), deps, req)
	if err != nil || res.Diags.HasErrors() {
		t.Fatalf("build: %v %v", err, res.Diags)
	}
	full := time.Since(start)
	t.Logf("1,020 elements: %d artifacts in %v", len(res.Artifacts), full)
	// SC-006's 10 s full-build bound is a developer-hardware bound. On hosted
	// CI runners (≈3× slower per core) the build is dominated by d2's own
	// layout engine, so CI records the time instead of failing on it; the
	// edit-to-browser budget below stays strict everywhere (research R2).
	if full > 10*time.Second {
		if os.Getenv("CI") == "true" {
			t.Logf("NOTE: full build took %v, over the 10s developer-hardware budget; not enforced on CI runners", full)
		} else {
			t.Errorf("full build took %v, budget 10s (SC-006)", full)
		}
	}

	f := filepath.Join(root, "s3.loko.hcl")
	src := readString(t, f)
	if err := os.WriteFile(f, []byte(strings.Replace(src, `"Container 4 of system 3"`, `"Container 4 of system 3, edited"`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	start = time.Now()
	serveReq := req
	serveReq.Formats = []string{"html"}
	if _, err := usecases.BuildArtifacts(context.Background(), deps, serveReq); err != nil {
		t.Fatal(err)
	}
	edit := time.Since(start)
	t.Logf("single-description rebuild (serve path, warm cache): %v", edit)
	// The watcher adds at most ~2×200 ms before the rebuild starts.
	if edit+400*time.Millisecond > 2*time.Second {
		t.Errorf("edit reflected in %v + poll latency, budget 2s (SC-006)", edit)
	}
}

// TestPerformanceWideView is the "very wide view" edge case: one system with
// 200 containers renders within the bound, and every container is labelled.
func TestPerformanceWideView(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("performance test: skipped under -short and -race, whose overhead invalidates wall-clock budgets")
	}
	root := writeLargeProject(t, 1, 200, 0)
	out := t.TempDir()
	start := time.Now()
	if r := runBuildIn(t, root, out, "svg"); r.code == usecases.ExitErrors {
		t.Fatalf("build failed:\n%s", r.stderr)
	}
	t.Logf("200-container view rendered in %v", time.Since(start))
	svg := readString(t, filepath.Join(out, "diagrams", "system-s0.svg"))
	for c := range 200 {
		if !strings.Contains(svg, fmt.Sprintf(">s0c%d<", c)) {
			t.Errorf("container s0c%d has no label in the wide view", c)
			break
		}
	}
}
