# Phase 0 Research: Constitution Compliance Refactor

**Feature**: 010-constitution-compliance
**Date**: 2026-05-20

This document resolves every NEEDS CLARIFICATION raised by the plan's Technical Context and records the architectural decisions that shape Phase 1. All decisions are recorded as `Decision / Rationale / Alternatives considered`.

---

## R1 — Counting convention for "effective lines"

**Decision**: An effective line is a source line that remains after dropping
(a) blank lines,
(b) single-line comments (`// …`) and multi-line comments (`/* … */`) — including the comment-only lines inside docstrings,
(c) the `package` declaration line, and
(d) every line inside an `import (…)` block (including the `import` keyword line itself).
The `tools/archcheck` line counter is the canonical implementation; manual estimates are non-authoritative.

**Rationale**: This convention is already defined verbatim in constitution Principle III and is already implemented in `tools/archcheck/lines.go` under feature 009. Reusing the same definition prevents two competing "what counts as a line?" answers in the codebase. The constitution version pins it (one source of truth).

**Alternatives considered**:
- *Raw `wc -l`*: fails because it punishes commenting; would discourage docstrings.
- *Tokens-per-function instead of lines*: more precise but harder for a reviewer to eyeball; rejected for ergonomics.
- *AST node count*: same ergonomic problem.

---

## R2 — How to enforce the rules: custom AST tool, off-the-shelf linter, or both?

**Decision**: Use both, with clearly separated roles:
1. `tools/archcheck` (custom Go AST binary) is the **authoritative** check. It owns: per-function effective-line budgets (CLI ≤ 50, MCP ≤ 30), per-file effective-line budgets (use-case ≤ 200, entity ≤ 300), layer-import rules, categorical exemptions, suppression-file evaluation, and machine-readable violation reports.
2. `golangci-lint depguard` is a **redundant fast-path** check that catches layer-import violations during normal lint runs. It is intentionally narrower than `archcheck` (it cannot do per-function line counting or categorical exemptions), but it catches the most common class of regression (a casual import in the wrong direction) instantly during `task lint`.

**Rationale**: A single off-the-shelf linter cannot do all of: per-function effective-line counts with the exact convention from R1, file-size budgets restricted to specific glob patterns, categorical filename exemptions (`*_cobra.go`, `*_test.go`, generated headers), and a suppression file. A custom tool can. But running an AST-aware tool on every save is heavy; `depguard` is fast enough to live inside the normal lint pass and catches the most common mistake instantly.

**Alternatives considered**:
- *`go-arch-lint` (only)*: cannot do per-function effective-line budgets; rules use yet another DSL we'd have to learn and maintain.
- *`depguard` (only)*: cannot do line budgets at all.
- *`archcheck` (only, no depguard)*: doubles developer feedback latency for the most common mistake (cross-layer import) because it would only fire during `task audit-constitution`, not during `task lint`.

---

## R3 — `depguard` configuration for layer rules

**Decision**: Encode the dependency-direction table from the constitution as one `depguard` rule per layer. Each rule names a list of allowed import patterns and an explicit `deny` for the forbidden patterns. Patterns are package import paths under `github.com/madstone-io/loko/…` (or whatever module path `go.mod` declares). The configuration lives in `.golangci.yml` and is the runtime artefact; `specs/010-constitution-compliance/contracts/structural-rules.yaml` is the human-readable mirror that ships with the spec and is referenced by the constitution governance footer.

**Rationale**: `depguard` v2 supports per-file-glob rule scoping (so the rule for `internal/core/entities/` only applies to those files), pattern-based allow/deny lists, and rich diagnostic messages. It is already a `golangci-lint` plugin in the project's toolchain. The runtime YAML is duplicated by design so the spec stays self-contained.

**Alternatives considered**:
- *`gomodguard`*: oriented toward module-level allowlists, not directory-level layer rules; awkward fit.
- *Hand-rolled `go list` parser*: reinvents `depguard` with no benefit.
- *Putting the rules only in `archcheck` (no `depguard`)*: see R2 — loses fast-path feedback during `task lint`.

---

## R4 — Suppression mechanism (FR-017) design

**Decision**: A single repo-root file `.archcheck-suppressions.yaml` enumerates each suppression entry as `{ rule, file, owner, expires_on, reason }`. The audit tool loads it at start, normalises paths, and silently drops matching violations from the failure report while still printing them under a "Suppressed (will fail after expiry)" footer. Any entry whose `expires_on` is in the past is converted back into a normal failure (the gate fails). Entries cannot be wildcarded across rules (`rule` is required and specific); they may use file globs to cover a small cluster of related files.

