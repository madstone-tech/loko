# Specification Quality Checklist: HCL Compiler Core

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-21
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

### Validation record (iteration 1 — all items pass)

- **Implementation neutrality**: the spec names no language, library, file format, or cloud
  provider. Technology-bearing terms from the design document were restated behaviourally:
  the configuration language as "declarative source files identified by a dedicated file-name
  suffix", the surgical-edit library as "the format command", the two serialisations as "two
  machine-readable encodings carrying equivalent information", and the layer-import rule as
  "the parsing or diagramming libraries". The only occurrences of a technology name are in the
  feature title and branch identifier, which are naming, not requirements.
- **Scope boundedness**: the spec covers Stage 1 of the v1.0 roadmap only. The Assumptions
  section names each deferred capability (rendering, semantic diff and rename declarations,
  infrastructure ingestion, policy, conversational authoring) and states why syntax reserved
  for them (views, exclusion patterns, physical claims) is validated here but consumed later.
- **Testability**: every functional requirement is stated as an observable behaviour. The
  error and warning lists (FR-028, FR-029) are enumerated rather than described, so SC-004 can
  assert 100% rule coverage against them.
- **Measurability**: success criteria carry explicit numbers — 100% coverage of the validation
  rule list, byte-identical repeat exports, sub-2-second check at 1,000 elements and
  sub-10-second at 5,000, a 30-minute first-architecture time, and idempotent formatting on
  100% of fixtures.
- **Deliberate non-obvious inclusions**: the deployment plane is P1 rather than deferred,
  because introducing it later would be a breaking language change for every file authors have
  written by then. Determinism is a functional requirement (FR-040) rather than a quality goal,
  because the golden-file test strategy depends on it. Retirement of the superseded model is a
  user story with acceptance scenarios (Story 6), so that removing the dual-source machinery is
  gated rather than left as an optional cleanup.
### Validation record (iteration 2 — after `/speckit-clarify`, 2026-09-22)

Re-evaluated all 16 items against the clarified spec. No state changes: 16/16 → 16/16 passing.
Five clarifications were integrated, each tightening a requirement that had been testable but
under-specified:

- **Instance addressing** was genuinely contradictory before this round — FR-024 implied no
  placement segment while the nested-groups edge case implied one. Resolved toward
  placement-independent instance addresses, which keeps re-parenting readable as a change rather
  than a remove-plus-add in the downstream diff feature. Added FR-012a (uniqueness) and two
  acceptance scenarios.
- **Format check mode** (FR-035a) and **machine-readable diagnostics** (FR-030a, FR-034a) closed a
  gap that would have forced continuous-integration jobs to scrape console text. FR-038 now states
  the three exit codes are exhaustive, so neither addition widens the exit-code contract.
- **Export schema version** (FR-036a–c) makes cross-release compatibility decidable. FR-036c is the
  load-bearing half: it forbids embedding the tool version, a timestamp, or a hostname, any of which
  would have silently broken SC-003's byte-identical guarantee.
- **Function set** (FR-017, FR-017a) replaced "a small set of string-manipulation functions" — which
  was not checkable — with exactly five named functions. Deliberately under-shipped, since adding a
  function later is painless and removing one is a breaking change.

New criteria SC-011 and SC-012 make the export-compatibility and CI-gating guarantees measurable.

### Validation record (iteration 3 — after `/speckit-analyze` remediation, 2026-09-22)

Re-evaluated all 16 items. No state changes: 16/16 → 16/16 passing. Nine analysis findings were
closed:

- **F1** (HIGH) — `syntax_error` existed in the diagnostics code enum but in no rule table and no
  task, so the rule-coverage test would have failed by construction. Added FR-028a distinguishing a
  parse failure from an undefined-construct error, a row in the data-model rule table, and a fixture
  task. This was the only finding that would have produced a red suite on day one.
- **F2** (MEDIUM) — the layer-rule guard test was numbered after the rules it was meant to fail
  against, so it would never have been observed red. Reordered.
- **G1** (MEDIUM) — views were decoded and typed but never validated, so FR-016's "validated" was
  unmet. Added resolution of view `include`/`exclude` plus a broken-reference fixture.
- **G2** (MEDIUM) — SC-011 promised verification by a pinned artefact that nothing created. Added the
  committed `export_v1.json` fixture and a version-refusal test.
- **G3** (MEDIUM) — SC-002 claimed coverage of "every reference position" without enumerating them.
  The resolution test now names all six positions.
- **C1** (LOW) — plan structure block now lists `json.go` and `diagnostics.go`.
- **A1** (LOW) — SC-005 was a 30-minute stopwatch claim with no verification method. Reframed as a
  documentation review with a concrete pass condition, paired with a docs task.
- **G4** (LOW) — FR-027's "no state file" had no asserting test; added one, so a future cache cannot
  be introduced unnoticed.
- **G5** (LOW) — added fixtures for the empty-architecture and unreadable-file edge cases.

Two findings were deliberately left unchanged: the FR-004/FR-012a overlap (FR-012a earns its own
diagnostic code because the clarified addressing rule makes it non-obvious) and the spec-versus-
contract terminology split (the spec is technology-agnostic by design; `data-model.md` bridges the
two vocabularies).

- **Deliberate exclusion**: the rename-declaration block is not specified here. It belongs to
  the diff feature, and its operands have evaluation semantics that cannot be meaningfully
  exercised without a diff engine to consume them.
