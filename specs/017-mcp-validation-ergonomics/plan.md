# Implementation Plan: MCP & Validation Ergonomics

**Branch**: `017-mcp-validation-ergonomics` | **Date**: 2026-10-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/017-mcp-validation-ergonomics/spec.md`

## Summary

Make a first MCP session work without trial and error. The language gets one definition, a data
table in `arch` (R1). The decoder derives from it, the edit schema is held to it by a parity test,
and a new read-only `schema` tool serves it, pointed to from the handshake instructions (R2, R3).

Warnings get a `planned` marker on instances (R4) and suppression by code, per run or on the
`project` block. Suppression is applied once, at the end of compilation, so every command and strict
mode agree (R5). Validate also gets a severity filter and a summary (R6), and MCP diagnostics
default to `file:line` (R7).

Two new MCP tools, `build` and `fmt`, reuse the CLI's use cases, wired in `cmd` (R8, R9). A
`project` target on `apply_edit`, plus a `bootstrap` hint from `describe`, let a client start from
an empty directory (R10).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, toolchain pinned)

**Primary Dependencies**: existing only: hcl/v2 and hclwrite (decoder, edits), toon-go (encoding),
in-process d2 (build). No new module.

**Storage**: files only. `*.loko.hcl` edits stay atomic through `SourceEditor.Commit`; build output
stays owned through `.loko-manifest`.

**Testing**: `go test ./...` with `-race` and the goroutine-leak gate. Goldens for `schema` and
summary output, fixtures for the two new codes, parity tests, MCP end-to-end tests.

**Target Platform**: macOS, Linux and Windows CLI; MCP over stdio.

**Project Type**: CLI compiler with an MCP server.

**Performance Goals**: `schema` full answer ≤ 3,000 TOON tokens, one block ≤ 500 (R2). MCP
validate for the reported model ≤ 50% of its 1.0.2 size (SC-004). Validate and build stay inside
the existing 015 and 014 performance budgets.

**Constraints**: byte-stable output (FR-006). No change to CLI output or exports when the new
options and attributes are unused (SC-007). `internal/mcp` imports no render adapter and no
entity package.

**Scale/Scope**: about 15 blocks and 50 attributes in the table; 3 new MCP tools (`schema`, `build`, `fmt`), 4 changed MCP
tools (`validate`, `describe`, `apply_edit`, `move`), 2 new language attributes, 2 new diagnostic codes, 4 new CLI flags.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle / rule | Status | How |
|---|---|---|
| **I. Clean Architecture** | ✅ | The language table, codes, `CheckIgnore` and `Without` are entities (stdlib). Suppression, summary, compact diagnostics, `BuildInProject` (confine, build, pick the inline view), `DescribeLanguage` and `Format` are use cases. Resolving the output directory touches the file system, so it is an `ArtifactStore` method implemented in `adapters/outputdir` (R8). The `build` tool's backends are wired in `cmd/wiring.go`, so `internal/mcp` imports no render adapter. |
| **II. Interface-First** | ✅ | No new port. `SourceFormatter` gains `Canonical` (in `hclsource`) and `ArtifactStore` gains `Confine` (in `outputdir`), so no use case touches the file system directly. |
| **III. Thin Handlers** | ✅ | `schema`, `build` and `fmt` handlers parse input, call one use case and encode the result, each ≤ 30 effective lines. The `build` tool's inline-view choice and refusal live in `BuildInProject`, not in the handler. CLI flags go through `*_cobra.go`. |
| **IV. Entity Validation** | ✅ | Allowed values and ignore rules live in `arch`; edit rules for `project` and `planned` live in `authoring`. Handlers do not validate. |
| **V. Test-First** | ✅ | Each phase opens with failing tests: parity, fixtures for the new codes, the summary golden, the compact-size test, MCP end-to-end tests for `build`, `fmt` and bootstrap. |
| **VI. Token Efficiency** | ✅ | This feature exists for it: compact diagnostics by default, a per-block `schema` filter, a summary capped at 20 addresses per code, and token budgets asserted in tests. |
| **VII. Simplicity** | ✅ | No new module or port, and no tool beyond the three the spec asks for. A bool and a list of strings, not a lifecycle vocabulary or a severity-override system (R4, R5). |
| Dependency Direction | ✅ | `authoring` keeps its copies instead of importing `arch`, held in step by parity tests (the 016 pattern). |
| File-size budgets | ✅ | `arch/schemas.go` is a pure-data file (exempt by name). New use-case files stay ≤ 200 effective lines: `describe_language.go`, `validate_filter.go`, `format_plan.go`, `build_in_project.go`. |
| Quality gates | ✅ | lint, test (with `go test -cover ./internal/core/...` > 80%), audit, race and goroutine-leak gate. ADR-0016 records the compact MCP default and the two attributes. |

**Language change governance**: two permanent v1.x attributes (`project.ignore_warnings`,
`instance.planned`), specified in [contracts/language.md](contracts/language.md) and documented in
`docs/language.md`.

**Output change**: compact MCP diagnostics are the one deliberate change to existing output (FR-013).
It is recorded in ADR-0016 and the changelog; `ranges: "full"` restores the 1.0.2 shape.

**Post-design re-check**: ✅ Passes. The design adds one entity file, two port methods, five
use-case files and three MCP tools; no new layer crossing.

## Project Structure

### Documentation (this feature)

```text
specs/017-mcp-validation-ergonomics/
├── spec.md
├── first-run-report.md   # the session report this feature answers
├── plan.md               # this file
├── research.md           # R1–R12
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── mcp-tools.md
│   ├── cli.md
│   └── language.md
├── checklists/
│   └── requirements.md
└── tasks.md              # /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/entities/arch/
├── schemas.go                  # NEW: Language table (R1)
├── diagnostic.go               # +2 codes, WarningCodes, CheckIgnore, Without (R5, R11)
├── source_model.go             # ProjectDecl.IgnoreWarnings
└── source_model_deployment.go  # InstanceDecl.Planned; IR instance Planned

