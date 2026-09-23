# Specification Quality Checklist: ViewModel Renderers

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-23
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Items marked incomplete require spec updates before `/speckit-clarify` or `/speckit-plan`.

### Validation record (iteration 1 — 16/16 pass)

- **Implementation neutrality**: the spec names no diagram syntax, template engine, markup format,
  file-watching library, or rendering library. The roadmap seed named several; each was restated
  behaviourally — the diagram library as "MUST NOT start an external process or require any
  separately installed executable", the deferred second diagram syntax as "a format whose
  specification is still experimental or whose rendering is unreliable in common viewers", and the
  theme mechanism as "replacement presentation files from a known directory". A scan for the seed's
  technology terms returns only a cross-reference path.
- **Testability**: every requirement is an observable behaviour. The determinism requirements
  (FR-021, FR-022) and the generated-file notice (FR-020) are stated so SC-002 through SC-004 can be
  measured by comparing bytes and scanning files, not by inspection.
- **Scope boundedness**: Assumptions name each deferred capability and, importantly, state that the
  authoring language does not change — if a view selection needs expressive power the language lacks,
  that is a different feature.

### Three premises in the roadmap seed that are no longer true

Checked against the merged codebase rather than taken from the seed. Each is reflected in the spec's
Assumptions rather than silently carried forward:

1. **"the `d2` library that is already a direct dependency"** — it is not. `go mod tidy` dropped it in
   the previous stage when the last code importing it was deleted. It has to be reintroduced, which
   is a dependency decision for the plan rather than a no-op.
2. **"split `templates.go` (1,753 lines)"** — 1,753 is the raw line count; it is 1,556 *effective*
   lines. The constitution amended in the previous stage caps adapter files at 400 effective lines, so
   this is now a hard gate rather than a suggestion, and the file must become at least four.
3. **"rewire the existing HTML site builder"** — the builder is parked under `internal/_parked/` and
   does not compile. It was written against the deleted model, so this is a rework with a reference
   implementation, not a rewiring.

### Deliberate choices worth review

- **Story 3 (byte-stable generated output) is P1, not polish.** Determinism is what makes committed
  diagrams reviewable; without it teams stop committing them and the drift problem returns through the
  back door. It ships with the first story rather than after.
- **A malformed theme override fails the build** (FR-035) rather than falling back to the built-in
  version. A theme that half-applies is harder to diagnose than one that refuses.
- **A declared view shadowing a derived one wins, with a warning** (FR-004). Silently dropping either
  would leave the author wondering where a picture went.
- **Stale generated files are removed** (FR-023) while files the tool did not generate are left alone.
  Without the first half, a deleted element leaves an orphan page that outlives it; without the
  second, the tool eats a user's unrelated files.
- **No output backend is reachable from the conversational interface** (FR-041). A tool that can
  author a rendered artifact recreates the two-sources-of-truth defect this release exists to remove.
