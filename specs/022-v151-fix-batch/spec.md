# Feature Specification: v1.5.1 Patch Bug-Fix Batch

**Feature Branch**: `022-v151-fix-batch`

**Created**: 2026-11-09

**Status**: Draft

**Input**: User description: "v1.5.1 bug fix batch from the third-party feedback report (loko-feedback.md): manifest location, missing help link, misleading dynamic error, trigger-kind note, and a documented CSS contract for themers"

## Clarifications

### Session 2026-11-09

- Q: What should the hidden state directory next to the output directory be named, given that the name becomes permanent user-visible surface once 1.5.1 ships? → A: `.dist.loko` (generalizes as `.<base>.loko` for any output directory name, e.g. `.release.loko` for `--out release`).
- Q: Under what conditions should a 1.5.1 build delete the legacy 1.5.0 manifest from inside the output directory? → A: Only on success — the legacy file is removed only when every earlier step of the build (write, prune, state-manifest write) succeeded.

## User Scenarios & Testing *(mandatory)*

This feature is a patch release: it corrects defects and misleading text in existing
behavior. It adds no new capability, changes no file format a consumer parses, and
introduces no new language surface. Each user story is an independently verifiable fix.

### User Story 1 - Deployment output carries no tool state (Priority: P1)

A user builds a project and publishes the output directory to a static host, a
container image, or a web server. The published directory must contain only the
generated site. In 1.5.0 the build records its ownership manifest
(`.loko-manifest`) inside the output directory, so every published copy carries
tool state that is meaningless to consumers and reveals internal paths.

**Why this priority**: The feedback report names this as a deployment defect
(issue 15). A copied manifest is harmless to rendering but is the only one of the
five defects with a production impact: it ships to every deployment.

**Independent Test**: Run a build; inspect the output directory for hidden state
files. Then rebuild after editing one source file; confirm the build still removes
files that the previous build generated but the new build no longer generates.

**Acceptance Scenarios**:

1. **Given** a project with no previous build, **When** the user runs a build, **Then**
   the output directory contains only generated artifacts and no manifest.
2. **Given** a project built with 1.5.0 (legacy manifest inside the output directory),
   **When** the user builds with 1.5.1, **Then** the 1.5.1 build removes the legacy
   manifest and continues to prune from it, so a checkout built under 1.5.0 does not
   lose pruning on the first new build.
3. **Given** a project built with 1.5.1, **When** a file that an earlier build
   generated disappears from the artifact set, **Then** the next build removes it.
4. **Given** an output directory whose state manifest in `.dist.loko` is missing or
   unreadable, **When** the user runs a build, **Then** the build succeeds and
   removes nothing owned by an unknown prior build (safe-degradation behavior is
   unchanged from 1.5.0).

---

### User Story 2 - A user can find the language reference from the CLI (Priority: P2)

A new user runs `loko --help` to learn the tool. In 1.5.0 the help text never
mentions where the language, the CLI reference, or the theming guides live, so the
user must find the documentation site or repository by search.

**Why this priority**: The help text is the first surface every new user sees
(issue 16). The fix is one line of text with no behavior change.

**Independent Test**: Run `loko --help` and check that it names where the
documentation lives (the `docs/` directory of the repository, with the language
reference called out).

**Acceptance Scenarios**:

1. **Given** any loko installation, **When** the user runs `loko --help`, **Then**
   the output includes a pointer to the documentation, and names the language
   reference specifically.
2. **Given** the same help output, **When** the user reads any command's help
   (`loko build --help`, `loko serve --help`), **Then** no duplicated doc pointer is
   added there (the root help carries it once).

---

### User Story 3 - The unsupported-block error states what to use instead (Priority: P2)

An author uses a `dynamic` block (a pattern from other HCL tools) in a `*.loko.hcl`
file. In 1.5.0 the error says the block type is not supported and explains v1.0
history, but does not say how to get the same result in loko's language. The author
stalls or searches.

