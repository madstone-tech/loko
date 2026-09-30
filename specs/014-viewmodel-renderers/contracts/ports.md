# Contract: new ports (`internal/core/usecases/ports.go`)

**Feature**: `014-viewmodel-renderers` | **Status**: normative

All new ports live in `ports.go` (Principle II). Adapters implement them, and wiring happens in
`cmd/` through a single composition helper per command. Types from `viewmodel` and `arch` are the
only ones that cross these interfaces.

```go
// Backend turns a projection into the bytes of one output format.
// It MUST read nothing but its arguments: no file system, no IR, no source (FR-011).
// It MUST NOT import or call another backend (FR-012).
// It MUST make no styling decision of its own (FR-008).
// Output MUST be a pure function of (in, opts): the same arguments give identical bytes (FR-021).
type Backend interface {
    Format() viewmodel.Format
    Render(ctx context.Context, in *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error)
}

// RenderOptions carries the non-architectural inputs a backend may need.
type RenderOptions struct {
    Theme []viewmodel.ThemeFile // sorted by Name; only the html backend reads it
}

// ProseReader returns the text of a docs reference, resolved against the project
// root with the same containment rule as the docs_not_found warning.
// found=false is not an error (FR-026).
type ProseReader interface {
    ReadProse(ctx context.Context, root, docs string) (text string, found bool, err error)
}

// ThemeSource returns the override files in <root>/templates/, sorted by name.
// A missing directory returns (nil, nil) (US7/AC2).
type ThemeSource interface {
    LoadTheme(ctx context.Context, root string) ([]viewmodel.ThemeFile, error)
}

// ArtifactStore commits a complete artifact set to an output directory,
// pruning files the previous build owned and nothing else (FR-023).
// It MUST write nothing if it cannot write everything.
type ArtifactStore interface {
    // sources feed the notice on .loko-manifest.
    Commit(ctx context.Context, outDir string, artifacts []viewmodel.Artifact, sources []string) (CommitReport, error)
}

type CommitReport struct {
    Written, Unchanged, Removed []string // each sorted, project-relative
}

// ChangeWatcher signals once per settled burst of changes to the watched set.
type ChangeWatcher interface {
    Watch(ctx context.Context, spec WatchSpec) (<-chan struct{}, error)
}

type WatchSpec struct {
    Root       string
    ExtraFiles func() []string // prose files from the last good compile; re-read on each tick
    Exclude    []string        // absolute directories never to watch, such as the output directory
}
// templates/ under Root is always watched.

// PreviewServer publishes build results to connected browsers.
type PreviewServer interface {
    Publish(artifacts []viewmodel.Artifact)  // switches to the ok state and broadcasts a reload
    Fail(diagnosticsText string)             // switches to the error state and broadcasts a reload
}
```

## Use cases

| Function | File | Responsibility |
|---|---|---|
| `ResolveViews(ir) ([]viewmodel.View, arch.Diagnostics)` | `resolve_views.go` | Derived views, declared views, shadowing (FR-001..006). |
| `SelectDeclared(ir, v) []arch.Address` | `select_declared.go` | include ∪ tags − exclude, with descendants. |
| `ProjectView(ir, view, prov) viewmodel.ViewModel` | `project_view.go`, `project_lift.go`, `project_deployment.go` | Visible set, nesting, lifting, boundary edges, styles (FR-007..010). |
| `ProjectPages(ir, prose, views) []viewmodel.ElementPage` | `project_pages.go` | Element pages and tables (FR-025, FR-026). |
| `Project(ir, prose, prov) (*viewmodel.Projection, arch.Diagnostics)` | `project.go` | Assembles the above. Pure (FR-010). |
| `BuildArtifacts(ctx, deps, req) (*BuildResult, error)` | `build_artifacts.go` | Compile → gate on errors → read prose and theme → project → expand formats → run backends in parallel → collision check → sorted artifact set. |
| `Build(ctx, deps, req) (*BuildResult, error)` | `build_site.go` | `BuildArtifacts`, then `ArtifactStore.Commit` only when there are no errors. The single call `cmd/build.go` makes. |
| `Serve(ctx, deps, req) error` | `serve_site.go` | Initial build, then a loop over watcher signals: build, then `Publish` or `Fail`. |

Every use-case function stays ≤ 60 effective lines and every file ≤ 200, per Constitution v1.3.0.
`BuildResult` carries `Artifacts`, `Diags`, `AddedFormats`, and an `ExitCode(strict)` method that
mirrors `CompileResult`.

## Implementation notes (2026-09-30)

- `ThemeError{File, Line, Message}` is in `ports.go`. The html backend returns it for a malformed
  override, and `BuildArtifacts` turns it into a `theme_invalid` diagnostic.
- `Serve` takes `ServeDeps{Build, Watcher, Preview, Describe}`. `Describe` renders diagnostics as
  text, and the caller supplies it, so core stays free of presentation.
  `usecases.Diagnostics = arch.Diagnostics` is an alias, so `cmd/` can name the type without
  importing entities.
- `CompileAndProject` exposes the projection to outer layers. The projection goldens in `cmd/` use
  it.
