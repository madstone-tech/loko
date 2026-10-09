# Feature Specification: MCP HCL Authoring

**Feature Branch**: `015-mcp-hcl-authoring`

**Created**: 2026-10-01

**Status**: Draft

**Input**: Stage 3 of [`specs/012-v1-architecture-dsl/roadmap.md`](../012-v1-architecture-dsl/roadmap.md). Depends on [`013-hcl-compiler-core`](../013-hcl-compiler-core/spec.md) and [`014-viewmodel-renderers`](../014-viewmodel-renderers/spec.md), both merged.

## Overview

loko's conversational interface starts but offers nothing. The tools it used to have wrote files into
the old file-tree model, and they were removed along with that model. Today an assistant connected
to loko can neither answer a question about an architecture nor change one.

This feature gives the assistant both again, on the new foundation. It can **read**: summarise a
project at a chosen level of detail, answer graph questions (what depends on this, what does this
reach, how do these two connect, what is isolated, what is most coupled), and report compile
diagnostics. It can **write**: add, change and remove elements, relationships, deployment instances
and their bindings, and rename an element. Writes go to the architecture source and nowhere else.

Two guarantees make the write side safe to hand to an assistant. A write never damages what a person
wrote: comments, ordering and formatting in hand-written source survive every edit byte for byte,
outside the lines the edit is about. And a write never leaves the architecture broken: every change
is compiled before anything is saved, and a change that would not compile saves nothing and explains
why.

The same graph questions become available on the command line, answered by the same engine, so a
person and an assistant asking the same question always get the same answer.

## Clarifications

### Session 2026-10-05

- Q: When the assistant needs to make several related changes, should it be able to submit them as
  one request that is applied all together or not at all? → A: Yes. Single edits, plus an optional
  batch of edits, applied in order, compiled once at the end, and saved all-or-nothing.
- Q: Before an edit is saved, should the assistant be able to ask for a preview that shows the exact
  change and compiles it but writes nothing? → A: Yes. An optional preview mode on every write
  returns the textual diff and any diagnostics, and writes nothing.
- Q: (Amendment, 2026-10-09, found while testing with a real architecture.) Can the assistant
  create and change declared views? → A: Yes. A view is an edit target like any other: add,
  update and remove, with `include` and `exclude` as element references and `tags` as strings.
  Descriptions list the declared views.
- Q: Should the assistant be allowed to write an element's prose file, or only set and clear the
  reference to it? → A: Never write it. Tools may set or clear the `docs` reference only; prose
  files are never created or changed by the conversational interface.
- Q: When the assistant removes an element that other declarations still point to, should loko
  refuse, or remove the dependents too when explicitly asked? → A: Refuse by default; an explicit
  cascade option also removes every dependent and lists each one in the result and in the preview.
- Q: When the assistant renames an element, may it also change the element's kind, or only its
  name? → A: Name and kind. The rename record covers both; parent rules are enforced by the compile
  check, and a batch can set the new parent in the same request.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ask the architecture a question (Priority: P1)

An engineer working with an assistant asks "what breaks if the orders database goes down?" The
assistant asks loko for everything that depends on that element, directly or indirectly, and answers
from the compiled architecture, not from guessing at files.

**Why this priority**: Reading comes first because it is useful on its own: it answers real questions
about any existing architecture with no risk to it, and every write workflow starts with a read.

**Independent Test**: Point the assistant at an existing fixture architecture and ask for dependents,
dependencies, a path between two elements, isolated elements, and the most-coupled elements. Every
answer matches what can be checked by hand against the source.

**Acceptance Scenarios**:

1. **Given** a compiled architecture, **When** the assistant asks for a project summary, **Then** it
   receives a compact overview, and can ask again for more detail (structure, then full) or for one
   element and its surroundings only.
2. **Given** an element, **When** the assistant asks what depends on it, **Then** it receives every
   element with a relationship into it, and, when asked for transitive results, everything that
   reaches it by any chain.
3. **Given** two elements, **When** the assistant asks how they connect, **Then** it receives a
   shortest chain of relationships between them, or a clear statement that none exists.
