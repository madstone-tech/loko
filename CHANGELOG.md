# Changelog

All notable changes to the loko project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`loko build` returns** (feature 014), rendering entirely from the compiled HCL:
  - D2 and SVG diagrams for a landscape, every system with containers, every container with
    components, and every environment, with no configuration.
  - Declared `view` blocks, rendered alongside the automatic views. A view that selects nothing
    warns with `view_empty`; one that replaces an automatic view warns with `view_shadowed`.
  - Markdown documents and a browsable site with a page per element: prose, uses and used-by
    tables, children, and the relevant diagram.
  - Byte-identical output across runs, machines and discovery orders. Every file carries a
    generated-file notice naming its sources. Owned files are tracked in `dist/.loko-manifest`, so
    only they are pruned.
  - Theme overrides from `<project>/templates/`. A malformed override fails with `theme_invalid`.
- **`loko serve` returns**: an in-memory preview on `127.0.0.1` that rebuilds as you save and
  reloads the browser. When the source does not compile, pages show the diagnostics instead of
  stale output.

### Changed

- **Go 1.27.** Building loko now requires Go 1.27 (`go 1.27.0`, toolchain go1.27.1); CI, releases
  and the container image build with 1.27. Tests that exercise concurrency run on
  `testing/synctest`'s fake clock, and every package that starts goroutines fails on a leaked one
  (Go 1.27 `goroutineleak` profile).
- **No `d2` binary is needed.** Diagrams render in-process with the d2 library (v0.7.1, dagre
  layout). No layer may start a process (`os/exec` is banned by archcheck).
- **Breaking (container image):** the image's entrypoint is now `/usr/local/bin/loko`. Run
  `docker run <image> build`, not `docker run <image> loko build`. The image no longer contains a
  `d2` binary or scaffold templates.
- Constitution v1.4.0: wiring may live in `main.go` or `cmd/`.

### Removed

- `loko watch`: it is part of `loko serve`.

## [0.2.0] - 2026-02-17

### Added

- **Functional Relationship Graph (US1.1 + US1.2)**: Fixed 4 MCP tools that previously returned empty results
  - `find_relationships`, `query_dependencies`, `query_related_components`, `analyze_coupling` now return real data
  - Frontmatter `relationships:` map parsed and added to architecture graph
  - D2 arrow syntax (`source -> target: label`) parsed and merged via concurrent worker pool (10 goroutines)
  - Union merge: frontmatter + D2 edges deduplicated by `sourceQualifiedID -> targetQualifiedID` key
  - `BuildArchitectureGraph` now accepts optional `D2Parser` via `NewBuildArchitectureGraphWithD2()`

- **Technology-Aware Template Selection (US2.1)**: `loko new component` auto-selects content template
  - 7 category-specific templates: `compute`, `datastore`, `messaging`, `api`, `event`, `storage`, `generic`
  - Technology → category mapping (e.g., "AWS Lambda" → compute, "DynamoDB" → datastore)
  - `--template` flag for explicit override
  - `Component.ContentTemplate` transient field propagated through scaffold pipeline

- **D2 Diagram Preview (US2.2)**: Show component position in container diagram after creation
  - `RenderDiagramPreview` use case with graceful degradation (no-op when d2 binary unavailable)
  - `PreviewRenderer` adapter generates minimal C4-conforming D2 snippet
  - `loko new component --preview` flag renders and prints SVG preview after scaffolding
  - MCP `create_component` tool accepts `preview: true` parameter, returns `diagram_preview` in response

- **Auto-Generated Component Lists (US2.3)**: Documentation tables auto-populated during `build_docs`
  - `GenerateComponentTable(container)` — Markdown table of components (Name, Technology, Description)
  - `GenerateContainerTable(system)` — Markdown table of containers (Name, Technology, Description)
  - Both tables are sorted alphabetically, pipe-character escaped, graceful on empty input
  - `{{component_table}}` placeholder in container templates, `{{container_table}}` in system templates
  - Injected during `RenderMarkdownDocs` use case

- **Drift Detection (US3.2)**: Detect inconsistencies between D2 and frontmatter
  - `DetectDrift` use case with `NewDetectDrift` / `NewDetectDriftWithD2` constructors
  - Detects `DriftOrphanedRelationship` (ERROR): frontmatter relationship to non-existent component
  - Detects `DriftDescriptionMismatch` (WARNING): D2 tooltip differs from frontmatter description
  - Detects `DriftMissingComponent` (ERROR): D2 arrow references non-existent component
  - `loko validate --check-drift` flag with severity-aware terminal output
  - Exit code 1 for ERROR-level drift; 0 for warnings-only or clean

