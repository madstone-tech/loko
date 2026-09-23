# Feature Specification: HCL Compiler Core

**Feature Branch**: `013-hcl-compiler-core`

**Created**: 2026-09-21

**Status**: Draft

**Input**: User description: `@specs/012-v1-architecture-dsl/design.md` — scoped to Stage 1 of [`specs/012-v1-architecture-dsl/roadmap.md`](../012-v1-architecture-dsl/roadmap.md) (`013-hcl-compiler-core`)

## Overview

Today an architecture is described in two places at once: prose files carry relationship
declarations, and diagram files carry the same relationships in arrow form. The two are merged when
artifacts are generated, and a separate drift check exists to report when they disagree. Because
nothing is authoritative, edits go stale silently, there is no way to ask what changed between two
revisions, and there is nothing well-defined to compare a real environment against.

This feature replaces the data model with a single authored artifact and a compiler. An author
writes the architecture in one declarative language; the tool reads it, resolves every reference,
and produces one fully-resolved, addressable value that every later capability (rendering,
diffing, reconciliation, policy) consumes. A broken reference becomes an error with a file, line,
and column instead of a quietly wrong diagram.

Stage 1 delivers the language, the compiler, validation, and the three commands that need nothing
else: check, format, and export. Rendering, diffing, infrastructure ingestion, and policy are
separate downstream features and are explicitly out of scope here.

## Clarifications

### Session 2026-09-22

- Q: When a deployment instance sits inside a nested placement group, should the placement group's path be part of the instance's stable address? → A: No. Instances are addressed directly under their environment (`deployment.prod.instance.api`); placement groups carry their own addresses and remain in the compiled result as placement data, but do not prefix the instances inside them. Instance names must therefore be unique within an environment, and re-parenting an instance reads as a change rather than a remove-plus-add.
- Q: Should the format command offer a mode that reports whether files would change, without rewriting them? → A: Yes. A check-only option lists every non-canonical path, writes nothing, and reuses the existing "errors present" exit code rather than introducing a fourth code.
- Q: Should the exported compiled architecture carry a version field identifying the shape of the export itself? → A: Yes — a single schema version integer, incremented only on an incompatible change to the export's shape, and deliberately independent of the product's release version so that exports stay byte-identical across releases that do not change the shape. Consumers reject an unrecognised version with a clear message. The producing tool's version is not embedded.
- Q: Should the check command be able to emit its diagnostics in a machine-readable form as well as human-readable text? → A: Yes. The check command gains a machine-readable output option emitting each diagnostic's severity, message, and file/line/column range in deterministic order, so that continuous-integration annotation and the later authoring interface consume one representation instead of parsing console text. The standard static-analysis interchange format remains deferred to the policy feature.
- Q: Which string-manipulation helpers should the language provide, and should the set be fixed for this release? → A: A fixed set of exactly five — join, split, lower, upper, replace — plus string interpolation. Any other function call is an unknown-function error that names the supported set. The set is deliberately under-shipped: adding a function later is painless, removing one is a breaking change.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Author an architecture in one place and have it checked (Priority: P1)

An architect describes their platform in declarative source files: the people and systems involved,
the containers inside each system, the components inside each container, external parties, and the
relationships between them. Relationships are written inside the element they originate from, and
they point at another element by name rather than by a repeated string. The architect runs the
check command and is told, precisely and with a file/line/column location, about anything that does
not hold together — a relationship pointing at something that does not exist, a component placed
directly under a system, an element nested inside itself.

**Why this priority**: This is the whole premise. Without a single authored source that can be
checked, nothing downstream is worth building. Everything else in v1 is a projection of the value
this story produces.

**Independent Test**: Write a small multi-system architecture by hand, run the check command, and
confirm it reports success. Introduce a typo in a reference and confirm the check fails with an
error naming the file, line, and column of the typo. This alone gives an author a linted,
trustworthy architecture description — value with no other feature present.

**Acceptance Scenarios**:

1. **Given** a project whose source files describe two systems, four containers, and the
   relationships between them, **When** the author runs the check command, **Then** the command
   reports no errors and exits with the success code.
2. **Given** a relationship whose target names an element that does not exist, **When** the author
   runs the check command, **Then** the command reports an unresolvable-reference error naming the
   file, line, and column of the offending reference, and exits with the error code.
