package hclsource

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/adapters/markdown"
	"github.com/madstone-tech/loko/internal/adapters/projectfs"
	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// reversedSource wraps the real source and reverses every slice of the model
// it returns, simulating source files discovered in a different order.
type reversedSource struct{ inner usecases.ArchitectureSource }

func (r reversedSource) Load(ctx context.Context, root string) (*arch.SourceModel, arch.Diagnostics, error) {
	m, diags, err := r.inner.Load(ctx, root)
	if m == nil {
		return m, diags, err
	}
	slices.Reverse(m.Files)
	slices.Reverse(m.Elements)
	for i := range m.Elements {
		slices.Reverse(m.Elements[i].Relations)
	}
	slices.Reverse(m.Environments)
	for i := range m.Environments {
		reverseGroups(m.Environments[i].Groups)
		slices.Reverse(m.Environments[i].Groups)
		slices.Reverse(m.Environments[i].Instances)
	}
	slices.Reverse(m.Views)
	return m, diags, err
}

func reverseGroups(gs []arch.GroupDecl) {
	for i := range gs {
		slices.Reverse(gs[i].Instances)
		reverseGroups(gs[i].Groups)
		slices.Reverse(gs[i].Groups)
	}
}

// renderBackends lists every real backend.
func renderBackends() []usecases.Backend {
	return []usecases.Backend{d2.NewSourceBackend(), d2.NewSVGBackend(), markdown.New(), html.New()}
}

// TestRenderDeterministicDiscoveryOrder is US3/AC2 and SC-002: the artifact
// set does not depend on the order source files and declarations were found
// in. It lives in the adapter layer because the reversing wrapper names
// arch.SourceModel, which cmd/ may not import (Constitution, Dependency
// Direction).
func TestRenderDeterministicDiscoveryOrder(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "..", "testdata", "projects")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "case-collision" {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			t.Parallel()
			build := func(src usecases.ArchitectureSource) *usecases.BuildResult {
				backends := renderBackends()
				res, err := usecases.BuildArtifacts(context.Background(), usecases.BuildDeps{
					Source: src, Prose: projectfs.Prose{}, Backends: backends,
				}, usecases.BuildRequest{Root: filepath.Join(root, e.Name()), BuildVersion: "1.0.0",
					Formats: usecases.SupportedFormats(backends)})
				if err != nil {
					t.Fatal(err)
				}
				return res
			}
			normal, reversed := build(New()), build(reversedSource{New()})
			if len(normal.Artifacts) != len(reversed.Artifacts) {
				t.Fatalf("%d artifacts vs %d reversed", len(normal.Artifacts), len(reversed.Artifacts))
			}
			for i, a := range normal.Artifacts {
				b := reversed.Artifacts[i]
				if a.Path != b.Path || !bytes.Equal(a.Bytes, b.Bytes) {
					t.Errorf("%s differs when sources are discovered in reverse order", a.Path)
				}
			}
		})
	}
}