4. **Given** an architecture, **When** the assistant asks for orphans or coupling, **Then** it
   receives the elements with no relationships, or the elements ranked by how many relationships
   enter and leave them.
5. **Given** an architecture that does not compile, **When** the assistant asks any question,
   **Then** it receives the diagnostics, each with file, line and column, instead of an answer built
   on a broken model.
6. **Given** any read, **When** the assistant does not ask for a format, **Then** the answer comes
   in the token-efficient format; JSON is available on request.

---

### User Story 2 - Build an architecture through conversation (Priority: P1)

An architect describes a new platform to an assistant: three systems, their containers, how they talk,
and where they run. The assistant creates each element, connects them, and places them in
environments. At the end the project compiles cleanly, and the architect can open the source and
read it as if a careful person had written it.

**Why this priority**: This is the exit criterion of the stage: an assistant can build a
multi-system architecture end to end. Without it, conversational authoring does not exist.

**Independent Test**: Starting from a project with only a `project` block, drive the write
operations to build a fixture-sized architecture (at least two systems, six containers, components,
relationships, two environments with nested placement and bindings). The result compiles with no
errors and matches a hand-built reference.

**Acceptance Scenarios**:

1. **Given** a project, **When** the assistant adds a system, a container inside it, and a
   relationship between two elements, **Then** each appears in the source in canonical form, and the
   project compiles.
2. **Given** an existing element, **When** the assistant changes its description, technology, owner,
   tags or prose reference, **Then** exactly that attribute changes in the source.
3. **Given** an environment, **When** the assistant adds an instance of a container inside a nested
   placement group and attaches a binding, **Then** the instance and binding appear in the right
   place, and the project compiles.
4. **Given** an element that nothing references, **When** the assistant removes it, **Then** its
   declaration disappears and nothing else in the source changes.
5. **Given** an element that other declarations still reference, **When** the assistant tries to
   remove it, **Then** nothing is saved and the response names each declaration that still refers to
   it.
6. **Given** an element that other declarations still reference, **When** the assistant removes it
   with cascade, **Then** it and every dependent declaration are removed, the response lists each
   one, and the project compiles.
7. **Given** any write, **When** it succeeds, **Then** the response says which files changed and
   includes any warnings the compiler reported.
8. **Given** any write, **When** the assistant asks for a preview, **Then** it receives the exact
   textual change and any diagnostics, nothing on disk changes, and applying the same request
   afterwards produces exactly the previewed change.

---

### User Story 3 - Never corrupt what a person wrote (Priority: P1)

A team hand-writes most of its architecture, with comments explaining decisions, deliberate ordering
and its own layout. It lets an assistant make small edits. After every edit, everything the team
wrote outside the edited declaration is exactly as it was.

**Why this priority**: P1 because it is the condition for letting an assistant near the source at
all. An assistant that quietly reformats a file, drops a comment, or reorders declarations destroys
trust in one edit, and an assistant that corrupts the source of truth is the failure that would sink
the tool.

**Independent Test**: Take hand-written fixture files full of comments, unusual spacing and
deliberate ordering. Apply every kind of edit. Compare each file before and after: every byte outside
the edited declaration is identical, and the compiled result differs only by the intended change.

**Acceptance Scenarios**:

1. **Given** a hand-written file with comments and custom spacing, **When** the assistant edits one
   element in it, **Then** every byte of the file outside that element's declaration is unchanged.
2. **Given** any sequence of edits, **When** each is applied, **Then** compiling after the edit
   yields the compiled architecture before the edit plus exactly the intended change, and nothing
   else.
3. **Given** an edit that would leave the project unable to compile, **When** the assistant submits
   it, **Then** nothing on disk changes, and the response carries the diagnostics explaining why.
4. **Given** source that was changed on disk after the assistant last read it, **When** the
   assistant submits an edit quoting the older revision, **Then** the edit is refused rather than
   overwriting the newer change, and the assistant is told to read again.
5. **Given** any request, **When** it would write anything other than architecture source (a
   diagram, a diagram source file, a document, a page), **Then** it is refused. No conversational
   tool can produce rendered output.

---

### User Story 4 - Rename without losing history (Priority: P2)