3. **Given** a component declared as belonging to a system rather than to a container, **When** the
   author runs the check command, **Then** the command reports a wrong-parent error with a source
   location.
4. **Given** two elements each declared as contained by the other, **When** the author runs the
   check command, **Then** the command reports a containment-cycle error.
5. **Given** an architecture in which element A depends on B, B on C, and C back on A, **When** the
   author runs the check command, **Then** the command reports no error, because dependency cycles
   between elements are a normal architecture and only containment must be acyclic.
6. **Given** source files spread across several files and nested directories under the project
   root, **When** the author runs the check command, **Then** all of them are discovered and merged
   into one architecture, and an element declared in one file may be referenced from another.
7. **Given** a project with several errors, **When** a continuous-integration job runs the check
   command requesting machine-readable output, **Then** every diagnostic is emitted with its
   severity, message, and file/line/column range in a structured form, in the same deterministic
   order as the text output, and the exit code matches what the text output would have produced.
8. **Given** an element whose declaration includes an attribute the language does not define,
   **When** the author runs the check command, **Then** the command reports an unknown-attribute
   error rather than silently ignoring it.

---

### User Story 2 - Describe where the architecture runs, per environment (Priority: P1)

The same logical architecture runs in more than one environment. The author declares each
environment separately, optionally organising it into a nesting of placement groups, and states
which logical element each running instance realises, along with the attributes that differ per
environment (sizing, limits) and an identifier claim that ties the instance to a physical resource
in that environment's infrastructure.

**Why this priority**: The deployment description is part of the language, not an add-on. It must
be designed and validated now, because the downstream reconciliation feature consumes it and
changing the language later would invalidate every file authors have written. It is P1 alongside
Story 1 because shipping the logical plane alone would force a breaking language change.

**Independent Test**: Declare one logical architecture and two environments that instantiate it
differently, run the check command, and confirm both compile and appear in the exported result with
distinct per-environment attributes. Point two instances at the same physical identifier and
confirm the check rejects it.

**Acceptance Scenarios**:

1. **Given** an environment containing instances that each name a logical element, **When** the
   author runs the check command, **Then** the environment compiles and each instance is
   addressable in the exported result.
2. **Given** an instance naming a logical element that does not exist, **When** the author runs the
   check command, **Then** an unresolvable-reference error is reported with a source location.
3. **Given** two instances in the same environment both claiming the same physical resource
   identifier, **When** the author runs the check command, **Then** a duplicate-claim error is
   reported naming both claims.
4. **Given** an instance with no physical resource claim, **When** the author runs the check
   command, **Then** a warning — not an error — is reported, and the command still exits with the
   success code unless strict mode is requested.
5. **Given** two environments instantiating the same logical element with different per-environment
   attributes, **When** the author exports the result, **Then** both instances appear with their own
   attribute values and the shared logical element appears once.
6. **Given** an instance that the author moves from one placement group to another without renaming
   it, **When** the author exports the result, **Then** the instance's address is unchanged and only
   its recorded placement differs.
7. **Given** two instances in the same environment sharing a name but sitting in different placement
   groups, **When** the author runs the check command, **Then** a duplicate-name error is reported
   naming both source locations.

---

### User Story 3 - Consume the compiled architecture as a stable artifact (Priority: P1)

A user or an automated pipeline needs the compiled architecture as data: every element and
relationship keyed by a stable address, with all references already resolved, in a machine-readable
form. The same input always produces byte-identical output, so the artifact can be committed,
compared, and used as a CI gate.

**Why this priority**: Every downstream feature — rendering, diff, reconciliation, policy — consumes
this artifact rather than re-reading the source. It is also what makes the whole stage testable:
without a stable exported value there is nothing to assert against. P1 because it is the contract
the rest of v1 is built on.

**Independent Test**: Export a fixture project twice and confirm the two outputs are byte-identical.
Confirm every element, relationship, and deployment instance appears under its documented address.
This delivers standalone value as a CI artifact and as input to external tooling.

**Acceptance Scenarios**:

1. **Given** a compiled project, **When** the author exports it, **Then** every element,
   relationship, and deployment instance appears keyed by its address, with relationship endpoints
   already resolved to those addresses rather than left as unresolved names.
