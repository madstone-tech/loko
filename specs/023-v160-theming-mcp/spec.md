# Feature Specification: 1.6.0 Theming API, MCP Edit Feedback, Listing, and Layout Defaults

**Feature Branch**: `023-v160-theming-mcp`

**Created**: 2026-11-09

**Status**: Draft

**Input**: User description: "1.6.0 additive batch from the third-party feedback report: a supported theming API (interface font, brand block, footer setting, theme asset files, font override, build theme report), smaller MCP edit feedback (diagnostic delta, formatting reported apart), a redundant-relationship diagnostic, listing of views and elements, expanded outside endpoints in deployment views, and a readable default layout. Commit the custom-branding example as the theming test site."

## Clarifications

### Session 2026-11-09

- Q: Which layout engine should draw a view that has no layout setting? → A: elk, for every view without a layout on the view or the project.
- Q: How should deployment views draw links to elements outside the view? → A: Expanded by default (one node per outside element), with an option to collapse.
- Q: Which MCP tools should return only the diagnostics their change caused, with the full list on request? → A: Every tool that changes source: `apply_edit`, `move` and `fmt`.
- Q: When should `redundant_relationship` fire for a container and one of its own components? → A: Same target only: the component uses exactly the element that the container uses.
- Q: Should the new listing commands also list relationships, as well as views and elements? → A: Yes: views, elements and relationships.

## User Scenarios & Testing *(mandatory)*

1.6.0 is a minor release. Every change is an addition to the 1.x surface or a
documented change of a default. Nothing that 1.5.x accepts becomes an error.
Theme files written for 1.5 keep building.

### User Story 1 - Rebrand the site without copying built-in templates (Priority: P1)

A team rebrands the generated site for each project it delivers: its own logo,
interface font and footer text. In 1.5 the team must copy the whole `nav` block
to change the logo, redefine the code font to change the interface font, and
edit a template to change the footer. Each copy goes out of date when loko
changes its built-in templates. The build also does not say that it applied a
theme.

**Why this priority**: Theming is the largest group in the feedback report
(items 7, 11, 12, 13). A supported theming surface makes loko usable for teams
that publish sites for others.

**Independent Test**: Make a theme that sets only the logo block, the interface
font variable and a project footer. Build it. Check that the logo, font and
footer change, that search, tree, views list and sidebar footer still come from
the built-in templates, that code blocks keep the code font, and that the build
output names the theme files and blocks that it applied.

**Acceptance Scenarios**:

1. **Given** a theme that overrides only the brand block, **When** the user
   builds, **Then** the sidebar shows the theme's logo, and every other part of
   the sidebar is the built-in markup.
2. **Given** a theme that sets the interface font variable, **When** the user
   builds, **Then** navigation, labels, cards and other interface text use that
   font, and code and HCL snippets keep the code font.
3. **Given** a project that sets a footer text, **When** the user builds with
   the built-in theme or with a theme that does not override the footer block,
   **Then** every page shows that footer text.
4. **Given** a project with a `templates/` directory, **When** the user builds,
   **Then** the build output has one line that names the theme directory, each
   top-level theme file, the blocks each template file overrides, and the number
   of theme assets.
5. **Given** a project with no `templates/` directory, **When** the user
   builds, **Then** the build output has no theme line.
6. **Given** a theme written for 1.5 that overrides the whole `nav` block,
   **When** the user builds with 1.6, **Then** the build succeeds and renders
   that theme's nav as before.

---

### User Story 2 - A theme ships its own files and fonts (Priority: P1)

A theme needs its own logo image and web fonts. In 1.5 `templates/` accepts only
CSS, script and template files, so the logo becomes a large inline data string
and fonts cannot ship at all. The built-in font file still downloads four font
families that the theme does not use.

**Why this priority**: Without files, the theming surface of story 1 still
forces workarounds (items 8, 9). The two stories together make theming complete.