An element was named badly early on. The architect asks the assistant to rename it. Every reference
to it across every file updates, and the rename is recorded, so a later comparison between versions
reports a rename rather than one element disappearing and an unrelated one appearing.

**Why this priority**: Renames are common and painful by hand across many files. P2 because the
core write operations cover most authoring; this makes one frequent operation safe and keeps
history meaningful for the comparison stage that follows.

**Independent Test**: Rename an element that is referenced from several files, including as a
parent, a relationship target and a deployment instance's subject. The project compiles, every
reference uses the new name, and a rename record exists in the source.

**Acceptance Scenarios**:

1. **Given** an element referenced across several files, **When** the assistant renames it,
   **Then** every reference uses the new name and the project compiles.
2. **Given** a rename, **When** it is applied, **Then** the source records that the old address
   moved to the new one, and the record compiles and is validated.
3. **Given** a rename to a name already in use, **When** it is submitted, **Then** nothing is saved
   and the response says which declaration already uses the name.
4. **Given** a container the architect decides is really a component, **When** the assistant renames
   it to a component and sets a container parent in the same batch, **Then** every reference
   follows, the rename record covers the change of kind, and the project compiles.
5. **Given** a rename record whose old address still exists, or whose new address does not, **When**
   the project compiles, **Then** a diagnostic names the record.

---

### User Story 5 - Ask the same questions from the terminal (Priority: P2)

An engineer without an assistant asks the same graph questions on the command line and gets the same
answers the assistant would.

**Why this priority**: P2 because the assistant path is the stage's focus. Having one engine behind
both front ends is what keeps their answers from drifting apart, and it makes the query engine
testable and scriptable.

**Independent Test**: For each query kind, run the command and the conversational tool against the
same fixture with the same inputs, and compare: the answers are identical.

**Acceptance Scenarios**:

1. **Given** a fixture, **When** the engineer asks for dependents, dependencies, a path, orphans or
   coupling on the command line, **Then** the answer is identical to the conversational tool's
   answer for the same inputs.
2. **Given** an unknown element address, **When** it is queried, **Then** the error names the
   address and suggests the closest existing ones.
3. **Given** a query, **When** the engineer asks for JSON, **Then** the output is machine-readable
   and can be piped into other tools.

---

### Edge Cases

- **Empty project.** Only a `project` block exists. Reads return an empty-but-valid summary, and the
  first write creates the first element.
- **No file chosen for a new element.** The element goes in the file that declares its parent, or,
  for a top-level element, the file that declares the project.
- **A file outside the project.** A request names a target file outside the project root, or one
  that is not architecture source. It is refused.
- **Concurrent writers.** Two edits arrive close together for the same file. They are applied one at
  a time, and each is compiled against the result of the previous one.
- **Large architecture.** Reads and writes on a 1,000-element project complete within the stated
  performance bounds.
- **Cycles.** Elements form a dependency cycle. Transitive queries terminate and report each element
  once, and path queries return a shortest path.
- **A no-op edit.** Setting an attribute to its current value changes nothing on disk and says so.
- **A file that does not parse.** A write that targets, or would compile alongside, a file with
  syntax errors is refused with the parse diagnostics, and nothing is written.
- **Removing a placement group that still holds instances.** The removal is refused, naming the
  instances, unless they are removed in the same batch or the removal cascades.
- **Changes with no valid intermediate state.** Retargeting a relationship from a removed element to
  its replacement, or swapping two elements' parents, is submitted as one batch; each step alone
  would not compile.

## Requirements *(mandatory)*

### Functional Requirements

#### Reading

- **FR-001**: The system MUST provide a description of the compiled architecture at three levels of
  detail: summary, structure and full. Every level lists the declared views.
- **FR-002**: A description MUST be scopable to one element address, returning that element, what
  it contains, and its immediate relationships.
- **FR-003**: The system MUST answer, for any element: its direct dependents, its direct
  dependencies, and, when asked, the transitive closure of either.
- **FR-004**: The system MUST return a shortest relationship path between two elements, or state
  that none exists.
- **FR-005**: The system MUST list elements with no relationships, and rank elements by coupling:
  the relationships entering and leaving them.
