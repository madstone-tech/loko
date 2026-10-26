# Feature Specification: MCP & Validation Ergonomics

**Feature Branch**: `017-mcp-validation-ergonomics`

**Created**: 2026-10-25

**Status**: Draft

**Input**: A first modelling session through `loko mcp` (loko 1.0.1) by an AI coding assistant,
reported in [first-run-report.md](first-run-report.md). The model compiled, but getting there cost
far more turns than it should have:

- the assistant learned the language by trial and error: about eight refused `apply_edit` previews
  before it knew which attributes each block takes (report #1). It also never found the existing
  `external` element kind, so it modelled third-party systems as `system` and drew warnings for
  them;
- a 12-container model produced 33 warnings, most of them expected (instances whose infrastructure
  does not exist yet, external systems), and the MCP output repeated a full source range for each
  one (report #2);
- it could edit and validate the model through MCP, but could not render or format it (report #8);
- in an empty directory, `describe` dead-ended, and `apply_edit` could not create the first file
  (report #9).

Report #3 (subdirectory files refused under the default root) and the external part of #2 were fixed
in 1.0.2. Report items #4, #6, #7 and #10 are deferred to feature 018, and #5 (dynamic views) to 019.

## Clarifications

### Session 2026-10-25

- Q: Should compact MCP diagnostics (file and line only) be the default, or opt-in? → A: The default.
  A client asks for full ranges when it needs them. This changes the MCP response shape within 1.x,
  which the changelog records; the only MCP clients so far are assistants, and they act on file and
  line.
- Q: Is a project-level setting to switch off warning codes in scope? → A: Yes: a list of warning
  codes on the `project` block, with the same rules as the per-run ignore option. The setting
  travels with the project, so CLI, CI and MCP runs agree without repeating flags.
- Q: How does a client create the project declaration through MCP? → A: As a new `project` target
  of `apply_edit`, so there is one write path with one set of preview and revision checks. In an
  empty directory, `describe` returns a ready-to-send `apply_edit` request that creates it.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Learn the language from the server (Priority: P1)

An assistant connects to `loko mcp` for the first time. Before writing anything, it asks the server
what the language contains: which blocks exist and how they nest, which attributes each one takes
and which values they allow, how `apply_edit` addresses are written, and how `docs` paths resolve.
It then writes a correct model without a single refused edit caused by a wrong guess about the
language.

**Why this priority**: every other friction in the report starts here. An assistant that knows the
language uses `external` for third-party systems, uses the right attributes, and skips the
discovery loop that cost the most turns.

**Independent Test**: start the MCP server, call the new read-only tool, and compare its answer
with the language reference: every block, attribute and allowed value the compiler accepts is
listed, and nothing the compiler rejects is.

**Acceptance Scenarios**:

1. **Given** a running MCP server, **When** a client lists tools, **Then** a read-only tool
   describing the language is listed, and the server's handshake instructions tell the client to
   call it before its first edit.
2. **Given** that tool, **When** a client calls it, **Then** the answer lists every block type with
   where it may appear, and for each block every attribute with its value type, whether it is
   required, and its allowed values where they are a fixed set (for example `layout`: `dagre` or
   `elk`).
3. **Given** that tool, **When** a client reads it, **Then** it states the address grammar
   `apply_edit` accepts (for example `container.api`, `person.user.uses.checkout`,
   `deployment.prod.node.web.instance.api`, `view.overview`), the rule that `docs` paths are
   relative to the project root, and one minimal valid example per block.
4. **Given** a client asks about one block only, **When** it calls the tool with that block's name,
   **Then** only that block's part of the answer is returned.
5. **Given** an attribute is accepted by the compiler, **When** it is missing from the tool's answer
   (or the reverse), **Then** an automated check in the project's own test suite fails.

---

### User Story 2 - Warnings that point at real problems (Priority: P1)

An architect validates a model that includes instances whose infrastructure has not been built yet.
They mark those instances as planned, and the warnings about missing bindings stop. When they want
an overview, they ask for a summary instead of a list. In CI, they ignore codes the team has decided
not to act on, and the remaining warnings still fail a strict run.

**Why this priority**: in the reported session most of the 33 warnings were expected. Warnings that
always fire teach people to ignore all of them, including the ones that matter.

**Independent Test**: validate a fixture with planned instances, then run validate with the summary,
severity and ignore options, from the CLI and through MCP, and compare the reported diagnostics and
exit codes.

**Acceptance Scenarios**:

1. **Given** an instance marked as planned and without a binding, **When** the model is validated,
   **Then** no `unbound_instance` warning is reported for it.
2. **Given** an instance marked as planned that does have a binding, **When** the model is
   validated, **Then** a warning says the planned marker is out of date.
3. **Given** a model with warnings, **When** validate is asked for a summary, **Then** the output
   groups diagnostics by code, with a count and the affected addresses under each code, errors
   first.
4. **Given** a model with warnings and errors, **When** validate is asked for errors only, **Then**
   only errors are shown, and the exit code is the same as without the filter.
5. **Given** a model whose only warnings are of an ignored code, **When** validate runs in strict
   mode with that code ignored, **Then** it exits as if there were no warnings.
6. **Given** a client asks to ignore an error code, **When** validate runs, **Then** the request is
   refused: errors cannot be ignored.
7. **Given** an MCP client calls validate without asking for full ranges, **When** the response is
   returned, **Then** each diagnostic carries its file and line only, and the response for the
   reported 33-warning model is at most half the size it was in 1.0.2.

---

### User Story 3 - Render and format without a shell (Priority: P2)

An assistant that has no shell access edits a model through MCP, then renders it and checks a
diagram, and formats the source, all through MCP tools.

**Why this priority**: today such a client can change the model but never see the result. Clients
with a shell can already run `loko build` and `loko fmt`, so this matters less than Stories 1 and 2.

**Independent Test**: through MCP alone, build a fixture project and format a deliberately
misformatted file; compare the written output and formatted source with what the CLI produces.

**Acceptance Scenarios**:

1. **Given** a valid model, **When** a client calls the build tool with formats and an output
   directory, **Then** the same files `loko build` would write are written, and the response lists
   their paths.
2. **Given** a client asks for one view inline, **When** the build tool returns, **Then** the
   response includes that view's diagram source or image, so the client can inspect it without
   reading files.
3. **Given** a model with errors, **When** the build tool is called, **Then** nothing is written
   and the diagnostics are returned, as `loko build` does.
4. **Given** an output directory outside the project, **When** the build tool is called, **Then**
   the call is refused and nothing is written.
5. **Given** a misformatted source file, **When** the format tool is called in preview mode,
   **Then** it returns the change as a diff and writes nothing; **When** it is called to apply,
   **Then** the file is rewritten exactly as `loko fmt` would rewrite it.

---

### User Story 4 - Start a project from nothing (Priority: P2)

An assistant is pointed at an empty directory. `describe` tells it how to start, with the minimal
valid project declaration. The assistant creates the project through MCP, with the same preview and
confirmation flow as any other edit, and carries on modelling.

**Why this priority**: it happens once per project, but today it dead-ends and pushes the client to
write files by hand, outside the preview and revision safety of `apply_edit`.

**Independent Test**: in an empty directory, call `describe`, then create the project through MCP,
then add a person and a system; validate reports no errors.

**Acceptance Scenarios**:

1. **Given** a directory with no source files, **When** a client calls `describe`, **Then** the
   response says the project is empty, shows the minimal valid declaration
   (`project "name" { description = "..." }`), and includes a ready-to-send `apply_edit` request
   that creates it.
2. **Given** an empty project, **When** a client sends `apply_edit` with the `project` target, a
   name, a description and a file name, **Then** a preview is shown first, and on confirmation the
   file is written with exactly one project declaration.
3. **Given** a project already exists, **When** a client tries to create another, **Then** the
   request is refused and names the file that declares the existing one.
4. **Given** an empty project, **When** a client tries to add an element before the project exists,
   **Then** the refusal says to create the project first and shows how.

---

### Edge Cases

- The language tool is called with the name of a block that does not exist: the refusal lists the
  block names that do.
- Validate is given an ignore code that is not a warning loko emits: it is refused and lists the
  warning codes, so a typo cannot silently ignore nothing.
- A summary is requested for a model with no diagnostics: it says so in one line.
- A project-level warning setting names a code that no longer exists (after an upgrade): validate
  warns about the unknown code instead of failing.
- The build tool is asked for one view inline that does not exist: refused, listing the view names.
- The format tool runs on a file with syntax errors: nothing is written and the syntax error is
  returned.
- Two MCP clients edit the same empty directory: the second project creation is refused as stale,
  as any other concurrent edit is.

## Requirements *(mandatory)*

### Functional Requirements

**Language description (Story 1)**

- **FR-001**: The MCP server MUST offer a read-only tool that describes the language: every block
  type and where it may be nested, and for each block every attribute with its value type, whether
  it is required, and its allowed values when they form a fixed set.
- **FR-002**: The tool's answer MUST include the `apply_edit` address grammar with one example per
  addressable target, the rule for resolving `docs` paths, and one minimal valid example per block.
- **FR-003**: The tool MUST accept an optional block name and then return only that block.
- **FR-004**: The answer MUST be derived from the same definition the compiler and the `apply_edit`
  refusal messages use, so the three cannot disagree. A test MUST fail when they do.
- **FR-005**: The server's handshake MUST carry instructions that point clients to the tool before
  their first edit.
- **FR-006**: The answer MUST be byte-stable for a given loko version, like every other loko output.

**Warning ergonomics (Story 2)**

- **FR-007**: An instance MUST accept an optional marker saying it is planned. A planned instance
  without a binding MUST NOT raise `unbound_instance`. A planned instance with a binding MUST raise
  a new warning saying the marker is out of date.
- **FR-008**: `apply_edit` and the language tool MUST support the planned marker like any other
  instance attribute.
- **FR-009**: Validate MUST offer, on the CLI and through MCP, a summary mode that groups diagnostics
  by code, with a count and the affected addresses for each code, errors first.
- **FR-010**: Validate MUST offer a severity filter that limits which diagnostics are shown. It MUST
  NOT change the exit code.
- **FR-011**: Validate MUST offer an option to ignore named warning codes. Ignored warnings MUST be
  left out of the output and MUST NOT count towards strict mode. Ignoring an error code, or a code
  loko does not emit, MUST be refused with the list of warning codes.
- **FR-012**: The ignore option MUST also apply to `loko build`, so strict builds and strict
  validation agree.
- **FR-013**: MCP diagnostics MUST, by default, give the file and line of each diagnostic and no
  other position fields. A client MUST be able to ask for full ranges, which are then given exactly
  as in 1.0.2. CLI output is unchanged.
- **FR-014**: The `project` block MUST accept a list of warning codes to switch off for every run,
  with the same rules as FR-011, except that a code loko does not emit is reported as a warning
  rather than refused (so an upgrade that retires a code does not break the project). Per-run
  ignore options add to the project's list.

**Render and format through MCP (Story 3)**

- **FR-015**: The MCP server MUST offer a build tool that takes output formats, an output directory
  and an optional view, writes exactly what `loko build` writes for the same inputs, and returns the
  written paths.
- **FR-016**: The build tool MUST optionally return one named view's diagram inline, as diagram
  source or as an image.
- **FR-017**: The build tool MUST refuse an output directory outside the project root and MUST
  follow the same ownership rules as `loko build` for the output directory (it never removes files
  it did not write).
- **FR-018**: The build tool MUST write nothing when the model has errors, and return the
  diagnostics.
- **FR-019**: The MCP server MUST offer a format tool that previews the formatting change as a diff,
  or applies it, with the same result as `loko fmt`. Applying MUST use the same revision check as
  `apply_edit`, so it cannot overwrite a concurrent change.

**Empty-project bootstrap (Story 4)**

- **FR-020**: `describe` on a directory with no source files MUST say the project is empty, show the
  minimal valid project declaration, and return a ready-to-send `apply_edit` request that creates
  it.
- **FR-021**: `apply_edit` MUST accept a `project` target, so a client can create the project
  declaration with a name, an optional description and a target file, with the same preview,
  confirmation and revision safety as any other edit. The same target MUST allow changing the
  project's attributes later.
- **FR-022**: Creating a project MUST be refused when one already exists, naming the file that
  declares it.
- **FR-023**: Adding an element to a project with no project declaration MUST be refused, and the
  refusal MUST show how to create the project.

**Documentation**

- **FR-024**: The language reference, the MCP reference and the CLI reference MUST describe every
  new tool, option, marker and warning code, and the diagnostics reference MUST list the new
  warning.

### Key Entities

- **Language description**: the set of block types, their nesting, their attributes (name, value
  type, required, allowed values), the address grammar, the `docs` rule and one example per block.
  Derived from the language definition; the same for every project.
- **Planned marker**: a property of a deployment instance saying its infrastructure is not built
  yet, so no binding is expected.
- **Diagnostic summary**: diagnostics grouped by code, each group with a severity, a count and the
  addresses it affects.
- **Warning suppression**: a set of warning codes left out of output and strict-mode counting, given
  per run, or for the whole project on the `project` block (FR-014).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Re-running the reported first session's model from scratch through MCP takes no
  refused edits caused by a wrong guess about the language (down from about eight).
- **SC-002**: With the expected warnings handled (externals modelled as `external`, unbuilt
  instances marked planned), the reported 12-container model shows at most 10 warnings, down from
  33, and every remaining warning is one the team must act on.
- **SC-003**: A summary of the reported model fits on one screen (at most 30 lines).
- **SC-004**: MCP validate output for the reported model is at most half its 1.0.2 size by default.
- **SC-005**: A client with no shell access can go from an empty directory to a rendered diagram
  using MCP tools only.
- **SC-006**: Every attribute the compiler accepts appears in the language description, and none it
  rejects, checked on every test run.
- **SC-007**: Existing projects and existing CLI invocations produce the same output as in 1.0.2
  when none of the new options or markers are used. The one deliberate change is the compact MCP
  diagnostics default (FR-013), recorded in the changelog.

## Assumptions

- The language description covers loko 1.x as shipped plus this feature's planned marker. Attributes
  added in feature 018 extend it automatically, because it is derived from the language definition
  (FR-004).
- The build tool writes to the same default output directory as `loko build` when none is given.
- The image returned inline by the build tool is the same SVG `loko build` writes; no new image
  format is added.
- The new planned-marker warning follows existing naming (for example `planned_instance_bound`) and
  is a warning, not an error.
- `--severity` and `--ignore` are new CLI flags on `validate` (and `--ignore` on `build`); existing
  flags keep their meaning.
- Out of scope: changing how `docs` paths resolve, `node` attributes, view `title` and
  `description`, layout hints (018), and dynamic views (019).