**Independent Test**: Put an image and a font file under the theme's assets
directory, and a replacement font stylesheet in the theme. Build. Check that the
files are in the output, that the theme's CSS can load them by relative URL, and
that the site no longer loads the built-in font families.

**Acceptance Scenarios**:

1. **Given** files under `templates/assets/`, **When** the user builds,
   **Then** each file is in the output under a theme assets path, byte for byte,
   and a stylesheet in the theme can refer to it with a relative URL.
2. **Given** a theme asset whose path is unsafe, is a symbolic link, or differs
   from another asset path only in letter case, **When** the user builds,
   **Then** the build fails with `theme_invalid` and names the file. Nothing is
   written. An asset can never replace a generated file, because assets are
   published under their own theme path.
3. **Given** a theme that provides `fonts.css`, **When** the user builds,
   **Then** the site loads the theme's `fonts.css` and not the built-in one, and
   the built-in font files are not written to the output.
4. **Given** a theme without `fonts.css`, **When** the user builds, **Then** the
   built-in fonts load as in 1.5.
5. **Given** a theme asset is removed, **When** the user rebuilds, **Then** the
   build removes that file from the output (ownership pruning covers theme
   files).

---

### User Story 3 - An assistant gets only the feedback its edit caused (Priority: P2)

An AI assistant edits the model through the MCP edit tool. In 1.5 every edit
preview returns every diagnostic in the model, so an edit that sets one
description repeats the whole warning list, and a larger model costs much more
context on every call. The preview diff also shows lines that only changed
alignment, so a one-attribute edit looks like a three-line change.

**Why this priority**: The MCP surface is a main consumer (constitution
principle VI). Repeated diagnostics are the largest avoidable token cost in an
editing session (item 3).

**Independent Test**: On a model with several existing warnings, preview an
edit that sets one description. Check that the result lists no unrelated
diagnostics, gives the counts of the unchanged ones, and separates the semantic
change from alignment-only changes. Then ask for the full list and check that
the 1.5 shape comes back.

**Acceptance Scenarios**:

1. **Given** a model with existing diagnostics, **When** an assistant previews
   or applies an edit, **Then** the result lists only the diagnostics that the
   edit added or removed, plus a count by severity of the diagnostics that did
   not change.
2. **Given** the same edit, **When** the assistant asks for all diagnostics,
   **Then** the result lists every diagnostic as in 1.5.
3. **Given** a `move` or `fmt` call on the same model, **When** the assistant
   previews or applies it, **Then** the result uses the same delta and the same
   full-list option.
4. **Given** an edit that changes one attribute in a block whose other
   attributes are realigned by canonical formatting, **When** the assistant
   previews it, **Then** the result separates lines that changed in meaning from
   lines that changed only in whitespace.
5. **Given** an edit that the compiler refuses, **When** the assistant previews
   it, **Then** the result lists every error that blocks the edit (refusals keep
   their 1.5 content).

---

### User Story 4 - Diagrams are readable without tuning the layout (Priority: P2)

An author builds a model of about ten containers and forty relationships with no
`layout` setting. The default layout engine draws the views very wide with
unreadable labels. Setting `layout = "elk"` fixes most of this, but a new author
does not know the setting exists.

**Why this priority**: Unreadable first diagrams are the first impression of the
tool (item 1). The fix changes rendered output, so it needs an explicit decision
record.

**Independent Test**: Build a model with no layout setting and a view above the
readability limit. Check which engine draws it and that a view or project that
sets `layout` keeps its choice.

**Acceptance Scenarios**:

1. **Given** a view that neither the view nor the project gives a layout,
   **When** the user builds, **Then** the view is drawn with elk.
2. **Given** a project or view that sets `layout`, **When** the user builds,
   **Then** the set engine draws it, as in 1.5.
3. **Given** a project that wants the 1.5 drawing, **When** it sets
   `layout = "dagre"`, **Then** every view draws as in 1.5.

---

### User Story 5 - The author learns about a redundant relationship (Priority: P2)