- **Test Coverage**: Core package coverage improved from 58.1% to 80.7%
  - New test files for `CreateComponent`, `CreateContainer`, `FindRelationships`, `SearchElements`
  - New test files for `ScaffoldEntity`, `RenderMarkdownDocs`, `UpdateDiagram`, `BuildDocs`
  - Entity tests for `Component` helper methods, `D2Relationship`, `DriftIssue`

- **Documentation**:
  - `docs/guides/relationships.md` — Frontmatter syntax, D2 arrows, union merge, troubleshooting
  - `docs/guides/templates.md` — Technology-to-template mapping, override flag, custom templates
  - `docs/guides/data-model.md` — Source of truth hierarchy, drift detection workflow
  - `docs/cli-reference.md` — Complete CLI reference with new flags
  - Updated `docs/guides/mcp-integration-guide.md` — v0.2.0 fixes for relationship tools

### Fixed

- `find_relationships`, `query_dependencies`, `query_related_components`, `analyze_coupling` MCP tools
  returning empty results in all cases (relationship graph was not being populated from D2 or frontmatter)

## [Unreleased]

### Added

- **Architecture Graph Improvements**: Comprehensive improvements to graph implementation
  - Qualified hierarchical node IDs to prevent collisions in multi-system projects
  - O(1) performance for dependency queries via IncomingEdges and ChildrenMap
  - Thread-safe GraphCache for MCP session optimization
  - Type-safe graph operations with C4Entity interface
  - Comprehensive documentation in ADR-0004

### Changed

- **BREAKING**: Node IDs now use qualified hierarchical format
  - **Before**: `"auth"` (short ID, collision-prone)
  - **After**: `"backend/api/auth"` (qualified ID, unique)
  - Migration path: Use ShortIDMap for backward compatibility
  - See `docs/migration-001-graph-qualified-ids.md` for migration guide

- **Performance**: Dependency queries optimized from O(E) to O(1)
  - `GetIncomingEdges()`: 68.56 ns/op (1,000,000x faster)
  - `GetChildren()`: 101.8 ns/op (5,000x faster)
  - `AnalyzeDependencies()`: 171.125µs (11,600x faster)

- **Type Safety**: Replaced runtime type assertions with compile-time checks
  - `GraphNode.Data` changed from `any` to `C4Entity` interface
  - MCP tool arguments use typed structs instead of `map[string]any`
  - Zero runtime type assertions in core/usecases package

- **Validation**: Architecture validation now filters to components only
  - Isolated component checks skip systems/containers
  - High coupling checks skip systems/containers
  - Prevents false positives for non-component entities

### Fixed

- **Critical Bug**: Node ID collisions in multi-system projects
  - Multiple systems with identically-named components no longer overwrite each other
  - Example: `backend/api/auth` and `admin/ui/auth` now coexist without data loss

- **Thread Safety**: Clarified ArchitectureGraph concurrency model
  - Graph is immutable after construction
  - GraphCache provides thread-safe concurrent access
  - Documentation prevents misuse patterns

### Performance

Benchmark results on Apple M3 Pro:
- `GetIncomingEdges`: 68.56 ns/op
- `GetChildren`: 101.8 ns/op
- Target: <1ms for all operations (achieved: 68ns avg)

### Documentation

- Added ADR-0004: Architecture Graph Conventions
  - Node ID format and collision prevention
  - Thread safety model and lifecycle
  - Relationship scope (component-level only)
  - Migration guidance for existing projects

- Enhanced godoc with examples
  - `AddNode()`: Qualified ID usage examples
  - `AddEdge()`: Component relationship examples
  - `GetDependencies()`: Dependency query examples

- Package-level documentation
  - Thread safety guarantees
  - Graph lifecycle stages
  - Immutability contract

### Migration Guide

See `docs/migration-001-graph-qualified-ids.md` for step-by-step migration from short IDs to qualified IDs.

**Breaking changes affect**:
- Projects with multiple systems using same component names
- MCP tools relying on short ID lookups
- Custom code directly accessing graph node IDs

**Backward compatibility**:
- `ShortIDMap` enables short ID resolution
- `ResolveID()` handles unambiguous short IDs
- No changes needed for single-system projects

---

## [0.1.0] - Previous Release

Initial release with core features.

[Unreleased]: https://github.com/madstone-tech/loko/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/madstone-tech/loko/releases/tag/v0.1.0
