package usecases

import (
	"context"
	"sort"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// Concrete test doubles for the render ports (Principle VII: no mocking
// library) and small builders for IR literals.

// fakeBackend records the projection it was given and returns canned
// artifacts.
type fakeBackend struct {
	format    viewmodel.Format
	artifacts []viewmodel.Artifact
	err       error

	mu   sync.Mutex
	got  *viewmodel.Projection
	opts RenderOptions
	runs int
}

func (b *fakeBackend) Format() viewmodel.Format { return b.format }

func (b *fakeBackend) Render(_ context.Context, in *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.got, b.opts = in, opts
	b.runs++
	return b.artifacts, b.err
}

type fakeProse struct {
	files map[string]string // docs value -> text; absent means not found
	reads []string
}

func (p *fakeProse) ReadProse(_ context.Context, _, docs string) (string, bool, error) {
	p.reads = append(p.reads, docs)
	text, ok := p.files[docs]
	return text, ok, nil
}

type fakeTheme struct {
	files []viewmodel.ThemeFile
	err   error
}

func (t *fakeTheme) LoadTheme(context.Context, string) ([]viewmodel.ThemeFile, error) {
	return t.files, t.err
}

type fakeStore struct {
	commits int
	outDir  string
	got     []viewmodel.Artifact
	err     error
}

func (s *fakeStore) Commit(_ context.Context, outDir string, as []viewmodel.Artifact, _ []string) (CommitReport, error) {
	s.commits++
	s.outDir, s.got = outDir, as
	var written []string
	for _, a := range as {
		written = append(written, a.Path)
	}
	return CommitReport{Written: written}, s.err
}

type fakeWatcher struct {
	ch   chan struct{}
	spec WatchSpec
}

func (w *fakeWatcher) Watch(_ context.Context, spec WatchSpec) (<-chan struct{}, error) {
	w.spec = spec
	return w.ch, nil
}

type fakePreview struct {
	mu        sync.Mutex
	published [][]viewmodel.Artifact
	failed    []string
	events    chan string
}

func (p *fakePreview) Publish(as []viewmodel.Artifact) {
	p.mu.Lock()
	p.published = append(p.published, as)
	p.mu.Unlock()
	if p.events != nil {
		p.events <- "publish"
	}
}

func (p *fakePreview) Fail(text string) {
	p.mu.Lock()
	p.failed = append(p.failed, text)
	p.mu.Unlock()
	if p.events != nil {
		p.events <- "fail"
	}
}

// ---- IR literal builders ----

func el(kind arch.ElementKind, name string, parent arch.Address, tags ...string) arch.Element {
	sort.Strings(tags)
	return arch.Element{
		Address: arch.NewElementAddress(kind, name),
		Kind:    kind,
		Name:    name,
		Parent:  parent,
		Tags:    tags,
		Range:   arch.SourceRange{File: string(kind) + ".loko.hcl", StartLine: 1, StartColumn: 1},
	}
}

func rl(source arch.Address, local string, target arch.Address, desc, tech string) arch.Relationship {
	return arch.Relationship{
		Address:     arch.NewRelationshipAddress(source, local),
		Source:      source,
		Target:      target,
		LocalName:   local,
		Description: desc,
		Technology:  tech,
	}
}

// irOf sorts elements and relationships by address, as BuildIR does.
func irOf(elems []arch.Element, rels []arch.Relationship, envs ...arch.Environment) *arch.IR {
	sort.Slice(elems, func(i, j int) bool { return elems[i].Address < elems[j].Address })
	sort.Slice(rels, func(i, j int) bool { return rels[i].Address < rels[j].Address })
	return arch.NewIR(arch.Project{Name: "test"}, elems, rels, envs, nil, nil)
}

// A small shared architecture:
//
//	person p ──▶ container c1 (system s1) ──▶ container c3 (system s2)
//	component k1 (in c1) ──▶ c3, k1 ──▶ k1 (self), c1 ──▶ c2 (both in s1)
var (
	aP  = arch.NewElementAddress(arch.KindPerson, "p")
	aS1 = arch.NewElementAddress(arch.KindSystem, "s1")
	aS2 = arch.NewElementAddress(arch.KindSystem, "s2")
	aC1 = arch.NewElementAddress(arch.KindContainer, "c1")
	aC2 = arch.NewElementAddress(arch.KindContainer, "c2")
	aC3 = arch.NewElementAddress(arch.KindContainer, "c3")
	aK1 = arch.NewElementAddress(arch.KindComponent, "k1")
)

func sampleIR() *arch.IR {
	return irOf(
		[]arch.Element{
			el(arch.KindPerson, "p", ""),
			el(arch.KindSystem, "s1", ""),
			el(arch.KindSystem, "s2", "", "pci"),
			el(arch.KindContainer, "c1", aS1),
			el(arch.KindContainer, "c2", aS1),
			el(arch.KindContainer, "c3", aS2),
			el(arch.KindComponent, "k1", aC1),
		},
		[]arch.Relationship{
			rl(aP, "uses", aC1, "uses", "HTTPS"),
			rl(aC1, "calls", aC3, "calls", "gRPC"),
			rl(aK1, "also", aC3, "also", "gRPC"),
			rl(aC1, "x", aC2, "x", ""),
			rl(aK1, "self", aK1, "self", ""),
		},
	)
}
