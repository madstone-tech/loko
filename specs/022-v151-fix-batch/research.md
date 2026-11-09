# Phase 0 Research: v1.5.1 Patch Bug-Fix Batch

**Branch**: `022-v151-fix-batch` | **Date**: 2026-11-09

All NEEDS CLARIFICATION items from the plan workflow are resolved. Every decision
below cites the runtime or source evidence gathered this session on loko 1.5.0
(`af84abc`).

## R1: Manifest location and state-directory name

**Decision**: Write the manifest to `<parent(outDir))/.<base>.loko/manifest`
(default: `.dist.loko/manifest`). Keep the file name `manifest`; the directory
carries the identity. Read order: state manifest first, then the legacy in-output
`.loko-manifest`. Remove the legacy file only after artifacts are written, pruning
has run, and the new manifest is in place (`store.go` tail of `Commit`).

**Rationale**:
- `Commit` (internal/adapters/outputdir/store.go:30) receives only `outDir`, so the
  state location must derive from it. Deriving from the parent directory keeps the
  state beside its output: renaming or moving the project tree carries both together.
- Verified consumer-facing defect: a build of
  `examples/simple-project` writes `dist/.loko-manifest`
  (manifest constant at internal/core/entities/viewmodel/commit_plan.go:11).
- Sibling projects each hold their own `.dist.loko`; no global keying needed
  (spec edge case, clarified session 2026-11-09).
- The legacy fallback is mandatory (spec FR-002): an existing checkout built under
  1.5.0 must keep pruning on the first 1.5.1 build.

**Alternatives considered**:
- *`.loko-manifests` directory next to `outDir`*: one directory for several outputs,
  but separates the manifest from its output dir and collides with a project dir
  named `loko`.
- *XDG cache keyed by escaped output path*: spreads state across the user's cache
  tree; overkill for a patch and breaks "state travels with the project".
- *Keep manifest inside `dist/` but rename*: does not fix the deployment defect.
- *User-deleted legacy file (build prints a note)*: leaves the defect in place
  until the user acts; rejected by clarify session (success-only auto-retire chosen).

## R2: Where the help pointer goes

**Decision**: One paragraph in `rootCmd.Long` (cmd/root.go:28-40), naming
`docs/language.md` as the language reference and the repository `docs/` directory.
No per-subcommand duplication (spec acceptance 2.2).

**Rationale**: `loko --help` output was verified to contain no documentation
pointer. `rootCmd.Long` is the root help body; `Short` stays as-is. The pointer
words itself like the README table (README.md:205), which already calls
`docs/language.md` "The complete `*.loko.hcl` language".

**Alternatives considered**:
- *Add the pointer to every subcommand's `Long`*: duplicates one line eight times;
  rejected by spec.
- *Add a `docs` subcommand*: new feature in a patch; violates FR-008.

## R3: The `dynamic` hint text and surface

**Decision**: Extend the `deliberatelyExcluded` map entry for `"dynamic"`
(internal/adapters/hclsource/schema.go:34) from
`dynamic block generation is deliberately excluded from v1.0` to
`dynamic block generation is deliberately excluded from v1.0; to repeat a
structure, use a view block with step blocks`.
`schema.go:109-111` interpolates the reason into
`%q is not supported: %s.`, so the map entry is the single edit; the diagnostic
code (`CodeUnknownBlock`) and location are unchanged.

**Rationale**: The map is the single source of the detail text for excluded
blocks; verified the error renders verbatim through the map (`schema.go:110`).
Runtime probe this session reproduced the current text end-to-end.

**Alternatives considered**:
- *Change `functions.go:47` (unknown-function detail)*: that is a different error
  surface (functions, not blocks); editing it is out of scope for this issue.
- *Emit a second diagnostic with the hint*: doubles output for one fix; the single
  detail sentence teaches in place, matching the map's design comment
  ("the error teach[es] rather than merely refuse[s]").

## R4: The `kind` note text

**Decision**: Replace the note for the relationship `kind` attribute
(internal/core/entities/arch/schemas.go:133) from `"sync by default"` to text that
states both facts: sync draws source → target by default, and `kind = "trigger"`
reverses the drawn direction (the declared source becomes the rendered receiver).

**Rationale**: Verified anchor; `RelationshipKinds = [sync, async, trigger]`
(internal/core/entities/arch/render_attrs.go:29). The reversal behavior itself is
confirmed correct for 1.5.x (runtime probe this session: a `uses` declaration with
`kind = "trigger"` renders with swapped endpoints) and is NOT changed here
(spec assumption) — only the note.