**Why this priority**: The message text is verbatim from the feedback report
(issue 17). Error text quality directly affects authoring sessions driven by an
assistant, which retry or stall on unhelpful errors.

**Independent Test**: Put a `dynamic` block in a source file, run `loko validate`,
and check that the error names the supported way to author a generated or repeated
structure (a `view` block with `step` blocks).

**Acceptance Scenarios**:

1. **Given** a source file with a `dynamic` block, **When** the user runs
   `loko validate`, **Then** the error keeps its existing reason text and adds one
   sentence that names the supported construct (`view` with `step` blocks).
2. **Given** a source file with no unsupported blocks, **When** the user runs
   `loko validate`, **Then** output is unchanged.

---

### User Story 4 - The relationship `kind` note states the trigger direction (Priority: P3)

An author reads the language schema (MCP `schema` output or the language reference)
to learn what `kind = "trigger"` does to a relationship. In 1.5.0 the note says only
"sync by default", so the author cannot learn that a trigger reverses the drawn edge:
the source of the `uses` declaration becomes the receiver in the rendered diagram.

**Why this priority**: The schema note is a documentation string inside an existing
surface (issue 2, docs half). Wrong or incomplete meaning shown by the `schema` tool
leads an assistant to model event flows backwards.

**Independent Test**: Run `loko schema` (the MCP schema tool) and check that the
`kind` attribute note states the default direction and the trigger reversal.

**Acceptance Scenarios**:

1. **Given** the `schema` tool output, **When** the reader selects the relationship
   `kind` attribute, **Then** the note states that relationships draw source to
   target by default and that `trigger` reverses the drawn direction.
2. **Given** a relationship with `kind = "trigger"`, **When** the project is built,
   **Then** the rendered edge direction is unchanged from 1.5.0 (this fix is
   documentation only; rendering behavior is not modified).

---

### User Story 5 - A themer can rely on documented class names (Priority: P3)

A user themes the generated site by overriding its CSS classes. In 1.5.0 the CSS
class names that templates emit are not documented anywhere, so a theme that styles
`.brand-link` or `.pager` breaks silently when an update renames a class. Template
blocks are already protected: overriding a block that no template defines fails the
build with `theme_invalid`. The CSS layer has no equivalent contract.

**Why this priority**: This fix is documentation (issue 14, CSS half). It completes
the theming contract so the error message and the guide cover the same surface.

**Independent Test**: Read the theming guide and the site contract; both state the
stable class names, CSS custom properties, and the stability promise.

**Acceptance Scenarios**:

1. **Given** the theming guide, **When** a themer reads it, **Then** it lists the
   CSS classes and custom properties the site templates emit, and states the
   stability commitment (renames called out in release notes, not silent).
2. **Given** a theme that overrides a template block that the built-in theme does
   not define, **When** the user runs a build, **Then** the build fails with
   `theme_invalid` as in 1.5.0 (existing behavior, unchanged; documented as the
   blocks-half of the contract).

---

### Edge Cases

- What happens when the state manifest of a 1.5.1 build exists but the legacy
  1.5.0 manifest also exists? The legacy manifest is removed only after the build fully
  succeeds (writes, pruning, and the state-directory manifest in place); a failed
  build leaves both files, and the next retry prunes from the newer state
  manifest.
- What happens when two projects share a parent directory and both use the default
  output name? Each project's state directory `.<base>.loko` lives next to its own
  output directory, so two sibling projects each hold their own `.dist.loko`; they
  never collide because the state directory is resolved relative to each output
  directory's own parent, not a global name.
- What happens when the output directory is relocated or renamed (parent directory
  moved)? The state directory travels with it (it sits next to the output
  directory), so pruning continues to work after a rename of the containing
  directory.
- What happens when an MCP client asks to build into the project root itself? The
  state directory would land in the root's parent, outside the project. The MCP
  build refuses an output directory that resolves to the project root, so a build
  never writes outside the project (output confinement from feature 017).
- What happens when a `dynamic` block appears with children or with several
  occurrences in one file? The error text is identical for each occurrence; the fix
  adds text only, it does not change error structure or codes.