internal/core/entities/authoring/
├── edit.go                     # TargetProject
├── edit_schema.go              # project attrs; instance planned (ValueBool)
└── edit_address.go             # AddressForms (R2)

internal/core/usecases/
├── ports.go                    # SourceFormatter.Canonical, ArtifactStore.Confine
├── compile_architecture.go     # CompileRequest.Ignore; suppression last
├── validate_project.go         # ValidateRequest, filters
├── validate_filter.go          # NEW: summary, severity, DiagnosticViews (R6, R7)
├── validate_warnings.go        # planned → no unbound_instance; planned_instance_bound
├── validate_project_block.go   # NEW or folded in: ignore_warnings checks
├── describe_language.go        # NEW: DescribeLanguage (R2)
├── describe.go                 # Bootstrap hint (R10)
├── format_plan.go              # NEW: AuthoringService.Format (R9)
├── build_in_project.go         # NEW: BuildInProject, PickInlineView (R8)
└── language_parity_test.go     # NEW: arch.Language ↔ authoring ↔ decoder

internal/adapters/hclsource/
├── schema.go                   # attribute lists derived from arch.Language
├── decode_project.go           # ignore_warnings
├── decode_deployment.go        # planned
├── edit_blocks.go              # project target placement, no_project
└── format.go                   # Canonical

internal/adapters/outputdir/
└── confine.go                  # NEW: Store.Confine (R8)

internal/mcp/
├── server.go                   # initialize instructions (R3)
└── tools/
    ├── schema.go               # NEW
    ├── build.go                # NEW
    ├── fmt.go                  # NEW
    ├── validate.go             # summary, severity, ignore, ranges
    ├── helpers.go              # ranges handling shared by every tool
    └── schemas.go              # input schemas

cmd/
├── wiring.go                   # newMCPTools gains build (with newBuildDeps) and fmt
├── validate.go, validate_cobra.go  # --summary, --severity, --ignore
└── build_cobra.go, serve_cobra.go  # --ignore

testdata/projects/first-run/              # NEW: the reported model's shape (SC-002–SC-004)
testdata/projects/first-run-remodelled/   # NEW: the SC-002 target
docs/language.md, docs/mcp-integration.md, docs/cli-reference.md
docs/adr/0016-mcp-ergonomics.md
CHANGELOG.md
```

**Structure Decision**: the existing single-module layout. Each change lands in the layer the
constitution assigns it, and no new package is created.

## Phases (for `/speckit-tasks`)

1. **Foundation**: `arch.Language`, the decoder deriving from it, parity tests. No behaviour change;
   all goldens unchanged.
2. **Story 1 (P1)**: `DescribeLanguage`, `AddressForms`, the `schema` tool, handshake instructions.
3. **Story 2 (P1)**: `planned`, `ignore_warnings`, `CompileRequest.Ignore`, the filters and summary,
   compact MCP diagnostics, CLI flags, the `first-run` fixture.
4. **Story 3 (P2)**: `SourceFormatter.Canonical`, `AuthoringService.Format`, the `fmt` tool;
   `ArtifactStore.Confine`, `BuildInProject` and the `build` tool.
5. **Story 4 (P2)**: `TargetProject`, `no_project` and `project_exists`, the `describe` bootstrap.
6. **Polish**: docs, ADR-0016, changelog, the token and size budget tests, quickstart run.

Stories 1 and 2 are independent after Phase 1. Story 4 uses Phase 1's table for its new target.

## Complexity Tracking

No constitution violations to justify.