2. **Given** the same unchanged source files, **When** the author exports twice, **Then** the two
   outputs are byte-identical, including the ordering of every collection.
3. **Given** a project whose source files are reordered or renamed without changing their content,
   **When** the author exports, **Then** the output is unchanged.
4. **Given** a project, **When** the author requests either supported export encoding, **Then** both
   carry the same information and both are deterministic.
5. **Given** a project that fails validation, **When** the author requests an export, **Then** no
   artifact is produced and the diagnostics are reported instead.
6. **Given** an exported artifact, **When** a consumer reads it, **Then** a schema version integer is
   present and is checked before any other field is interpreted.
7. **Given** an exported artifact whose schema version the running tool does not recognise, **When** a
   consumer reads it, **Then** the read is refused with a message naming the version found and the
   versions supported, and no partial result is returned.

---

### User Story 4 - Keep source files canonically formatted (Priority: P2)

The author runs a format command that rewrites their source files into one canonical style, so that
hand-written files and machine-written files are indistinguishable and formatting never shows up in
a review diff.

**Why this priority**: Important but not blocking. The architecture can be authored and checked
without it. It becomes load-bearing once the downstream authoring feature writes into the same
files, which is why it belongs in this stage rather than later.

**Independent Test**: Take a badly-indented source file, run the format command, and confirm it is
rewritten canonically; run the command again and confirm the file is unchanged.

**Acceptance Scenarios**:

1. **Given** a source file with inconsistent indentation and alignment, **When** the author runs the
   format command, **Then** the file is rewritten in canonical style.
2. **Given** an already-canonical file, **When** the author runs the format command, **Then** the
   file is byte-unchanged.
3. **Given** a source file containing comments and a deliberate declaration order, **When** the
   author runs the format command, **Then** every comment is preserved in place and no declaration
   is reordered.
4. **Given** a source file that does not parse, **When** the author runs the format command, **Then**
   a parse error is reported with a source location and the file is left untouched.
5. **Given** a project containing two unformatted files, **When** a continuous-integration job runs
   the format command in check-only mode, **Then** both paths are listed, no file on disk is
   modified, and the command exits with the "errors present" code.
6. **Given** a project whose files are all canonically formatted, **When** the format command is run
   in check-only mode, **Then** no path is listed and the command exits with the success code.

---

### User Story 5 - Declare the architecture's compatible tool versions (Priority: P3)

The author records in the source which tool versions can read this architecture, so a colleague with
an older build is told so plainly rather than getting confusing errors.

**Why this priority**: A small, cheap guard. It matters only once more than one version exists in
the wild, so it is the lowest priority in the stage — but it must be in the language from the first
release, because adding it later means older builds cannot honour it.

**Independent Test**: Declare a version constraint the running tool does not satisfy and confirm the
check command reports it as an error rather than proceeding.

**Acceptance Scenarios**:

1. **Given** a declared version constraint that the running tool satisfies, **When** the author runs
   the check command, **Then** the constraint contributes no diagnostic.
2. **Given** a declared version constraint that the running tool does not satisfy, **When** the
   author runs the check command, **Then** an error is reported naming the constraint and the
   running version, and no artifact is produced.
3. **Given** no declared version constraint, **When** the author runs the check command, **Then** no
   version diagnostic is produced.

---

### User Story 6 - Retire the superseded dual-source model (Priority: P2)

A maintainer removes the parts of the product that existed only to reconcile two sources of truth —
the file-tree reader, the separate configuration format, the prose-scaffolding operations, the
drift-detection machinery, and the standalone HTTP interface — along with the tests covering them,
so the codebase has one model and no way to reintroduce a second.

**Why this priority**: The new compiler can be built alongside the old model, so this is not
blocking. But leaving both in place means the old paths remain reachable and the defect this release
exists to remove survives. It ships in this stage, not a later cleanup.

**Independent Test**: Confirm the superseded code paths are gone, that the project builds and its
remaining tests pass, and that no command offers drift detection.

**Acceptance Scenarios**:

1. **Given** the completed stage, **When** a maintainer inspects the codebase, **Then** the file-tree
   project reader, the separate configuration format handling, the template-scaffolding dependency,
   the standalone HTTP interface, the element- and relationship-scaffolding operations, and the
   drift-detection machinery are all removed.