**Rationale**: A flat, single-file design is the lightest weight option that still satisfies the audit-quality requirements (owner + expiry, surface in tooling, not silent). Expiry-as-failure trick converts the suppression file into a self-cleaning ledger: an unmaintained suppression eventually becomes a CI failure, forcing either renewal or fix.

**Alternatives considered**:
- *Inline `//nolint:archcheck-XYZ`-style pragmas*: scatters suppression across the codebase, no central ledger, easy to forget.
- *Per-rule allowlist files*: more files, more boilerplate, no advantage.
- *No suppression mechanism* (force a full clean-up first): rejected — see Complexity Tracking justification in plan.md.

---

## R5 — CI integration approach

**Decision**: Add one new job step `Audit constitution` to `.github/workflows/ci.yml`, which runs `task audit-constitution`. The task builds `tools/archcheck` (cached across runs) and invokes it with the structural-rules YAML and the suppressions file. The step is **required** for the branch protection rules of `main`; failure blocks the merge button. Existing `Build` and `Test` jobs are not modified. The audit step also publishes the JSON violation report as a workflow artefact so a contributor can download it instead of scrolling through logs.

**Rationale**: One additional step is the smallest possible change to the existing CI; it composes cleanly with the existing `Test` and `Lint` steps without reordering them. Branch protection — not workflow YAML — is what makes the step gating, matching the existing pattern for `Test`.

**Alternatives considered**:
- *Run `archcheck` inside the `Lint` step*: couples two distinct concerns and makes failures harder to attribute on the run summary page.
- *Run `archcheck` as a pre-commit hook only*: leaves the gate at the wrong level (developer machine, not CI); easy to bypass with `--no-verify`.

---

## R6 — Test-coverage gating

**Decision**: FR-016 ("coverage must not decrease for touched packages") is enforced by a small CI script that runs `go test -coverprofile=cover.out ./...`, parses the per-package coverage with `go tool cover`, and compares each touched package's number against the value recorded at the merge-base of the PR. A regression beyond a 0.5-pp tolerance fails the CI job. The list of "touched packages" is computed from `git diff --name-only $merge_base`.

**Rationale**: Treats coverage as a per-package floor rather than a global one. A global floor would mask a major regression in one package by improvements in another; the per-package floor matches the spec's intent ("packages touched by the refactor").

**Alternatives considered**:
- *Global floor only*: rejected — masks per-package regressions.
- *Codecov / Coveralls*: adds a third-party dependency to the gating path; current tooling can do this in ~30 lines of bash + `go tool cover`.

---

## R7 — Constitution amendment v1.1.0 → v1.2.0

**Decision**: Ship a constitution amendment as part of this feature. The amendment tightens one row of the dependency-direction table:

| Layer | May Import (v1.2.0) | Must Not Import (v1.2.0) |
|-------|---------------------|---------------------------|
| `internal/mcp/` | core/usecases, adapters | `internal/core/entities/` directly; api; cmd |
| `internal/api/` | core/usecases, adapters | `internal/core/entities/` directly; mcp; cmd |

(The `cmd/` row already forbids direct entity imports under v1.1.0 — no change.)

The amendment is accompanied by `docs/adr/0010-tighten-outer-layer-entity-import-rule.md` and a `SYNC IMPACT REPORT` header update mirroring the v1.0.0 → v1.1.0 precedent. The constitution version goes to **1.2.0** (MINOR — tightens a principle's surface area without redefining it incompatibly with the principle's existing intent).

**Rationale**: The user's feature description names all three outer entry-points as bound by the no-direct-entity-import rule. Feature 009 deliberately deferred the mcp/api tightening to keep that feature's scope narrow; this feature owns the deferred work. Doing the amendment under 010 closes the gap noted in the v1.1.0 SYNC IMPACT REPORT ("Tightening that is a future-feature concern").

**Alternatives considered**:
- *Scope FR-008 down to `cmd/` only*: rejected — user-stated requirement is broader; would leave the same can-kick for a third feature.
- *Ship the amendment under a separate constitution-only PR*: rejected — amendment and the code it enables must land together to avoid a window in which the rule is in the constitution but unenforced.

---

## R8 — Decomposition strategy for `cmd/new.go` (504 lines)

**Decision**: The handler `cmd/new.go` is split into:
- `cmd/new.go` (orchestration only; one Cobra `RunE` per subcommand, each ≤ 50 effective lines): parses flags, loads config, dispatches to a use case, renders output.
- `internal/core/usecases/scaffold_project.go`: top-level project scaffold (creates project root, `loko.toml`, top-level layout).
- `internal/core/usecases/scaffold_system.go`: add-system flow.
- `internal/core/usecases/scaffold_container.go`: add-container flow (under a chosen system).
- `internal/core/usecases/scaffold_component.go`: add-component flow (under a chosen container).