- What happens when a theme references a class the docs list but an old built-in
  theme lacked? The class list documents the current (1.5.x) site; the guide states
  that classes added after 1.5.0 are additive.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The build MUST keep the ownership manifest outside the output
  directory, in a hidden state directory named `.<base>.loko` (for the default
  output directory `dist`: `.dist.loko`; for `--out release`: `.release.loko`),
  placed next to the output directory. An MCP build MUST refuse an output
  directory that resolves to the project root, because its state directory would
  fall outside the project.
- **FR-002**: The build MUST treat a manifest written by 1.5.0 or earlier (inside
  the output directory) as a valid manifest for pruning, and MUST remove that legacy
  file only after every earlier step of the build has succeeded — artifacts written,
  orphans pruned, and the state-directory manifest in place. A build that fails at
  any earlier step MUST leave the legacy file in place, so a retry still prunes
  from the newest valid state.
- **FR-003**: The build MUST prune previously owned files exactly as 1.5.0 did, and
  MUST prune nothing when the manifest is missing, unreadable, or corrupt.
- **FR-004**: The root CLI help MUST point to the repository documentation and name
  the language reference.
- **FR-005**: The not-supported-block error for `dynamic` MUST keep its current
  reason text and MUST add one sentence that names `view` with `step` blocks as the
  supported authoring path.
- **FR-006**: The schema note for the relationship `kind` attribute MUST state the
  default direction (source to target) and that `trigger` reverses the drawn
  direction.
- **FR-007**: The theming guide MUST document the stable CSS class names and custom
  properties of the generated site, and MUST state the stability commitment for
  them.
- **FR-008**: The release MUST change no rendered output other than through the
  fixes above (diagram bytes, site rendering, markdown, and export output are
  unchanged except where a fix inherently touches text output).

### Key Entities *(include if feature involves data)*

- **Ownership manifest**: A sorted list of paths the last build wrote, written next
  to the output directory since 1.5.1. Drives pruning. Previously stored inside the
  output directory.
- **State directory**: The hidden directory `.<base>.loko` (for the default output
  directory `dist`: `.dist.loko`), placed in the output directory's parent. Holds
  the manifest. Not published when the output directory itself is copied or
  deployed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A build of any project produces an output directory whose file list
  contains zero hidden dot-files.
- **SC-002**: Rebuilding a project after removing one artifact source removes that
  artifact (pruning works through the new manifest location), verified on a checkout
  first built under 1.5.0.
- **SC-003**: `loko --help` output contains a documentation pointer; a reader can
  find the language reference path from that output in one step.
- **SC-004**: `loko validate` on a file with a `dynamic` block returns an error whose
  text includes the supported-construct sentence.
- **SC-005**: The `kind` attribute note in the schema output states both the default
  direction and the trigger reversal.
- **SC-006**: The theming guide names the stable class contract, and the release
  notes for 1.5.1 list every fix with its motivating feedback item.

## Assumptions

- Patch discipline: no new user-facing capability, no language surface, no new
  diagnostic codes. The manifest relocation changes a file location, which is a
  defect fix with an explicit migration path (FR-002), not an interface change.
- Consumers that published `dist/` under 1.5.0 shipped the manifest unknowingly;
  after the upgrade the copied legacy manifest is removed from the source tree on
  the next local build. Already-deployed copies are not retroactively fixed.
- The state directory is hidden (dot-prefixed) and holds only the manifest in this
  release; later releases may add build caches there without further migration.
- The class list in the theming guide documents the current site; a full CSS
  naming audit (new classes, renaming for stability) is out of scope for a patch.
- The rendered edge direction for `trigger` relationships is confirmed correct for
  1.5.x and is not changed here; only the note text is fixed.
- Terminology: this spec says "manifest" for the ownership manifest and "state
  directory" for `.<base>.loko`; it avoids "state file" and "manifest file" as
  synonyms.
- Release process: `release/1.5.1` branch and tag per the project's release
  convention; CHANGELOG gains a `### Fixed` section in the established style.
