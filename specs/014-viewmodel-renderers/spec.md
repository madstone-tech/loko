# Feature Specification: ViewModel Renderers

**Feature Branch**: `014-viewmodel-renderers`

**Created**: 2026-09-23

**Status**: Draft

**Input**: Stage 2 of [`specs/012-v1-architecture-dsl/roadmap.md`](../012-v1-architecture-dsl/roadmap.md). Depends on [`013-hcl-compiler-core`](../013-hcl-compiler-core/spec.md), merged.

## Overview

The compiler landed in the previous stage: an architecture is described in one authored artifact and
compiles to a resolved, addressable value. Nothing yet draws it. `loko build` and `loko serve` were
removed along with the model they used to read, so today a user can check and export their
architecture but cannot see it.

This feature restores them on the new foundation, and changes the relationship between the model and
what is drawn. Previously a diagram was a second place the architecture was described, which is what
made the two drift apart. Now every rendered artifact is a projection of the compiled result and
nothing generated is ever edited by hand — the structural guarantee that replaces drift detection.

Between the compiled architecture and any output sits one intermediate step. The compiled result
resolves to a set of **views**, each a slice of the architecture at one level of detail; each view
projects to a **view model** that already decides which elements and connections appear, how they
nest, and how they are styled. Output backends consume a view model and know nothing about the
authoring language, the C4 model, or each other. Adding an output format later means adding one
backend, not touching the middle.

Views are available without configuring anything: one for the whole landscape, one per system, one
per container, and one per environment. A user who wants something narrower declares it, and that
declaration is already part of the language.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See the architecture without configuring anything (Priority: P1)

An author who has a checked architecture runs the build command and gets diagrams. They did not
declare any views, name any files, or choose any layout. They get a landscape picture of the whole
architecture, one picture per system showing what is inside it, one per container showing its
internals, and one per environment showing where things run.

**Why this priority**: This is the feature. A user who has to configure views before seeing anything
has been handed a second authoring task, which is what the previous release got wrong. Zero-config
output is what makes the compiled model worth having.

**Independent Test**: Take the architecture from the previous stage's documentation, run the build
command with no arguments, and confirm a diagram appears for the landscape, for each system, for
each container, and for each environment. Value is delivered with no other story present.

**Acceptance Scenarios**:

1. **Given** an architecture with two systems and six containers and no declared views, **When** the
   author runs the build command, **Then** diagrams are produced for the landscape, for each of the
   two systems, and for each container that has components, with no configuration.
2. **Given** an architecture with two environments, **When** the author runs the build command,
   **Then** one deployment picture is produced per environment, each showing only that environment's
   instances and their placement nesting.
3. **Given** a system with no containers, **When** the author runs the build command, **Then** no
   empty container view is produced for it and no error is reported.
4. **Given** an architecture where an element is connected to something outside the current view's
   boundary, **When** that view is produced, **Then** the connection is shown reaching the boundary
   rather than being silently dropped.
5. **Given** an architecture that fails to compile, **When** the author runs the build command,
   **Then** no artifact is written, the diagnostics are reported, and any previously generated output
   is left exactly as it was.

---

### User Story 2 - Narrow the picture to one concern (Priority: P2)

An architect wants a diagram of just the payment path across three systems, for a design review. They
declare a view naming what to include and exclude — using the declaration form the language already
has — and it is produced alongside the automatic ones.

**Why this priority**: The escape hatch for the real case. The automatic views cover the common need;
this covers the presentation, the review, and the "explain this one flow" request. It depends on
nothing in Story 1 beyond the projection step, and the declaration is already parsed and validated.

**Independent Test**: Declare a view naming two systems and one excluded container, build, and
confirm a diagram containing exactly the expected elements appears.

**Acceptance Scenarios**:

1. **Given** a declared view naming elements to include, **When** the author builds, **Then** a
   diagram is produced containing those elements and the connections between them.
2. **Given** a declared view that also names elements to exclude, **When** the author builds, **Then**
   the excluded elements do not appear, and connections that terminated on them are shown reaching the
   boundary.
