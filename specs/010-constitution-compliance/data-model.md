# Phase 1 Data Model: Constitution Compliance Refactor

**Feature**: 010-constitution-compliance
**Date**: 2026-05-20

This document models the conceptual entities the structural compliance check operates on. These are **tooling-internal** types (not domain entities): they live in `tools/archcheck` and never cross the import boundary into `internal/core/`. They are documented here because the spec's "Key Entities" section names them and because the `contracts/` files schema-ise them.

---

## Entity: SourceFileCategory

**What it represents**: The classification the audit tool assigns to a single Go source file. The category determines which size budget (if any) applies and which set of layer-import rules applies.

**Values** (enum):
- `cli_handler` — file under `cmd/`, not `*_cobra.go`, not `*_test.go`, not generated.
- `mcp_tool` — file under `internal/mcp/tools/`, not data-file, not `*_test.go`, not generated.
- `api_handler` — file under `internal/api/`, not `*_test.go`, not generated. (Budget deferred per Principle III; layer rule still applies.)
- `use_case` — file under `internal/core/usecases/`, not `*_test.go`, not generated. Includes `ports.go`.
- `entity` — file under `internal/core/entities/`, not `*_test.go`, not generated.
- `adapter` — file under `internal/adapters/<name>/`, not `*_test.go`, not generated.
- `data_file` — filename matches `schemas.go | registry.go | helpers.go | constants.go` (per Principle III categorical exemption). Subject to layer rules but **not** size budgets.
- `cobra_wiring` — filename matches `*_cobra.go`. Subject to layer rules but **not** size budgets.
- `test_file` — filename matches `*_test.go`. Exempt from production-code size budgets; subject to a separate (more generous) test-file budget — initial value: **none** (no budget on test files in 1.0 of this rule set; revisit if test files become a problem).
- `generated` — file's first non-blank line matches the `// Code generated .* DO NOT EDIT\.` regex. Subject to layer rules but **not** size budgets.
- `other` — anything not matching the above (e.g., `tools/archcheck/` files themselves, `cmd/loko/main.go`, package-private test helpers). Subject to layer rules only.

**Derivation**: The audit tool's `categorize(path string, header []byte) SourceFileCategory` function applies the rules above in priority order (generated > test > cobra-wiring > data-file > path-based bucket).

**Invariants**:
- Every Go source file in the repository has exactly one category.
- Category is computable from path + first-line read, without parsing the file.

---

## Entity: LayerRule

**What it represents**: A statement of the form "files in category X may import only from import-path patterns Y₁, Y₂, … and must not import from import-path patterns Z₁, Z₂, …".

**Fields**:
- `name` — short stable identifier, e.g. `entities-pure`, `usecases-only-entities`, `outer-no-entities`. Used in violation messages and suppression entries.
- `applies_to` — list of `SourceFileCategory` values this rule binds.
- `allow` — list of Go import-path glob patterns the file MAY import (stdlib is always implicitly allowed).
- `deny` — list of Go import-path glob patterns the file MUST NOT import. `deny` wins over `allow` on overlap.
- `message` — human-readable explanation appended to violation diagnostics.

**Rule set** (initial, mirrors constitution v1.2.0 dependency table):
| Rule | Applies to | Deny pattern | Notes |
|------|------------|--------------|-------|
| `entities-pure` | `entity` | `github.com/madstone-io/loko/internal/**` | Entities import stdlib only |
| `usecases-only-entities` | `use_case` | `github.com/madstone-io/loko/internal/**` except `…/internal/core/entities/**` | |
| `adapters-only-core` | `adapter` | `github.com/madstone-io/loko/internal/mcp/**`, `…/internal/api/**`, `github.com/madstone-io/loko/cmd/**` | |
| `outer-no-entities` (CLI) | `cli_handler` | `github.com/madstone-io/loko/internal/core/entities/**` | Constitution v1.1.0 |
| `outer-no-entities` (MCP) | `mcp_tool` | `github.com/madstone-io/loko/internal/core/entities/**` | Constitution v1.2.0 — new |
| `outer-no-entities` (API) | `api_handler` | `github.com/madstone-io/loko/internal/core/entities/**` | Constitution v1.2.0 — new |
| `outer-no-cross-talk` | `cli_handler`, `mcp_tool`, `api_handler` | The two **other** outer packages (e.g., MCP files must not import API, etc.) | |

**Invariants**:
- Every category has at least one rule (`other` has an empty rule that always passes — explicit, not implicit).
- No two rules in the same category may both `allow` and `deny` the same exact pattern.