2. **Given** the removals, **When** the project is built and its test suite is run, **Then** both
   succeed, with the tests covering the removed code deleted rather than adapted.
3. **Given** the completed stage, **When** a user lists available commands, **Then** no
   drift-detection command or option is offered.
4. **Given** the completed stage, **When** the structural compliance check is run, **Then** it passes,
   including a rule that confines the parsing library to the adapter layer.

---

### Edge Cases

- **No source files.** The project root contains no architecture source. The check command reports a
  clear "no architecture source found" message rather than silently succeeding on an empty model.
- **Duplicate declarations.** The same element address is declared in two files. This is an error
  naming both source locations, not a merge.
- **Empty architecture.** A project declares itself but no elements. This compiles and exports an
  empty element set; warnings may note it, but it is not an error.
- **Reference to a container from the wrong plane.** A logical relationship targets a deployment
  instance, or a deployment instance targets a relationship. Rejected with a type error.
- **Self-relationship.** An element declares a relationship targeting itself. Permitted — a component
  calling itself recursively is describable — but flagged as a warning.
- **Deeply nested placement groups.** Environments nest placement groups several levels deep. Each
  group yields its own distinct address and its position in the nesting is preserved in the compiled
  result, but the instances inside are still addressed directly under the environment. Nothing is
  flattened or lost.
- **Same instance name in two placement groups.** Two instances in one environment share a name
  while sitting in different placement groups. Rejected as a duplicate-name error, because instance
  addresses do not include the placement path.
- **Very large project.** Several thousand elements across hundreds of files. Discovery, compile, and
  export complete within the stated performance bound; diagnostics remain individually locatable.
- **Unreadable or malformed file.** A source file has bad permissions or contains invalid syntax.
  Reported as a diagnostic against that file; other files still parse so the author sees every error
  in one run rather than one per invocation.
- **Call to an unavailable function.** An author reaches for a helper the language does not provide
  (formatting, trimming, a substring). Reported as an unknown-function error naming the five
  supported functions, so the message itself documents the available set.
- **Non-source files present.** Prose files, generated output, and unrelated content sit under the
  project root. Discovery ignores everything that is not architecture source.
- **Prose file referenced but absent.** An element points at a prose document that does not exist on
  disk. Reported as a warning, not an error — the architecture is still valid.
- **Warnings under strict mode.** A project with warnings and no errors exits successfully by
  default, and exits with the distinct warnings code when strict mode is requested.

## Requirements *(mandatory)*

### Functional Requirements

#### Source discovery and project definition

- **FR-001**: The system MUST discover all architecture source files recursively beneath the project
  root, identified by a dedicated file-name suffix, and merge them into a single architecture.
- **FR-002**: The system MUST ignore any file under the project root that does not carry the
  architecture-source suffix.
- **FR-003**: The system MUST accept project-level settings — at minimum a project name, a
  description, and a tool-version constraint — declared in the same language as the architecture
  itself, replacing the previous separate configuration format.
- **FR-004**: The system MUST report an error when the same address is declared more than once,
  naming every source location involved.
- **FR-005**: The system MUST report a clear, distinct message when no architecture source is found
  beneath the project root.

#### Language — logical plane

- **FR-006**: The language MUST provide element kinds for people, systems, containers, components,
  and external parties.
- **FR-007**: Each element MUST accept, at minimum, a description, an owner, a technology, a set of
  tags, and a reference to a prose document.
- **FR-008**: Relationships MUST be declared inside the element they originate from, carry a local
  name, and name their target by reference.
- **FR-009**: Each relationship MUST accept a description and a technology.
- **FR-010**: Containers MUST declare the system that contains them, and components MUST declare the
  container that contains them, each by reference.

#### Language — deployment plane

- **FR-011**: The language MUST provide an environment kind that accepts, at minimum, a provider, an
  account identifier, and a region.
- **FR-012**: An environment MUST support an optional nesting of placement groups, of arbitrary
  depth, each containing further placement groups or instances. Each placement group MUST carry its
  own address reflecting its position in the nesting, but a placement group MUST NOT contribute a
  segment to the addresses of the instances inside it.
