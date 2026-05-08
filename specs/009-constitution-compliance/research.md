# Phase 0 Research: Constitution Compliance Refactor

This document resolves the open technical decisions identified in `plan.md`. Every decision is recorded as **Decision / Rationale / Alternatives considered** so that the chain of reasoning is traceable from the constitution and spec down to the executed implementation.

## R1. Audit-tool implementation: shell vs. third-party arch-linter vs. custom Go AST binary

**Decision**: Build a small custom Go AST binary at `tools/archcheck/`. It reads the rule set from `specs/009-constitution-compliance/contracts/structural-rules.yaml` (canonical location is the project's `.specify/memory/constitution.md`; the YAML is a machine-consumable mirror), walks the module via `go/parser` + `go/ast`, and emits a structured violation report (JSON for CI, text for humans).

**Rationale**:

- The spec requires per-function line budgets for handlers (FR-001, FR-002). Shell pipelines (the current `scripts/audit-constitution.sh`) cannot identify function boundaries reliably; doing so via `awk` would re-implement a Go parser badly.
- The same binary needs to enforce layer-import rules (FR-005..FR-008). Both checks reach for the AST anyway, so combining them keeps the rule set in one place.
- A custom Go binary lets us match the project's existing "effective-line" counting convention exactly (ignore blank lines, `//` comments, `/* */` blocks, `import (...)` groups, the `package` declaration). This is necessary to verify against the existing `scripts/audit-constitution.sh` output during the migration window — we run both, expect identical whole-file counts, then retire the shell version.
- Keeping the binary in `tools/archcheck/` (outside `internal/`) makes it explicit that this is build-time tooling, not product code, and avoids confusing the very rules it enforces.

**Alternatives considered**:

- **`go-arch-lint`** (a popular open-source Go architecture linter) — handles layer-import rules well, but does not enforce per-function effective-line budgets and brings ~20 transitive deps into the toolchain. Rejected.
- **`golangci-lint` `depguard` rule alone** — enforces import boundaries declaratively, but offers no per-function size enforcement. We still adopt `depguard` as a *redundant* layer-import check (cheap insurance), but it cannot be the only mechanism.
- **`golangci-lint` `funlen` linter** — counts function length, but its definition of "length" is not the project's "effective-lines" definition; it counts statements or non-blank-non-comment lines depending on configuration. Adopting it would diverge from the existing convention and force a retroactive recount of every "known violation" — extra churn for no benefit. Rejected as the primary check; not adopted as a redundant check either, because `archcheck` already covers this.
- **Continuing to extend `scripts/audit-constitution.sh`** — would require an inline Go-tokeniser in shell, which is not realistic, and would still leave layer-import enforcement unsolved. Rejected.

## R2. Layer-import enforcement: depguard vs. archcheck-only vs. both

**Decision**: Both. `archcheck` is the authoritative enforcement (single rule source, deterministic, versioned with the codebase); `depguard` (configured in `.golangci.yml`) is a redundant secondary check that runs as part of `make lint` for fast IDE feedback.

**Rationale**:

- `depguard` runs in `golangci-lint`, which contributors invoke continuously via editor integrations and `make lint`. Failing fast in the editor is high-leverage feedback.
- `archcheck` is the source of truth because it shares its rule file with the file-size and function-size checks; one violation report covers all four rule families and the messages are uniform.
- Running both is safe: the rule data is mirrored from `structural-rules.yaml` into `.golangci.yml` (a small one-time mechanical translation). If the two ever diverge, `archcheck` (run in CI) wins and the divergence becomes a visible failure that prompts re-syncing the depguard config.

**Alternatives considered**:

- **`depguard` only**: cheaper but doesn't cover file-size or function-size rules; would still need a separate tool for those. Rejected.
- **`archcheck` only**: forces contributors to remember to run a project-specific binary instead of the standard Go linter. Rejected for ergonomic reasons.

## R3. Effective-line counting at function granularity

**Decision**: For each function declaration in a Go file, `archcheck` computes effective lines as `count_effective_lines(text_between(funcDecl.Pos(), funcDecl.End()))`, where `count_effective_lines` reproduces the rule already implemented in `scripts/audit-constitution.sh`:

- Drop blank lines (`/^\s*$/`).
- Drop the `package` declaration line(s).
- Drop single-line imports (`import "..."`).
- Drop `import (...)` blocks (everything from the opening to the matching `)`).
- Drop single-line comments (`//`).
- Drop block comments (`/* ... */`), including multi-line.

The `package` clause is never inside a function so it drops out trivially. `import` blocks are likewise outside function bodies. The interesting per-function exclusions in practice are blank lines and comments.

**Rationale**:

- Bit-for-bit consistency with the existing whole-file counter is required during the migration window so that whole-file counts agree on the same files. Once `scripts/audit-constitution.sh` is retired, the function-level counts are the only ones that matter for handlers.
- Using token positions from `go/ast` rather than re-tokenising the file ourselves avoids subtle bugs (string literals containing `//`, raw string literals containing `/* */`, etc.).

**Alternatives considered**:

- **Cyclomatic complexity** (`gocyclo`) — measures branches, not size. Different concept, not what the spec asks for. Rejected.
- **AST node count** — clean but disconnected from the lived experience of "this function is too long when I look at it." Rejected.
- **Statement count** — closer, but still differs from the existing convention. Rejected for migration-friction reasons.

## R4. Pre-existing oversized non-handler files: in-scope or follow-up?

**Decision**: All in scope for this feature. Each gets a dedicated splitting task in Phase 2 (`/speckit.tasks`).

The full list of files exceeding the new budgets at the start of this feature:

| File | Current effective lines (whole-file) | Budget | Action |
|------|--------------------------------------|--------|--------|
| `internal/core/entities/graph.go` | 678 | 300 | Split into `graph.go` (Graph type, builders), `graph_edges.go` (edge ops), `graph_query.go` (queries/traversal). |
| `internal/core/usecases/query_architecture.go` | 527 | 200 | Split by query type: `query_architecture.go` (top-level dispatcher), `query_dependencies.go`, `query_related_components.go`, `query_project_overview.go`. |
| `internal/core/usecases/build_docs.go` | 479 | 200 | Split into `build_docs.go` (orchestration), `build_docs_d2.go` (D2 step), `build_docs_markdown.go` (markdown step), `build_docs_assets.go` (template/asset emission). |
| `internal/core/usecases/build_architecture_graph.go` | 469 | 200 | Split into `build_architecture_graph.go` (orchestration), `architecture_graph_systems.go`, `architecture_graph_components.go`, `architecture_graph_relationships.go`. |
| `internal/core/usecases/ports.go` | 426 | 200 | Split by port: `ports_repository.go`, `ports_renderer.go`, `ports_template.go`, `ports_encoder.go`, `ports_clock.go`, `ports_filesystem.go`. |
| `internal/core/usecases/validate_architecture.go` | 368 | 200 | Split into `validate_architecture.go` (orchestration), `validate_drift.go`, `validate_relationships.go`, `validate_components.go`. |
| `internal/core/usecases/scaffold_entity.go` | 364 | 200 | Split into `scaffold_project.go`, `scaffold_system.go`, `scaffold_container.go`, `scaffold_component.go` (matches the pattern requested in the spec for new/build decomposition). |
| `internal/core/entities/system.go` | 255 | 300 | Already within budget — no action. |
| `cmd/new.go` | 356 (raw) — measured effective ≤ 200 | per-function 50 | Decompose so each handler function ≤ 50 effective lines; logic moved to `usecases/scaffold_*.go`. |
| `cmd/build.go` | 232 (raw) — measured effective ≤ 150 | per-function 50 | Decompose so each handler function ≤ 50 effective lines; logic delegated to `usecases/build_docs.go`. |
| `internal/mcp/tools/graph_tools.go` | 3 tools in one file | per-function 30 | Split into `query_dependencies.go`, `query_related_components.go`, `analyze_coupling.go`. |
| `internal/mcp/tools/create_component.go` | preview logic inline | per-function 30 | Move preview generation into a use case (or extend `create_component` use case). |
| `internal/mcp/tools/{create_system,update_system,update_component,build_docs,validate_diagram,...}.go` | inline `InputSchema` blocks | per-function 30 | Move schemas to `internal/mcp/tools/schemas.go` (already a data-file). |

**Rationale**: SC-001 ("zero violations across the entire codebase") is non-negotiable. Either the audit tool detects every file, or it is weakened to look the other way at known problem files — and the latter is exactly the `KNOWN_VIOLATIONS` mechanism the spec sets out to retire. Including these files in scope is what makes the feature internally consistent.

**Alternatives considered**:

- **Defer non-handler splits to v0.3.0** — would require keeping `KNOWN_VIOLATIONS` populated, contradicting the spec's intent. Rejected.
- **Weaken the file-size rule** — would also contradict the spec. Rejected.
- **Apply the new file-size rule only to files modified in this branch** — clever but creates two classes of files in the codebase, which is the worst kind of inconsistency. Rejected.

## R5. Exemption mechanism after the refactor

**Decision**: Two exemption categories, encoded in `structural-rules.yaml`:

1. **Data-file exemption** (file-size only — layer rules still apply): files whose content is type/data declarations rather than logic. Match by basename: `schemas.go`, `registry.go`, `helpers.go`, `constants.go`. Plus `*_cobra.go` (Cobra flag wiring is data-shaped) and `*_test.go` (tests are exempt from production-code budgets).
2. **Generated-file exemption** (file-size only): files whose first non-blank line matches the regex `^// Code generated.+DO NOT EDIT\.$` (the Go convention). No file currently matches this; the rule is forward-looking.

The pre-existing `KNOWN_VIOLATIONS` allowlist in `scripts/audit-constitution.sh` is **deleted at the end of the refactor**. If a new file legitimately needs to exceed a budget after the migration, the path forward is (a) refactor the file, (b) add a permanent exemption category in `structural-rules.yaml` plus an ADR explaining it, or (c) amend the constitution to relax the limit. Per-file ad-hoc allowlist entries are not permitted post-migration.

**Rationale**:

- The known-violations allowlist exists today precisely because the original audit lacked the granularity to enforce the right thing. Once we have per-function checks and a richer rule engine, the allowlist's reason-for-being is gone.
- Keeping data-file and generated-file exemptions as *categories* (matched by pattern) rather than *individual entries* (matched by path) is the difference between a rule and an escape hatch. The categories are auditable in code review; ad-hoc entries are not.

**Alternatives considered**:

- **Keep the per-file allowlist mechanism** — ergonomic for one-off cases but creates a long-lived "TODO list of debt" file that nobody owns. Rejected.
- **No exemptions at all** — would force `*_cobra.go` files (e.g., `new_cobra.go` 215 lines, all flag declarations) to be split artificially, fighting against Cobra's idiom. Rejected.

## R6. Constitution amendment process

**Decision**: This feature ships the amendment to `.specify/memory/constitution.md` from v1.0.0 → v1.1.0 in the same PR sequence as the audit tool. The amendment changes:

- Section "Quality gates": replace "Handler line counts within limits (CLI < 150, MCP < 100); known violations documented in scripts/audit-constitution.sh" with "Handler function sizes within limits (CLI handler functions ≤ 50 effective lines, MCP tool handler functions ≤ 30) and core file sizes within limits (use-case files ≤ 200, entity files ≤ 300); enforced by `make audit-constitution`."
- Add a new bullet: "Layer-import rules are mechanically enforced by `tools/archcheck` and `golangci-lint depguard`; violations block CI."
- Bump the footer: `**Version**: 1.1.0 | **Ratified**: 2026-02-06 | **Last Amended**: 2026-05-08`.
- Cross-reference ADR-0009 and `specs/009-constitution-compliance/`.

**Rationale**: The constitution itself defines the amendment process: "Amendments require: documented rationale, review of impact on existing code, and a migration plan if breaking." All three are present:

- *Rationale*: production design document (cited in `spec.md`).
- *Impact review*: R4 above lists every file affected.
- *Migration plan*: the Phase 2 task list (to be generated by `/speckit.tasks`).

**Alternatives considered**:

- **Ship the audit tool with v1.0.0 limits, defer amendment**: contradicts spec FR-001..FR-004. Rejected.
- **Major version bump (v2.0.0)**: not warranted; the architecture is unchanged, only quantitative limits and the enforcement mechanism changed. Rejected.

## R7. CI gating step format

**Decision**: A new top-level step in `.github/workflows/ci.yml`, sequenced after `make build` and before `make test`, runs `make audit-constitution`. The Makefile target invokes `tools/archcheck` with `--format=json --report=audit-report.json` and a non-zero exit on any violation. The job uploads `audit-report.json` as an artefact on failure so the contributor can see every violation, not just the first.

For pull-request UX, the CI step also writes a GitHub-Actions-annotated text format on stderr (`::error file=path::message`) so violations appear inline on the diff in the PR Files-changed view.

**Rationale**:

- Failing before tests run gives the fastest feedback for a structural problem (no need to wait for the full test suite to learn that an import is misplaced).
- The JSON artefact is what `archcheck` produces natively; the text annotations are a presentation layer for GitHub specifically. Both come from the same run, so they cannot disagree.

**Alternatives considered**:

- **Run audit only on `main`**: defeats the purpose; PRs would land violations and the gate would only catch them post-merge. Rejected.
- **Run audit as a separate workflow**: adds latency (separate runner spin-up). Rejected.

---

## Summary of decisions

| # | Decision | Drives |
|---|----------|--------|
| R1 | Custom Go AST binary `tools/archcheck` | FR-001..FR-004, FR-014 |
| R2 | `archcheck` + `depguard` redundant pair | FR-005..FR-008, FR-013 |
| R3 | Reuse existing effective-line convention at function granularity | FR-001..FR-002, FR-019 |
| R4 | All pre-existing oversized files refactored in this feature | SC-001, FR-003..FR-004 |
| R5 | Categorical exemptions only (data-files, generated-files); per-file allowlist removed | FR-018 |
| R6 | Constitution amendment to v1.1.0, ADR-0009 | FR-019, Constitution Check |
| R7 | CI gating step before tests, JSON report + GitHub annotations | FR-013, FR-014, SC-004 |