3. **Given** a declared view selecting elements by tag, **When** the author builds, **Then** every
   element carrying that tag appears.
4. **Given** a declared view whose selection resolves to nothing, **When** the author builds, **Then**
   a warning names the empty view and no diagram file is produced for it.
5. **Given** a declared view with the same name as an automatic one, **When** the author builds,
   **Then** the declared view wins and a warning names the shadowed automatic view, so the author is
   not left wondering where the other picture went.

---

### User Story 3 - Trust that generated output is generated (Priority: P1)

A team commits the generated diagrams so that a pull request shows a visibly changed picture when
someone rewires the architecture. For that to work, two things must hold: rebuilding an unchanged
architecture must change nothing, and nobody must be tempted to edit the output by hand.

**Why this priority**: P1 because it is a precondition, not a nicety. If output churns between runs
the diffs are unreviewable and the team stops committing them; if generated files look editable
someone will edit one, and the second source of truth is back. This is the structural guarantee that
replaces drift detection, so it ships with the first story rather than after it.

**Independent Test**: Build twice without touching the source and confirm every produced file is
byte-identical. Open any produced file and confirm it says, at the top, that it is generated and
which source it came from.

**Acceptance Scenarios**:

1. **Given** an unchanged architecture, **When** the author builds twice, **Then** every produced
   file is byte-identical between the two runs.
2. **Given** the same architecture built on a different machine, or with its source files discovered
   in a different order, **When** the outputs are compared, **Then** they are byte-identical.
3. **Given** any generated file, **When** it is opened, **Then** its first lines state that it is
   generated, that it must not be edited, and which authored artifact it came from.
4. **Given** an architecture where one element's description changes, **When** the author rebuilds,
   **Then** only the files affected by that element differ, and unrelated files are byte-unchanged.
5. **Given** a previously generated output directory, **When** the author rebuilds after deleting an
   element, **Then** the file for the deleted element is removed rather than left behind as a stale
   orphan.

---

### User Story 4 - Read the architecture as prose and pictures together (Priority: P2)

A newcomer opens a generated site and reads their way through the architecture: a page per element
with its prose inlined, tables of what it uses and what uses it, its components listed, and the
relevant diagram embedded. They navigate by clicking, not by grepping.

**Why this priority**: The diagrams alone answer "what is connected to what". The prose answers
"why", and that is what a newcomer actually needs. P2 rather than P1 because the diagrams have
standalone value and this depends on them.

**Independent Test**: Build the site for a documented architecture, open it in a browser, and
navigate from the landscape to a container's page and read its prose.

**Acceptance Scenarios**:

1. **Given** an architecture whose elements reference prose files, **When** the author builds the
   site, **Then** each element has a page with its prose inlined and its diagram embedded.
2. **Given** an element with relationships, **When** its page is produced, **Then** it carries a
   generated table of what it uses and what uses it, each entry linking to the other element's page.
3. **Given** an element whose prose file is missing, **When** its page is produced, **Then** the page
   is still produced with its generated content and the absence is noted on the page rather than
   failing the build.
4. **Given** a built site, **When** a reader follows any internal link, **Then** it resolves to a
   page that exists.

---

### User Story 5 - See changes as they are made (Priority: P2)

An author editing their architecture keeps a browser open beside their editor. Saving a file updates
the browser without them touching it. When they save something that does not compile, the browser
shows the diagnostic instead of going blank or silently showing stale output.

**Why this priority**: It shortens the edit loop from minutes to seconds, which is most of what makes
the tool pleasant. It is P2 because everything it shows is already produced by Story 1; this is the
delivery mechanism.

**Independent Test**: Start the watch command, edit a description, save, and observe the browser
update without interaction.

**Acceptance Scenarios**:

1. **Given** the watch command is running, **When** the author saves a change to a source file,
   **Then** the browser reflects the change without manual reload.
2. **Given** the watch command is running, **When** the author saves a file that does not compile,
   **Then** the diagnostic is displayed with its file, line, and column, and the last good output is
   not silently served as if current.
3. **Given** a compile error has been displayed, **When** the author fixes it and saves, **Then** the
   view recovers without restarting the command.
