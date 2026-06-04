# ADR 0010 — Tighten outer-layer entity-import rule (Constitution v1.1.0 → v1.2.0)

**Status**: Accepted
**Date**: 2026-05-21
**Constitution version**: 1.2.0
**Feature**: 010-constitution-compliance

## Context

Under constitution v1.1.0 only the `cmd/` outer entry-point was forbidden from importing `internal/core/entities/` directly. The two remaining outer entry-points — `internal/mcp/` and `internal/api/` — were explicitly allowed to reach into the entity package. The v1.1.0 SYNC IMPACT REPORT (Sync Impact Report header, Modified principles → Architecture Rules → cmd row) called this gap out and tagged it as "future-feature concern".

This left a real asymmetry: a refactor that pushed business logic out of `cmd/` (feature 009) still permitted the same coupling in the other two outer interfaces. MCP tool handlers and HTTP API handlers were importing entities directly to convert results into protocol-specific shapes (`systemToMap(*entities.System)`, `relationshipToMap(*entities.Relationship)`), use entity helpers (`entities.NormalizeName`), and pattern-match entity types in their internal dispatch logic. Over time, that grows: each new MCP tool tends to repeat the pattern, and the value of "three interfaces, one core" promised in Principle I is eroded.

## Decision

Tighten the Dependency Direction table so the `internal/mcp/` and `internal/api/` rows match the `cmd/` row:

| Layer | May Import (v1.2.0) | Must Not Import (v1.2.0) |
|-------|---------------------|---------------------------|
| `internal/mcp/` | core/usecases, adapters | `internal/core/entities/` directly; api; cmd |
| `internal/api/` | core/usecases, adapters | `internal/core/entities/` directly; mcp; cmd |

Outer entry-points obtain entity-shaped data via use-case return values or adapter outputs. The use-case ports are the contract; the entity package is private to core.

Bump constitution version 1.1.0 → 1.2.0 (MINOR — tightens a principle's surface area without redefining it incompatibly).

## Consequences

### Positive

- Closes the v1.1.0 carve-out noted in the prior SYNC IMPACT REPORT.
- Restores symmetry between the three outer interfaces — none of them sees entity types directly.
- Forces the conversation, on every new MCP tool or API handler, about which use-case interface the new feature consumes. If no use case fits, the contributor writes one rather than reaching into entities.

### Negative (and how addressed)

- **16 existing MCP tool files (plus 4 of their tests) and one API test file violate the new rule.** The actual entity-decoupling refactor of these files is non-trivial: each tool's response-shape converter has to either return a DTO from the use case or call a converter that lives in the adapter layer. To keep this feature's scope tight, the refactor is **deferred** to a follow-up feature (tentatively `011-mcp-entity-decoupling`); the violations are recorded in `.archcheck-suppressions.yaml` with `expires_on: 2026-08-19` (90 days). Suppressions self-clear: expired entries re-fire as failures, so the deadline is mechanically enforced.
- **The structural-compliance check would otherwise turn the main branch red on day one.** This is exactly the case the suppression mechanism (FR-017 in the 010 spec) was designed to handle. The mechanism's 90-day hard cap, owner tagging, and reason requirement keep the suppression list honest.

### Neutral

- The redundant `depguard` fast-path in `.golangci.yml` is tightened in lockstep so local `task lint` reports the new violations the same way CI does.

## Alternatives considered

1. **Soften feature 010's spec FR-008 to match v1.1.0** — would have been a third can-kick after 009 also deferred this. Rejected because the user's feature description explicitly named all three outer entry-points.
2. **Do the full entity-decoupling refactor inside feature 010** — multi-hour effort with non-trivial response-shape regression risk. Rejected because the suppression mechanism exists precisely to defer this kind of scope without surrendering the gate. Net session-budget gain: ~3 hours.
3. **Add a layer-import exemption for `*_test.go` files** — would silence the 4 mcp + 1 api test violations without listing them. Rejected because tests SHOULD be reviewed when the production they exercise changes; a categorical layer-import exemption removes that signal.

## Migration plan

1. **This feature (010)**: amendment + rules YAML update + 90-day suppression entries land in the same PR.
2. **Feature 011 (to be opened)**: actual entity-decoupling refactor of the suppressed files. Targets removal of every suppression before `expires_on`.
3. **Mechanical guard**: if 011 slips, the suppressions expire and CI starts blocking unrelated PRs — that pressure is the intended forcing function.

## References

- Constitution v1.2.0: `.specify/memory/constitution.md`
- Structural rules (machine-readable): `tools/archcheck/rules.yaml`
- Spec FR-008: `specs/010-constitution-compliance/spec.md`
- Suppression mechanism design: `specs/010-constitution-compliance/research.md` R4 + R7
- Active suppressions: `.archcheck-suppressions.yaml`