A container and one of its own components both declare a relationship to the
same target. The system view merges them into one edge labelled
"2 relationships" and drops both descriptions. Nothing tells the author that one
of the two declarations is redundant.

**Why this priority**: Silent loss of descriptions makes diagrams less useful
and the cause is easy to fix once reported (item 4).

**Independent Test**: Declare the same target from a container and from one of
its components. Validate. Check that one diagnostic points at the container's
relationship and names the component's one.

**Acceptance Scenarios**:

1. **Given** a container and its own component that both use the same target,
   **When** the user validates, **Then** an information diagnostic
   `redundant_relationship` points at the container's relationship and names the
   component relationship that already covers it.
2. **Given** the diagnostic, **When** the user runs validate with `--strict`,
   **Then** it stays information and does not change the exit code. Only
   `missing_docs` escalates under `--strict` (ADR-0017).
3. **Given** the diagnostic, **When** the project lists it in
   `ignore_warnings` or the user passes it to `--ignore`, **Then** it is not
   reported.
4. **Given** a container and a component of a different container that use the
   same target, **When** the user validates, **Then** no `redundant_relationship`
   is reported.

---

### User Story 6 - List the views, elements and relationships (Priority: P3)

An author wants to confirm that a view exists, that a removed element is gone,
or that a relationship was removed. In 1.5 this needs a full export and a
script.

**Why this priority**: Small and self-contained (item 18). It saves a script for
a common check.

**Independent Test**: Run the listing commands on an example project and compare
the output with the views, elements and relationships in the source.

**Acceptance Scenarios**:

1. **Given** a project, **When** the user lists views, **Then** the output has
   one row per view: its name, kind (generated or declared) and title, sorted by
   name.
2. **Given** a project, **When** the user lists elements, **Then** the output has
   one row per element: its address, kind and title, sorted by address.
3. **Given** a kind filter, **When** the user lists elements of one kind,
   **Then** only those elements are listed.
4. **Given** a project, **When** the user lists relationships, **Then** the
   output has one row per relationship: its address, source, target, kind and
   description, sorted by address.
5. **Given** the machine-readable format option, **When** the user lists views,
   elements or relationships, **Then** the output is JSON, and the MCP query tool
   offers the same listing.

---

### User Story 7 - Deployment views show each outside endpoint (Priority: P3)

A deployed container talks to a person and two external systems. The deployment
view draws all three links to one dashed "outside this view" node labelled
"2 relationships", so the reader cannot see which element each link reaches.

**Why this priority**: A readability fix for one view kind (item 6).

**Independent Test**: Build a deployment view whose instances link to three
different elements outside the view. Count the outside nodes and check their
labels.

**Acceptance Scenarios**:

1. **Given** a deployment view with links to three outside elements, **When**
   the user builds with outside endpoints expanded, **Then** the view has one
   outside node per element, each labelled with that element's title.
2. **Given** the same view with outside endpoints collapsed, **When** the user
   builds, **Then** the view draws one outside node as in 1.5.
3. **Given** a project that does not choose, **When** the user builds,
   **Then** deployment views draw outside endpoints expanded.

---

### Edge Cases

- A theme overrides both the new brand block and the whole `nav` block. The
  `nav` override wins, because it replaces the block that calls the brand block.
- A project footer contains characters that are special in HTML. The footer is
  escaped and shown as plain text.
- A theme asset has a name that is unsafe as an output path (absolute path,
  `..`, control characters). The build fails with `theme_invalid`.
- A theme asset directory holds a very large file. The build copies it. The
  build does not limit file sizes.
- A theme provides `fonts.css` and also expects the built-in fonts. The theme
  must load them itself. The built-in font files are not written.
- An edit changes no diagnostic at all. The result has an empty delta and the
  counts only.
- An edit removes the element that a diagnostic was about. The diagnostic is in
  the removed half of the delta.
- A component's relationship and its container's relationship to the same
  target have different kinds (sync and async). They are still redundant. The
  diagnostic names both kinds.
