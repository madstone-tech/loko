# Feature Specification: Language Polish

**Feature Branch**: `018-language-polish`

**Created**: 2026-10-26

**Status**: Draft

**Input**: The items of the first modelling session report
([017 first-run report](../017-mcp-validation-ergonomics/first-run-report.md)) that 017 deferred:

- #6: deployment `node` blocks take no attributes, so a cloud layout (account → region → VPC →
  functions and tables) can be shown only by node names; `instance` cannot say what runs in one
  deployment when it differs from the container;
- #7: a `view` cannot say what it shows or why; the rendered page shows only its identifier;
- #4: `docs` paths resolve from the project root, not from the file that declares the element, so
  moving a source file together with its docs folder breaks every link;
- #10: a container view of about ten containers rendered very wide and short, with unreadable
  labels, and nothing told the author which setting would fix it.

Dynamic views (#5) are feature 019.

## Clarifications

### Session 2026-10-26

- Q: How do `docs` paths resolve? → A: From the file that declares the element first. When nothing
  is there but the same path exists from the project root (the 1.1.0 rule), loko uses that file and
  gives an information hint with the file-relative path to write instead. Existing models keep
  working, and a source file moved together with its docs folder keeps its links.

### Session 2026-10-27

- Q: Where do a deployment node's `description` and `tags` appear? → A: The diagram labels the
  node with its title and `[technology]` only; the description and tags appear in a "Nodes" table
  on that deployment's site page, and in the export.
- Q: Is a view's `title` drawn in the exported D2 and SVG? → A: Yes, as a heading at the top of the
  canvas, so an SVG embedded on its own says what it shows. The description stays on the site page
  and in the export. Views without a title are unchanged.
- Q: Does a project-level `layout` silence the wide-view hint? → A: Yes. Any explicit `direction` or
  `layout` on the view, or `layout` on the project, means the author chose; no hint.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Describe deployment infrastructure (Priority: P1)

An architect models a serverless system deployed as AWS account → region → VPC → Lambda functions
and a DynamoDB table. Each deployment node says what it is (a title, the technology, a short
description, tags), and the deployment diagram draws that information on each node, the way
element boxes already show their title and technology. Where one deployment runs a container on
different technology than its declaration (a local process in `dev`, Lambda in `prod`), the
instance says so, and that deployment's diagram shows the instance's technology.

**Why this priority**: deployment diagrams are the view that most needs to say what each box is,
and today they cannot. It was the largest gap in the report after the 017 items.

**Independent Test**: add the four attributes to nodes of a fixture and `technology` to one
instance; build; the deployment diagram shows them, and every other diagram is byte-identical.

**Acceptance Scenarios**:

1. **Given** a node with `title`, `technology`, `description` and `tags`, **When** the project is
   built, **Then** the deployment diagram labels the node with its title and `[technology]`, and
   the deployment's site page lists the node with its description and tags in a "Nodes" table.
2. **Given** a node with no new attributes, **When** the project is built, **Then** its diagram is
   unchanged from 1.1.0.
3. **Given** an instance with `technology = "local process"` of a container whose technology is
   "Lambda", **When** the `dev` deployment diagram is built, **Then** that instance shows
   "local process", and the `prod` instance of the same container shows "Lambda".
4. **Given** an assistant using MCP, **When** it asks the `schema` tool about `node` and
   `instance`, **Then** the new attributes are listed, and `apply_edit` can set and remove each one
   on a group and an instance.

---

### User Story 2 - Views that say what they show (Priority: P1)

An architect declares `view "phase_1"` with a title ("Phase 1: ingestion only") and a description
of what the view leaves out and why. The rendered diagram page shows the title as its heading and
the description as a caption; the site navigation lists the view by its title.

**Why this priority**: views are what readers open first, and an identifier such as `phase_1`
tells them nothing. Small change, high reader value.

**Independent Test**: give one view a title and a description, build the site, and check the
view's page heading, caption and navigation entry; views without them render as before.

**Acceptance Scenarios**:

1. **Given** a view with a `title`, **When** the site is built, **Then** the page heading and the
   navigation entry show the title, and the view's name still identifies it in addresses, file
   names and links.
2. **Given** a view with a `description`, **When** the site is built, **Then** the description is
   shown under the heading.
3. **Given** a view with a `title`, **When** its SVG is opened on its own, **Then** the title is
   drawn as a heading at the top of the diagram, and the description is not drawn.
4. **Given** a view with neither, **When** the site is built, **Then** the output is unchanged
   from 1.1.0.
5. **Given** an empty `title`, **When** the model is compiled or an edit sets it, **Then** it is
   refused, as for an element's empty title.

---

### User Story 3 - Docs links that survive moving files (Priority: P2)

An architect keeps each subsystem's source file next to its documentation:
`architecture/payments/payments.loko.hcl` and `architecture/payments/docs/overview.md`. The element
declares `docs = "docs/overview.md"`. Whatever the rule is, the author knows it before writing the
path (from the language reference and the `schema` tool), and when a path does not resolve, the
warning names the exact location loko looked at.

**Why this priority**: the report found the rule by trial and error. The resolution change, if
any, has to keep existing models working, which makes it the riskiest item here.

**Independent Test**: a fixture with source files in subdirectories and docs next to them; the
documented rule resolves every path, and a wrong path produces a warning naming the location
checked.

**Acceptance Scenarios**:

1. **Given** a `docs` path that does not resolve, **When** the model is validated, **Then** the
   `docs_not_found` warning names both locations loko checked (from the declaring file and from
   the project root), relative to the project root.
2. **Given** `architecture/payments/payments.loko.hcl` declaring `docs = "docs/overview.md"` and
   the file `architecture/payments/docs/overview.md`, **When** the model is built, **Then** the
   element's page shows that document, with no diagnostic.
3. **Given** a 1.1.0 model whose file in `architecture/` declares
   `docs = "architecture/docs/people.md"`, **When** it is validated, **Then** the document still
   resolves, and an information hint suggests `docs/people.md`; `--strict` still exits `0`.
4. **Given** an existing 1.1.0 model whose `docs` paths resolve today, **When** it is validated
   and built with 1.2, **Then** every path still resolves, and the site links the same documents,
   unless a different file exists at the file-relative location, which is then linked and reported
   by a `docs_ambiguous` warning naming both.
5. **Given** the `schema` tool and the language reference, **When** an author reads the `docs`
   attribute, **Then** the resolution rule is stated with an example from a subdirectory file.

---

### User Story 4 - Hint when a view renders too wide (Priority: P3)

An architect builds a container view of ten containers and six actors, with no `direction` or
`layout` set. The diagram comes out more than twice as wide as tall and several thousand pixels wide. The build reports an
information-level hint for that view, naming the measured width-to-height ratio and the settings
to try (`direction` and `layout = "elk"`). The build otherwise succeeds unchanged, and loko does
not change the layout on its own.

**Why this priority**: the author can already fix a wide view; they just did not know which
setting to try. Useful, but the smallest gap.

**Independent Test**: build a fixture whose view is known to render wide; one hint names that
view; a narrow view in the same project gets none; exit codes, including under `--strict`, are
unchanged.

**Acceptance Scenarios**:

1. **Given** a view whose rendered diagram is wider than the limit, **When** it is built, **Then**
   one information-level hint names the view, the ratio and the settings to try.
2. **Given** the same project built with `--strict`, **When** the build finishes, **Then** the
   exit code is the same as without the hint: hints never count as warnings.
3. **Given** a view that sets `direction` or `layout`, or a project that sets `layout`, **When**
   the view renders wide, **Then** no hint is given: the author already chose.
4. **Given** the hint's code in `ignore_warnings` or an ignore option, **When** the project is
   built, **Then** the hint is not reported.

---

### Edge Cases

- A node `title` that is empty is refused, as for elements and views.
- Node `tags` follow the same rules as element tags (strings, duplicates reported as for elements).
- Nested nodes each carry their own attributes; a child does not inherit its parent's.
- An instance's `technology` with no container technology to override is allowed: it states the
  technology for that deployment.
- A view's `title` may repeat another view's title; names stay the unique identifiers.
- Views generated by loko (not declared) have no title or description, and render as before.
- `docs` paths that leave the project root stay refused as today, under both resolutions; a
  file-relative path may use `../` as long as it stays inside the project.
- An element declared in a file at the project root resolves the same under both rules, so no hint
  is given.
- The `docs_not_found` warning names both locations checked, the file-relative one first (the one
  the author should write).
- The width hint is measured on the laid-out diagram, so `validate` (which does not render) and a
  build without SVG output do not report it; `build` and the MCP `build` tool do. `serve` shows
  errors only (017), so it does not either.
- A generated view (landscape, system, container, deployment) has no block to set `direction` on;
  its hint suggests `layout = "elk"` on the project, or declaring a view of the same elements.
- A view that renders empty or with one element gets no width hint.
- The new attributes are formatted by `fmt` in the same canonical order as other attributes.

## Requirements *(mandatory)*

### Functional Requirements

**Deployment nodes and instances**

- **FR-001**: A deployment `node` MUST accept optional `title`, `technology`, `description` and
  `tags`, with the same value rules as the attributes of the same names on elements.
- **FR-002**: The deployment diagram MUST label a node with its title (its name when no title) and
  its technology when set, as `[technology]`, and nothing else; the description MUST NOT be drawn on
  the diagram.
- **FR-002a**: Each deployment's site page MUST list its nodes that have a title, technology,
  description or tags in a "Nodes" table (name, title, technology, description, tags), nested nodes
  shown by their path. A deployment with no such node MUST render as in 1.1.0.
- **FR-003**: An `instance` MUST accept an optional `technology`; on that deployment's diagram, the
  instance MUST show it in place of the element's technology.
- **FR-004**: The compiled model export MUST include the new node and instance attributes, and
  omit them when unset, so 1.1.0 exports of models that do not use them are unchanged.

**Views**

- **FR-005**: A declared `view` MUST accept optional `title` and `description`. An empty title MUST
  be refused.
- **FR-006**: The rendered site MUST use a view's title as its page heading and navigation label,
  and its description as a caption under the heading. The view's name MUST remain its identifier in
  addresses, file names and links.
- **FR-007**: A view's diagram files (D2, SVG) MUST draw its title as a heading at the top of the
  canvas. The description MUST NOT be drawn on the canvas. Diagrams of views without a title MUST
  be byte-identical to 1.1.0.

**Docs paths**

- **FR-008**: A `docs` path MUST resolve from the directory of the file that declares the element.
  When no file exists there but the path exists from the project root, loko MUST use the
  root-relative file and report an information diagnostic naming the file-relative path to write
  instead.
- **FR-009**: Every `docs` path that resolves in 1.1.0 MUST still resolve, with no error. It MUST
  get no new warning unless the file-relative and the root-relative paths both exist and are
  different files: then the file-relative one MUST win and loko MUST report a `docs_ambiguous`
  warning naming both, since the linked document changes.
- **FR-010**: `docs_not_found` MUST name both locations loko checked (from the declaring file, then
  from the project root), relative to the project root.
- **FR-011**: The rule MUST be stated in the language reference, in the `schema` tool's answer and
  in the `docs` attribute's description.

**Wide-view hint**

- **FR-012**: The tool MUST support an information severity below warning. Information diagnostics
  MUST NOT change any exit code, including under `--strict`, and MUST NOT count as warnings in
  summaries or totals.
- **FR-013**: When a rendered view is at least 3,000 px wide and at least 2.5 times wider than
  tall, and neither the view sets `direction` or `layout` nor the project sets `layout`, the build
  MUST report one information diagnostic for it, naming the view, its measured size and ratio, and
  the settings to try.
- **FR-014**: The hint MUST be suppressible like a warning: by `ignore_warnings` and by the ignore
  option of `build` (CLI `--ignore`, MCP `ignore`). Information diagnostics MUST also be hidden by
  the severity option where a command has one (`validate --severity warning`, MCP `validate`
  `severity: "warning"`); `build` gains no severity option.
- **FR-015**: The tool MUST NOT change a view's direction or layout automatically.

**Language table and tools**

- **FR-016**: Every new attribute MUST come from the single language table, so the decoder, the
  edit schema, the `schema` tool, `fmt` and `apply_edit` agree, and the existing parity tests cover
  them.
- **FR-017**: `apply_edit` MUST be able to set and remove each new attribute on its target (group,
  instance, view), with the same preview, revision and refusal behaviour as other attributes.
- **FR-018**: Models that use none of the new attributes MUST compile, export and render
  byte-identically to 1.1.0.

### Key Entities

- **Deployment node**: a place where instances run (account, region, network). Gains title,
  technology, description and tags.
- **Instance**: an element deployed in an environment. Gains an optional technology that overrides
  the element's for that deployment.
- **View**: a named selection of elements. Gains title and description, used for reading, never as
  identifiers.
- **Docs path**: a path from an element to its prose, resolved from the declaring file, with the
  project root as a fallback that comes with a hint.
- **Information diagnostic**: a finding below warning severity. It never fails a run; it advises.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The cloud layout from the report (account → region → VPC → functions, table, bucket)
  can be modelled so that every node on its deployment diagram shows what it is, with no
  information carried only in node names.
- **SC-002**: Every declared view with a title is listed by that title in the site navigation and
  shown with it as the page heading.
- **SC-003**: All `docs` paths in the repository's examples and test projects resolve after the
  change exactly as before (0 new `docs_not_found` warnings), and a misplaced path's warning names
  the location checked in 100% of cases.
- **SC-004**: The report's wide container view gets exactly one hint, and every view in the
  repository's examples that gets one is at least 3,000 px wide when rendered (unreadable when shown
  at page width); views where the author chose `direction` or `layout` get none.
- **SC-005**: Exit codes of `validate` and `build`, with and without `--strict`, are unchanged on
  every example and test project.
- **SC-006**: Projects that use none of the new attributes produce byte-identical exports and
  diagrams to 1.1.0 (the determinism job and the export goldens pass unchanged).

## Assumptions

- Node and view `title` and `description` are plain strings, like element `title` and
  `description`; markdown in descriptions follows the same rules as element descriptions.
- The width limit is fixed, chosen during planning from measurements of the example views
  (research R10: at least 3,000 px wide and 2.5 times wider than tall); it is not configurable in
  this release.
- The width hint is computed from the layout the renderer already produces; no extra layout run.
- Information diagnostics are shown by the CLI and MCP like warnings, with their own severity
  label, and are included in `--summary` under their own heading.
- The new attributes are additive and need no ADR; the information severity and any change to
  `docs` resolution are recorded in an ADR, since they change diagnostic and resolution behaviour.
- This release is 1.2.0 (new language attributes).