4. **Given** the watch command is running, **When** the author saves a file that is not architecture
   source, **Then** no rebuild occurs.
5. **Given** several files are saved in quick succession, **When** the watcher reacts, **Then** one
   rebuild occurs rather than one per file.

---

### User Story 6 - Install one thing (Priority: P2)

A user installs the tool and renders diagrams. They do not install a diagram tool, a runtime, or
anything else, and they do not hit an error telling them a binary is missing from their path.

**Why this priority**: The previous release required a separate diagram binary on the path, which is
a support burden, a container-image burden, and the first thing to fail in CI. P2 because rendering
works either way; this is about what installing costs.

**Independent Test**: On a machine with no diagram tool installed, and with the path emptied, render
a diagram successfully.

**Acceptance Scenarios**:

1. **Given** a machine with no separate diagram tool installed, **When** the author renders a
   diagram, **Then** it succeeds.
2. **Given** the tool is run in a minimal container with no other executables present, **When** the
   author builds, **Then** every output format is produced.
3. **Given** any build, **When** it runs, **Then** no external process is started.

---

### User Story 7 - Change how it looks without forking (Priority: P3)

A team wants the generated site in their own colours and with their own header. They place override
files in a known directory and their versions are used instead of the built-in ones. They do not
maintain a fork, and they override only what they want to change.

**Why this priority**: Real adoption need, lowest priority because the built-in appearance is usable
and nothing else depends on this.

**Independent Test**: Override one part of the output, rebuild, and confirm the override is used
while everything not overridden is unchanged.

**Acceptance Scenarios**:

1. **Given** an override directory containing a replacement for one part of the output, **When** the
   author builds, **Then** the override is used and every other part is unchanged.
2. **Given** no override directory, **When** the author builds, **Then** the built-in appearance is
   used with no warning.
3. **Given** an override that is malformed, **When** the author builds, **Then** the error names the
   override file and the build fails rather than silently falling back — a theme that half-applies is
   harder to diagnose than one that refuses.
4. **Given** an override directory, **When** the author builds twice, **Then** the output is
   byte-identical between runs, exactly as it is without overrides.

---

### Edge Cases

- **Nothing to draw.** The architecture compiles but declares no elements. The build succeeds,
  produces no diagram files, and says so rather than emitting an empty picture.
- **A single element.** One system, nothing else. A landscape view is produced containing it; no
  container or component views are produced.
- **Very wide view.** One system contains 200 containers. The view is produced, every container's
  label appears in the rendered diagram, and it completes within the stated performance bound.
- **Deeply nested placement.** An environment nests placement groups five deep. The deployment view
  shows every level of nesting rather than flattening it.
- **Cyclic connections.** Elements form a dependency cycle. The diagram draws it; cycles between
  elements are legal.
- **Self-connection.** An element connects to itself. The diagram shows a self-loop rather than
  dropping the connection or failing.
- **Name that is not safe in a file name.** An element's name contains characters a file system
  rejects. The file name is derived deterministically and remains unique, and the mapping is stable
  across runs.
- **Two elements whose names differ only by case.** On a case-insensitive file system their pages
  would collide. The collision is detected and reported rather than one silently overwriting the
  other.
- **Output directory is not empty.** It contains unrelated files a user put there. The build replaces
  what it owns and leaves the rest alone.
- **Output directory cannot be written.** Reported as an error naming the directory, with nothing
  partially written.
- **Watch loop triggered by its own output.** The output directory sits inside the project root. The
  watcher ignores generated output rather than rebuilding forever.

## Requirements *(mandatory)*

### Functional Requirements

#### Views

- **FR-001**: The system MUST derive a set of views from the compiled architecture without
  configuration: one landscape view of the whole architecture, one view per system showing the
  elements it contains, one view per container that has components, and one view per environment.
- **FR-002**: The system MUST NOT produce a view that would contain no elements, and MUST NOT report
  that as an error.
- **FR-003**: The system MUST include declared views alongside the derived ones, honouring their
  include, exclude, and tag selections.
- **FR-004**: A declared view whose name collides with a derived one MUST take precedence, and the
  shadowing MUST be reported as a warning.