---

## Entity: SizeBudget

**What it represents**: A maximum line count, measured in effective lines (per R1), associated with a `SourceFileCategory` and a granularity (`per_function` or `per_file`).

**Fields**:
- `category` — the `SourceFileCategory` this budget applies to.
- `granularity` — `per_function` | `per_file`.
- `limit` — positive integer (effective lines).
- `name` — short stable identifier for diagnostics, e.g. `cli-handler-func-size`, `usecase-file-size`.

**Budget set** (initial, mirrors constitution v1.2.0):
| Budget | Category | Granularity | Limit |
|--------|----------|-------------|-------|
| `cli-handler-func-size` | `cli_handler` | per_function | 50 |
| `mcp-handler-func-size` | `mcp_tool` | per_function | 30 |
| `usecase-file-size` | `use_case` | per_file | 200 |
| `entity-file-size` | `entity` | per_file | 300 |

(`api_handler` per-function budget is intentionally absent — deferred per Principle III. `entity`, `use_case`, `adapter` per-function budgets are intentionally absent — only the file-size budget applies there.)

**Invariants**:
- For any (category, granularity) pair there is at most one budget.
- `limit > 0`.

---

## Entity: ViolationReport

**What it represents**: A single failure raised by the audit tool against a single file (and optionally a single function within it).

**Fields**:
- `rule` — name of the violated rule or budget (`outer-no-entities`, `cli-handler-func-size`, etc.).
- `file` — repo-relative path.
- `function` — present only for `per_function` size violations; absent for layer and per-file size violations.
- `kind` — `layer_import` | `function_size` | `file_size`.
- `measured` — present only for size violations: the actual effective-line count.
- `limit` — present only for size violations: the budget that was exceeded.
- `import_path` — present only for `layer_import`: the offending import.
- `message` — human-readable diagnostic, formatted for terminal output.

**Output formats**:
- Text (default, for terminal + CI log): one line per violation, format: `<file>:<line> [<rule>] <message>`.
- JSON (`--format=json`): one `ViolationReport` per object, top-level array. Schema in `contracts/violation-report.schema.json`.

**Invariants**:
- A `ViolationReport` is **never** emitted for a file/function that has a non-expired matching `SuppressionEntry`. Suppressed items appear under a separate "Suppressed (N entries)" footer with their expiry dates.

---

## Entity: SuppressionEntry

**What it represents**: A scoped, owner-tagged, dated exemption for a single pre-existing violation that lets the gate land green on `main` without blocking work outside this feature's scope.

**Fields**:
- `rule` — name of the rule/budget being suppressed. Wildcards not allowed.
- `file` — repo-relative path (may be a glob, scoped narrowly to a small cluster of related files).
- `function` — optional; only relevant when `rule` is per-function.
- `owner` — GitHub handle (string starting with `@`) responsible for clearing the suppression.
- `expires_on` — ISO-8601 date (`YYYY-MM-DD`). When today's date is past `expires_on`, the entry is invalid and the underlying violation re-fires as a normal failure.
- `reason` — short prose explaining why the violation cannot be fixed inside this feature.

**State machine**:
```
                +-------------+
created  --->   |   active    |  -- expires_on reached -->  expired (=> violation re-fires)
                +-------------+
                       |
                       v
                  fix landed
                       |
                       v
                   removed
```

**Invariants**:
- `expires_on` must be ≤ 90 days from creation date (enforced by the audit tool at load time; longer-lived suppressions require a constitution amendment, not a longer expiry).
- Every active suppression must reference an existing rule or budget name (typo-protection).
- An expired suppression that has not been renewed or removed within 30 days emits a "stale suppression" warning in addition to the underlying-violation failure.

---

## Cross-entity relationships

```
SourceFileCategory ─┬─> (matched by) ──> LayerRule (1..n per category)
                    └─> (matched by) ──> SizeBudget (0..n per category, max 1 per granularity)

ViolationReport ── raised by ──> (LayerRule | SizeBudget) on a (file, optional function)

SuppressionEntry ── targets ──> (rule_name, file [, function]) ──> suppresses matching ViolationReport
```

---

## What is *not* modelled here

- **Domain entities** (`System`, `Container`, `Component`, `Relationship`) — these are unchanged by the refactor. They live in `internal/core/entities/` and are subject to the entity file-size budget but otherwise untouched.
- **External configuration** — the audit tool reads only its own rule and suppression files; it does not consume `loko.toml`.
- **Persistence** — the audit tool is stateless; violation reports are written to stdout (or a CI artefact) but never persisted in a database.