- **FR-006**: The system MUST compile on request and return every diagnostic with its source file,
  line and column.
- **FR-007**: A read against an architecture that fails to compile MUST return the diagnostics
  rather than a partial answer.
- **FR-008**: Reads MUST default to the token-efficient output format, and MUST offer JSON on
  request.
- **FR-009**: Every read result MUST be deterministic: the same architecture and inputs always yield
  identical output.

#### Writing

- **FR-010**: The system MUST support adding, updating and removing elements (person, system,
  container, component, external), relationships, environments, placement groups, deployment
  instances, their bindings, and declared views.
- **FR-011**: An update MUST change only the attributes named in the request.
- **FR-012**: The system MUST compile the result of every write before saving anything. A write
  whose result has compile errors MUST save nothing, and MUST return the diagnostics.
- **FR-012a**: A write MAY be a batch: an ordered list of edits applied in sequence. Intermediate
  states need not compile; only the final result is compiled. The batch is saved all-or-nothing: if
  any edit cannot be applied, or the final result has compile errors, no file changes, and the
  response identifies the failing edit or the diagnostics.
- **FR-012b**: Every write, single or batch, MUST offer a preview mode that performs the same
  application and compile check, returns the exact textual change per file and any diagnostics, and
  writes nothing to disk. Applying the same request without preview afterwards MUST produce exactly
  the previewed change, provided no file changed on disk in between (FR-016).
- **FR-013**: An edit MUST leave every byte of every file outside the edited declarations
  unchanged: comments, ordering, spacing and formatting included. A declaration includes its
  directly attached leading comments; removing a declaration also removes one adjacent blank line.
  For a rename (FR-022), every declaration that references the renamed element is an edited
  declaration, and within it only the reference changes.
- **FR-014**: New declarations MUST be written in canonical form, identical to what the formatter
  would produce.
- **FR-015**: Compiling after a successful edit MUST yield the previous compiled architecture plus
  exactly the requested change, and nothing else.
- **FR-016**: A write MUST be refused if any file it would change has changed on disk since the
  content the request was based on, and the response MUST say so.
- **FR-017**: Writes to the same project MUST be applied one at a time.
- **FR-018**: By default, a removal that would leave dangling references MUST be refused, naming
  every declaration that still refers to the target.
- **FR-018a**: A removal MAY request cascade. A cascading removal also removes every declaration
  that depends on the target: relationships into or out of it, its children and their dependents,
  and deployment instances of it with their bindings. It then compiles and saves all-or-nothing like
  any write. The result, and the preview, MUST list every declaration removed.
- **FR-019**: A write MUST report which files it changed, and any warnings the compiler produced.
- **FR-020**: A write MUST target only architecture source files inside the project root.
- **FR-021**: A no-op write MUST change nothing on disk and say so.

#### Renaming

- **FR-022**: The system MUST rename an element, changing its name, its kind, or both, and update
  every reference to it in every file. A kind change that breaks a parent rule (for example, a
  component without a container parent) fails the compile check and saves nothing, unless the same
  batch also sets a valid parent.
- **FR-023**: A rename MUST record, in the source, that the old address moved to the new one.
- **FR-024**: The architecture language MUST accept the rename record, and MUST validate it: its old
  address must no longer be declared, its new address must exist, and no address may be the source
  of two records. A violation MUST be a diagnostic with a source range.
- **FR-025**: A rename to an address already in use MUST be refused, naming the existing
  declaration.

#### What the conversational interface may write

- **FR-026**: No conversational tool MUST be able to write anything other than architecture source:
  no diagram, no diagram source, no document, no page.
- **FR-026a**: An edit MAY set or clear an element's prose reference. No conversational tool may
  create, change or delete the prose file it points to, even inside the project root.
- **FR-027**: The v0 file-scaffolding tools MUST NOT return. The conversational interface consists
  of the read and write tools in this specification.

#### Command line

- **FR-028**: The system MUST provide a query command offering the same queries as the
  conversational interface: dependents, dependencies (direct or transitive), path, orphans and
  coupling.
- **FR-029**: The command and the conversational tool MUST produce identical results for identical
  inputs, because both MUST call the same shared operation.
