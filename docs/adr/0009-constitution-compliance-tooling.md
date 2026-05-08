# ADR 0009: Constitution-Compliance Tooling

## Status

Accepted, ratified in constitution v1.1.0 (2026-05-08).

## Context

The loko architectural constitution (`.specify/memory/constitution.md`) defines layer-import rules and handler-size limits, but until v1.0.0 these were enforced only by prose review and a per-file shell audit (`scripts/audit-constitution.sh`). Two structural gaps motivated change:

1. **Granularity**: handler size was capped per-file (CLI < 150 lines, MCP < 100), which lets one bloated function hide among small ones. The intent was always per-function — the production design document `docs/superpowers/specs/2026-05-08-loko-production-design.md` (Constitution Compliance section) makes this explicit and tightens the limits to CLI ≤ 50 / MCP ≤ 30 effective lines per function. The production design also adds whole-file budgets for the inner core (use-case files ≤ 200, entity files ≤ 300).
2. **Layer-import enforcement**: the constitution's dependency-direction table was not mechanically checked. Violations were caught only when a reviewer noticed; in practice a `KNOWN_VIOLATIONS` allowlist in the shell script accumulated debt rather than reduced it.

## Decision

Three coupled decisions:

1. **Custom Go AST audit binary at `tools/archcheck/`**, driven by a declarative YAML rule set at `specs/009-constitution-compliance/contracts/structural-rules.yaml`. The binary checks file-size budgets, function-size budgets (per `*ast.FuncDecl`), and layer-import rules in a single pass, emitting a structured violation report (JSON for CI artefacts, text + GitHub annotations for inline PR review). Effective-line counting reproduces the existing `scripts/audit-constitution.sh` rule (drop blanks, comments, `package` decl, `import (...)` blocks) at function granularity.

   *Rejected alternatives*: `go-arch-lint` (no per-function size enforcement; ~20 transitive deps); `golangci-lint funlen` (different "length" definition; would diverge from the existing convention); extending the shell script (per-function granularity needs a real parser).

2. **`golangci-lint depguard` as a redundant secondary check** for layer-import rules, configured in `.golangci.yml`. The same six-row layer table is encoded once in `structural-rules.yaml` and emitted into the depguard config via `tools/archcheck --emit-golangci-config`. A CI cross-check fails if the two diverge. Rationale: developers already invoke `golangci-lint` continuously through editor integrations; failing fast in the editor is high-leverage feedback. `archcheck` remains authoritative.

3. **Categorical exemptions only; per-file allowlist retired.** `structural-rules.yaml` exempts data-files (`schemas.go`, `registry.go`, `helpers.go`, `constants.go`), Cobra flag-wiring files (`*_cobra.go`), test files (`*_test.go`), and generated files (`// Code generated ... DO NOT EDIT.`) from file-size and function-size budgets — but never from layer-import rules. The `KNOWN_VIOLATIONS` array in `scripts/audit-constitution.sh` is removed. New legitimate exemptions require editing the YAML rule set plus an ADR; ad-hoc per-file entries are not permitted.

   The constitution itself is amended to v1.1.0 in the same PR sequence to align prose authority with executable rules. The amendment process (documented rationale, impact review, migration plan) is satisfied by `specs/009-constitution-compliance/`.

## Consequences

**Positive:**

- One tool, one rule file, three rule families. Adding a layer or changing a budget is a YAML edit plus a CI re-run.
- Per-function granularity prevents the "one bloated function hidden in a small file" failure mode.
- The audit step in CI produces inline GitHub annotations on the PR Files-changed view, so violations land where a reviewer is already looking.
- Retiring the `KNOWN_VIOLATIONS` allowlist removes a long-lived TODO list of debt that nobody owned. The audit either passes cleanly or it doesn't.

**Negative:**

- Adopting tighter limits requires refactoring pre-existing oversized files (cmd/new.go, cmd/build.go, several MCP tools, several oversized core files). Tracked under feature 009 (~108 tasks).
- A custom audit tool is project-maintained code, not an off-the-shelf linter. Long-term cost: keep `tools/archcheck` deps minimal, lean on stdlib `go/ast`.
- The depguard mirror of the layer rules introduces a sync risk; mitigated by the CI cross-check that asserts equivalence.

**Mitigations:**

- A baseline-violations.json snapshot (committed to `specs/009-constitution-compliance/`) records the starting state so progress is measurable per-PR.
- The legacy `scripts/audit-constitution.sh` becomes a `--legacy-shim` wrapper for one release, then is deleted; no muscle-memory disruption during the migration window.
- Feature 009's spec deliberately does *not* tighten the rules for `internal/mcp/` or `internal/api/` direct entity imports; that is a future amendment to avoid expanding scope by ~18 files of mcp/api work in one feature.