- **FR-012a**: Instance names MUST be unique within an environment, regardless of which placement
  group each instance sits in. A duplicate instance name within one environment MUST be reported as
  an error naming both source locations.
- **FR-013**: An instance MUST reference exactly one logical element and MUST accept a map of
  environment-specific attributes.
- **FR-014**: An instance MUST support one or more claims over physical resources, each identifying
  a claim kind and selecting resources by exact identifier, a list of identifiers, a wildcard
  pattern, or a tag map.
- **FR-015**: The language MUST provide a project-level way to declare physical resource patterns
  that are deliberately not modelled, so that later features consuming the compiled result can
  exclude them.

#### Language — views, locals, and deliberate exclusions

- **FR-016**: The language MUST accept explicit view declarations that name a filtered subset of the
  architecture. In this stage they are parsed, validated, and carried through to the compiled
  result; rendering them is a downstream feature.
- **FR-017**: The language MUST support local values, string interpolation, and exactly five
  string-manipulation functions: join, split, lower, upper, and replace. No other function is
  provided in this release.
- **FR-017a**: A call to any function outside the set in FR-017 MUST be reported as an
  unknown-function error carrying a source location and naming the five supported functions, rather
  than being silently ignored or attempted.
- **FR-018**: The language MUST NOT provide iteration constructs, dynamic block generation,
  input-variable declarations, or a module system. Their use MUST be reported as an unknown-block
  error rather than silently ignored.

#### Reference resolution

- **FR-019**: A reference to another element MUST be written as a structured path rather than a
  repeated string, and MUST be resolvable across file boundaries.
- **FR-020**: Reference resolution MUST occur in two passes — first collecting every declared
  address, then resolving references against that collection — so that declaration order and file
  order have no effect on the outcome.
- **FR-021**: An unresolvable reference MUST be reported as an error carrying the file, line, and
  column of the reference.
- **FR-022**: A reference whose target is of a kind the context does not accept MUST be reported as a
  type error with a source location.

#### Compiled result

- **FR-023**: Compilation MUST produce a single fully-resolved result containing every element keyed
  by address, every relationship with both endpoints resolved to addresses, the environment
  hierarchy, and the declared views.
- **FR-024**: Addresses MUST be stable and derived from the declaration, following documented forms
  for elements, for relationships nested inside an element, for placement groups, and for deployment
  instances. An instance's address MUST be independent of its placement group, so that moving an
  instance between placement groups preserves its identity.
- **FR-025**: The compiled result MUST NOT be mutated after construction; every consumer receives
  the same value.
- **FR-026**: Every command and every downstream consumer MUST obtain the architecture through
  compilation; none may read or parse the source files directly.
- **FR-027**: The system MUST NOT write or require any lock file, state file, or database. The source
  files are the only persisted representation.

#### Validation

- **FR-028**: The system MUST report as errors: a source file that does not parse; an unresolvable
  reference; a component whose parent is not a container; a container whose parent is not a system;
  two claims over the same physical resource identifier; two instances sharing a name within one
  environment; a containment cycle; an unknown block, attribute, or function; and an unsatisfied
  tool-version constraint.
- **FR-028a**: A file that does not parse MUST be reported as a distinct parse error, separate from
  the unknown-block and unknown-attribute errors that apply to files which parse but declare
  something the language does not define. The distinction matters because the two have different
  remedies and because a file that does not parse cannot be checked for anything else.
- **FR-029**: The system MUST report as warnings: an element with no relationships; a system with no
  containers; an element with no prose document; a deployment instance with no physical claim; a
  self-relationship; and a prose reference pointing at a file that does not exist.
- **FR-030**: Every diagnostic MUST carry a severity, a human-readable message, and a file, line, and
  column range.
- **FR-030a**: Diagnostics MUST be representable in a machine-readable form carrying the same
  severity, message, and source range as the human-readable form, so that consumers never parse
  formatted console text to recover them.
- **FR-031**: The system MUST report every diagnostic it can find in a single run, rather than
  stopping at the first error.
- **FR-032**: Dependency cycles between elements MUST be accepted without diagnostic; only
  containment is required to form a single-parent acyclic tree.
- **FR-033**: Diagnostics MUST be emitted in a deterministic order for identical input.

#### Commands

- **FR-034**: The system MUST provide a check command that compiles the project and prints all
  diagnostics, with a strict option that escalates the exit code when warnings are present.
