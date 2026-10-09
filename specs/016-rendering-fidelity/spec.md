# Feature Specification: Rendering Fidelity

**Feature Branch**: `016-rendering-fidelity` (to be created when this feature starts)

**Created**: 2026-10-09

**Status**: Draft

**Input**: UX testing of feature 015. Two architectures were modelled entirely through the MCP
tools: a food-delivery platform designed from scratch, and an existing serverless AWS system
(Lambda, API Gateway, SQS, SNS, DynamoDB, S3) with nine hand-drawn D2 diagrams, modelled from its
Terraform. The latter is kept, anonymized, as the fixture `testdata/projects/serverless-reference`. Both models compiled and answered queries
well. The rendered diagrams, though, were clearly worse than the hand-drawn ones they should
replace:

- every container was the same blue rectangle;
- boxes showed raw identifiers (`read_api`, `portal`);
- container views were a 1600×370 strip of unreadable text;
- asynchronous and trigger connections could not be told apart, and had to be drawn against the
  data flow.

Findings G2, G3, G4, G6 and G9 in [the UX report](ux-findings-serverless.md).

## Clarifications

### Session 2026-10-09

- Q: Should tags on relationships be usable in queries, or only carried into the export? → A: Carried
  and shown: relationship tags are exported and rendered on the relationship in diagrams and pages;
  queries do not filter on them in this feature.
- Q: Which element should a trigger relationship be written on? → A: On the invoked element,
  targeting its trigger (`container "sync" { uses "consume" { target = container.sync_queue,
  kind = "trigger" } }`). One rule holds for every relationship: the declaring element depends on
  its target. Queries are unchanged; only the drawn arrow points from the trigger.
- Q: Should container and component views switch to top-down by default, or only on request? → A:
  New default. Container and component views render top-down unless a view sets
  `direction = "right"`; this one-time change to existing output is recorded in the changelog.
- Q: Which element kinds may carry a `shape`? → A: Containers and external systems. Systems and
  people may not; setting `shape` on them is a compile error.
- Q: When an element has a title, what does its diagram box show? → A: The title prominently, with
  the address name in smaller text beneath it, in diagrams and on site and markdown pages.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Diagrams people can read (Priority: P1)

An architect opens the container diagram of a system with about fifteen containers. They can read
every box and every label at normal size, and the diagram fits a screen or a printed page without
scrolling sideways.

**Why this priority**: A diagram that cannot be read cannot replace a hand-drawn one, and replacing
hand-drawn diagrams is the reason the tool exists. This is the single largest gap the testing
found.

**Independent Test**: Render the serverless-reference container view and compare it with the
system's hand-drawn container diagram. Text is legible at 100% zoom, and the aspect ratio is within a readable
range.

**Acceptance Scenarios**:

1. **Given** a system with 15 containers and 25 relationships, **When** its container view is
   rendered, **Then** the diagram's width-to-height ratio is at most 2:1 and no label is smaller
   than the body text of the generated site.
2. **Given** a view, **When** its author sets a layout direction (top-down or left-right), **Then**
   the diagram is laid out in that direction.
3. **Given** no direction is set, **When** a container or component view is rendered, **Then** it
   is laid out top-down; the landscape keeps its current direction.

---

### User Story 2 - Names people recognise (Priority: P1)

An architect gives each element a display title ("Read API Lambda") in addition to its address
name (`read_api`). Diagrams, site pages and markdown show the title; addresses, references and
queries keep using the name.

**Why this priority**: Every box in a rendered diagram today shows an identifier, which makes the
output look like a debugging dump rather than documentation.

**Independent Test**: Set titles on three elements and build. Their boxes, page headings and
markdown headings show the titles; `loko query` output and exports still show addresses.

**Acceptance Scenarios**:

1. **Given** an element with a title, **When** any diagram, site page or markdown document
   mentions it, **Then** the title is shown prominently, with the address name in smaller text
   beneath it.
2. **Given** an element without a title, **When** it is rendered, **Then** its name is shown, as
   today.
3. **Given** the MCP tools, **When** an assistant sets or clears a title, **Then** it is an
   ordinary attribute edit, with no new tool.

---

### User Story 3 - Shapes that say what a thing is (Priority: P2)

An architect marks containers, and external systems, as databases, queues, topics, functions or
buckets. Each is drawn in
a shape that conveys its role: a cylinder for a database, a queue shape for a queue. Readers no
longer need to read every label to find the data stores.

**Why this priority**: The hand-drawn reference diagram uses cylinders and queues, and readers rely
on them. Without shapes, a data store and a Lambda look identical.

**Independent Test**: Mark the serverless reference's DynamoDB table as a database and its two SQS queues as
queues, then render. Those three nodes have distinct shapes, and every other node is unchanged.

**Acceptance Scenarios**:

1. **Given** an element marked as a database, **When** any view including it is rendered, **Then**
   it is drawn as a cylinder.
2. **Given** an element with no marking, **When** it is rendered, **Then** it keeps its kind's
   default shape.
3. **Given** an unknown shape value, **When** the project is compiled, **Then** a diagnostic names
   the allowed values.

---

### User Story 4 - Connections that say how things talk (Priority: P2)

