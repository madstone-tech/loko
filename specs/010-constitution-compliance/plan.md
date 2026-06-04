# Implementation Plan: Constitution Compliance Refactor

**Branch**: `010-constitution-compliance` | **Date**: 2026-05-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/010-constitution-compliance/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command.

## Summary

This feature operationalises the loko constitution by (a) refactoring the two CLI handlers (`cmd/new.go`, `cmd/build.go`) and the oversized MCP tool handlers under `internal/mcp/tools/` so that every handler function fits within its per-function effective-line budget, (b) extracting the displaced logic into narrowly-scoped use-case files under `internal/core/usecases/` while keeping every core file within its file-size budget, and (c) installing an automated structural-compliance check — a custom Go AST audit binary (`tools/archcheck`) plus a redundant `golangci-lint depguard` configuration — as a gating step in CI so the constitution is enforced mechanically on every pull request. The refactor is behaviour-preserving: command names, flags, exit codes, produced files, and MCP tool input/output schemas are byte-equivalent across the change. A small suppression mechanism (file-scoped, owner-tagged, dated) is added so the gate can land green on `main` without first requiring a separate sweep of pre-existing violations outside the named scope.

## Technical Context

**Language/Version**: Go 1.25+ (matches `go.mod`)
**Primary Dependencies**: cobra + viper (CLI), lipgloss (TUI), fsnotify (watcher), d2 CLI (behind `DiagramRenderer` port), ason (behind `TemplateEngine` port), TOON v3 (behind `OutputEncoder` port). New for this feature: a Go AST-based custom audit tool (`tools/archcheck`, already scaffolded under feature 009 — extended here) and `golangci-lint depguard` rule (already in toolchain via `task lint`). No new third-party runtime dependency.
**Storage**: filesystem only (`loko.toml`, `relationships.toml`, `*.md`, `dist/*.d2` artefacts). Refactor adds no storage. The compliance check uses a small YAML rule file (`specs/010-constitution-compliance/contracts/structural-rules.yaml`) and a suppression file (`.archcheck-suppressions.yaml` at repo root).
**Testing**: `go test ./...` for unit + integration; existing CLI smoke tests under `tests/integration/cli/`; MCP tool tests under `internal/mcp/tools/*_test.go`; new audit-tool tests under `tools/archcheck/*_test.go` (`go test ./tools/archcheck/...`).
**Target Platform**: macOS / Linux developer workstations and CI runners (GitHub Actions Ubuntu); single static Go binary for `loko`; `tools/archcheck` is a separate small build-time binary.
**Project Type**: single repository — Go monorepo with three consumer interfaces (CLI, MCP server, HTTP API) over a shared core (entities + use cases + adapter ports).
**Performance Goals**: structural compliance check runs in **< 30 s** on a contributor laptop (SC-009); per-file AST parse and per-file dependency-graph evaluation amortised across worker goroutines if needed.
**Constraints**: pure refactor — **no observable behaviour change** (FR-015); test coverage of touched packages must not decrease (FR-016); CI gate must produce file-line-rule diagnostics good enough that a contributor unfamiliar with the project can act on the failure without consulting docs (SC-010).
**Scale/Scope**: ~30–50 production files touched across `cmd/`, `internal/mcp/tools/`, `internal/core/usecases/`, `internal/core/entities/`; one new tool package (`tools/archcheck`); one new contract directory (`contracts/`); one new CI step; one constitution amendment (1.1.0 → 1.2.0) to tighten the outer-layer rule (see Constitution Check below).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture (NON-NEGOTIABLE) | ✅ PASS | The refactor *is* the operationalisation of this principle. Moves logic from `cmd/` and `internal/mcp/tools/` into `internal/core/usecases/`; honours the dependency table; introduces no new core-layer external dependency. |
| II. Interface-First | ✅ PASS | New use cases consume existing ports (`ProjectRepository`, `TemplateEngine`, `DiagramRenderer`, `OutputEncoder`); no new ports introduced (no new external dependency). |
| III. Thin Handlers | ✅ PASS by construction | FR-001 (CLI ≤ 50 eff lines/function), FR-002 (MCP ≤ 30 eff lines/function), and FR-011 (decompose oversized MCP handlers) mirror this principle exactly. |
| IV. Entity Validation | ✅ PASS | No validation logic moves into use cases or handlers. Entity constructors remain authoritative. |
| V. Test-First | ✅ PASS | New use cases land with unit tests using concrete mock ports; new `tools/archcheck` rules land with table-driven tests over fixture trees; FR-016 forbids coverage regression. |
| VI. Token Efficiency | ✅ PASS (out of scope) | No change to output formats or MCP responses; refactor is structural only. |
| VII. Simplicity & YAGNI | ⚠ TENSION → JUSTIFIED | The suppression mechanism (FR-017) is a new piece of infrastructure that did not exist before. Justified because without it the CI gate cannot be turned green on `main` without first executing a wider, unbounded clean-up. See Complexity Tracking. |
| Architecture Rules — File-Size Budgets | ✅ PASS | FR-003 (use-case ≤ 200), FR-004 (entity ≤ 300) mirror the rules table verbatim. |
| Architecture Rules — Dependency Direction | ⚠ TENSION → REQUIRES AMENDMENT | Spec FR-008 forbids direct entity-layer imports from **all three** outer entry-points (`cmd/`, `internal/mcp/`, `internal/api/`). Constitution v1.1.0 forbids it only for `cmd/`. This feature must either (a) ship a constitution amendment to v1.2.0 that tightens `internal/mcp/` and `internal/api/`, or (b) scope FR-008 down to match v1.1.0. Decision: ship the amendment as a deliverable of this feature (see Complexity Tracking + research.md R7). |
| Quality Gates — Before Every Commit | ✅ PASS | New `depguard` config strengthens, does not replace, existing `task lint`. |
| Quality Gates — Before Every PR | ✅ PASS | This feature *adds* the `task audit-constitution` gate; nothing existing weakens. |

