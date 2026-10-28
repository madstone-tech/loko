# Feature Specification: Site Redesign

**Feature Branch**: `021-site-redesign`

**Created**: 2026-10-28

**Status**: Draft

**Input**: User description: seed prompt `docs/superpowers/specs/2026-10-28-021-site-redesign-seed.md`. Rework the generated HTML site (`loko build --format html`, `loko serve`) so a reader can tell, without the diagram, where an element sits, what it relates to and what it does. Restyle the site with Madstone branding. The project owner approved these decisions with mockups in brainstorming on 2026-10-27. The feature depends on 020 (diagram look): it reuses 020's icon set, kind palette and SVG selectors.

## Context

The landscape and system pages work. The container and component pages give no context outside the diagram:

- The sidebar lists only views, so the reader cannot browse the hierarchy (system → container → component).
- An element page shows no picture of its neighbourhood.
- "Part of" goes up one level only, and children are not listed.
- The Uses and Used by tables are flat: they show no kind, and they repeat a neighbour once per relationship.
- The sidebar and index show raw ids (`api_gateway`), not names.
- In the dark theme, dark-blue links on navy are hard to read.
- Diagrams are scaled to page width, so the labels of a wide view shrink to about 6px.

The owner's goal for an element page, in their words: "how it relates to its parent and siblings, and its function or any prose worth mentioning."

## Clarifications

### Session 2026-10-28

- Q: When a component's `docs` points at the same file its parent uses, what should the component's page show? → A: A link to the parent's prose only. The component counts as having no prose, so its In context diagram fills the main column.
- Q: How much of the Architecture tree should each page include in its HTML? → A: A pruned tree: the top level, the children of each ancestor of the current element, and the current element's own children. Collapsed branches link to their element's page.
- Q: How strict should compatibility be for users who override built-in template blocks? → A: Every block name in the 014 table keeps resolving and new blocks are added. Where a block's meaning changes (for example `relation-table` now renders Talks to rows), the change goes in the changelog and an ADR. Not a breaking change.
- Q: How much weight may the bundled fonts add to the site? → A: Latin-subset WOFF2, only the weights used, a hard cap of 500 KB for all fonts together; system fonts fall back for characters outside the subset.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Understand an element without reading the diagram (Priority: P1)

A reader opens a container or component page. Without studying a diagram, they learn what the element does, what it is part of, which siblings it has, and what it talks to.

**Why this priority**: This is the owner's stated goal and the main gap in the current site. Element pages are most of the site's pages.

**Independent Test**: Build the `examples/microservices` project and open a component page. Check that the page shows the parent, the siblings with their relation to this component, one row per neighbour, and the "In context" diagram with this element outlined.

**Acceptance Scenarios**:

