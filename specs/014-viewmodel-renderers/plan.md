# Implementation Plan: ViewModel Renderers

**Branch**: `014-viewmodel-renderers` | **Date**: 2026-09-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/014-viewmodel-renderers/spec.md`

**Stage**: 2 of 6 in the [v1.0 roadmap](../012-v1-architecture-dsl/roadmap.md). It depends on
[`013-hcl-compiler-core`](../013-hcl-compiler-core/plan.md), which is merged.

## Summary

This feature brings back `loko build` and `loko serve`, now as pure projections of the compiled IR.
The pipeline has a new middle stage:

```
IR ──ResolveViews──▶ []View ──Project──▶ Projection{Views []ViewModel, Pages []ElementPage}
                                              │
                    ┌─────────────┬───────────┼─────────────┐
                    ▼             ▼           ▼             ▼
                 d2 backend   svg backend  md backend   html backend   ──▶ []Artifact ──Commit──▶ dist/
```

Views are derived with no configuration: a landscape, one per system with containers, one per
container with components, and one per environment. Declared `view` blocks join them and win on a
name collision. Each view projects to a `ViewModel` that has already decided visibility, nesting,
lifted edges, boundary edges and style. Backends see nothing else.

The technical approach turns on five decisions, all taken in [research.md](./research.md):

1. **d2 runs in-process.** `oss.terrastruct.com/d2` v0.7.1 comes back as a library with the embedded
   dagre layout (ELK was measured and rejected as too slow for SC-006, R2). The `exec.LookPath("d2")` adapter is deleted, so no process is ever started (R1).
2. **One projection algorithm for every view kind.** A visible set, lifting to the nearest visible
   ancestor, and a single synthetic `outside` node for boundary connections (R4).
3. **Format dependencies are declared in core, not taken between backends.** `html` and `md`
   require `svg`, and the build expands the set. No backend imports another (R5).
4. **The output directory is owned through a manifest.** Commits are staged, identical bytes are
   skipped, only previously owned files are pruned, and a failed build leaves the directory as it
   was (R8).
5. **Watch is polling, not fsnotify.** fsnotify is on the constitution's "removed in v1.0" list, and
   a 200 ms stat loop with a one-tick settle meets the 2 s budget on every file system (R12). Serve
   is in-memory, with SSE reload injected only at serve time (R13).

## Technical Context

**Language/Version**: Go 1.25+ (matches `go.mod`; toolchain go1.25.4)

**Primary Dependencies**:
- `oss.terrastruct.com/d2` v0.7.1. **Reintroduced** (feature 013 dropped it). The packages used are
  `d2lib`, `d2layouts/d2dagrelayout`, `d2renderers/d2svg`, and `lib/textmeasure`.
- `github.com/yuin/goldmark` v1.7.17. Promoted from indirect (via d2) to direct, so it adds no new
  module. Raised from v1.7.4 for GO-2026-5320 (research R10).

Both are confined to `internal/adapters/**`. The rest is retained unchanged: cobra, lipgloss,
toon-go, hcl/v2, go-cty, and yaml.v3. The standard library covers `html/template`, `embed`,
`net/http` (SSE), and polling. **Not** added: fsnotify, any websocket library, or a mermaid
renderer.

**Storage**: The file system only, meaning the output directory plus its `.loko-manifest`. There is
no cache on disk. The SVG render cache is in-memory and lives only as long as the process (R2).

**Testing**: `go test` with golden fixtures. There are projection goldens (JSON) under
`testdata/projects/<fixture>/expected/projection.json`, run from `cmd/projection_golden_test.go`
because the parser lives in an adapter that core tests cannot import, and backend goldens under
`internal/adapters/{d2,markdown,html}/testdata/golden/`. End-to-end build tests in `cmd/` run over
the new `testdata/projects/*` fixtures. `-update` regenerates the goldens. Mocks are concrete
structs with no mocking library (Principle VII). A link checker and a notice scanner run over every
fixture build.

**Target Platform**: A single static binary for macOS, Linux and Windows with **no runtime
dependency** (US6). CI adds a `macos-latest` job for the determinism suite (SC-002).

**Project Type**: Compiler + CLI (a single Go module with clean architecture).

**Performance Goals**: A 1,000-element build in under 10 s, and a single-file edit under `serve`
reflected in under 2 s (SC-006). Achieved with parallel per-view layout and a content-hash SVG
cache (R2). It is verified by extending feature 013's performance harness.

**Constraints**: Byte-identical output across runs, machines, and discovery orders (FR-021). No
external process (FR-014). Backends read only the projection (FR-011) and never each other
(FR-012). Exactly three exit codes (FR-038). Adapter files ≤ 400 effective lines, use-case files
≤ 200 and functions ≤ 60, entity files ≤ 200, CLI functions ≤ 50 (Constitution v1.3.0).

**Scale/Scope**: The benchmark covers views up to 200 nodes and projects of 1,000 elements. Four
backends. Two commands. About 14 new use-case files, 1 new entity package (about 7 files), and 6
new or rewritten adapter packages.

No `NEEDS CLARIFICATION` remains. The open points (the view rules for each kind, the backend
interface shape, theme granularity, how watching works, and the notice placement for SVG and HTML)
are resolved in research R3–R13.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design (below).*

| Principle / rule | Status | How this plan satisfies it |
|---|---|---|
| **I. Clean Architecture** | ✅ | `viewmodel` entities are stdlib-only. Projection lives in `usecases`. d2, goldmark, `net/http` and the file system live in adapters. `cmd/` wires. Core imports nothing outward. |
| **II. Interface-First** | ✅ | New ports (`Backend`, `ProseReader`, `ThemeSource`, `ArtifactStore`, `ChangeWatcher`, `PreviewServer`) are all in `ports.go` ([contracts/ports.md](contracts/ports.md)). No use case names a concrete adapter. Wiring happens in `cmd/wiring.go`. The constitution text says "only in `main.go`", which the existing `validate` and `export` commands already contradict, so task T023a amends it to v1.4.0 ("`main.go` or `cmd/`, never `internal/**`") **before** Phase 3. |
| **III. Thin Handlers** | ✅ | `cmd/build.go` and `cmd/serve.go` parse flags, call `usecases.Build` or `usecases.Serve`, and render diagnostics, at ≤ 50 lines per function. Flag wiring is in `*_cobra.go`. |
| **IV. Entity Validation** | ✅ | `viewmodel.ParseFormats`, `(*ViewModel).Validate` (referential invariants), and `Segment` (the path escape) live in entities. Use cases trust constructed values. |
| **V. Test-First** | ✅ | Every phase in tasks opens with failing goldens or table tests: projection goldens before `Project`, backend goldens before each backend, and e2e determinism tests before `Commit`. Core coverage > 80%. |
| **VI. Token Efficiency** | ✅ n/a | Build output is for humans and version control, not LLM context. MCP is untouched, and FR-041 keeps backends out of it. |
| **VII. Simplicity & YAGNI** | ✅ | No feature flags. No mermaid. No disk cache. No fsnotify. The markdown library comes from an existing transitive module. The parked hand-written markdown parser is not restored. |
| Dependency Direction table | ✅ | Enforced by archcheck and depguard, extended with a goldmark confinement, an `os/exec` ban, and an mcp → render-adapter ban (R15). |
| File-size budgets | ✅ | `_parked/html/templates.go` (1,753 lines) becomes embedded `.gohtml`/`.css`/`.js` assets plus Go files each under 400 lines. The split happens during the rework, as the spec requires. |
| Removed in v1.0 list | ✅ | fsnotify, ason, the HTTP API, and veve-cli/PDF are **not** reintroduced. The loopback preview server is an adapter used only by `serve`, not the removed HTTP API (R13, ADR-0013). |
| Quality gates | ✅ | `task lint`, `task test` and `task audit-constitution` must be green. ADR-0013 is written for the new architectural decisions. |

**Governance follow-up**: task T023a makes a **MINOR** amendment (v1.4.0) before any user-story work.
It names `main.go` and `cmd/` as the composition root (Principles I and II). It also fixes stale text:
Principle VII's "no runtime dependencies except d2 (and optionally veve-cli)" becomes "no runtime
dependencies"; the File Organization tree gains the new adapter directories and `entities/viewmodel`;
goldmark joins the dependency table; and the Quality Gates entity budget changes from 300 to 200,
matching Architecture Rules. Decision logic stays in core: `viewmodel.PlanCommit` decides what to
write, skip and prune, and `viewmodel.ValidateTheme` rejects unknown override files. The adapters
only perform I/O, apart from template parse errors, which only the parser can see (analysis finding
K1). The machine-readable mirror in
`specs/009-constitution-compliance/contracts/structural-rules.yaml` is synced in the same PR.

**Post-design re-check (after Phase 1)**: ✅ Still passes. The design introduced no new principle
tension. The largest adapter file is expected to be `adapters/html/render.go` (≈250 lines), and the
largest use-case file `project_view.go` (≈180 lines, split into `project_lift.go` and
`project_deployment.go` to stay under 200).

## Project Structure

### Documentation (this feature)

```text
specs/014-viewmodel-renderers/
├── spec.md
├── plan.md              # this file
├── research.md          # Phase 0: R1–R15
├── data-model.md        # Phase 1: viewmodel entities, diagnostics, paths
├── quickstart.md        # Phase 1: validation scenarios
├── contracts/
│   ├── cli.md           # loko build / loko serve
│   ├── ports.md         # new ports and use-case map
│   ├── output-layout.md # dist/ layout, manifest, notice, determinism
│   └── theme.md         # templates/ override contract
├── checklists/
└── tasks.md             # Phase 2 (/speckit-tasks; not created here)
```

### Source Code (repository root)

```text
internal/core/entities/viewmodel/          # NEW — stdlib only, ≤ 200 lines/file
├── view.go            # View, ViewID, ViewKind, sort order
├── model.go           # ViewModel, Node, Edge, NodeRole, Validate()
├── style.go           # Style, EdgeStyle, Shape, StyleFor() convention table
├── page.go            # ElementPage, LinkRef, RelationRow, Projection, ProjectHeader
├── artifact.go        # Format, ParseFormats, Requires, Artifact, Manifest, ThemeFile
├── notice.go          # generated-file notice text
└── paths.go           # Segment, DiagramFile, ViewPage, ElementPage, Rel

internal/core/entities/arch/diagnostic.go  # EDIT — view_empty, view_shadowed, output_path_collision, theme_invalid

internal/core/usecases/
├── ports.go                 # EDIT — Backend, RenderOptions, ProseReader, ThemeSource, ArtifactStore, ChangeWatcher, PreviewServer
├── resolve_views.go         # NEW — derived + declared views, shadowing
├── select_declared.go       # NEW — include ∪ tags − exclude
├── project.go               # NEW — Project(): assembles Projection
├── project_view.go          # NEW — visible sets per kind, nesting, nodes
├── project_lift.go          # NEW — lifting, edge merge, boundary edges
├── project_deployment.go    # NEW — group tree, instances, instance edges
├── project_pages.go         # NEW — ElementPage tables, diagram choice
├── provenance.go            # NEW — address → source file table from SourceModel
├── build_artifacts.go       # NEW — compile → project → render (parallel) → collisions
├── build_site.go            # NEW — Build(): BuildArtifacts + Commit
├── serve_site.go            # NEW — Serve(): initial build + watch loop
└── performance_test.go      # EDIT — projection timings (render timings live in cmd/)

internal/adapters/
├── d2/                      # REWRITE — renderer.go (exec) deleted
│   ├── emit.go              #   ViewModel → D2 source (shared by both backends)
│   ├── source_backend.go    #   Backend "d2"
│   ├── svg_backend.go       #   Backend "svg": d2lib + dagre + d2svg, worker pool, hash cache
│   └── testdata/golden/
├── markdown/                # NEW — Backend "md"
├── html/                    # NEW (reworked from internal/_parked/html)
│   ├── render.go            #   Backend "html": pages, index, assets
│   ├── theme.go             #   built-in + override resolution, theme_invalid
│   ├── prose.go             #   goldmark (raw HTML off, GFM tables)
│   ├── theme/*.gohtml, style.css, custom.css, site.js   # embedded assets
│   └── testdata/golden/
├── outputdir/               # NEW — ArtifactStore: staging, skip-identical, manifest prune
├── projectfs/               # NEW — ProseReader, ThemeSource
├── watch/                   # NEW — ChangeWatcher: 200 ms poller, one-tick settle
└── devserver/               # NEW — PreviewServer: loopback net/http, SSE, error state

cmd/
├── build.go, build_cobra.go # NEW
├── serve.go, serve_cobra.go # NEW
├── wiring.go                # NEW — backend registry + adapter construction (composition root, per Constitution v1.4.0)
├── projection_golden_test.go # NEW — SC-007 projection goldens over testdata/projects/*
├── build_test.go            # NEW — e2e: determinism, notice scan, prune, local change, empty PATH
├── build_perf_test.go       # NEW — SC-006
└── root_test.go             # EDIT — build/serve present; watch still removed

internal/_parked/html/       # DELETE once reworked; README updated
testdata/projects/           # NEW — two-systems, declared-views, deployment-nested, edge-cases, docs-site, empty, single-element
tools/archcheck/rules.yaml, .golangci.yml     # EDIT — R15 rules
.github/workflows/ci.yml     # EDIT — macos-latest determinism job
docs/adr/0013-viewmodel-renderers.md          # NEW
docs/cli-reference.md, docs/quickstart.md, docs/language.md (view semantics), README.md   # EDIT
.specify/memory/constitution.md               # EDIT — v1.4.0 MINOR (composition root) + stale-text fixes
```

**Structure Decision**: This follows the existing clean-architecture layout. New entity types get
their own package (`entities/viewmodel`) so `arch` stays the IR and nothing else, and the IR's
export shape is untouched. Each backend is its own adapter package, which makes FR-012's
"removing one changes no other" checkable by deletion. The `d2` and `svg` backends share one package
because both are thin layers over the same D2 emitter. Neither calls the other.

## Complexity Tracking

No constitution violations. The table is intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
