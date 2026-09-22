---

description: "Task list for 013-hcl-compiler-core"
---

# Tasks: HCL Compiler Core

**Input**: Design documents from `/specs/013-hcl-compiler-core/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/)

**Tests**: Test tasks are **included and mandatory**. Constitution Principle V (Test-First) is
non-negotiable for this repository, and SC-004 requires a fixture per validation rule. Tests are
written before the implementation they cover.

**Organization**: Tasks are grouped by user story. Note the honest caveat below.

**Revision**: renumbered and extended after `/speckit-analyze` (2026-09-22) to close findings F1, F2,
G1–G5. 94 → 100 tasks.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US6)
- Exact file paths are given in every task

## Path Conventions

Single Go module at the repository root. Layout per [plan.md](./plan.md) § Project Structure:
`internal/core/entities/`, `internal/core/usecases/`, `internal/adapters/hclsource/`,
`internal/adapters/encoding/`, `cmd/`, `tools/`.

---

## ⚠️ Read before starting: story independence in this feature

The template assumes each user story is an independently deliverable slice. **That is only partly
true here, and pretending otherwise would produce a misleading plan.** This feature replaces a
compiler's data model: three of the six stories are P1 because they share one pipeline, and none of
them can ship without the shared entity and parser foundation in Phase 2.

What is genuinely independent:

- **US4** (formatting) and **US5** (version constraint) are real independent slices — each can be
  built, tested, and demonstrated alone once Phase 2 exists.
- **US2** (deployment plane) is independently *testable* — its rules have their own fixtures — but it
  is not separately *shippable*, because the same `validate` command surfaces it.
- **US1** and **US3** are two halves of one thing: US1 is the logical plane and its diagnostics, US3
  is the compiled artefact those diagnostics are produced alongside. The MVP is **Phase 2 + US1 +
  US3** together.
- **US6** (deleting the v0 model) is a teardown, not a feature slice. It must run **last**, after
  every other phase is green, because it removes 44 files' worth of dependencies and the build cannot
  be green mid-way. See [research.md](./research.md) R9.

Phase 2 is therefore unusually large. That is the accurate shape of the work, not padding.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Dependencies and tooling in place before any code is written

- [ ] T001 Add `github.com/hashicorp/hcl/v2` and promote `github.com/zclconf/go-cty` to a direct dependency in `go.mod`, then run `go mod tidy` and commit the resulting `go.sum`
- [ ] T002 Create the package directory `internal/adapters/hclsource/` with a `doc.go` stating that this is the only package permitted to import HCL or cty, per FR-044
- [ ] T003 Write the layer-rule guard test in `tools/archcheck/archcheck_test.go` proving the audit fires when a forbidden import is placed in `internal/core/entities/`, per quickstart Scenario 6. **Run it and confirm it fails** before T004/T005 — it is the only evidence the new rule is actually wired rather than decorative
- [ ] T004 [P] Add the layer rule to `specs/010-constitution-compliance/contracts/structural-rules.yaml`: no package outside `internal/adapters/**` may import `github.com/hashicorp/hcl/**`, `github.com/zclconf/go-cty/**`, or `oss.terrastruct.com/d2/**`; confirm T003 now passes
- [ ] T005 [P] Mirror the same rule into `.golangci.yml` under `depguard` as the fast-path check

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The entity vocabulary and the parser front end. Every user story depends on all of it.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Entities — tests first

- [ ] T006 [P] Write table tests for address construction, parsing, and byte-wise ordering in `internal/core/entities/address_test.go`, covering all six address forms in data-model.md §1 and the name pattern `^[A-Za-z_][A-Za-z0-9_-]*$`
- [ ] T007 [P] Write table tests for `Diagnostics.SortedForOutput()` (file, then start byte, then code) and `ExitCode(strict bool)` returning exactly 0/1/2 in `internal/core/entities/diagnostic_test.go`

### Entities — implementation

- [ ] T008 [P] Implement `Address` with constructors per form and byte-wise `Compare` in `internal/core/entities/address.go` — constructors only, never string concatenation by callers
- [ ] T009 [P] Implement `SourceRange` (File, StartLine, StartColumn, StartByte, EndLine, EndColumn, EndByte) in `internal/core/entities/source_range.go`; `File` is project-relative with forward slashes on every platform per FR-036c
- [ ] T010 Implement `Severity`, `Diagnostic` (Severity, Code, Summary, Detail, Range, Related), `Diagnostics`, `HasErrors`, `SortedForOutput`, and `ExitCode` in `internal/core/entities/diagnostic.go`; the `Code` set must match the enum in `contracts/diagnostics.schema.json` **exactly**, including `syntax_error`
- [ ] T011 [P] Implement the unresolved `SourceModel` types — `ProjectDecl`, `ElementDecl`, `RelationDecl`, `Reference`, `ViewDecl`, `IgnorePattern`, `Value` — in `internal/core/entities/source_model.go`
- [ ] T012 [P] Implement the unresolved deployment declarations — `EnvironmentDecl`, `GroupDecl`, `InstanceDecl`, `ClaimDecl` — in `internal/core/entities/source_model_deployment.go` (separate file to stay inside the 300 effective-line entity budget)
- [ ] T013 [P] Implement IR root and indexes — `IR` with `SchemaVersion`, `Project`, and **ordered slices only, never maps**, plus unexported lookup indexes and `Lookup`/`ElementsByKind`/`OutgoingFrom`, no mutating methods — in `internal/core/entities/ir.go`
- [ ] T014 [P] Implement `Element` and `Relationship` IR types in `internal/core/entities/ir_element.go`; `Tags` is sorted and de-duplicated
- [ ] T015 [P] Implement `Environment`, `Group`, `Instance`, `Claim` in `internal/core/entities/ir_deployment.go`; `Environment.Instances` is **flat and complete** with placement recorded via `Group.Contains` and `Instance.PlacedIn`, per data-model.md §3
- [ ] T016 [P] Implement `View` in `internal/core/entities/ir_view.go`

### Ports

- [ ] T017 Declare the `ArchitectureSource` port — `Load(ctx, root string) (*entities.SourceModel, entities.Diagnostics, error)` — in `internal/core/usecases/ports.go`, and add a concrete mock (not a mocking library) in `internal/core/usecases/mocks_test.go`

### Parser adapter — tests first

- [ ] T018 [P] Create golden-fixture harness with a `-update` flag in `internal/adapters/hclsource/golden_test.go`, reading `testdata/<case>/{input/,expected_ir.json,expected_diagnostics.json}`
- [ ] T019 [P] Write discovery tests in `internal/adapters/hclsource/discover_test.go` covering: nested directories merged (FR-001), non-`*.loko.hcl` files ignored (FR-002), and no-source-found producing `no_source_found` (FR-005)
- [ ] T020 [P] Add a `syntax_error` fixture in `internal/adapters/hclsource/testdata/syntax_error/` with an unclosed block, asserting the diagnostic uses code `syntax_error` — **not** `unknown_block` — and that a second, valid file in the same project still parses and still reports its own diagnostics (FR-028a, FR-031)
- [ ] T021 [P] Add two edge-case fixtures in `internal/adapters/hclsource/testdata/`: `empty_architecture/` (a `project` block and nothing else — compiles, exports an empty element set, is **not** an error) and `unreadable_file/` (a file with no read permission — reported as a diagnostic against that file while the others still parse)

### Parser adapter — implementation

- [ ] T022 Implement recursive `*.loko.hcl` discovery returning a sorted file list in `internal/adapters/hclsource/discover.go`
- [ ] T023 Implement parsing via `hclparse.Parser`, retaining the file cache for the run so diagnostics can render source snippets, in `internal/adapters/hclsource/parse.go`; a file that fails to parse yields `syntax_error` and parsing continues with the remaining files
- [ ] T024 [P] Implement `hcl.Range` → `entities.SourceRange` conversion, normalising paths to project-relative with forward slashes, in `internal/adapters/hclsource/ranges.go`
- [ ] T025 Define block and attribute schemas for every construct in `contracts/language.md` in `internal/adapters/hclsource/schema.go`, using `PartialContent` so unexpected blocks and attributes become diagnostics rather than fatal errors (FR-031)
- [ ] T026 Implement static traversal extraction with `hcl.AbsTraversalForExpr` → `entities.Reference{Raw, Range}` in `internal/adapters/hclsource/traversal.go`; references are **never evaluated** per research R2, and a quoted string in a reference position is a type error
- [ ] T027 [P] Wire the five functions `join`, `split`, `lower`, `upper`, `replace` from `cty/function/stdlib` into an `hcl.EvalContext` in `internal/adapters/hclsource/functions.go`; any other call emits `unknown_function` whose detail names all five (FR-017a)
- [ ] T028 Implement `project`, `locals`, `view`, and `reconcile` decoding in `internal/adapters/hclsource/decode_project.go`; `locals` may reference other locals but never elements
- [ ] T029 Implement human-readable diagnostic rendering with source snippets and carets via `hcl.NewDiagnosticTextWriter` in `internal/adapters/hclsource/render_diagnostics.go`; colour is suppressed when stdout is not a terminal or `NO_COLOR` is set

**Checkpoint**: entities compile with tests green, and the parser can discover, parse, and report syntax diagnostics. User story work can begin.

---

## Phase 3: User Story 1 — Author an architecture in one place and have it checked (Priority: P1) 🎯 MVP

**Goal**: An architect writes logical elements and relationships across several files and gets a
precise, complete diagnostic report from `loko validate`.

**Independent Test**: Write the two-system project from quickstart Scenario 1, run `loko validate`,
get exit 0. Break a reference, get exit 1 with the file, line, and column of the traversal.

### Tests for User Story 1

- [ ] T030 [P] [US1] Add golden fixtures for the happy-path logical project and the multi-file merge case in `internal/adapters/hclsource/testdata/logical_basic/` and `testdata/logical_multifile/`
- [ ] T031 [P] [US1] Write resolution table tests over `SourceModel` literals — no files on disk — in `internal/core/usecases/resolve_references_test.go`, covering resolution across files, order independence (FR-020), and `wrong_reference_kind` (FR-022). Include **one `unresolved_reference` case per reference position in the language** — `container.system`, `component.container`, `uses.target`, `instance.of`, `view.include`, and `view.exclude` — so SC-002's "every reference position" claim is measured rather than asserted
- [ ] T032 [P] [US1] Add a view fixture in `internal/adapters/hclsource/testdata/view_broken_reference/` and a table test in `internal/core/usecases/resolve_references_test.go` asserting that an unresolvable reference in a view's `include` or `exclude` is an error, not silently dropped (FR-016)
- [ ] T033 [P] [US1] Write structural validation table tests in `internal/core/usecases/validate_structure_test.go` for `wrong_parent_kind` (component under a system, container under a container), `duplicate_declaration` naming both sites, and `containment_cycle`
- [ ] T034 [P] [US1] Write a test in `internal/core/usecases/validate_structure_test.go` asserting that the relationship cycle `api → queue → worker → api` produces **no** diagnostic (FR-032) — the single most likely design error in this feature
- [ ] T035 [P] [US1] Write warning tests in `internal/core/usecases/validate_warnings_test.go` for `orphan_element`, `empty_system`, `missing_docs`, `docs_not_found`, and `self_relationship`
- [ ] T036 [P] [US1] Add golden fixtures producing `unknown_block` (including `for_each`, `dynamic`, `variable`, `module` by name), `unknown_attribute`, and `unknown_function` in `internal/adapters/hclsource/testdata/unknown_constructs/`
- [ ] T037 [P] [US1] Write a test asserting five independent errors yield five diagnostics from one invocation (FR-031, SC-008) in `internal/core/usecases/compile_architecture_test.go`

### Implementation for User Story 1

- [ ] T038 [US1] Implement decoding for `person`, `system`, `container`, `component`, `external` and nested `uses` blocks in `internal/adapters/hclsource/decode_logical.go`, populating `BodyRanges` per attribute for precise diagnostics
- [ ] T039 [US1] Implement the symbol table and reference resolution in `internal/core/usecases/resolve_references.go` — pass 1 collects every declared address, pass 2 resolves; emits `unresolved_reference` and `wrong_reference_kind` with the reference's own range
- [ ] T040 [US1] Extend `internal/core/usecases/resolve_references.go` to resolve view `include` and `exclude` references through the same symbol table, so views are validated and not merely carried through (FR-016)
- [ ] T041 [US1] Implement parent-kind rules and the containment cycle check in `internal/core/usecases/validate_structure.go`; the cycle traversal must walk the `Parent` edge **only** and must be a separate function from any relationship traversal
- [ ] T042 [US1] Implement warnings in `internal/core/usecases/validate_warnings.go`; `docs_not_found` stats the referenced path relative to the project root
- [ ] T043 [US1] Implement `CompileArchitecture` in `internal/core/usecases/compile_architecture.go`, orchestrating source → resolve → validate → IR and accumulating diagnostics across all stages without early return (FR-031); keep under the 200 effective-line use-case budget by delegating each step
- [ ] T044 [US1] Implement JSON diagnostic serialisation conforming to `contracts/diagnostics.schema.json` in `internal/adapters/encoding/diagnostics.go`, including the `summary` counts
- [ ] T045 [US1] Implement the `loko validate` handler in `cmd/validate.go` with `--strict`, `--format text|json`, and `--path`; handler must be ≤ 50 effective lines and contain no logic beyond flag parsing, calling the use case, and formatting (FR-039)
- [ ] T046 [US1] Wire flags and registration in `cmd/validate_cobra.go` and `cmd/root.go`, asserting exit codes are identical for both `--format` values (FR-034a)
- [ ] T047 [US1] Add an end-to-end test in `cmd/validate_test.go` covering quickstart Scenario 1: clean project exits 0, `--strict` with warnings exits 2, broken reference exits 1

**Checkpoint**: `loko validate` works end to end on the logical plane.

---

## Phase 4: User Story 2 — Describe where the architecture runs, per environment (Priority: P1)

**Goal**: Two environments instantiate the same logical elements with different attributes and
bindings, and their rules are enforced.

**Independent Test**: Declare one architecture and two environments per quickstart Scenario 2; both
compile, instance addresses carry no node segment, and a duplicate instance name is rejected.

### Tests for User Story 2

- [ ] T048 [P] [US2] Add golden fixtures for a nested `node` tree with instances and bindings in `internal/adapters/hclsource/testdata/deployment_nested/`
- [ ] T049 [P] [US2] Write an address test asserting an instance inside `node "vpc-main" { node "subnet-a" { … } }` resolves to `deployment.prod.instance.api` with `placedIn` `deployment.prod.node.vpc-main.subnet-a` in `internal/core/usecases/build_ir_deployment_test.go`
- [ ] T050 [P] [US2] Write a re-parenting test asserting an instance moved between nodes keeps its address and changes only `placedIn` (FR-024, quickstart Scenario 2.6) in `internal/core/usecases/build_ir_deployment_test.go`
- [ ] T051 [P] [US2] Write deployment validation tests in `internal/core/usecases/validate_deployment_test.go` for `duplicate_claim` (two bindings on the same exact identifier), `duplicate_instance_name` across two different nodes naming both sites (FR-012a), and `unbound_instance` as a **warning** that leaves exit 0 without `--strict`

### Implementation for User Story 2

- [ ] T052 [US2] Implement `deployment`, nested `node`, `instance`, and `binding` decoding in `internal/adapters/hclsource/decode_deployment.go`; binding selectors accept `address` (exact or glob), `addresses`, or `tags`, with kind taken from the block label (`terraform` or `cloudformation`)
- [ ] T053 [US2] Implement deployment validation in `internal/core/usecases/validate_deployment.go` — `duplicate_claim`, `duplicate_instance_name`, `unbound_instance`, and `of` resolving to a logical element
- [ ] T054 [US2] Implement deployment IR construction in `internal/core/usecases/build_ir_deployment.go`: flatten instances onto `Environment.Instances` regardless of nesting depth, assign group addresses from the node path, and set `Group.Contains` and `Instance.PlacedIn` cross-references
- [ ] T055 [US2] Add an end-to-end test in `cmd/validate_test.go` for quickstart Scenario 2, asserting the exported instance address and `placedIn` values

**Checkpoint**: the deployment plane compiles, validates, and addresses correctly.

---

## Phase 5: User Story 3 — Consume the compiled architecture as a stable artifact (Priority: P1) 🎯 MVP

**Goal**: `loko export` produces a deterministic, versioned, fully-resolved artefact.

**Independent Test**: Export twice and `cmp` the outputs; perturb file discovery order and `cmp`
again; confirm `schemaVersion` is present and no run-varying value leaked.

### Tests for User Story 3

- [ ] T056 [P] [US3] Write determinism tests in `internal/core/usecases/build_ir_test.go`: two compiles of the same source produce deeply equal IRs, and a shuffled discovery order produces an identical IR
- [ ] T057 [P] [US3] Write byte-equality tests in `internal/adapters/encoding/export_test.go` for both JSON and TOON, plus a **shuffled-file-order** case — repeat-run equality alone will not catch a map that escaped into the IR (research R6)
- [ ] T058 [P] [US3] Write a leak test in `internal/adapters/encoding/export_test.go` asserting no timestamp, hostname, tool version, or absolute path appears in the export (FR-036c)
- [ ] T059 [P] [US3] Write a schema-conformance test validating exported JSON against `specs/013-hcl-compiler-core/contracts/ir.schema.json` in `internal/adapters/encoding/export_schema_test.go`
- [ ] T060 [P] [US3] Commit a pinned artefact `internal/adapters/encoding/testdata/export_v1.json` at `schemaVersion: 1` and write a test in `internal/adapters/encoding/version_compat_test.go` asserting it reads cleanly, and that a copy with the version hand-bumped to `2` is refused with a message naming both the version found and the versions supported (SC-011, FR-036b)
- [ ] T061 [P] [US3] Write a test asserting `export` writes **no artefact** and leaves any `--out` file untouched when compilation reports errors (FR-037) in `cmd/export_test.go`
- [ ] T062 [P] [US3] Write a statelessness test in `cmd/export_test.go` asserting that `validate` and `export` create no file inside the project root — no lock file, cache, or state file (FR-027), so a future cache cannot be added unnoticed
- [ ] T063 [P] [US3] Extend the empty-architecture fixture from T021 with an expected export, asserting an architecture with no elements exports an empty element set rather than erroring

### Implementation for User Story 3

- [ ] T064 [US3] Implement logical IR construction with sorted ordering in `internal/core/usecases/build_ir.go`: carry elements, relationships, **views**, and `Ignores` into the IR (FR-016 requires views to reach the compiled result, not merely be validated), then sort every slice by address byte-wise at construction, sort and de-duplicate tags, sort attributes by key, claims by kind then address, and `Ignores` lexically (FR-015, FR-040)
- [ ] T065 [US3] Set `SchemaVersion = 1` on the IR and add a consumer-side version check that refuses an unrecognised value naming the version found and the versions supported (FR-036a, FR-036b) in `internal/core/entities/ir.go`
- [ ] T066 [US3] Implement deterministic JSON encoding of the IR in `internal/adapters/encoding/json.go`, serialising in slice order and performing no sorting of its own
- [ ] T067 [US3] Extend the TOON encoder for the IR in `internal/adapters/encoding/toon.go`, carrying information equivalent to the JSON form (FR-036)
- [ ] T068 [US3] Implement `ExportIR` in `internal/core/usecases/export_ir.go`: compile, suppress the artefact entirely on error, and return diagnostics
- [ ] T069 [US3] Implement the `loko export` handler in `cmd/export.go` with `--format json|toon`, `--out`, and `--path`; artefact to stdout, diagnostics to stderr so `loko export --format json | conftest test -` works unfiltered
- [ ] T070 [US3] Wire flags and registration in `cmd/export_cobra.go` and `cmd/root.go`

**Checkpoint**: MVP complete — `validate` and `export` both work on the full language.

---

## Phase 6: User Story 4 — Keep source files canonically formatted (Priority: P2)

**Goal**: `loko fmt` canonicalises source without disturbing comments or ordering, and gates CI.

**Independent Test**: Format a messy file, confirm canonical output with comments intact; run again
and confirm no change; run `--check` on a messy file and confirm exit 1 with nothing written.

### Tests for User Story 4

- [ ] T071 [P] [US4] Write formatting fixtures in `internal/adapters/hclsource/testdata/format/` pairing messy input with canonical output, including inline comments, comments between blocks, and a deliberate declaration order
- [ ] T072 [P] [US4] Write an idempotence test asserting a second `fmt` run leaves every fixture byte-unchanged (SC-007) in `internal/adapters/hclsource/format_test.go`
- [ ] T073 [P] [US4] Write a test asserting an unparseable file is reported with a source location and left **byte-unchanged** in `internal/adapters/hclsource/format_test.go`
- [ ] T074 [P] [US4] Write `--check` tests in `cmd/fmt_test.go`: two unformatted files list both paths, write nothing, and exit 1; an all-canonical project exits 0 (FR-035a)

### Implementation for User Story 4

- [ ] T075 [US4] Implement canonical formatting via `hclwrite.Format` over the parsed token stream in `internal/adapters/hclsource/format.go`, preserving comments and declaration order and refusing to touch files that do not parse
- [ ] T076 [US4] Implement `FormatSources` in `internal/core/usecases/format_sources.go` supporting both write mode and check mode, returning the list of non-canonical paths
- [ ] T077 [US4] Implement the `loko fmt` handler in `cmd/fmt.go` with `--check` and `--path`; `--check` reuses exit code 1 and must not introduce a fourth code (FR-038)
- [ ] T078 [US4] Wire flags and registration in `cmd/fmt_cobra.go` and `cmd/root.go`

**Checkpoint**: formatting works and is CI-gateable.

---

## Phase 7: User Story 5 — Declare the architecture's compatible tool versions (Priority: P3)

**Goal**: `loko_version` is honoured, so an older build says so plainly.

**Independent Test**: Set a constraint the running tool cannot satisfy and confirm a clear error with
exit 1 and no artefact.

### Tests for User Story 5

- [ ] T079 [P] [US5] Write table tests for the constraint evaluator in `internal/core/entities/version_constraint_test.go` covering `=`, `!=`, `>`, `>=`, `<`, `<=`, `~>`, comma-separated conjunctions, pre-release ordering, and a malformed constraint producing a clear parse error
- [ ] T080 [P] [US5] Write a test asserting an absent `loko_version` produces no diagnostic, and an unsatisfiable one produces `version_unsatisfied` naming both the constraint and the running version, in `internal/core/usecases/validate_structure_test.go`

### Implementation for User Story 5

- [ ] T081 [US5] Implement the semantic-version constraint evaluator using the standard library only in `internal/core/entities/version_constraint.go` (research R5 — no `hashicorp/go-version`, since core takes no dependencies)
- [ ] T082 [US5] Evaluate the project's constraint against the build version during compilation and emit `version_unsatisfied` in `internal/core/usecases/compile_architecture.go`

**Checkpoint**: version constraints enforced.

---

## Phase 8: User Story 6 — Retire the superseded dual-source model (Priority: P2)

**Goal**: One model, no drift machinery, green build.

**⚠️ MUST RUN LAST.** This removes 44 non-test files' worth of dependencies. The build will be red
between T084 and T091; do not interleave it with other phases. See [research.md](./research.md) R9.

**Independent Test**: `task test`, `task lint`, and `task audit-constitution` all pass; `loko --help`
offers no drift option and no removed command; `go build ./...` is green.

### Implementation for User Story 6

- [ ] T083 [US6] Delete the drift machinery: `internal/core/usecases/detect_drift.go`, the drift entity, `internal/adapters/d2/d2_relationship.go`, and every `--check-drift` flag and its help text (FR-043)
- [ ] T084 [US6] Delete the v0 model adapters — `internal/adapters/filesystem/project_repo.go`, `relationship_repo.go`, `config.go`, the whole `internal/adapters/config/` and `internal/adapters/ason/` packages — and remove the `ProjectRepository`, `ConfigLoader`, and `TemplateEngine` ports from `internal/core/usecases/ports.go`
- [ ] T085 [US6] Delete the HTTP interface: `internal/api/` (including `handlers/`, `middleware/`, `static/`, and swagger), plus `cmd/api.go` and `cmd/api_cobra.go` (FR-041)
- [ ] T086 [US6] Delete the scaffolding and v0-tree use cases: all `scaffold_*`, all `create_*`, `delete_relationship`, `find_relationships`, `list_relationships`, `search_elements`, `query_architecture*`, `build_docs*`, `render_markdown_docs`, `update_diagram`, `init_project`, `select_template`, `enhance_component_diagram`, `render_diagram_preview`, and `validate_architecture*` under `internal/core/usecases/`
- [ ] T087 [US6] Delete the MCP tools backed by the removed use cases under `internal/mcp/tools/`, updating `registry.go`, and keep the MCP server harness and `graph_cache.go` in place for the authoring stage
- [ ] T088 [US6] Delete the CLI commands backed by the removed use cases: `cmd/build.go`, `serve.go`, `watch.go`, `new.go`, `init.go` and their `*_cobra.go` files, and update the registration list in `cmd/root.go`
- [ ] T089 [US6] Delete every test file covering the removed code under `internal/core/usecases/`, `internal/adapters/`, `internal/mcp/tools/`, `internal/api/`, and `cmd/` — do **not** adapt them to keep passing (FR-042)
- [ ] T090 [US6] Verify `internal/adapters/d2/parser.go` and `d2_import_test.go` are still present and still compile; they are parked for the diagram-import stage (FR-045), as is `internal/adapters/html/` left on disk unwired
- [ ] T091 [US6] Run `go mod tidy` in the repository root to drop the orphaned modules (ason, and TOML handling if nothing else imports it), then confirm `go build ./...` is green
- [ ] T092 [US6] Run `task test`, `task lint`, and `task audit-constitution`; confirm the layer rule from T003–T005 passes and coverage on `internal/core/` is above 80%
- [ ] T093 [US6] Add an end-to-end test in `cmd/root_test.go` asserting `loko --help` lists exactly `validate`, `fmt`, `export`, `version`, `completion`, `mcp` and offers no drift option

**Checkpoint**: one model, green build, no reachable drift detection.

---

## Phase 9: Polish & Cross-Cutting Concerns

- [ ] T094 [P] Write `docs/adr/0012-hcl-source-of-truth.md` covering HCL as layer 0, the static-traversal reference decision (research R2), the `SourceModel`/IR split (R7), and the deletion of the dual-source model (R9) — required by the constitution's Governance section for any new architectural decision
- [ ] T095 [P] Amend `.specify/memory/constitution.md`: update the Technology Stack section (TOML configuration, ason template engine, shelled-out d2 CLI are removed or superseded) and the External Dependencies table, and bump the version with a sync-impact note
- [ ] T096 [P] Build the synthetic project generator in `tools/genfixture/main.go` producing N elements across M files, for the performance bounds in SC-006
- [ ] T097 Add a performance test asserting 1,000 elements across 200 files compiles in under 2 seconds and 5,000 elements in under 10 seconds in `internal/core/usecases/performance_test.go`
- [ ] T098 Add `TestValidationRuleCoverage` in `internal/core/usecases/rule_coverage_test.go`, enumerating every code in `contracts/diagnostics.schema.json` — all 19, including `syntax_error` — and failing if any lacks a fixture. This is the mechanical check behind SC-004; without it, "100% of the rule list" is a claim rather than a fact
- [ ] T099 [P] Write `docs/language.md` from `contracts/language.md` as user-facing documentation, with a worked example for every block, attribute, and function so SC-005's documentation review can pass, plus a "Migrating from v0.2" page describing the MCP-driven recipe and noting that the `v0.2.x` tag stays installable
- [ ] T100 [P] Update `README.md` for the new command set, stating plainly that `build`, `serve`, and `init` return in the next release; then run every scenario in [quickstart.md](./quickstart.md) end to end against the built binary and record the results

---

## Dependencies & Execution Order

### Phase dependencies

- **Phase 1 (Setup)**: no dependencies
- **Phase 2 (Foundational)**: depends on Phase 1 — **blocks every user story**
- **Phase 3 (US1)**: depends on Phase 2
- **Phase 4 (US2)**: depends on Phase 2; shares `build_ir` scaffolding with US3 but has its own files
- **Phase 5 (US3)**: depends on Phase 2; `build_ir.go` (T064) pairs with `build_ir_deployment.go` (T054); T063 depends on the T021 fixture
- **Phase 6 (US4)**: depends on Phase 2 only — genuinely parallel with US1/US2/US3
- **Phase 7 (US5)**: depends on Phase 2 only — genuinely parallel with US1/US2/US3
- **Phase 8 (US6)**: depends on **every** other phase being green
- **Phase 9 (Polish)**: depends on Phase 8

### Critical path

`T001 → T008/T009/T010 → T013 → T017 → T022–T026 → T038–T043 → T064 → T068 → T069 → Phase 8`

### Within each user story

- Tests before implementation (Constitution Principle V)
- Entities before use cases before adapters before commands
- Use-case files stay ≤ 200 effective lines; split by step rather than growing one file

### Parallel opportunities

- **Phase 1**: T004 and T005 together, **after** T003 is written and confirmed failing
- **Phase 2**: T006–T007 together; then T008, T009, T011, T012, T013, T014, T015, T016 all together (different files); T018–T021 together; T024 and T027 together
- **Phase 3**: T030–T037 all together
- **Phase 4**: T048–T051 all together
- **Phase 5**: T056–T063 all together
- **Phase 6**: T071–T074 all together
- **Phase 7**: T079–T080 together
- **Phases 6 and 7 can run alongside Phases 3–5** with separate developers
- **Phase 9**: T094, T095, T096, T099, T100 together

---

## Parallel Example: Phase 2 entities

```bash
# After T006 and T007 are written and failing, launch the entity implementations together:
Task: "Implement Address in internal/core/entities/address.go"
Task: "Implement SourceRange in internal/core/entities/source_range.go"
Task: "Implement SourceModel types in internal/core/entities/source_model.go"
Task: "Implement deployment declarations in internal/core/entities/source_model_deployment.go"
Task: "Implement IR root and indexes in internal/core/entities/ir.go"
Task: "Implement Element and Relationship in internal/core/entities/ir_element.go"
Task: "Implement deployment IR types in internal/core/entities/ir_deployment.go"
Task: "Implement View in internal/core/entities/ir_view.go"
```

---

## Implementation Strategy

### MVP scope

The MVP is **Phase 1 + Phase 2 + Phase 3 (US1) + Phase 5 (US3)** — roughly T001–T047 plus T056–T070.
US1 alone is not a usable increment, because a compiler that validates but cannot emit its compiled
value gives downstream stages nothing to consume and gives the test suite nothing to assert against.
Adding US3 is what makes the golden-file strategy work at all.

### Incremental delivery

1. Phases 1–2 → foundation, nothing user-visible yet
2. + Phase 3 + Phase 5 → **MVP**: `loko validate` and `loko export` on the logical plane
3. + Phase 4 → deployment plane; the language is now complete
4. + Phase 6 → `loko fmt`, CI-gateable
5. + Phase 7 → version constraints
6. + Phase 8 → one model, drift machinery gone
7. + Phase 9 → ADR, constitution amendment, docs, performance verification

### Do not start Phase 8 early

It is tempting to delete the old model first to stop maintaining two worlds. Resist it: the deletion
removes the dependency base for 44 files, so starting it before the new path is green means debugging
new compiler code against a build that will not compile for unrelated reasons.

### Parallel team strategy

After Phase 2: Developer A takes US1 + US3 (the critical path), Developer B takes US2, Developer C
takes US4 + US5. All three converge before Phase 8, which one person should perform in a single pass.

---

## Notes

- `[P]` means different files with no incomplete dependencies
- Tests are mandatory here, not optional — Constitution Principle V
- Commit after each task or logical group
- Every new use-case file is subject to the 200 effective-line budget; every entity file to 300; CLI
  handler functions to 50. `task audit-constitution` enforces all three.
- Stop at any checkpoint to validate independently