Each use-case file is targeted ≤ 200 effective lines (FR-003). If a single scaffold flow exceeds the budget, it is split further by sub-step (e.g., `scaffold_project_files.go` for the file-writing sub-step) — splits are deliberate and cohesion-preserving, not mechanical.

**Rationale**: The four-way split mirrors the four logical subcommands of `loko new` (`project | system | container | component`); each file is then independently testable with concrete mock `ProjectRepository` + `TemplateEngine` ports. This matches the constitution's "interface-first" principle and the spec's "narrative coherence" acceptance scenario (Story 4 AS3).

**Alternatives considered**:
- *Single big `scaffold.go` with internal sub-functions*: would itself exceed 200 lines; defeats the purpose.
- *One file per public method (`scaffold_new_project`, `scaffold_new_system`…)*: same as the chosen split; semantically equivalent.

---

## R9 — Decomposition strategy for `cmd/build.go` (251 lines)

**Decision**: `cmd/build.go` becomes a thin Cobra `RunE` that parses flags and delegates to an existing/extended `internal/core/usecases/build_docs.go`. The existing companion files (`build_docs_diagrams.go`, `build_docs_tables.go`) absorb their respective sub-step logic that today lives inline in `cmd/build.go`. Any further sub-step that pushes `build_docs.go` past 200 effective lines is split into a new `build_docs_<step>.go` sibling (e.g., `build_docs_render.go`, `build_docs_assets.go`).

**Rationale**: `build_docs.go` is the natural home for build orchestration; the file family already exists, so this is incremental rather than greenfield. The split-by-sub-step pattern is the one the constitution's File-Size Budgets row explicitly suggests as an example.

**Alternatives considered**:
- *One monolithic `build_docs.go`*: would exceed 200 lines; rejected by FR-003.
- *Class-style "Builder" type with methods spread across files*: more ceremony than needed; the project consistently uses free functions over structs-with-methods for use cases.

---

## R10 — MCP-tool decomposition strategy

**Decision**: For every MCP tool handler currently exceeding 30 effective lines, the per-tool file structure becomes:
- The handler function (≤ 30 eff lines): unmarshal protocol input → call a use case → marshal protocol output.
- A small "request" struct + "response" struct that owns nothing but JSON tags (lives in the same file or a sibling `*_schemas.go` — `*_schemas.go` is a categorically exempt data-file under Principle III).
- Domain logic moves to an `internal/core/usecases/` use case (creating a new one if no existing use case fits — see R8/R9 patterns).

If two MCP tools share a use case (likely for `build_docs` and the CLI's `build`), the use case is written once and called by both adapters. This is the constitution's "three interfaces share one core" promise in action.

**Rationale**: Standard MCP-adapter pattern matching constitution Principle III. The schemas-file carve-out is already enshrined in the categorical exemptions (`schemas.go`, `registry.go`, `helpers.go`, `constants.go`) so no new exemption is needed.

**Alternatives considered**:
- *Inline JSON-unmarshal calls in handler*: violates the 30-eff-line budget on tools with non-trivial schemas.
- *One mega-file per tool category*: defeats the per-handler budget and hurts grep-ability.

---

## R11 — Behaviour preservation verification

**Decision**: Behaviour preservation (FR-015) is verified by three layers:
1. **Unit tests**: existing per-package tests must all stay green.
2. **CLI golden-file tests** (under `tests/integration/cli/`): for each CLI command in scope, capture pre-refactor `stdout + stderr + exit-code + tree-of-files-produced` into a golden file; the test re-runs the command and diffs against the golden file.
3. **MCP smoke tests** (`specs/010-constitution-compliance/mcp-smoke.md`): for each MCP tool, a small JSON-RPC fixture is replayed against the running server and the response is byte-compared to a pre-refactor capture.

The golden files and JSON fixtures are captured **before** any production code is moved, so the diff at the end of the refactor is genuinely a refactor signal, not an accidental retrofit.

**Rationale**: Three layers because each catches a different class of regression: unit tests catch logic bugs, golden-file tests catch shell-surface regressions (a printf format that drifts), MCP smoke tests catch protocol-surface regressions.

**Alternatives considered**:
- *Unit tests only*: misses shell- and protocol-surface drift, which is exactly the failure mode a "no observable behaviour change" refactor is supposed to prevent.
- *Big-bang manual QA at the end*: not reproducible, not scalable, not what FR-016 asks for.

---

*All NEEDS CLARIFICATION items from the plan's Technical Context are resolved above. Proceed to Phase 1.*
