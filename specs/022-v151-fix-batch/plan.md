# Implementation Plan: v1.5.1 Patch Bug-Fix Batch

**Branch**: `022-v151-fix-batch` | **Date**: 2026-11-09 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/022-v151-fix-batch/spec.md`

## Summary

Five defect fixes from the third-party feedback report, shipped as patch 1.5.1:

1. The ownership manifest moves out of the output directory into a state directory
   `.<base>.loko` next to it (`.dist.loko` for the default `dist`), with a 1.5.0
   fallback and success-only legacy retirement (issue 15).
2. Root CLI help gains a documentation pointer naming `docs/language.md` (issue 16).
3. The unknown-block error for `dynamic` names the supported construct: a `view`
   block with `step` blocks (issue 17).
4. The relationship `kind` schema note states the default direction and the trigger
   reversal (issue 2, docs half).
5. The theming guide documents the stable CSS class contract (issue 14, CSS half).

Patch discipline (spec FR-008): see spec Assumptions; enforced by task T020.

## Technical Context

**Language/Version**: Go 1.27+ (`go 1.27.0` in go.mod, toolchain pinned)

**Primary Dependencies**: hcl/v2 (source), cobra (CLI), toon-go (encoding) — no new
dependencies; all fixes touch existing packages only.

**Storage**: File system only. One new directory per project: `.<base>.loko`
beside the output directory, holding one file: `manifest`.

**Testing**: `go test` table-driven with `t.Parallel()`; golden files under
`testdata/`; quality gates:
`task test && task lint`, coverage > 80% on `internal/core/`.

**Target Platform**: Single static binary; macOS (arm64) and Linux CI.

**Project Type**: CLI + MCP server + embedded HTTP preview (one binary).

**Performance Goals**: No change. Manifest read/write is one small file per build.

**Constraints**: Constitution budgets (use-case files ≤ 200 / entity files ≤ 200 /
adapter files ≤ 400 effective lines; handler functions ≤ 50 CLI / ≤ 30 MCP).
Patch-version discipline: consumers pin 1.5.x, so no output-shape surprises.

**Scale/Scope**: 5 fixes across 4 packages + 2 docs files + CHANGELOG. The largest
single change is the manifest relocation in `internal/adapters/outputdir/`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|---|---|---|
| I. Clean Architecture (dependency direction) | PASS | Manifest logic stays in `adapters/outputdir/`; the state-directory name is derived there from `outDir`, no core changes. Text fixes are in adapters/cmd. |
| II. Interface-First | PASS | `Store.Commit` already goes through `ArtifactStore` port; the new state path is an adapter-internal detail. No new port. |
| III. Thin Handlers | PASS | `cmd/root.go` help text is a string change. No MCP handler changes (schema note lives in `internal/core/entities/arch/schemas.go`, a data-file exempt from budgets). |
| IV. Entity Validation | PASS | No validation changes; `DecodeManifest` safety-widening behavior is reused unchanged. |
| V. Test-First | COMMITTED | Tests written before each fix; store tests gain migration cases (legacy manifest fallback, retirement, failure-retention). |
| VI. Token Efficiency | PASS | No output shapes removed. The schema note grows by one sentence (additive). Help text grows by one line. |
| VII. Simplicity & YAGNI | PASS | No shims: the 1.5.0 fallback is a spec requirement (FR-002), not a compat flag. No abstraction for the state-directory naming; one function computes it. |
| File budgets | PASS | `store.go` is currently ~167 lines; the change stays well under the 400 adapter budget. No handler functions touched. |
| ADR needed? | NO | No architectural decision: the manifest path is data placement, specified in the spec (clarified session 2026-11-09), not a new pattern. FR-002's migration path is documented in the spec and CHANGELOG. |
| *Post-Phase-1 re-check* | PASS | Design adds: helper functions `stateDir`/`stateFile` + legacy fallback inside the existing adapter (no new port, no layer crossing), contract `contracts/output-state.md` amending 014's output-layout contract normatively, and 2 new contracts + quickstart. No budgets affected; no new dependencies; rendering behavior untouched (FR-008 verification is golden-diff review). |

## Project Structure

### Documentation (this feature)

```text
specs/022-v151-fix-batch/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   ├── output-state.md          # state directory + manifest contract (amends 014)
│   ├── cli-language-notes.md    # help pointer + dynamic hint + kind note contract
│   └── theme-classes.md         # stable CSS class contract (amends 021, guide)
├── checklists/
│   └── requirements.md  # Spec quality checklist (all passing)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
cmd/
└── root.go                      # Fix 2: docs pointer in Long help text (root.go:26-40)

internal/adapters/hclsource/
└── schema.go                    # Fix 3: dynamic hint in deliberatelyExcluded (schema.go:34)
    └── functions.go             # (unchanged: function detail text is a different surface)

internal/core/entities/arch/
└── schemas.go                   # Fix 4: kind attr note (schemas.go:133; data-file, budget-exempt)

internal/adapters/outputdir/
├── store.go                     # Fix 1: stateDir()/stateFile()/readManifest/write path
├── store_test.go                # Fix 1 tests: migration, retirement, failure retention
└── doc.go                       # Fix 1 doc comment

docs/
├── guides/site-theming.md       # Fix 5: stable CSS class contract section
└── adr/                         # (no new ADR; 0013/0014 contracts updated instead)

specs/014-viewmodel-renderers/contracts/
└── output-layout.md             # Fix 1: contract amendment for the new manifest location

CHANGELOG.md                     # ### Fixed section, house style (CHANGELOG.md:8-67)
```

**Structure Decision**: All five fixes are single-site edits in existing files at
their existing canonical locations (verified anchors in research.md). No new
packages, no new ports, no file moves. The only new on-disk artifact is the state
directory next to the output directory, implemented wholly inside
`adapters/outputdir`.
