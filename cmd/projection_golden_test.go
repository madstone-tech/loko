package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

var update = flag.Bool("update", false, "rewrite golden files from current behaviour")

// TestProjectionGolden pins the intermediate value for fixtures that between
// them cover every view kind: landscape, system, container, deployment and
// declared (SC-007).
func TestProjectionGolden(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"two-systems", "declared-views", "deployment-nested", "edge-cases"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(fixtures, name)
			proj, diags, err := usecases.CompileAndProject(context.Background(), hclsource.New(),
				newProseReader(), usecases.CompileRequest{Root: root, BuildVersion: buildVersion()})
			if err != nil || proj == nil {
				t.Fatalf("CompileAndProject: %v %v", err, diags)
			}
			got, err := json.MarshalIndent(proj, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, '\n')
			path := filepath.Join(root, "expected", "projection.json")
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading golden (run with -update to create): %v", err)
			}
			if !bytes.Equal(got, want) {
				actual := filepath.Join(t.ArtifactDir(), "projection.json")
				if err := os.WriteFile(actual, got, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Errorf("%s differs from golden; actual output: %s (keep it with go test -artifacts), then run with -update", path, actual)
			}
		})
	}
}