An architect marks a relationship as synchronous (the default), asynchronous, or a trigger. The
diagram draws asynchronous connections dashed. A trigger is drawn from the trigger to the thing it
invokes, which is the direction the hand-drawn diagrams use. Queries still treat the invoked thing
as depending on its trigger.

**Why this priority**: In both test architectures, triggers (a queue feeding a Lambda, a schedule
invoking a function) had to be written against the data flow so that dependency queries came out
right. The diagrams then pointed the wrong way. Choosing between correct queries and correct
diagrams should not be necessary.

**Independent Test**: Model `sync_queue → sync` as a trigger. The diagram draws the arrow from the
queue to the Lambda, and `query dependents container.sync_queue` lists `container.sync`.

**Acceptance Scenarios**:

1. **Given** a relationship marked async, **When** it is rendered, **Then** its line is dashed.
2. **Given** a relationship marked trigger, declared on the invoked element and targeting its
   trigger, **When** it is rendered, **Then** the arrow points from the trigger to the invoked
   element.
3. **Given** the same trigger relationship, **When** dependents and dependencies are queried,
   **Then** the results are those of an ordinary relationship: the invoked element depends on its
   trigger.
4. **Given** no kind, **When** a relationship is rendered, **Then** it is drawn exactly as today.

---

### Edge Cases

- **A title shared by two elements.** Titles need not be unique; addresses stay the identity.
  Diagrams show both titles, each with its address.
- **A very long title.** It wraps within its box; it is never truncated silently.
- **Kinds without a shape.** A person or a system cannot carry a `shape`; this is a compile
  error. An external system can, so a partner's database is drawn as one.
- **Existing projects.** A project using none of the new attributes builds byte-identically, except
  for the layout direction of container and component views (US1/AC3), which is a deliberate
  output change recorded in the changelog.
- **Shapes in views where a node is collapsed into its parent.** The parent's shape is used, not
  the child's.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every element MUST accept an optional display `title`. Diagrams, site pages and
  markdown MUST show the title prominently, with the address name in smaller text beneath it.
  Without a title, the name is shown as today.
- **FR-002**: Containers and external systems MUST accept an optional `shape` from a fixed set (at
  least: database, queue, topic, function, bucket). Each set member MUST render as a distinct
  shape. Invalid values, and `shape` on a system or person, MUST be compile errors that list the
  allowed values and kinds.
- **FR-003**: Relationships MUST accept an optional `kind`: `sync` (the default), `async` or
  `trigger`. Async relationships MUST render dashed. A trigger relationship is declared on the
  invoked element and targets its trigger, like any other relationship; it MUST render with the
  arrow pointing from the trigger to the invoked element. There is exactly one way to write a
  trigger.
- **FR-004**: Relationship kind MUST NOT change query results: dependents, dependencies, path,
  orphans and coupling MUST be identical with or without kinds.
- **FR-005**: Views MUST accept an optional layout `direction` (`down` or `right`). The default
  MUST be `down` for container and component views, and unchanged for the landscape and
  deployment views. Changing the default is a deliberate, one-time change to existing output and
  MUST be recorded in the changelog.
- **FR-006**: A container view of 15 containers and 25 relationships MUST render with a
  width-to-height ratio of at most 2:1.
- **FR-007**: All new attributes MUST be editable through `apply_edit` like existing ones, and
  returned by `describe` at `full`.
- **FR-008**: Rendering MUST stay byte-identical across runs and machines.
- **FR-009**: The new attributes MUST appear in the IR export as optional fields, omitted when
  unset, so existing exports are unchanged.
- **FR-010**: Relationships MAY carry free-form `tags`. Tags MUST be carried into the export and
  shown on the relationship in diagrams and site pages. Queries MUST NOT filter on them in this
  feature (a later feature may add filtering without breaking changes).

### Key Entities

- **Title**: an optional human name on any element; it never participates in references.
- **Shape**: a closed set of rendering roles for containers and external systems.
- **Relationship kind**: how a connection behaves (sync, async, trigger). It affects drawing only.
- **Layout direction**: a per-view rendering preference.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The serverless-reference container view, modelled with titles, shapes and relationship kinds,
  renders with every label legible at 100% zoom and an aspect ratio of at most 2:1.
- **SC-002**: Shown both diagrams side by side, a reviewer can identify the data stores, queues and
  async connections in the generated serverless-reference container diagram as quickly as in the hand-drawn
  one.
- **SC-003**: Projects that use none of the new attributes produce byte-identical exports, and
  byte-identical diagrams apart from the documented default direction change.
- **SC-004**: Query results on the serverless-reference model are identical before and after adding relationship
  kinds.

## Assumptions

- Layout stays on the in-process D2 engine with dagre (feature 014). ELK remains rejected unless
  dagre cannot meet FR-006, in which case the plan revisits it.
- The shape set maps onto shapes D2 already provides (cylinder, queue, …). No custom icon artwork
  is in scope; AWS service icons are a possible later feature.
- Dynamic (sequence) views, environment reuse and Terraform import are out of scope; each is its
  own feature. Import (finding G1) is the strongest candidate to follow this one.
- All additions are optional language attributes: v1.x additive changes under the compatibility
  commitments in `docs/language.md`.