**Alternatives considered**:
- *Also document `async`*: the note is one string for all kinds; mentioning async's
  dashed rendering is 1.6.0 scope with the rendering-fidelity work, per the release
  split decision.

## R5: The stable CSS contract

**Decision**: Add a "Stable classes, variables, and blocks" section to
`docs/guides/site-theming.md` listing the classes and custom properties the
built-in templates emit, with the stability commitment stated once. Also update
`site-theming.md`'s "Errors" section intro: the blocks-half of the contract
(`theme_invalid` on unknown `{{define}}`) is already implemented; the docs gain one
sentence pointing readers to the class list as the CSS-half contract.

**Rationale**: Verified in the guide: it documents overridable files and CSS
variables (`--bg`, `--card`, `--txt-1..3`, `--link`, `--accent`, `--kind-*`), but
the emitted class names (`.landscape`, `.systems`, `.views`, `.rows`, `.row`),
the sidebar/nav/pager classes, and the `.brand`/`.brand-link` markup the templates
render are documented nowhere. A custom brand theme built to test the feedback
styled exactly `.brand`, `.accent-bar`, `.pager`, `.theme-toggle` — four classes a
real theme needs — none documented.

**Alternatives considered**:
- *Emit a diagnostic when a theme styles an unknown class*: needs CSS parsing in
  the build; out of scope for a patch and arguably never (docs are the contract).
- *Generate the class list from templates at build time*: runtime cost and a new
  artifact for docs-only value; rejected (YAGNI).

## R6: Contract amendment surface

**Decision**: Amend
`specs/014-viewmodel-renderers/contracts/output-layout.md` (normative): the tree
drops `.loko-manifest` from `dist/` and gains the state directory; the ownership
section and the notice table row for `.loko-manifest` are updated to the new
location.

**Rationale**: The contract is the normative source for output layout; leaving it
stale would contradict the code. `docs/quickstart.md:136` and
`docs/adr/0013-viewmodel-renderers.md:41` mention `.loko-manifest` in prose about
ownership; quickstart's tree is updated too. ADR-0013 is a historical record of a
past decision; it stays untouched (ADRs are immutable once accepted).

**Alternatives considered**:
- *New ADR*: rejected in the plan's constitution check; data placement specified
  in the feature spec, not a new architectural pattern.
- *Edit ADR-0013 in place*: breaks the immutable-ADR convention.

## R7: Test and golden impact

**Decision**: Tests and goldens affected, verified by locating them:

| Change | Tests affected | Update needed |
|---|---|---|
| Fix 1 (manifest) | `internal/adapters/outputdir/store_test.go:90,160` reads/creates the legacy path; `TestManifestWrittenLast` asserts location | Rewrite to the new location + add migration cases |
| Fix 1 (manifest) | `cmd/root_test.go` asserts `built ... N removed` counts (pruning) | Possibly counts change when a legacy manifest file is retired; verify at run time |
| Fix 2 (help) | Golden help output in `cmd/testdata/` if any | Regenerate with `-update` if the Long text is golden-captured |
| Fix 3 (`dynamic`) | `internal/adapters/hclsource/decode_test.go:268` asserts current detail substring | Update the expected substring |
| Fix 4 (`kind`) | MCP schema goldens if the note is captured | Regenerate with `-update` |
| Fix 5 (docs) | None (docs only) | — |

Golden regeneration command (CONTRIBUTING.md:91):
`go test ./cmd/ ./internal/adapters/{d2,hclsource,html,markdown}/ -update`.

**Rationale**: Verified each of these anchors exists (grep this session). No other
packages import `vm.ManifestFile` outside `outputdir` (grep: `internal/`, `cmd/`).

**Alternatives considered**: none — the mapping is exhaustive per grep.

## R8: Release mechanics

**Decision**: After the batch merges to `main`, cut `release/1.5.1` from it, tag
`v1.5.1`, push tag first (memory convention); CHANGELOG gains `## [1.5.1]` with a
single `### Fixed` section listing the five fixes with their feedback issue
numbers, in the established house style (CHANGELOG.md:213-243 pattern: motivating
sentence first, bold one-line summaries).

**Rationale**: Confirmed precedent: 1.0.1 and 1.0.2 are Fixed-only releases;
tags v1.0.0 through v1.5.0 exist; the split decision (handoff) puts only these
five items in 1.5.1; the six v1.6.0 items stay out.

**Alternatives considered**:
- *Bundle the 1.6.0 theming work*: rejected; consumers pin 1.5.x and a patch that
  adds surface breaks that contract (docs/language.md:487-493).