1. **Given** a component inside a container inside a system, **When** the reader opens the component's page, **Then** the header shows a kind label with the parent (`COMPONENT · IN GATEWAY`), the title with the element's icon tile in its kind colour, a metadata line (technology, full id, source file and line), and the description as a lead.
2. **Given** the same page, **When** the reader looks at the right rail, **Then** it shows, in this order: Part of (the parent with kind, technology and description, and the system above it), Siblings (the parent's other children, the current one marked, each with "calls" or "called by" when they are related), Talks to (one row per neighbour with a direction arrow, technology chip, description and a "sibling" marker where it applies), and In context.
3. **Given** an element with three relationships to the same neighbour, **When** the reader views Talks to, **Then** that neighbour appears in one row, not three.
4. **Given** an element with prose in its `docs`, **When** the reader opens its page, **Then** the main column shows that prose and a note naming the source file to edit.
5. **Given** an element with no prose, **When** the reader opens its page, **Then** the "In context" diagram moves into the main column, so the page is never empty.
6. **Given** the "In context" diagram, **When** it is shown, **Then** it is the parent's existing diagram with every other node faded and this element outlined. No new diagram is rendered for the element.
7. **Given** a system or container page, **When** the reader opens it, **Then** the page also lists its children as rows with name, kind, technology and description.

---

### User Story 2 - Browse the architecture hierarchy (Priority: P1)

A reader moves through the model as a tree (people, systems, containers, components), always knows where they are, and steps between siblings.

**Why this priority**: Without a tree, a reader cannot reach a component page except through a diagram link. Navigation applies to every page.

**Independent Test**: Open any component page. Check that the sidebar tree shows the current element highlighted with its branch expanded, the breadcrumb leads up to the project, and the arrow keys move to the previous and next sibling.

**Acceptance Scenarios**:

1. **Given** any page, **When** it loads, **Then** the sidebar shows the project name and description, a search box, an Architecture tree with 020's icons, a Views list below the tree, and a footer saying the pages are generated from `*.loko.hcl` and must be edited at the source.
2. **Given** the sidebar and any list on the site, **When** an element is shown, **Then** it shows its name, not its id.
3. **Given** a component page, **When** it loads, **Then** the tree highlights the current element and expands its branch, and a breadcrumb runs up to the project (`order-management › API Gateway › Gateway › Router`).
4. **Given** an element with siblings, **When** the reader presses → or ←, **Then** the site goes to the next or previous sibling. A pager at the bottom of the page offers the same moves.
5. **Given** the reader is typing in the search box, **When** they press → or ←, **Then** the cursor moves in the text and the page does not change.

---

### User Story 3 - Read the overview and view pages (Priority: P2)

A reader opens the site's index to see the size and shape of the architecture, then opens a view and reads its diagram at a legible size.

**Why this priority**: The index is the first page a reader sees, and view pages carry the diagrams. They work today, so the gap is smaller than for element pages.

**Independent Test**: Build any example and open the index and one wide view. Check the stat cards, the row lists, and that the reader can zoom the diagram until its labels are readable.

**Acceptance Scenarios**:

1. **Given** the index page, **When** it loads, **Then** it shows stat cards for systems (and externals), containers (and data stores: databases, queues and topics), relationships (and async), and open warning and info diagnostics.
2. **Given** the index page, **When** it loads, **Then** it shows the landscape in a captioned figure with an "Open view" link, and lists systems and views as rows with name, description and chips (child counts, external, automatic, declared, sequence, deployment), in place of the raw table of ids.
3. **Given** a view page, **When** it loads, **Then** the header shows the view kind and mode (`SYSTEM VIEW · AUTOMATIC`) with counts in the metadata line, and the diagram sits on a light panel in a captioned figure with links to the SVG and the D2 source.
4. **Given** a view diagram, **When** the reader clicks or scrolls the wheel on it, drags it, or presses reset, **Then** it zooms, pans, or returns to its first fit.
5. **Given** a view page, **When** it loads, **Then** an "In this view" row list names every element in the view. This list is the diagram's text alternative.
6. **Given** a dynamic view (019), **When** its page loads, **Then** it keeps its Steps table.

---

### User Story 4 - Read the site in the Madstone look, in light or dark (Priority: P2)

A reader sees the site in Madstone branding, in the theme that matches their system, and can change it.

**Why this priority**: The brand and a readable dark theme matter for every page, but the site is usable without them.

**Independent Test**: Open the site with the system set to dark, then to light, then with no preference and with scripts off. Check the theme each time, use the toggle, reload, and check the choice is kept.

**Acceptance Scenarios**:

1. **Given** the system prefers dark, **When** the site loads, **Then** it uses the dark theme. **Given** the system prefers light, **Then** it uses the light theme.
2. **Given** the preference cannot be determined (no preference, scripts off, print, an older browser), **When** the site loads, **Then** it uses the light theme.
3. **Given** the reader uses the theme toggle, **When** they reload or open another page, **Then** their choice is kept.
4. **Given** scripts are off, **When** the site loads, **Then** every page renders in the system theme, or light, and all content and links are reachable.
5. **Given** either theme, **When** a diagram is shown, **Then** it sits on a light "paper" panel, and element tiles and chips use 020's kind colours so the page matches its diagram.

---

### User Story 5 - Use the site with assistive tech, offline, in print and on a phone (Priority: P3)

A reader uses a screen reader, keyboard, reduced-motion setting, printer, phone, or no network, and the site still works.

**Why this priority**: These are quality requirements that apply to all the stories above. They are tested separately.

**Independent Test**: Run a contrast check on both themes, tab through a page, print an element page and a view page, open a page at phone width, and open the site with the network off.

**Acceptance Scenarios**:

1. **Given** any text in either theme, **When** its contrast is measured against its background, **Then** it is at least 4.5:1.
2. **Given** keyboard navigation, **When** any control has focus, **Then** a focus ring is visible.
3. **Given** the reader prefers reduced motion, **When** the page loads, **Then** nothing blinks or animates. Without that preference, the cursor blink stops after a few seconds.
4. **Given** a printed page, **When** it is printed, **Then** it uses the light palette, hides the sidebar, toggles and pager, places the rail under the prose, keeps cards, table rows and diagrams whole, and prints diagrams full-width as vectors without glows or the dot grid.
5. **Given** a phone-width screen, **When** a page loads, **Then** the rail stacks under the description, the sidebar sits behind a menu button, and the page does not scroll sideways.
6. **Given** no network, **When** the site is opened from disk or `loko serve`, **Then** every font, icon and script loads and the page looks the same as online.

---

### Edge Cases

- A system with no containers, or a container with no components: the page shows no Children list (not an empty one) and the rail omits Siblings when the parent has no other children.
- A top-level element (person or system) has no parent: Part of is omitted, Siblings are the other top-level elements of the same kind, and the breadcrumb is the project and the element.
- An element with no relationships: Talks to is omitted.
- A neighbour related in both directions: one row, with a two-way arrow.
- A relationship between siblings: the Talks to row carries the "sibling" marker, and the Siblings list shows "calls" or "called by" (or both).
- A parent with no diagram (no view includes it): In context is omitted, and the page does not fail.
- An element that appears in more than one view: In context uses the parent's own view. If the parent has no view of its own, it uses the first view that contains the element, in a stable order.
- A component whose `docs` points at its parent's file (as in `examples/microservices`): the page links to the parent's prose and does not inline it, and its In context diagram fills the main column.
- Very long names or ids: they wrap or truncate with the full text available, and do not cause a sideways scroll.
- Arrow keys on the first or last sibling: nothing happens (no wrap).
- A project with one element and one view: every page still renders with its chrome.
- A user's theme override in `<project>/templates/` defines an old block name: it still builds, because no 014 block name is removed. A block name that never existed is still reported with its file, as today (014).
- A name or prose with characters outside the Latin subset (for example CJK or Cyrillic): those characters are drawn in the system fallback font. Nothing is missing and the layout holds.
- A projected count that is zero (for example, no async relationships): the stat card shows 0, not a blank.

## Requirements *(mandatory)*

### Functional Requirements

**Navigation (every page)**

- **FR-001**: Every page MUST have a sidebar with the project name and description, a search box, an Architecture tree (people, systems, containers, components, each with its 020 icon), a Views list, and a footer saying the pages are generated from `*.loko.hcl` and must be edited at the source.
- **FR-002**: Each page MUST include a pruned tree: the top-level elements, the children of each ancestor of the current element, and the current element's own children. The tree MUST highlight the current element with its branch expanded. A collapsed branch MUST link to its element's page, where that branch is expanded. The tree MUST work without scripts. On the index and view pages, the tree MUST show the top level only.
- **FR-003**: Every page MUST have a breadcrumb that runs up to the project.
- **FR-004**: Element pages MUST have a previous/next pager between siblings. The ← and → keys MUST do the same, except while focus is in a text field. The pager MUST NOT wrap from the last sibling to the first.
- **FR-005**: The site MUST show element and view names, not ids, everywhere except the metadata line, which shows the full id.
- **FR-006**: The index, element and view pages MUST all use the same chrome (sidebar, tree, breadcrumb).

**Element page**

- **FR-007**: An element page header MUST show an accent bar, a kind label naming the parent (`COMPONENT · IN GATEWAY`), the title with the element's icon tile in its kind colour, a muted metadata line (technology, full id, source file and line), and the description as a lead.
- **FR-008**: The main column MUST show the element's own prose and a note naming its source file. When the element has no prose of its own, the In context diagram MUST appear in the main column.
- **FR-009**: When an element's `docs` resolves to a file that its parent also uses, the page MUST link to the parent's prose and MUST NOT inline it. For FR-008, such an element counts as having no prose of its own.
- **FR-010**: The right rail MUST show, in this order and omitting empty sections: Part of, Siblings, Talks to, In context.
- **FR-011**: Part of MUST show the parent with its kind, technology and description, and the system above it when the parent is a container.
- **FR-012**: Siblings MUST list the parent's other children with the current one marked, each with "calls", "called by" or both when a relationship links it to this element.
- **FR-013**: Talks to MUST show one row per neighbour, whatever the number of relationships. Each row MUST show a direction arrow (out, in or both), a technology chip, a description, and a "sibling" marker when the neighbour is a sibling.
- **FR-014**: In context MUST reuse the parent's existing diagram with every other node faded and this element outlined. It MUST NOT render a new diagram per element.
- **FR-015**: A system or container page MUST list its children as rows with name, kind, technology and description.

**Overview page**

- **FR-016**: The index MUST show stat cards for: systems (and externals), containers (and databases and queues), relationships (and async), and open warning and info diagnostics.
- **FR-017**: The index MUST show the landscape in a captioned figure with an "Open view" link.
- **FR-018**: The index MUST list systems and views as rows with name, description and chips (child counts, external, automatic, declared, sequence, deployment).

**View page**

- **FR-019**: A view page header MUST follow FR-007's layout, with the view kind and mode as the label (`SYSTEM VIEW · AUTOMATIC`) and element and relationship counts in the metadata line.
- **FR-020**: The diagram MUST sit on a light panel in a captioned figure with links to its SVG and D2 source.
- **FR-021**: The diagram MUST support zoom (click and wheel), pan (drag) and reset. Without scripts, it MUST show at its fitted size.
- **FR-022**: A view page MUST have an "In this view" row list of its elements, which serves as the diagram's text alternative.
- **FR-023**: Dynamic view pages MUST keep their Steps table (019).

**Look and typography**

- **FR-024**: The site MUST have light and dark themes. It MUST follow the system colour-scheme preference, and MUST use light when the preference cannot be determined (no preference, scripts off, print, older browsers).
- **FR-025**: A toggle MUST let the reader override the theme, and the choice MUST persist across pages and reloads in that browser.
- **FR-026**: The site MUST use the Madstone surfaces: dark page `#0F1117` and cards `#1A1D24`; light page `#F9FAFB` and white cards. The sidebar MUST be Bedrock `#1E1E1E` with a faint dot grid in both themes, with the Madstone Eye glyph and a `LOKO` wordmark at its top.
- **FR-027**: Cipher Green MUST be used only as a signal (active tree item, current sibling, accent bar, active toggles, hover and focus rings), never as a large fill. Links MUST use Signal Blue.
- **FR-028**: The site MUST use small radii (6px cards, 4px buttons, 3px tags), borders for cards in dark mode, light shadows only in light mode, the brand glyphs `▸` `▾` `→`, sentence-case labels and no emoji.
- **FR-029**: Element tiles and chips MUST use 020's kind palette and icons, and diagrams MUST sit on a light panel in both themes.
- **FR-030**: Headings and the page title MUST use Orbitron; prose, leads and descriptions MUST use Source Sans 3; the sidebar, kind labels, metadata, ids, chips and code MUST use JetBrains Mono; Audiowide MUST be used only for the wordmark and stat numbers. This exception to the brand's monospace-body rule MUST be recorded in a decision record.
- **FR-031**: All fonts, the Eye glyph and icons MUST be bundled into the site with their licences, and every bundled file MUST carry the generated-file notice like any other output (014 FR-020). The site MUST NOT load anything from the network at view time. Fonts MUST be WOFF2, cut down to the Latin character set, with only the weights the design uses, and MUST total at most 500 KB. Each font stack MUST end in a system font, which draws any character outside the subset.

**Accessibility**

- **FR-032**: Every text colour in both themes MUST have at least 4.5:1 contrast with its background. In particular: the faintest text tone MUST be used only for decoration; light mode MUST show Cipher Green text as `#128200` or as a green rail beside dark text; light-mode links MUST use `#2D66AE`. These overrides MUST stay in loko and MUST NOT change the brand tokens.
- **FR-033**: The cursor blink MUST stop after a few seconds and MUST NOT run when the reader prefers reduced motion.
- **FR-034**: Every focusable control MUST show a visible focus ring.
- **FR-035**: Kind MUST be shown by colour, icon and text together, never by colour alone.
- **FR-036**: Every diagram MUST be in a figure with a caption.

**Print and narrow screens**

- **FR-037**: The site MUST have a print style: light palette; no sidebar, toggles or pager; the rail under the prose; cards, table rows and diagrams not split across pages; diagrams full-width as vectors; no glows or dot grid.
- **FR-038**: At phone width, the rail MUST stack under the description, the sidebar MUST collapse behind a menu button, and the page MUST NOT scroll sideways.

**Compatibility and constraints**

- **FR-039**: The site MUST stay fully static and self-contained. Its scripts MUST stay small, with no third-party libraries, and cover only pan and zoom, the theme toggle, the tree, keyboard paging, the menu button and search.
- **FR-040**: Two builds of the same project MUST produce byte-identical output on any machine.
- **FR-041**: `loko serve` live reload MUST keep working.
- **FR-042**: Theme overrides in `<project>/templates/` (014) MUST keep working. Every block name in the 014 theme contract MUST still exist, so an override that defines only old block names still builds. New blocks MAY be added for the new sections (rail, tree, breadcrumb, pager, stat cards). Any block whose content or meaning changes MUST be listed in the changelog and recorded in a decision record, and the theme contract MUST be updated to list every block.
- **FR-043**: Search, the markdown output, and MCP and CLI outputs MUST NOT change.
- **FR-044**: The added asset weight of the site (fonts, brand assets, scripts and styles) MUST be measured and reported in the release notes. A test MUST fail if the fonts total more than 500 KB.

### Key Entities

- **Element page context**: what a page needs beyond the element itself: its ancestors (for the breadcrumb and Part of), its siblings with their relation to it, its children, its neighbours rolled up to one entry each with direction and sibling flag, its previous and next sibling, and the diagram to use for In context.
- **Neighbour**: another element linked to this one by one or more relationships, with a direction (out, in or both), its technology, its description and whether it is a sibling.
- **Overview counts**: systems and externals, containers and databases or queues, relationships and async relationships, and open warning and info diagnostics.
- **Navigation tree**: the model's people, systems, containers and components in a stable order, each with its name, kind, icon and page link. Each page gets a pruned copy along its own path (FR-002).
- **Theme**: light or dark, chosen from the system preference, with light as the fallback and an optional reader override that persists.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On every container and component page in the examples, a reader can name the element's parent, its siblings and every neighbour without opening a diagram.
- **SC-002**: Every element in the model is reachable from any page through the tree in at most one click per level of nesting (system, container, component). No page's tree has more rows than the top level plus the children along its own path.
- **SC-003**: No neighbour appears more than once in any element's Talks to list.
- **SC-004**: All text in both themes measures at least 4.5:1 contrast, and every interactive element has a visible focus style; both are checked automatically. A manual accessibility audit of the index, an element page and a view page in both themes reports no contrast or focus failures.
- **SC-005**: The 1,020-element benchmark project builds in under 10 seconds.
- **SC-006**: Two builds of the same project produce byte-identical output.
- **SC-007**: With the network off, every page loads with its fonts, icons and scripts, and makes zero network requests.
- **SC-008**: At 375px width, no page scrolls sideways.
- **SC-009**: A wide view's labels can be zoomed to at least 12px on screen.
- **SC-010**: Every existing override in the examples and the 014 theme tests still builds without changes, and the changelog names each block whose content changed.
- **SC-011**: The bundled fonts total 500 KB or less, and a name with non-Latin characters still renders legibly.

## Assumptions

- **Same chrome everywhere**: the seed proposed that the index and view pages get the element page's full chrome, and this spec adopts it (FR-006).
- The approved mockups (`.superpowers/brainstorm/15936-1793152578/content/site-mockup-madstone-v4.html`, local only) are the visual reference where this spec is silent.
- The brand source of truth is `../mdstn-branding` (`tokens/madstone.css`). Its fonts are OFL and its Eye glyph can be redistributed with loko under its licence.
- The theme choice is stored in the reader's browser only. It is not a project setting.
- Siblings of a top-level element are the other top-level elements of the same kind (people with people, systems with systems).
- The fixed sibling order is the order the tree uses, which is the model's stable order.
- Search keeps its current behaviour and index. Only its place in the sidebar changes.
- The new derived data (siblings, children, neighbour rollup, counts, breadcrumb) is computed once per build from the existing model. No new storage and no new third-party module are needed.

### Dependencies

- 020 (diagram look): the Lucide icon set, the Signal Blue kind palette, and the SVG selectors in `specs/020-diagram-look/contracts/svg-output.md`, which In context uses to fade and outline nodes.
- 014 (view-model renderers): the theme override contract in `specs/014-viewmodel-renderers/contracts/theme.md`.
- 019 (dynamic views): the Steps table.

### Out of Scope

- User customisation through a tokens or styles file (a later feature).
- A dark-aware diagram palette.
- Per-element generated diagrams.
- Changes to the Madstone brand tokens.