- **FR-005**: A declared view that resolves to no elements MUST be reported as a warning and MUST NOT
  produce an output file.
- **FR-006**: Every view MUST have a stable identity derived from what it depicts, so that the file
  it produces keeps the same name across runs.

#### Projection

- **FR-007**: Each view MUST project to a single intermediate value carrying the elements it shows,
  the connections between them, how those elements nest inside one another, and the visual treatment
  each has been assigned.
- **FR-008**: The projection MUST decide visual treatment; output backends MUST NOT make styling
  decisions of their own.
- **FR-009**: A connection whose other end lies outside the view MUST be represented as reaching the
  view boundary rather than being dropped.
- **FR-010**: The projection MUST be a pure function of the compiled architecture: the same compiled
  input always yields the same intermediate value.

#### Output backends

- **FR-011**: Every output backend MUST consume only the intermediate value and MUST NOT read the
  authored source, the file system, or the compiled architecture directly.
- **FR-012**: Backends MUST NOT depend on one another; adding or removing one MUST NOT require
  changing another.
- **FR-013**: The system MUST provide backends for diagram source, rendered vector images, prose
  documents, and a browsable site.
- **FR-014**: Diagram rendering MUST NOT start an external process or require any separately
  installed executable.
- **FR-015**: The system MUST select which formats to produce from a user-supplied list, and MUST
  report an unknown format by name alongside the list of supported ones.

#### Styling by convention

- **FR-016**: Element type MUST determine shape, consistently across every view and every backend.
- **FR-017**: An element representing something outside the team's control MUST be visually
  distinguished from one inside it.
- **FR-018**: An element's tags MUST be carried into the output in a form that allows a user's own
  styling to target them.
- **FR-019**: Styling MUST be deterministic: the same element always receives the same treatment.

#### Generated-output guarantees

- **FR-020**: Every generated file MUST open with a notice stating that it is generated, that it must
  not be edited, and which authored artifact it was produced from.
- **FR-021**: Building an unchanged architecture twice MUST produce byte-identical files, across runs,
  machines, and source-file discovery orders.
- **FR-022**: A change to one element MUST leave files unrelated to that element byte-unchanged.
- **FR-023**: The system MUST remove generated files whose source no longer exists, and MUST NOT
  remove or modify files in the output directory that it did not generate.
- **FR-024**: No artifact MUST be produced when the architecture fails to compile, and any previously
  generated output MUST be left unchanged.

#### The site

- **FR-025**: The site MUST provide a page per element carrying that element's prose, its diagram, a
  table of what it uses and what uses it, and a list of what it contains.
- **FR-026**: An element whose prose file is absent MUST still get a page, with the absence noted on
  the page rather than failing the build.
- **FR-027**: Every internal link in the site MUST resolve to a page that exists.
- **FR-028**: Two elements whose file names would collide MUST be reported as an error naming both,
  rather than one overwriting the other.

#### Watch and serve

- **FR-029**: The system MUST provide a mode that serves the site locally and updates the browser when
  the architecture changes, without manual reload.
- **FR-030**: A change that fails to compile MUST display the diagnostics with their source positions,
  and MUST NOT present stale output as if it were current.
- **FR-031**: Recovery from a failed compile MUST NOT require restarting the command.
- **FR-032**: Changes to files that are not architecture source MUST NOT trigger a rebuild, and
  generated output MUST NOT trigger a rebuild of itself.
- **FR-033**: Several changes arriving together MUST result in one rebuild.

#### Theming

- **FR-034**: The system MUST use replacement presentation files from a known directory when present,
  falling back to built-in ones for anything not overridden.
- **FR-035**: A malformed override MUST fail the build with an error naming the offending file, rather
  than silently falling back to the built-in version.

#### Commands

- **FR-036**: The system MUST provide a build command taking the set of formats to produce and the
  destination directory.
- **FR-037**: The system MUST provide a watch-and-serve command.
- **FR-038**: Both commands MUST use the existing three exit codes and MUST NOT introduce a fourth.
- **FR-039**: Command handlers MUST contain no logic beyond argument handling and invoking the shared
  operations.

