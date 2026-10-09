package usecases

import (
	"context"
	"fmt"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// AuthoringDeps is what reading and editing a project needs. Outer layers
// build it once (cmd/wiring.go) and pass it to every read and write.
type AuthoringDeps struct {
	Root         string
	Source       OverlaySource
	Editor       SourceEditor
	BuildVersion string
	// Revisions remembers the revisions handed out, so a write can find the
	// per-file hashes behind a short token. Nil (as for `loko query`) means
	// reads do not remember.
	Revisions *RevisionMemory
}

// AuthoringService applies edits to one project, one write at a time (FR-017):
// concurrent calls queue on the mutex, so each plans against the previous
// one's result instead of racing it.
type AuthoringService struct {
	mu   sync.Mutex
	deps AuthoringDeps
}

// NewAuthoringService returns a service for deps.Root.
func NewAuthoringService(deps AuthoringDeps) *AuthoringService {
	if deps.Revisions == nil {
		deps.Revisions = &RevisionMemory{}
	}
	return &AuthoringService{deps: deps}
}

// Deps returns the dependencies, for the read use cases that share them.
func (s *AuthoringService) Deps() AuthoringDeps { return s.deps }

// EditInput is one edit as an outer layer receives it (the MCP contract's
// Edit object). Apply validates it into an authoring.Edit; outer layers never
// construct entities themselves.
type EditInput struct {
	Op      string         `json:"op"`
	Target  string         `json:"target"`
	Address string         `json:"address"`
	Binding *BindingInput  `json:"binding,omitempty"`
	Set     map[string]any `json:"set,omitempty"`
	Clear   []string       `json:"clear,omitempty"`
	Cascade bool           `json:"cascade,omitempty"`
	To      string         `json:"to,omitempty"`
	File    string         `json:"file,omitempty"`
}

// BindingInput picks a binding inside an instance.
type BindingInput struct {
	Kind  string `json:"kind"`
	Index int    `json:"index,omitempty"`
}

// ApplyRequest is one apply_edit call.
type ApplyRequest struct {
	Edits        []EditInput
	BaseRevision string
	Preview      bool
}

// EditResult is the outcome of a write. OK false means nothing was written.
type EditResult struct {
	OK       bool                   `json:"ok" toon:"ok"`
	Preview  bool                   `json:"preview,omitempty" toon:"preview,omitempty"`
	NoOp     bool                   `json:"noop,omitempty" toon:"noop,omitempty"`
	Files    []authoring.FileChange `json:"files,omitempty" toon:"files,omitempty"`
	Removed  []string               `json:"removed,omitempty" toon:"removed,omitempty"`
	Refusal  *Refusal               `json:"refusal,omitempty" toon:"refusal,omitempty"`
	Diags    arch.Diagnostics       `json:"diagnostics,omitempty" toon:"diagnostics,omitempty"`
	Revision string                 `json:"revision" toon:"revision"`
}

// Refusal says why a write changed nothing, and for which edit in the batch.
type Refusal struct {
	Reason     string   `json:"reason" toon:"reason"`
	Edit       int      `json:"edit" toon:"edit"`
	Detail     string   `json:"detail" toon:"detail"`
	Dependents []string `json:"dependents,omitempty" toon:"dependents,omitempty"`
}

// compile loads the project, through the overlay when one is given, and runs
// the full compiler pipeline on it.
func compileOverlay(ctx context.Context, deps AuthoringDeps, overlay []authoring.FileContent) (*CompileResult, error) {
	src := overlaySource{src: deps.Source, overlay: overlay}
	return CompileArchitecture(ctx, src, CompileRequest{Root: deps.Root, BuildVersion: deps.BuildVersion})
}

// compileIR compiles the project and builds its IR; the IR is nil when the
// project has errors.
func compileIR(ctx context.Context, deps AuthoringDeps) (*arch.IR, *CompileResult, error) {
	res, err := compileOverlay(ctx, deps, nil)
	if err != nil {
		return nil, nil, err
	}
	if res.HasErrors() {
		return nil, res, nil
	}
	return BuildIR(res.Model, res.Resolved), res, nil
}

// currentRevision returns the project's current revision token, remembering
// the revision behind it.
func currentRevision(ctx context.Context, deps AuthoringDeps) (string, error) {
	rev, err := deps.Editor.Revision(ctx, deps.Root)
	if err != nil {
		return "", fmt.Errorf("hashing source: %w", err)
	}
	deps.Revisions.remember(rev)
	return rev.Token(), nil
}

// overlaySource adapts an OverlaySource plus a fixed overlay to the
// ArchitectureSource the compiler takes.
type overlaySource struct {
	src     OverlaySource
	overlay []authoring.FileContent
}

func (o overlaySource) Load(ctx context.Context, root string) (*arch.SourceModel, arch.Diagnostics, error) {
	return o.src.LoadOverlay(ctx, root, o.overlay)
}