- **FR-030**: An unknown address MUST produce an error naming it and suggesting the closest
  existing addresses.
- **FR-031**: The query command MUST use the existing three exit codes and no fourth.

#### Architecture rules

- **FR-032**: Conversational and command-line handlers MUST contain no logic beyond translating
  input and output and invoking shared operations, within the existing per-function size budgets.

### Key Entities

- **Description**: A view of the compiled architecture at one level of detail (summary, structure,
  full), optionally scoped to an address.
- **Query**: A graph question (dependents, dependencies, path, orphans, coupling) with its inputs and
  a deterministic, ordered answer.
- **Edit**: A requested change to one declaration: add, update or remove an element, relationship,
  environment, placement group, instance or binding. It names its target address, the attributes to set, and
  optionally the file to place a new declaration in. Edits can be submitted alone or as an ordered
  batch that is compiled once and saved all-or-nothing.
- **Edit result**: The files changed, the compiler's warnings, or (on refusal) the diagnostics and
  reasons, with nothing saved. In preview mode it carries the textual diff of each file that would
  change, and nothing is saved.
- **Rename record**: A declaration in the source stating that an old address now lives at a new one;
  the two may differ in name, kind, or both.
  It is validated by the compiler and consumed by the comparison stage that follows.
- **Revision** (content fingerprint): What an edit was based on. Every read returns one; every
  write quotes it, which lets the write detect that a file changed underneath it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Starting from an empty project, an assistant using only the conversational tools
  builds an architecture of at least two systems, six containers, components, relationships and two
  environments with nested placement and bindings. The result compiles with zero errors and matches
  a hand-built reference.
- **SC-002**: Across every edit kind applied to hand-written fixtures, 100% of bytes outside the
  edited declarations are unchanged, verified by comparing every file before and after.
- **SC-003**: A randomized round-trip test applies thousands of generated edit sequences. For every
  edit, compiling afterwards yields the prior compiled architecture plus exactly the intended
  change; zero divergences are tolerated.
- **SC-004**: 100% of write attempts whose result would not compile leave every file on disk
  byte-identical, verified by test.
- **SC-005**: No conversational tool produces any file other than architecture source, verified by
  inspecting what every tool can write.
- **SC-006**: For every query kind and fixture, the command-line answer and the conversational
  answer are byte-identical.
- **SC-007**: On a 1,000-element architecture, any read completes in under 1 second and any single
  edit, including its compile check, in under 2 seconds.
- **SC-008**: A summary-level description of a 1,000-element architecture fits in at most 300 tokens
  in the token-efficient format, small enough for an assistant to load at the start of every
  conversation.

## Assumptions

- **Scope boundary.** This is stage 3 of the v1.0 roadmap. The comparison engine that consumes
  rename records (stage 4), infrastructure ingestion (stage 5) and the policy engine (stage 6) are
  later features. This feature adds the rename record to the language and validates it; what a
  comparison does with it belongs to stage 4.
- **The language gains exactly one construct.** The rename record is the only addition. Like the
  rest of the language it is permanent surface, so its form is chosen deliberately. Everything else
  an edit can express already exists.
- **Writes are surgical edits, not regeneration.** Edits change the specific declarations involved
  and leave the rest of each file untouched. Rewriting a whole file from the compiled model would
  satisfy "compiles afterwards" but would destroy comments and layout, which FR-013 forbids.
- **Prose files are not written** (confirmed in Clarifications). An edit may set or clear an
  element's prose reference, but prose is authored by people: the rule that the conversational
  interface writes architecture source and nothing else (FR-026, FR-026a) admits no exception.
  Assistant co-authoring of prose would be a separate feature with its own safeguards.
- **One project per server.** The conversational server serves the project it was started in, as it
  does today.
- **The token-efficient default stays.** As established by the earlier decision to default
  conversational reads to the token-efficient format, reads default to it and offer JSON.
- **Governance constraints apply unchanged.** Dependency direction, ports and adapters, the
  thin-handler budgets (≤ 30 lines per conversational tool function, ≤ 50 per CLI function) and the
  file-size budgets govern this work.