#### Scope boundaries

- **FR-040**: The system MUST NOT provide a diagram format whose specification is still experimental
  or whose rendering is unreliable in common viewers; such formats are deferred.
- **FR-041**: No output backend MUST be reachable from the conversational interface. Generated
  artifacts are produced by the build command only.

### Key Entities

- **View**: A named slice of the architecture at one level of detail. Either derived automatically
  from the structure or declared by the author. Has a stable identity that determines its output file
  name.
- **View model**: The intermediate value a view projects to. Carries the elements shown, the
  connections between them, the nesting between elements, boundary markers for connections that leave
  the view, and the visual treatment assigned to each element and connection. The single input to
  every output backend.
- **Boundary connection**: A connection with one end outside the view, represented so the reader can
  see that something leaves without seeing where it goes.
- **Backend**: A component turning one view model into the bytes of one output format. Knows nothing
  about the authoring language, the architecture model, or other backends.
- **Generated artifact**: A file produced from a view model, carrying a notice naming its source, owned
  by the tool and never edited by hand.
- **Theme override**: A user-supplied replacement for part of the built-in presentation, resolved
  ahead of the built-in version.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An author with a compiled architecture and no configuration gets a complete set of
  diagrams from a single command, with zero decisions required beforehand.
- **SC-002**: Building the same unchanged architecture twice produces byte-identical output, verified
  across every fixture project, on differing source-discovery orders, and on two operating systems.
- **SC-003**: Changing one element's description leaves at least 90% of previously generated files
  byte-unchanged, so a review diff shows only what actually changed.
- **SC-004**: 100% of generated files carry a notice naming the artifact they came from, verified by
  scanning every produced file.
- **SC-005**: Rendering succeeds on a machine with an emptied executable search path, and the
  distributed artifact requires no companion installation.
- **SC-006**: Building an architecture of 1,000 elements completes in under 10 seconds, and a
  single-file edit under the watch command is reflected in under 2 seconds.
- **SC-007**: Every view kind — landscape, per system, per container, per environment, and declared —
  has at least one fixture pairing an input architecture with its expected output; coverage of the
  view-kind list is 100%.
- **SC-008**: Every internal link in a generated site resolves, verified by a link check over each
  fixture site with zero broken links.
- **SC-009**: A team can change the site's appearance without modifying or forking the tool, verified
  by overriding presentation from a directory outside it.
- **SC-010**: The build and watch commands are available again, and the tool's documented command set
  matches what it actually offers.

## Assumptions

- **Scope boundary.** This is stage 2 of the v1.0 roadmap. Semantic diff, infrastructure ingestion and
  reconciliation, the policy engine, and the conversational authoring interface are separate later
  features. This feature consumes the compiled architecture and produces output; it does not change
  the authoring language.
- **The authoring language does not change.** Declared views, tags, prose references, and the
  deployment nesting already exist and are already validated. If a view selection turns out to need
  expressive power the language lacks, that is a language change and belongs to its own feature.
- **A separately installed diagram tool is no longer required.** The previous release shelled out to
  one. Removing that is part of this feature, and the library needed to do it must be reintroduced as
  a dependency — it was dropped when the old code that used it was deleted.
- **The parked site builder is a starting point, not a drop-in.** It was written against the deleted
  model and must be reworked to consume view models. Its single large presentation file exceeds the
  file-size budget introduced in the current constitution and must be split during the rework, not
  after.
- **One diagram format is deferred.** A second diagram syntax whose specification is experimental and
  which renders poorly in common viewers is out of scope; the projection step is what makes adding it
  later a backend-only change.
- **Users are comfortable running a command-line tool and opening a local browser.** No hosted
  service, no account, no build server.
- **Generated output is committed.** The whole determinism requirement exists because teams put these
  files in version control so that a pull request shows a changed picture. If output were never
  committed, byte-stability would be a nicety rather than a requirement.
- **Governance constraints apply unchanged.** The architecture constitution — dependency direction,
  ports-and-adapters separation, and the file- and function-size budgets as amended in the current
  version — governs this work and gates it.