- A component uses a component inside its container's target, while the
  container uses the target container itself. The targets differ, so no
  `redundant_relationship` is reported.
- A container view with expanded outside endpoints. The option applies only to
  deployment views in this release.
- A view list on a project with no declared views. The list holds the generated
  views only.

## Requirements *(mandatory)*

### Functional Requirements

**Theming (stories 1 and 2)**

- **FR-001**: The built-in stylesheet MUST provide an interface font variable
  that styles navigation, labels, cards and other interface text. Code and HCL
  snippets MUST keep the code font variable.
- **FR-002**: The default value of the interface font variable MUST keep the 1.5
  look of the built-in theme. Sites built with no theme MUST render as in 1.5.
- **FR-003**: The built-in layout MUST render the sidebar logo through its own
  overridable block. The navigation block MUST call that block.
- **FR-004**: The language MUST accept an optional footer text on the project.
  The built-in footer block MUST show it as plain text when it is set, and the
  1.5 footer when it is not set.
- **FR-005**: When a project has a `templates/` directory, the build output MUST
  include one line that names it, each top-level theme file, the blocks each
  template file overrides, and the number of theme assets. Blocks are listed
  when the `html` format is built. The CLI and the MCP build tool MUST report
  the same content.
- **FR-006**: Files under `templates/assets/` MUST be copied to the output under
  a theme assets path, byte for byte, with their relative paths kept.
- **FR-007**: A theme asset whose path is unsafe, is a symbolic link, or differs
  from another asset path only in letter case MUST fail the build with
  `theme_invalid` that names the file, and nothing MUST be written. The theme
  assets path MUST keep every asset apart from generated files.
- **FR-008**: A `fonts.css` in the theme MUST replace the built-in font
  stylesheet, and the built-in font files MUST NOT be written.
- **FR-009**: Theme files written to the output MUST be owned like generated
  files: listed in the manifest and pruned when removed.
- **FR-010**: The theming guide MUST document the new variable, block, footer
  setting, assets directory and font override, and MUST add them to the stable
  theming contract.
- **FR-011**: The repository MUST include a branded example project that uses
  the brand block, the interface font variable, the footer setting, a theme
  asset and a font override, and that builds and validates clean.

**MCP edit feedback (story 3)**

- **FR-012**: Every MCP tool that changes source (`apply_edit`, `move`, `fmt`)
  MUST return, by default, only the diagnostics that its change added or
  removed, plus a count by severity of unchanged diagnostics. This applies to
  previews and to applied changes.
- **FR-013**: The same tools MUST accept an option that returns every
  diagnostic in the 1.5 shape.
- **FR-014**: `apply_edit` and `move` MUST separate diff lines that changed in meaning
  from lines that changed only in whitespace from canonical formatting. Applied
  files MUST stay in canonical form.
- **FR-015**: A refused edit MUST keep its 1.5 content: every blocking error.
- **FR-016**: The change of default in FR-012 MUST be recorded in a decision
  record and the changelog (constitution principle VI).

**Layout default (story 4)**

- **FR-017**: A view with no layout from the view or the project MUST be drawn
  with elk.
- **FR-018**: A layout set on the view or the project MUST win over the default,
  and `layout = "dagre"` MUST give the 1.5 drawing.
- **FR-019**: The changes of default rendering (FR-017 elk, FR-025 expanded
  outside endpoints) MUST be recorded in a decision record and the changelog.
- **FR-020**: The wide-view hint MUST keep giving correct advice under the new
  default.

**Redundant relationships (story 5)**

- **FR-021**: Validation MUST report an information diagnostic
  `redundant_relationship` when a container and one of its own components use
  exactly the same target element. It MUST point at the container's
  relationship and name the component's relationship. A component that uses a
  child of the container's target (for example a component inside it) MUST NOT
  trigger the diagnostic.
- **FR-022**: The diagnostic MUST follow the existing severity rules (information
  never changes an exit code, `--strict` included) and ignore rules
  (`ignore_warnings`, `--ignore`).

