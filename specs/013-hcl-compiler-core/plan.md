# Implementation Plan: HCL Compiler Core

**Branch**: `013-hcl-compiler-core` | **Date**: 2026-09-22 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/013-hcl-compiler-core/spec.md`

**Stage**: 1 of 6 in the [v1.0 roadmap](../012-v1-architecture-dsl/roadmap.md). Design source:
[design.md](../012-v1-architecture-dsl/design.md).

## Summary

Replace loko's file-tree data model with a single authored artifact and a stateless compiler. HCL v2
becomes layer 0: `*.loko.hcl` files are discovered recursively, parsed, resolved, and compiled into
an immutable address-keyed IR that every later capability consumes. `loko validate`, `loko fmt`, and
`loko export` ship on top of it; the v0 dual-source machinery — the file-tree repository, the TOML
configuration adapter, the template engine, the HTTP interface, the scaffolding use cases, and drift
detection — is deleted in the same change.

The technical approach turns on three decisions taken in [research.md](./research.md):

1. **References are static traversals, never evaluated** (R2). `hcl.AbsTraversalForExpr` yields a
   dotted address plus a range; resolution happens in core against a symbol table. This keeps
   business logic out of the adapter, makes declaration order irrelevant for free, and makes the diff
   stage's `moved` block an ordinary use of the mechanism rather than an exception to the evaluator.
2. **The parser emits a flat, unresolved `SourceModel`; core builds the IR** (R7). Every rule in
   FR-028 and FR-029 is then unit-testable from a struct literal with no files on disk, and the
   golden-file suite stays focused on syntax.
3. **The IR holds ordered slices, never maps** (R6). Determinism is a structural property rather than
   a discipline each encoder must remember, which is the precondition for the entire golden-file test
   strategy.

## Technical Context

**Language/Version**: Go 1.25+ (matches `go.mod`; toolchain go1.25.4)

**Primary Dependencies**: `github.com/hashicorp/hcl/v2` (parse, evaluate, `hclwrite` formatting) and
`github.com/zclconf/go-cty` (value model; supplies all five functions of FR-017 from
`cty/function/stdlib`). Both confined to `internal/adapters/hclsource` by the new layer rule. Retained
unchanged: cobra + viper, lipgloss, toon-go, stretchr/testify. **Not** added: `go-git` (diff stage),
`hashicorp/go-version` (replaced by a stdlib constraint evaluator, research R5).

**Storage**: Filesystem only. No lock file, state file, or database (FR-027). Git remains the history.

**Testing**: `go test` with golden fixtures under `testdata/` (`-update` regenerates), table tests over
`SourceModel` literals for rules, and byte-equality tests for determinism. Concrete mock structs, no
mocking library (Principle VII).

**Target Platform**: Single static binary — macOS, Linux, Windows. No external process dependency in
this stage.

**Project Type**: Compiler + CLI (single Go module, clean architecture).

**Performance Goals**: 1,000 elements across 200 files compiles in < 2 s; 5,000 elements in < 10 s
(SC-006). Sequential parsing is expected to clear this comfortably; concurrency is not planned and
would be added only against a measured failure.

**Constraints**: Byte-identical output across runs, machines, and file-system orderings (FR-040).
Every diagnostic carries file/line/column (FR-030). Exactly three exit codes (FR-038). No package
outside `internal/adapters/...` may import the HCL or D2 libraries (FR-044).

**Scale/Scope**: Deletes roughly 6,700 LOC of v0 model code plus its tests; adds the compiler, IR,
validation, and three commands. 44 non-test files reference the three ports being removed.

## Constitution Check

*GATE: evaluated before Phase 0 and re-evaluated after Phase 1 design.*

| Principle | Assessment | Verdict |
|---|---|---|
| **I. Clean Architecture** | HCL and cty confined to `internal/adapters/hclsource`; entities stdlib-only via `entities.SourceRange` rather than `hcl.Range` (R8); resolution and validation in core (R7). Dependency direction preserved. | **PASS** |
| **II. Interface-First** | New port `ArchitectureSource` declared in `usecases/ports.go` before the adapter exists; `OutputEncoder` reused unchanged. Ports for the deleted model are removed, not stubbed. | **PASS** |
| **III. Thin Handlers** | Three new CLI handlers (`validate`, `fmt`, `export`), each parse-flags → call use case → format output. Well inside 50 effective lines. MCP handlers are only deleted in this stage. | **PASS** |
| **IV. Entity Validation** | Address construction, version-constraint evaluation, and diagnostic construction live in entities. Graph-shaped rules (containment cycle, duplicate claims) are use cases, since they span entities — consistent with existing `validate_circular.go`. | **PASS** |
| **V. Test-First** | Golden fixtures and rule tables written before implementation; > 80% coverage on `internal/core/`. Deleted code's tests are deleted, not adapted (FR-042). | **PASS** |
| **VI. Token Efficiency** | TOON retained as an export encoding alongside JSON (FR-036). Progressive detail levels belong to the `describe` tool in the authoring stage, not here. | **PASS** |
| **VII. Simplicity & YAGNI** | No iteration constructs, no modules, five functions only (FR-017/FR-018). No compatibility shim for the v0 model — it is deleted outright (R9). Version constraints hand-rolled rather than ported behind a new port (R5). | **PASS** |

**Layer rules**: the Dependency Direction table is satisfied. One rule is **added** by this feature per
FR-044, to `specs/010-constitution-compliance/contracts/structural-rules.yaml`: no package outside
`internal/adapters/**` may import `github.com/hashicorp/hcl/**`, `github.com/zclconf/go-cty/**`, or
`oss.terrastruct.com/d2/**`. Mirrored into `.golangci.yml` `depguard` as the fast path.

**File-size budgets**: use-case files ≤ 200 effective lines forces the compile pipeline to be split by
step rather than written as one `compile.go` — reflected in the structure below.

**Constitution and technology-stack drift**: the constitution's *Technology Stack* and *External
Dependencies* sections still describe TOML configuration, the ason template engine, and a shelled-out
d2 CLI, all of which this feature removes or supersedes. That is a documentation amendment, not a
violation — recorded in Complexity Tracking and to be carried out with the ADR below rather than left
to drift.

**ADR required**: `docs/adr/0012-hcl-source-of-truth.md` — adopting HCL as layer 0, the static-traversal
reference decision, and the deletion of the dual-source model. Required by the Governance section for
any new architectural decision.

*Post-Phase-1 re-check*: no new violations. The `SourceModel`/IR split introduced in Phase 1 strengthens
Principle I rather than straining it, and adds no dependency.

## Project Structure

### Documentation (this feature)

```text
specs/013-hcl-compiler-core/
├── plan.md                        # This file
├── spec.md                        # Feature specification (clarified 2026-09-22)
├── research.md                    # Phase 0 — 12 decisions, incl. the deletion sequencing risk
├── data-model.md                  # Phase 1 — SourceModel, IR, addresses, rule table
├── quickstart.md                  # Phase 1 — runnable validation scenarios
├── contracts/
│   ├── language.md                # The *.loko.hcl grammar (user-facing, permanent surface)
│   ├── cli.md                     # Commands, flags, exit codes, removed commands
│   ├── ir.schema.json             # Export shape, schemaVersion 1
│   └── diagnostics.schema.json    # Machine-readable diagnostics
├── checklists/requirements.md     # Spec quality checklist (16/16)
└── tasks.md                       # Phase 2 — created by /speckit-tasks, not here
```

### Source Code (repository root)

```text
internal/core/entities/                 # stdlib only
├── address.go                          # Address forms, parsing, byte-wise ordering
├── diagnostic.go                       # Severity, Diagnostic, Diagnostics, exit-code mapping
├── source_range.go                     # SourceRange (local copy of hcl.Range — R8)
├── source_model.go                     # Unresolved declarations from the parser
├── source_model_deployment.go          # Unresolved deployment declarations (split for the 300-line budget)
├── ir.go                               # IR root + indexes
├── ir_element.go                       # Element, Relationship
├── ir_deployment.go                    # Environment, Group, Instance, Claim
├── ir_view.go                          # View
└── version_constraint.go               # ~>, >=, … evaluator, stdlib only (R5)

internal/core/usecases/                 # ≤ 200 effective lines per file
├── ports.go                            # + ArchitectureSource; − ProjectRepository/ConfigLoader/TemplateEngine
├── compile_architecture.go             # Orchestrates: source → resolve → validate → IR
├── resolve_references.go               # Symbol table, reference resolution (FR-019–022)
├── validate_structure.go               # Parent kinds, containment cycle (FR-028)
├── validate_deployment.go              # Duplicate claims, duplicate instance names (FR-012a, FR-028)
├── validate_warnings.go                # FR-029
├── build_ir.go                         # Construction + sorted ordering (FR-040)
├── build_ir_deployment.go              # Environment/group/instance construction and placement cross-refs
├── export_ir.go                        # Compile then encode; suppress artefact on error (FR-037)
└── format_sources.go                   # fmt and fmt --check orchestration (FR-035, FR-035a)

internal/adapters/hclsource/            # ONLY package importing hcl/cty
├── discover.go                         # Recursive *.loko.hcl discovery, sorted
├── parse.go                            # hclparse, file cache, syntax diagnostics
├── schema.go                           # Block/attribute schemas per kind
├── decode_logical.go                   # person/system/container/component/external + uses
├── decode_deployment.go                # deployment/node/instance/binding
├── decode_project.go                   # project/view/reconcile/locals
├── traversal.go                        # AbsTraversalForExpr → Reference (R2)
├── functions.go                        # The five cty stdlib functions (R4)
├── ranges.go                           # hcl.Range → entities.SourceRange (R8)
├── render_diagnostics.go               # Human-readable output with source snippets
└── format.go                           # hclwrite canonical formatting

internal/adapters/encoding/             # existing; extended
├── toon.go                             # extended for the IR
├── json.go                             # new — deterministic JSON encoding of the IR
└── diagnostics.go                      # new — machine-readable diagnostics (FR-030a)

cmd/                                    # ≤ 50 effective lines per handler
├── validate.go / validate_cobra.go     # rewritten
├── fmt.go / fmt_cobra.go               # new
├── export.go / export_cobra.go         # rewritten
└── root.go                             # command registration updated

tools/genfixture/
└── main.go                             # synthetic project generator for the SC-006 bounds
```

**Deleted in this stage** (research R9, FR-041–FR-043): `internal/adapters/filesystem/project_repo.go`,
`relationship_repo.go`, `config.go`; `internal/adapters/config/`; `internal/adapters/ason/`;
`internal/api/` and `cmd/api.go`; all `scaffold_*`, `create_*`, `delete_relationship`,
`find_relationships`, `list_relationships`, `search_elements`, `query_architecture*`, `build_docs*`,
`render_markdown_docs`, `update_diagram`, `init_project`, `detect_drift`, `validate_architecture*` use
cases; the MCP tools backed by them; `cmd/build.go`, `serve.go`, `watch.go`, `new.go`, `init.go`; and
every test covering the above.

**Parked, not deleted** (FR-045): `internal/adapters/d2/parser.go` and `d2_import_test.go` for the
diagram-import stage. `internal/adapters/html/` stays on disk unwired for the renderer stage.

**Structure Decision**: The existing clean-architecture layout is kept exactly; this feature swaps the
organs inside it. The one new package is `internal/adapters/hclsource`, which exists specifically so
that the HCL dependency has a single home the layer rule can name. `internal/core/entities` gains the
IR and grows, but every file stays inside the 300-effective-line budget by splitting the IR across
four files rather than one.

## Order of Work

Follows the roadmap's "compiler core first" instruction, with the deletion last so the build is only
briefly red:

1. **Entities** — addresses, ranges, diagnostics, version constraints, `SourceModel`, IR types. Pure,
   fully testable, no dependency on anything else.
2. **Parser adapter** — discovery, parse, schemas, traversal extraction, functions. Golden fixtures for
   syntax and unknown-construct diagnostics.
3. **Core pipeline** — symbol table, resolution, validation rules, IR construction and ordering. Table
   tests from `SourceModel` literals.
4. **Encoding + export** — deterministic JSON and TOON, `schemaVersion`, byte-equality tests.
5. **Formatting** — `hclwrite` round-trip, idempotence, `--check`.
6. **Commands** — `validate`, `fmt`, `export`; text and JSON diagnostic rendering.
7. **Deletion** — remove the v0 model, its use cases, its MCP tools, its commands, and its tests; `go
   mod tidy`; add the layer rule to the structural-rules file and `depguard`; run
   `task audit-constitution`.
8. **ADR + constitution amendment** — `docs/adr/0012-hcl-source-of-truth.md`, and update the
   constitution's Technology Stack and External Dependencies sections.

## Complexity Tracking

| Violation / deviation | Why needed | Simpler alternative rejected because |
|---|---|---|
| Version-constraint evaluator hand-written in entities | The check is one of FR-028's errors, so it belongs in the validation layer, and core may not take a dependency | Putting `hashicorp/go-version` behind a port means a port, adapter, mock, and wiring for arithmetic on three integers — Principle VII forbids abstractions for one-time operations |
| Two models (`SourceModel` and `IR`) rather than one | Keeps resolution and validation in core while the HCL types stay in the adapter (R7); makes every rule testable without files | A single model means the adapter returns the finished IR, moving resolution and validation into the adapter — a direct Principle I and IV violation |
| `build`, `serve`, `watch`, `init` non-functional until the renderer stage | Deleting `ProjectRepository` reaches 44 files; keeping them alive requires retaining the v0 model (R9) | Keeping both models behind a flag *is* the dual-source-of-truth state this release exists to remove, and Principle VII forbids compatibility shims |
| Constitution's Technology Stack section becomes stale on merge | TOML, ason, and the shelled-out d2 CLI are removed or superseded by this feature and the next | Amending the constitution before the code lands would describe a state that does not yet exist; the amendment ships with this feature's ADR |

## Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Containment cycle check wrongly applied to relationships, breaking legal architectures (FR-032) | Medium | High — rejects valid input | Explicit fixture in the golden suite and in quickstart Scenario 1.5; the two traversals are deliberately separate functions in `validate_structure.go` |
| A Go map escapes into the IR and output is nondeterministic on some runs | Medium | High — breaks golden tests intermittently and pushes users to hand-edit artefacts | IR exposes slices only (R6); shuffled-discovery-order test in the determinism suite, not just repeat-run equality |
| Deletion breaks the build in ways not visible until late | High | Medium | Deletion is step 7, after the new path is green; the 44-file blast radius is enumerated in R9 rather than discovered |
| `hclwrite` round-trip loses a comment in an edge case | Low | High — the authoring stage depends on it | Idempotence and comment-preservation fixtures now; the full round-trip property test is the authoring stage's gate |
| Scope creep from "while we're in here" renderer work | Medium | Medium | `build`/`serve` are explicitly out; the renderer stage owns them |

## Handover to `/speckit-tasks`

Phase 2 is **not** performed by this command. When generating tasks, note that user stories map to the
work order above as: Story 1 → steps 1–3, 6; Story 2 → steps 1–3; Story 3 → steps 1, 3, 4; Story 4 →
step 5; Story 5 → step 1 (`version_constraint.go`); Story 6 → steps 7–8. Steps 1 and 2 are
parallelisable after entities land; step 7 must be last.