- **FR-034a**: The check command MUST offer a machine-readable output option that emits every
  diagnostic in the structured form of FR-030a, in the deterministic order of FR-033. Exit codes MUST
  be identical whichever output form is selected. The standard static-analysis interchange format is
  out of scope for this stage.
- **FR-035**: The system MUST provide a format command that rewrites source files canonically while
  preserving every comment, the declaration order, and the meaning of the file, leaving
  already-canonical files byte-unchanged and refusing to modify a file that does not parse.
- **FR-035a**: The format command MUST offer a check-only option that writes nothing, lists the paths
  of every file that is not canonically formatted, and exits with the "errors present" code when any
  file would change or the success code when none would, so that a continuous-integration job can
  fail on unformatted source without inspecting version control.
- **FR-036**: The system MUST provide an export command that writes the compiled result in either of
  two machine-readable encodings carrying equivalent information.
- **FR-036a**: Every exported artifact MUST carry a schema version integer identifying the shape of
  the export. This integer MUST be incremented only when the export's shape changes incompatibly, and
  MUST be independent of the product's release version, so that releases which do not change the
  shape continue to produce byte-identical exports.
- **FR-036b**: Any consumer reading a previously exported artifact MUST verify the schema version and
  MUST refuse an unrecognised version with a message naming the version found and the versions
  supported, rather than attempting a partial or best-effort read.
- **FR-036c**: The exported artifact MUST NOT embed the producing tool's version string, a timestamp,
  a hostname, a file-system path outside the project, or any other value that varies between runs or
  machines for identical input.
- **FR-037**: The export command MUST produce no artifact when compilation reports errors.
- **FR-038**: Commands MUST use exactly three distinct exit codes: success, errors present, and
  warnings present under strict mode. No command in this stage may introduce a fourth code; the
  format command's check-only option reuses the "errors present" code.
- **FR-039**: Command handlers MUST contain no logic beyond argument handling and invoking the shared
  operations, so that later interfaces call exactly the same operations.

#### Determinism

- **FR-040**: All collections in the compiled result and in every exported encoding MUST be emitted
  in a defined sorted order, so that identical input yields byte-identical output across runs,
  machines, and file-system orderings.

#### Retirement of the superseded model

- **FR-041**: The system MUST remove the file-tree project reader, the separate configuration-format
  handling, the template-scaffolding dependency, the standalone HTTP interface and its generated
  schema, every element- and relationship-scaffolding operation, and all drift-detection machinery.
- **FR-042**: Tests covering removed behaviour MUST be deleted alongside it, not adapted to keep
  passing.
- **FR-043**: The system MUST NOT offer any drift-detection command or option after this stage.
- **FR-044**: The structural compliance check MUST be extended with a rule preventing any package
  outside the adapter layer from importing the parsing or diagramming libraries, and MUST pass.
- **FR-045**: The diagram-parsing code that a later release will reuse for diagram import MUST be
  retained rather than deleted.

### Key Entities

- **Project**: The unit being compiled. Has a name, a description, a tool-version constraint, and a
  root directory beneath which source is discovered.
- **Element**: A node in the logical architecture — a person, system, container, component, or
  external party. Identified by an address; carries a description, owner, technology, tags, a prose
  reference, and its containment parent where applicable.
- **Relationship**: A directed edge, declared inside its source element, carrying a local name, a
  resolved target address, a description, and a technology. Addressed relative to its source.
- **Environment**: A named deployment target with a provider, account, and region, containing a tree
  of placement groups and instances.
- **Placement group**: An optional grouping inside an environment, nestable to arbitrary depth. Has
  its own address reflecting its position in the nesting, and records which instances it holds, but
  does not contribute a segment to those instances' addresses.
- **Instance**: A realisation of one logical element inside one environment, carrying
  environment-specific attributes, physical resource claims, and a reference to the placement group
  holding it. Named uniquely within its environment and addressed directly under it.
- **Claim**: An assertion that an instance corresponds to one or more physical resources, identified
  by kind and by an exact identifier, a list, a wildcard pattern, or a tag map.
- **View**: A named, filtered subset of the architecture, validated here and rendered downstream.
- **Address**: The stable identity of any element, relationship, environment, placement group, or
  instance. Derived from the declaration and used as the key throughout the compiled result.