**Verdict**: gates pass under the documented amendment plan. Proceed to Phase 0.

### Post-Phase-1 Re-evaluation

Re-checked after Phase 1 artefacts (research.md, data-model.md, contracts/, quickstart.md) were written:

- **No new principle violations introduced by the design.** The `tools/archcheck` extensions remain outside `internal/` and so are not bound by the dependency-direction rules; the new rule and suppression file formats are pure data.
- **Suppression mechanism (FR-017)** is now fully scoped in `contracts/suppression-file-schema.yaml` and `data-model.md`: 90-day hard cap, owner-tagged, expiry-as-failure semantics. The Simplicity & YAGNI tension flagged before Phase 0 stays at "JUSTIFIED" — the design does not grow the mechanism beyond what FR-017 requires.
- **Constitution amendment 1.1.0 → 1.2.0** remains a planned deliverable; the rule set in `contracts/structural-rules.yaml` already encodes the v1.2.0 outer-layer-tightening, so the amendment and the enforcement land atomically (`research.md R7`).
- **No new ports introduced.** The use cases extracted from `cmd/new.go` and `cmd/build.go` consume existing ports only (`ProjectRepository`, `TemplateEngine`, `DiagramRenderer`, `OutputEncoder`); Interface-First is satisfied without extension.

Gates remain green. Proceed to `/speckit.tasks`.

## Project Structure

### Documentation (this feature)

```text
specs/010-constitution-compliance/
├── plan.md              # This file
├── spec.md              # Feature specification (already exists)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── structural-rules.yaml          # Machine-readable mirror of the rules
│   ├── archcheck-cli.md               # CLI contract for tools/archcheck
│   ├── suppression-file-schema.yaml   # Schema for .archcheck-suppressions.yaml
│   └── violation-report.schema.json   # JSON schema for machine-readable diagnostics
└── checklists/
    └── requirements.md   # Already created during /speckit.specify
```

### Source Code (repository root)