**Listing (story 6)**

- **FR-023**: The CLI MUST list views (name, kind, title), elements (address,
  kind, title) and relationships (address, source, target, kind, description),
  each sorted, with an element-kind filter for elements.
- **FR-024**: Listing MUST support text and JSON output, and the MCP query tool
  MUST offer the same listing.

**Outside endpoints (story 7)**

- **FR-025**: Deployment views MUST draw one outside node per outside element,
  labelled with the element's title, by default. An option MUST draw one
  collapsed outside node as in 1.5.
- **FR-026**: The choice MUST be settable on the project and on each
  `deployment` block. The `deployment` value wins over the project value.
  Declared views are never deployment views, so the `view` block does not take
  the setting.

**Release**

- **FR-027**: Every addition to the language MUST be documented in
  `docs/language.md` and the MCP schema output.
- **FR-028**: The CHANGELOG MUST gain a `## [1.6.0]` section with every change,
  each citing its feedback item.

### Key Entities *(include if feature involves data)*

- **Theme**: The override files under `templates/`: template files, stylesheets,
  a script, an optional font stylesheet and an optional assets directory. Applied
  completely or not at all.
- **Theme report**: What the build says it applied: the theme directory, each
  file, and the blocks each template file overrides.
- **Project footer**: An optional plain-text attribute on the project, shown by
  the built-in footer block.
- **Diagnostic delta**: For one source change (`apply_edit`, `move` or `fmt`),
  the diagnostics added, the diagnostics removed, and counts by severity of the
  rest.
- **Outside endpoint mode**: Collapsed (one outside node) or expanded (one node
  per outside element), for deployment views.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The branded example changes its logo, interface font and footer
  with zero copied built-in blocks other than the brand block.
- **SC-002**: The branded example's stylesheet holds no embedded image data; its
  logo and fonts are separate files in the output.
- **SC-003**: A site built with a theme font override loads zero built-in font
  files.
- **SC-004**: On a model with at least 10 existing diagnostics, an edit preview
  that sets one description returns at most the diagnostics that the edit
  changed, and its encoded size is at most half the 1.5 size.
- **SC-005**: A one-attribute edit preview reports exactly one line that changed
  in meaning.
- **SC-006**: On `testdata/projects/first-run` (12 containers, 22
  relationships, no layout set), every view renders at a page-width text size of
  at least 6 px.
- **SC-007**: A site built from a 1.5 project with no theme, an explicit
  `layout` and collapsed outside endpoints renders byte-identical pages to 1.5,
  except where a change in this release adds content.
- **SC-008**: Every feedback item in scope (1, 3, 4, 6, 7, 8, 9, 11, 12, 13,
  18) has a passing acceptance scenario and a CHANGELOG entry.

## Assumptions

- Scope: feedback items 1, 3, 4, 6, 7, 8, 9, 11, 12, 13 and 18. Item 10
  (diagram theming) needs its own decision record and is out of scope. Item 2
  (trigger value rename or warning) and item 5 (stale MCP server) are out of
  scope. Semantic diff is a later release.
- Theme assets land under a dedicated theme path inside the output assets
  directory, so they can never overwrite a generated file by accident. The exact
  path is a plan decision.
- Item 13's second proposal (warn on an unknown block) already exists as
  `theme_invalid` and is not new work.
- Formatting stays canonical. The edit tool reports alignment-only lines apart,
  rather than leaving files unformatted.
- The listing commands extend the existing `query` family rather than adding a
  new top-level command, so the command surface stays the same. The plan may
  revisit this.
- The diagnostic delta matches diagnostics before and after a change by code,
  severity, address, file and detail. Line and column positions are not part of
  the match, because an edit that adds a line moves every diagnostic below it.
- The branded example is the fictional "Example Co." theme, kept generic.
- Release process: PR into `main`, then tag `v1.6.0`, as for 1.5.1.
