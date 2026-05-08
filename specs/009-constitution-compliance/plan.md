# Implementation Plan: Constitution Compliance Refactor

**Branch**: `009-constitution-compliance` | **Date**: 2026-05-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-constitution-compliance/spec.md`

## Summary

Bring the loko codebase into compliance with the architectural constitution by (a) decomposing oversized CLI handlers (`cmd/new.go`, `cmd/build.go`) into core use cases, (b) thinning the oversized MCP tool handlers, (c) splitting oversized core files (`entities/graph.go`, `usecases/query_architecture.go`, etc.) so every file fits its budget, and (d) replacing the existing per-file shell audit with a richer mechanical check that verifies layer-import rules and per-function size budgets. The check runs locally via `make audit-constitution` and as a gating step in `.github/workflows/ci.yml`. The constitution itself is amended to v1.1.0 to reflect the tightened limits (CLI handler functions ≤ 50 effective lines, MCP tool handler functions ≤ 30, use-case files ≤ 200, entity files ≤ 300). External behaviour — command names, flags, exit codes, MCP tool names, schemas, and produced files — is preserved exactly.

## Technical Context

**Language/Version**: Go 1.25+ (matches `go.mod`)
**Primary Dependencies**: cobra + viper (CLI), lipgloss (TUI), fsnotify (watcher), d2 CLI (behind `DiagramRenderer` port), ason (behind `TemplateEngine` port), TOON v3 (behind `OutputEncoder` port). For this feature: a Go AST-based custom audit tool (no new third-party runtime dependency); `golangci-lint` (already present in the toolchain via `make lint`) is extended with a `depguard` rule for layer-import enforcement.
**Storage**: filesystem only (`loko.toml`, `relationships.toml`, `*.md`, `dist/*.d2` artefacts). Refactor adds no storage.
**Testing**: Go standard `testing` package, table-driven tests, concrete mock implementations of ports (no mocking libraries). Coverage target: > 80% on `internal/core/` (constitution requirement preserved).
**Target Platform**: macOS / Linux developer machines and Linux CI runners (single static Go binary `loko`).
**Project Type**: Single Go module with CLI (`cmd/`), MCP server (`internal/mcp/`), HTTP API server (`internal/api/`), and core domain (`internal/core/{entities,usecases}` + `internal/adapters/`). No frontend, no separate backend.
**Performance Goals**: Audit tool finishes in < 5 s for the current ~10.5 k LOC codebase so it does not slow down `make lint` or CI; refactor itself is a no-op for runtime performance (FR-015 demands behavioural equivalence).
**Constraints**: Behavioural equivalence to pre-refactor for all external surfaces (FR-015); pre-existing test suite must pass without weakening or removing tests (FR-016); test coverage of touched handlers must not decrease (FR-017); the audit must produce file-path + offender + rule for every violation (FR-014).
**Scale/Scope**: ~10.5 k LOC; ~20 cmd files (`*.go` + `*_cobra.go` pairs); ~28 MCP tool files; ~20 use cases; ~19 entity files. Five files are currently above the new budgets and four are at the boundary; the `KNOWN_VIOLATIONS` allowlist tracks seven of them today.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution (`.specify/memory/constitution.md` v1.0.0, ratified 2026-02-06) defines five gates relevant to this feature:

| Gate | Constitution rule | Verdict for this feature |
|------|-------------------|--------------------------|
| **Layered architecture** | Entities → use cases → adapters → outer; each layer has explicit allowed imports. | **PASS.** Feature does not change the architecture; it makes it mechanically verifiable. |
| **Thin outer wrappers** | "CLI commands, MCP tools, and API handlers are thin wrappers that delegate to use cases." | **PASS (re-asserted).** Feature pulls scaffolding and build orchestration out of `cmd/` into `usecases/scaffold_*.go` and `usecases/build_docs.go`. |
| **Test coverage > 80% on `internal/core/`** | Mandatory per principle. | **PASS.** FR-016/FR-017 require no decrease in coverage; new use-case packages inherit existing tests and add `*_test.go` for split units. |
| **Handler line counts** (CLI < 150, MCP < 100, per-file effective lines) | Mandatory per principle. | **AMENDMENT REQUIRED.** This feature tightens the limits to per-function (CLI ≤ 50, MCP ≤ 30) and adds per-file budgets for use cases (≤ 200) and entities (≤ 300). Tracked in **Complexity Tracking** below. |
| **ADR for new architectural decisions** | Required for any new architectural decision. | **GATED ON NEW ADR.** This feature introduces (a) a custom Go AST audit tool, (b) `depguard` import-rule layer enforcement, and (c) tightened limits. A single ADR `docs/adr/0009-constitution-compliance-tooling.md` is part of the deliverables (see Phase 1). |

**Pre-design gate result: PASS, conditional on the constitution amendment and ADR being included as deliverables.** Both are listed in Phase 1; the gate is re-evaluated at the end of Phase 1.

## Project Structure

### Documentation (this feature)

```text
specs/009-constitution-compliance/
├── plan.md                      # This file
├── spec.md                      # /speckit.specify output (already written)
├── research.md                  # Phase 0 output
├── data-model.md                # Phase 1 output
├── quickstart.md                # Phase 1 output
├── contracts/
│   ├── structural-rules.yaml    # The rule set the audit tool consumes
│   └── ci-step.contract.md      # Contract for the gating CI step
├── checklists/
│   └── requirements.md          # Already created by /speckit.specify
└── tasks.md                     # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

The repository already follows the constitutional layout. The refactor relocates code within it; no new top-level directories are created beyond `tools/archcheck/` for the audit binary.

```text
cmd/                                  # Outer layer: CLI entry points (cobra)
├── new.go                            # Refactor target: 356 → ≤ 50 lines per func
├── new_cobra.go                      # Flag wiring (data-only; exempt under data-file rule)
├── build.go                          # Refactor target: 232 → ≤ 50 lines per func
├── build_cobra.go
├── validate.go, watch.go, mcp.go, serve.go, api.go, init.go, root.go,
│   export_cobra.go, completion_cobra.go, ...

internal/
├── core/
│   ├── entities/                     # ≤ 300 lines per file
│   │   ├── graph.go                  # Refactor target: 678 → split into graph/, graph_edges.go, graph_query.go
│   │   ├── system.go (255), component.go (236), project.go (233),
│   │   ├── relationship.go (220), search.go (178), container.go (175), ...
│   ├── usecases/                     # ≤ 200 lines per file
│   │   ├── scaffold_entity.go        # Existing 364-line file; split into:
│   │   │   ├── scaffold_project.go
│   │   │   ├── scaffold_system.go
│   │   │   ├── scaffold_container.go
│   │   │   └── scaffold_component.go
│   │   ├── build_docs.go             # 479 → split into build_docs.go (orchestration) +
│   │   │                             #   build_docs_d2.go (renderer step) +
│   │   │                             #   build_docs_markdown.go (renderer step)
│   │   ├── query_architecture.go     # 527 → split by query type
│   │   ├── build_architecture_graph.go (469), ports.go (426),
│   │   ├── validate_architecture.go (368), search_elements.go (240),
│   │   ├── render_markdown_docs.go (213), ...
│   └── (no `ports/` subdir — `ports.go` lives directly in `usecases/`; refactor splits it)
├── adapters/                         # Implementations of ports (cobra-free, mcp-free)
│   ├── d2/                           # DiagramRenderer
│   ├── ason/                         # TemplateEngine
│   ├── toon/                         # OutputEncoder
│   ├── fs/                           # filesystem ProjectRepository
│   └── ...
├── mcp/
│   ├── server.go, graph_cache.go
│   └── tools/                        # ≤ 30 lines per handler function
│       ├── graph_tools.go            # Already flagged: split into 3 files (one per tool)
│       ├── create_component.go (preview logic → use case)
│       ├── create_system.go, update_system.go, build_docs.go, ...
│       └── schemas.go, registry.go, helpers.go (data-files; exempt)
└── api/                              # HTTP handlers (≤ 50 lines per handler function)

scripts/
└── audit-constitution.sh             # Wraps the new Go binary; preserved as the entry point
                                      # invoked by Makefile + CI; KNOWN_VIOLATIONS allowlist removed.

tools/
└── archcheck/                        # NEW: custom Go AST audit binary
    ├── main.go                       # Reads contracts/structural-rules.yaml,
    ├── filesize.go                   #   walks the module, emits structured violations.
    ├── funcsize.go                   # Per-function effective-line counter (Go AST).
    ├── layer.go                      # Layer-import checker.
    └── archcheck_test.go

docs/
├── adr/
│   └── 0009-constitution-compliance-tooling.md   # NEW ADR
└── superpowers/specs/2026-05-08-loko-production-design.md  # Source-of-truth for the limits

.github/workflows/
└── ci.yml                            # Add `make audit-constitution` as a gating step

.specify/memory/
└── constitution.md                   # AMENDED to v1.1.0 (new limits, ADR ref, migration note)

Makefile                              # Add `audit-constitution` target wrapping `tools/archcheck` + golangci-lint depguard
.golangci.yml                         # NEW (or extended): `depguard` rule encoding the layer table
```

**Structure Decision**: Single Go module, no subprojects. The repository already implements the constitution's layered layout; this feature *enforces* the layout rather than introducing a new one. The only new top-level directory is `tools/archcheck/` for the audit binary, which is itself outside the constitution's enforcement scope (it is build-time tooling, not product code).

## Phase 0 — Research summary

Five open decisions resolved in [research.md](./research.md):

1. **Audit tool choice**: custom Go AST binary (`tools/archcheck`) instead of `go-arch-lint` or pure shell. Reason: per-function effective-line counting requires AST anyway, and we already need a Go program to keep the rule data in `contracts/structural-rules.yaml` declarative.
2. **Layer-import enforcement**: `golangci-lint` `depguard` linter, already wired through `make lint`, plus a redundant check in `archcheck` so the rules can be evaluated independently of `golangci-lint` version drift.
3. **Effective-line counting**: `archcheck` reuses the existing rule (drop blank lines, comments, `import (...)` blocks, `package` decl) and applies it at function granularity using `go/ast` token positions; output is byte-equivalent to `scripts/audit-constitution.sh` for whole-file counts so the migration is verifiable.
4. **Pre-existing oversized non-handler files** (graph.go 678, query_architecture.go 527, build_docs.go 479, build_architecture_graph.go 469, ports.go 426, validate_architecture.go 368, scaffold_entity.go 364): all are in scope. Each gets a dedicated split task in Phase 2; the refactor is purely structural (move types, keep public API stable).
5. **`KNOWN_VIOLATIONS` allowlist mechanism**: removed at the end of the refactor. The new audit either passes cleanly or fails CI; an exemption mechanism remains, but only for `*_test.go` files and the generated `*_cobra.go` Cobra flag-wiring files (treated as data-files, same as `schemas.go`/`registry.go`/`helpers.go`/`constants.go` today).

## Phase 1 — Design artefacts

- [data-model.md](./data-model.md): the rule entities (`LayerRule`, `FileSizeRule`, `FunctionSizeRule`, `Exemption`, `Violation`) consumed and emitted by `archcheck`.
- [contracts/structural-rules.yaml](./contracts/structural-rules.yaml): declarative rule set; the single source of truth for limits, layer table, and exemption patterns. Amending this file is the only way to change limits.
- [contracts/ci-step.contract.md](./contracts/ci-step.contract.md): contract that the CI gating step must satisfy (exit codes, output format, file-name conventions for the violation report).
- [quickstart.md](./quickstart.md): developer-facing instructions — how to run the audit locally, how to interpret a violation, how to add a legitimate exemption, how to amend the rules.
- [docs/adr/0009-constitution-compliance-tooling.md](../../docs/adr/0009-constitution-compliance-tooling.md): records the audit-tool decision, the depguard choice, and the constitution amendment rationale.

### Agent context update

After writing the artefacts, run:

```bash
.specify/scripts/bash/update-agent-context.sh claude
```

This refreshes `CLAUDE.md` with the additions: Go AST audit tool, `depguard` layer rule, amended constitution v1.1.0, new `make audit-constitution` target.

## Constitution Re-Check (post-design)

| Gate | Verdict |
|------|---------|
| Layered architecture | **PASS.** Layer table in `structural-rules.yaml` mirrors the constitution table verbatim. |
| Thin outer wrappers | **PASS.** Plan moves all CLI orchestration into use cases (`scaffold_*.go`, `build_docs.go`); MCP tools become protocol adapters only. |
| Test coverage > 80% | **PASS.** Refactor moves bodies but preserves tests; new split files inherit and extend existing `*_test.go` partners. |
| Handler line counts | **PASS, after amendment.** Constitution v1.1.0 (delivered as part of this feature) replaces "CLI < 150 / MCP < 100 / per-file" with "CLI handler funcs ≤ 50, MCP tool funcs ≤ 30, usecase files ≤ 200, entity files ≤ 300, per-function for handlers, per-file for use cases and entities". Migration plan = the refactor itself. |
| ADR for new architectural decisions | **PASS.** ADR-0009 is a deliverable. |

Post-design gate: **PASS.**

## Complexity Tracking

> Filled because the feature amends the constitution.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Constitution amendment from v1.0.0 to v1.1.0 (tighter limits, per-function for handlers, new file budgets for usecase/entity) | The feature description in `spec.md` cites the production-design document `docs/superpowers/specs/2026-05-08-loko-production-design.md` (Constitution Compliance section) as the authority for the tighter limits. The whole point of the feature is to make those limits real. Without amendment, the audit tool would either over-enforce (failing CI on rules not in the constitution) or under-enforce (the prose limits in the spec would diverge from the executable rules). | Keeping the v1.0.0 limits and shipping the audit tool only: rejected because it would not deliver the spec's success criteria (SC-002 / SC-003 explicitly require ≤ 50 / ≤ 30). Putting the new limits only in `structural-rules.yaml` without amending the constitution: rejected because the constitution is the canonical reference and out-of-band rule overrides are an anti-pattern called out in the constitution itself ("ADR written for any new architectural decision"). |
| New `tools/archcheck/` Go binary (build-time tooling outside the layered architecture) | Per-function effective-line counting requires Go AST traversal, which is not expressible in shell or in `golangci-lint` config alone. Custom Go is also the right place to keep the layer-import logic, the file-size logic, and the exemption mechanism in one declarative-rule-driven engine. | Reusing `gocyclo`/`funlen` from `golangci-lint`: those measure cyclomatic complexity / non-comment-non-blank lines but not the project's specific "effective lines" definition (which excludes `import (...)` blocks and `package` decls), so they would diverge from `scripts/audit-constitution.sh` and the migration would be impossible to verify. Keeping the shell script: rejected because per-function granularity requires structured parsing. |
| Pre-existing oversized non-handler files (graph.go 678, ports.go 426, etc.) refactored in this feature even though `spec.md` only names `cmd/new.go` and `cmd/build.go` | FR-003 / FR-004 in `spec.md` apply to *every* use-case and entity file. The spec's "known violations" list is illustrative, not exhaustive (per Assumption: "the refactor will discover the precise list during execution by running the size check, and will fix all of them"). | Deferring those splits to a follow-up feature: rejected because SC-001 ("zero violations across the entire codebase") is a measurable success criterion of *this* feature; deferring would mean either failing SC-001 or weakening the audit so as not to detect the pre-existing files — both unacceptable. |