```text
cmd/                                         # CLI commands (thin: each handler func ≤ 50 eff lines)
├── build.go                                 # ← refactor target (currently 251 lines)
├── new.go                                   # ← refactor target (currently 504 lines)
├── init.go                                  # already refactored under 009 — re-verified here
├── validate.go                              # already refactored under 009 — re-verified here
└── …

internal/
├── core/
│   ├── entities/                            # ≤ 300 eff lines/file; no internal imports
│   │   ├── graph.go
│   │   ├── graph_id.go
│   │   ├── graph_traversal.go
│   │   └── …
│   └── usecases/                            # ≤ 200 eff lines/file; imports entities only
│       ├── scaffold_project.go              # ← NEW (extracted from cmd/new.go)
│       ├── scaffold_system.go               # ← NEW (extracted from cmd/new.go)
│       ├── scaffold_container.go            # ← NEW (extracted from cmd/new.go)
│       ├── scaffold_component.go            # ← NEW (extracted from cmd/new.go)
│       ├── build_docs.go                    # ← refactor target (absorbs cmd/build.go)
│       ├── build_docs_diagrams.go           # already exists
│       ├── build_docs_tables.go             # already exists
│       ├── ports.go                         # unchanged
│       └── …
├── adapters/                                # untouched in this feature
├── mcp/
│   └── tools/                               # ≤ 30 eff lines/handler func
│       ├── analyze_coupling.go              # ← refactor target if oversized (baseline check)
│       ├── build_docs.go                    # ← refactor target if oversized (baseline check)
│       └── …
└── api/                                     # honours new FR-008 (no direct entity imports)

tools/
└── archcheck/                               # build-time audit binary (extends 009 baseline)
    ├── main.go
    ├── lines.go                             # effective-line counter
    ├── types.go                             # rule + violation types
    ├── layer.go                             # ← NEW: layer-import rule engine
    ├── suppression.go                       # ← NEW: suppression-file loader
    ├── report.go                            # ← NEW: violation-report writer (text + JSON)
    └── *_test.go                            # table-driven rule tests

.archcheck-suppressions.yaml                 # ← NEW: scoped, dated suppressions for out-of-scope pre-existing violations
.golangci.yml                                # ← EDIT: depguard layer rules added
.github/workflows/ci.yml                     # ← EDIT: add `task audit-constitution` gating step
Taskfile.yml                                 # ← EDIT: add `audit-constitution` task
.specify/memory/constitution.md              # ← EDIT: bump to 1.2.0 (see research.md R7)
```

**Structure Decision**: single Go monorepo, no new top-level directory. The audit tool already lives at `tools/archcheck/` (scaffolded under feature 009); this feature extends it with the layer-rule engine, the suppression loader, and the violation reporter. Use-case extraction lands under `internal/core/usecases/scaffold_*.go` (per spec-described decomposition) and `internal/core/usecases/build_docs*.go` (extending the file family already present). No new package boundaries are introduced inside `internal/` — splits are purely file-level.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Suppression mechanism (FR-017) — adds a `.archcheck-suppressions.yaml` plus loader code | Without it, turning the CI gate on `main` would require simultaneously fixing every pre-existing violation in packages outside this feature's named scope. That broadens scope unboundedly and turns a focused refactor into a multi-week sweep. | "Fix everything first" was rejected because the named scope (cmd, MCP tools, core file sizes) is large enough on its own and stretches the feature past one mergeable unit. "Disable the gate until everything is clean" was rejected because it permanently delays the gate, defeating Story 3. |
| Constitution amendment 1.1.0 → 1.2.0 — tightens dependency-rule row for `internal/mcp/` and `internal/api/` to forbid direct `internal/core/entities/` imports | Spec FR-008 (user-stated requirement) is stricter than the current constitution. Either FR-008 must be softened or the constitution must be tightened. Tightening is the user's intent and the natural evolutionary step; the rules table already has a TODO note (see constitution Sync Impact Report) anticipating this. | "Soften FR-008 to match v1.1.0" was considered and rejected because the user explicitly named `cmd, internal/mcp, internal/api` in the feature description, and feature 009 already left this work deliberately deferred — re-deferring would be a third can-kick. The amendment is small (one row edit + version bump), scoped to this feature's surface, and ships with an ADR (`docs/adr/0010-tighten-outer-layer-entity-import-rule.md`). |

---

*Below this point: Phase 0 and Phase 1 outputs are written to companion files in this directory (research.md, data-model.md, contracts/, quickstart.md). This plan.md ends here per `/speckit.plan` workflow.*