- **Compiled result**: The immutable, fully-resolved value produced by compilation: elements by
  address, relationships with resolved endpoints, the environment hierarchy, and the views.
- **Diagnostic**: A severity, a message, and a file/line/column range. Produced by parsing,
  resolution, and validation alike.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An architecture is described in exactly one authored artifact; no rendered or
  secondary artifact is authored, and no reconciliation between two descriptions is required or
  offered.
- **SC-002**: 100% of references that do not resolve are reported as errors before any artifact is
  produced, each naming the file, line, and column of the reference — measured by a fixture set
  covering every reference position in the language.
- **SC-003**: Exporting the same unchanged project twice produces byte-identical output, verified on
  every fixture project in the test suite and across differing file-system orderings.
- **SC-004**: Every validation rule listed in the requirements has at least one fixture pairing an
  input that triggers it against the expected diagnostic, including the expected source range;
  coverage of the rule list is 100%.
- **SC-005**: The shipped documentation is sufficient, without reading source, to describe a
  two-system, six-container architecture with relationships and two environments and get a clean
  check result. Verified by a documentation review confirming that every block, attribute, and
  function in the language has a worked example, and that each error message names either the
  permitted values or the documentation section that lists them.
- **SC-006**: Checking a project of 1,000 elements across 200 source files completes in under 2
  seconds on a developer laptop, and a project of 5,000 elements completes in under 10 seconds.
- **SC-007**: Running the format command twice in succession leaves files byte-unchanged on the
  second run, for 100% of fixture files; comments and declaration order survive in 100% of cases.
- **SC-008**: A single check run reports every discoverable problem in the project; introducing five
  independent errors yields five diagnostics from one invocation, not one.
- **SC-009**: The superseded model layer is fully removed and the project builds and passes its test
  suite with no drift-detection capability reachable by any command or option.
- **SC-010**: The structural compliance check passes, including the new rule confining the parsing
  and diagramming libraries to the adapter layer.
- **SC-011**: An artifact exported by one release is read without error by any later release that has
  not incremented the export schema version, and is refused with a version-naming message by any
  release that has — verified by a fixture artifact pinned at the current schema version.
- **SC-012**: A continuous-integration pipeline can gate on unformatted source and on compile
  diagnostics using exit codes and machine-readable output alone, with no parsing of human-readable
  console text and no inspection of version-control state.

## Assumptions

- **Scope boundary.** This feature is Stage 1 of the v1.0 roadmap. Rendering and the site builder,
  semantic diff and rename declarations, infrastructure ingestion and reconciliation, the policy
  engine, and the conversational authoring interface are separate downstream features. Where the
  language must reserve syntax for them (views, resource-exclusion patterns, physical claims), the
  syntax is defined and validated here but only consumed later.
- **Rename declarations deferred.** The rename-declaration block belongs to the diff feature and is
  not part of this stage. Its operands have unusual evaluation semantics, so adding it here without
  the diff engine would mean specifying behaviour with nothing to exercise it.
- **No automated migration.** Existing projects on the previous model are not migrated by any
  command. The previous release remains installable, and a migration recipe is documented separately.
  Deciding this now avoids building a reader for the model being deleted.
- **Users are comfortable editing declarative configuration files** in a text editor and running a
  command-line tool. No graphical or web editor is provided.
- **Git remains the system of record** for history. The compiler is stateless; nothing is persisted
  outside the authored source files.
- **Existing consumers survive the transition.** The site builder, the diagramming output, the
  encoders, the conversational-interface harness, and the command shells remain in the codebase
  during this stage, even though they are rewired to the compiled result only in later stages. This
  stage must leave the project building and its remaining tests passing.
- **Determinism is a hard requirement, not a quality goal.** It is the precondition for the
  golden-file testing strategy and for reviewable generated output downstream, so it is specified as
  a requirement rather than left to implementation discretion.
- **Two export encodings.** Both the general-purpose machine-readable encoding and the compact
  encoding already used by the product are supported, carrying equivalent information.
- **Governance constraints apply unchanged.** The existing architecture constitution — dependency
  direction, ports-and-adapters separation, and the file- and function-size budgets — governs this
  work; the structural compliance check gates it.
