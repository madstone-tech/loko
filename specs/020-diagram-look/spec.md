# Feature Specification: Diagram Look

**Feature Branch**: `020-diagram-look`

**Created**: 2026-10-27

**Status**: Draft

**Input**: Brainstorming session of 2026-10-27 on styling and theming. The rendered SVG diagrams
are functional but look crude: the stock shapes for people, databases, queues, topics and
functions are basic outlines, and every line of an element's label (name, kind and technology,
description) is drawn in the same font, size and colour, so nothing stands out. A throwaway spike
on `examples/microservices` showed that a cleaner default is reachable without a new dependency:
each element drawn as a card with a small line icon, and the label lines restyled into a
hierarchy. This feature makes that the default look. Letting users override the style (icons,
fonts, colours through tokens) is a later feature; the HTML site redesign (021) builds on this
one's icons and palette.

## Clarifications

### Session 2026-10-27

- Q: On the kind line, does the technology keep its own casing or go to full capitals? → A: Only
  the kind is in capitals; the technology keeps the casing the author wrote
  (`CONTAINER · Go, gRPC`).
- Q: Do sequence-diagram participants (019) get the icon cards too? → A: Yes, the same cards and
  icons as other views; if the sequence layout cannot place the icon cleanly, participants in
  that style fall back to plain cards with the new palette and text hierarchy, and the plan
  records why.
- Q: Where does a titled element's name go? → A: At the end of the kind line, as written, only
  for titled elements: `SYSTEM · Kong · api_gateway`.
- Q: How much larger may a view get than today's rendering? → A: At most 10% in width and in
  height, per view.

### Session 2026-10-27 (analysis)

- Q: Must a card be no wider than its rendered text? → A: No. A card is no larger than d2 sizes it
  for the same label text. d2 measures every line at the name's size, so a smaller kind line or
  description leaves some space on the right. SC-004 bounds the result (FR-012).
- Q: Do nodes outside the view take the neutral colours? → A: No. They keep their current
  colours. Only boundaries move to the neutrals (FR-009, FR-013).
- Q: Does the `.d2` output keep the declared shape? → A: Yes, as a `shape-<shape>` class on the
  node, so the information stays in the source (FR-016, constitution Principle VI).

### Session 2026-10-27 (implementation)

- Q: The 1,020-element benchmark's SVGs grow 16.6%, over SC-005's 15%. Raise the limit, write the
  kind line in mixed case, or measure on the examples? → A: Raise SC-005 to 20% and keep the
  design. The examples grow 9.6–12.7% (research R10).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read an element at a glance (Priority: P1)

An architect opens a rendered diagram. Each element is a card in its kind's colour with a small
icon in the corner that says what it is (a person, a system, a container, a component, an
external system, a database, a queue, a topic, a function, a storage bucket). On the card, the
element's name stands out first, then a small line naming its kind and technology, then a
lighter description.

**Why this priority**: This is the complaint the feature answers: the shapes look crude and the
text blends. Without it nothing else in the feature matters.

**Independent Test**: Build `examples/microservices` and open its system and container views.
Every element is a card with an icon matching its kind or declared shape; on every card the name
is visibly heavier than the kind line, and the kind line visibly smaller than the description.

**Acceptance Scenarios**:

1. **Given** a container declared with `shape = "database"`, **When** its view is rendered,
   **Then** it is drawn as a card with the database icon, not as a cylinder.
2. **Given** a person, a system, a container, a component and an external system with no
   `shape`, **When** they are rendered, **Then** each is a card with its kind's icon.
3. **Given** a container with technology "Go, gRPC" and a description, **When** it is rendered,
   **Then** its card reads: the name in bold, then `CONTAINER · Go, gRPC` in a small
   letter-spaced line with only the kind in capitals, then the description in lighter, smaller
   text.
4. **Given** a system named `api_gateway` with title "API Gateway" and technology "Kong",
   **When** it is rendered, **Then** "API Gateway" is the bold first line and the kind line reads
   `SYSTEM · Kong · api_gateway`.
5. **Given** an element without a `title`, **When** it is rendered, **Then** its name is the bold
   first line and does not appear again on the kind line.

---

### User Story 2 - Share a diagram outside loko (Priority: P1)

An architect drops a rendered SVG into a pull request, a wiki page, a slide or a PDF. It looks
the same there as in loko's own site: the icons and the text hierarchy survive, nothing is
fetched from the network, and the text stays real, selectable text.

**Why this priority**: Diagrams are committed and embedded; a look that only works inside loko's
site would be a regression from today.

**Independent Test**: Open a rendered SVG directly in a browser with the network disabled, and
convert it with a standalone SVG rasteriser. Icons, card colours and the text hierarchy appear
in both.

**Acceptance Scenarios**:

1. **Given** a rendered SVG, **When** it is opened offline in a browser, **Then** every icon is
   drawn and no request is made.
2. **Given** a rendered SVG, **When** it is converted by a standalone SVG renderer that does not
   run scripts, **Then** the icons, colours and text hierarchy are present.
3. **Given** a rendered SVG, **When** a reader selects the text on a card, **Then** the name,
   kind line and description are selectable text, not outlines.

---

### User Story 3 - A palette that matches the Madstone brand and stays readable (Priority: P2)

The kinds keep the C4 convention that a higher level is darker (person and system darkest,
container in between, component lightest), but in the Madstone Signal Blue ramp, with external
systems in a dark neutral. Edges, edge labels and boundaries use brand neutrals instead of the
renderer's default blue. Every piece of text on a card is readable against its fill.

**Why this priority**: It aligns the diagrams with the site redesign (021) and the brand, but the
diagrams are already useful in any palette once User Story 1 lands.

**Independent Test**: Render each kind and measure the contrast of every text line against its
card fill; inspect the stroke and label colours of edges and boundaries.

**Acceptance Scenarios**:

1. **Given** any element card, **When** the contrast between each of its text lines and the fill
   is measured, **Then** it meets WCAG AA for its size.
2. **Given** a view with relationships, **When** it is rendered, **Then** edges and their labels
   use the neutral colours, not the renderer's default blue.
3. **Given** any rendered diagram, **When** its colours are listed, **Then** the brand's green
   accent does not appear.

---

### User Story 4 - Compact, legible cards (Priority: P2)

Cards are sized to their content: the icon does not push the text to the top of a card that is
mostly empty, and no text line on a card renders smaller than 11px when the diagram is shown at its
natural size.

**Why this priority**: The spike's cards were noticeably roomier than today's boxes, which makes
diagrams larger and the text smaller once they are scaled to fit a page.

**Independent Test**: Compare each view of `examples/microservices` before and after; cards are
no larger than today's boxes for the same label, within the tolerance in SC-004.

**Acceptance Scenarios**:

1. **Given** a container with a one-line description, **When** it is rendered, **Then** FR-012's
   rule on the empty band below the text holds.
2. **Given** any view, **When** it is rendered, **Then** its overall width and height are within
   the SC-004 tolerance of today's rendering.
3. **Given** any element card, **When** the diagram is shown at its natural size, **Then** its
   smallest line (the kind line) is at least 11px high.

---

### Edge Cases

- An element with no technology: the kind line shows only the kind (`SYSTEM`).
- An element with no description: the card ends after the kind line, with no blank line.
- An element whose label has more lines than expected (a long title wrapped over several
  lines): every title line is bold; only the kind line and description lines are restyled.
- Edge labels with a technology line: the technology line is small and letter-spaced; the
  relationship's description keeps its normal weight so it does not overflow the measured space.
- Edge labels with tags (`#async #pii`): the tag line is drawn smaller and lighter.
- Boundaries (the subject of a view, groups) keep their current shape and treatment; only their
  colours move to the neutrals. Nodes outside the view keep their current shape, treatment and
  colours.
- Deployment views: instances are drawn with the same element cards and icons as their elements.
  Environments and deployment nodes are boundaries (research R11).
- Dynamic views drawn as sequence diagrams (019): participants use the same icon cards as in
  other views. If the sequence layout cannot place the icon cleanly (overlapping the name or the
  lifeline), participants in that style are drawn as plain cards with the new palette and text
  hierarchy, never failing the build.
- An element kind or shape that has no icon assigned (a kind added later): it is drawn as a card
  with the system icon, and a test fails so the gap is noticed.
- A user's own CSS that targets the `kind-*` and `tag-*` classes on nodes keeps matching the same
  nodes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every element node (person, system, container, component, external system) MUST be
  drawn as a rounded rectangle card filled with its kind's colour, in place of the stock person,
  cylinder, queue, hexagon, step and stored-data shapes.
- **FR-002**: Every element card MUST carry one icon in its top-left corner, chosen by the
  element's declared `shape` first and its kind second, from this table:

  | Shape or kind | Icon (Lucide) |
  |---|---|
  | `database` | `database` |
  | `queue` | `list-ordered` |
  | `topic` | `megaphone` |
  | `function` | `square-function` |
  | `bucket` | `archive` |
  | person | `user` |
  | system | `boxes` |
  | container | `container` |
  | component | `blocks` |
  | external | `globe` |

- **FR-003**: Icons MUST be drawn in the card's text colour (white on dark fills, near-black on
  the light component fill).
- **FR-004**: Icons MUST be the upstream Lucide files, shipped inside loko with the Lucide
  licence notice, and embedded in each SVG so that no file or network resource is referenced.
- **FR-005**: An element card's label MUST read, top to bottom: the name (or title) in bold; a
  small, letter-spaced kind line made of the kind in capitals and, when present, the technology
  exactly as the author wrote it (`CONTAINER · Go, gRPC`); the description in a smaller, lighter
  weight. For a titled element the bold line is the title and the kind line ends with the element's name as written
  (`SYSTEM · Kong · api_gateway`); for an untitled element the name appears only on the bold
  line.
- **FR-006**: Edge labels MUST keep the relationship description at its current weight and size,
  and draw the technology line and the tag line smaller and lighter.
- **FR-007**: All label text MUST remain plain SVG text that renders in any SVG viewer, without
  scripts, embedded HTML or external fonts beyond the font the renderer already embeds.
- **FR-008**: Element fills MUST follow this palette, darker for higher levels:

  | Kind | Fill | Text |
  |---|---|---|
  | person | `#1A4275` | white |
  | system | `#1E4F8F` | white |
  | container | `#2D66AE` (Signal 700, chosen in planning so every line passes FR-010) | white |
  | component | `#93C5FD` | `#0F1117` |
  | external | `#525252`, dashed outline | white |

- **FR-009**: Boundaries MUST use a `#737373` outline with `#3D3D3D` text; edges MUST use a
  `#737373` stroke and `#3D3D3D` label text. The brand's green accent (`#18C800`) MUST NOT be used
  in diagrams.
- **FR-010**: Every text line on a card, including the faded kind line and the description, MUST
  meet a WCAG AA contrast ratio against the card fill for its rendered size.
- **FR-011**: The palette, the icon table and the label text styles MUST each be defined in one
  place, so a later customisation feature can replace them without touching the code that draws
  nodes.
- **FR-012**: Cards MUST be no larger than d2 sizes them for the same label text: the icon adds no
  padding. The icon MUST NOT leave an empty band below the text larger than the space above the
  name.
- **FR-013**: Boundaries (view subjects and groups) and nodes outside the view MUST keep their
  current shape and dashed treatment.
- **FR-014**: The `kind-*` and `tag-*` classes MUST stay on the same nodes as today.
- **FR-015**: Output MUST stay byte-identical across runs and machines for the same input.
- **FR-016**: The D2 source loko writes (`.d2` files) MUST stay valid input for the d2 command-line
  tool and produce the same cards, palette and label text when rendered by it. Icons and the label
  line styles are added only to loko's SVG output (planning, research R1: d2's own icon attribute
  enlarges every card, so icons are drawn after rendering). A node with a declared `shape` MUST
  carry a `shape-<shape>` class in the `.d2` output, so the shape stays in the source (research R12).
- **FR-017**: Every kind and shape the language accepts MUST have an icon; a test MUST fail when
  one is added without an icon.
- **FR-018**: Participants of a sequence-style dynamic view MUST be drawn with the same cards and
  icons as in other views, unless the plan shows the sequence layout cannot place an icon without
  overlapping the name or the lifeline; in that case they MUST be plain cards with the palette
  and text hierarchy of FR-005 and FR-008.

### Key Entities

- **Kind palette**: per element kind, the fill, outline and text colour of its card; plus the
  neutral colours for boundaries and edges.
- **Icon table**: per declared shape and per element kind, the icon drawn on the card.
- **Label text styles**: per label line role (name, kind line, description, edge technology, edge
  tags), its font face, relative size, letter-spacing and opacity.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On every view of every example project, 100% of element cards show an icon and
  the three-level text hierarchy.
- **SC-002**: 100% of text lines on element cards meet WCAG AA contrast against their fill.
- **SC-003**: Rendered SVGs open with icons and text hierarchy intact in a browser with the
  network off and in a standalone SVG rasteriser that runs no scripts.
- **SC-004**: No example view grows by more than 10% in width or in height compared with today's
  rendering (shrinking is allowed).
- **SC-005**: Building the 1,020-element benchmark project stays under the existing 10-second
  budget, and the rendered SVGs grow by no more than 20% in total size (amended from 15% during
  implementation; see the clarification below and research R10).
- **SC-006**: Two builds of the same project on different machines produce byte-identical SVGs.
- **SC-007**: The project owner, comparing before and after renders of `examples/microservices`,
  judges the new diagrams clearer to read; this was confirmed for the spike and is re-checked on
  the final output.

## Assumptions

- The font stays Source Sans Pro, which the renderer already embeds. Switching diagrams to the
  brand's JetBrains Mono is deferred to the customisation feature: monospace text widens every
  card, and the renderer needs the font file itself to measure text.
- User-facing customisation (overriding icons, fonts or colours, through a styles block or
  theme files) is out of scope; FR-011 only keeps the door open.
- The light "paper" background stays; diagrams are shown on a light panel in both site themes
  (021), so no dark-aware diagram is needed yet.
- The d2 command-line tool (checked with v0.9.0) compiles loko's `.d2` output; icons are not
  part of that output (FR-016, research R1).
- Lucide's ISC licence permits bundling and modification; its notice ships with loko.
- The markdown output embeds or links the same SVGs and needs no change of its own.
- Golden SVG files in the tests are regenerated and reviewed as part of this feature; that every
  diagram changes is expected.
- This feature and 019 ship together in the next minor release (1.3.0).
- Feature 016 (FR-006, SC-001) requires labels of at least 16px. That stays true for every
  label's base size and its name line. The kind line, the description and an edge's technology
  and tag lines are smaller by design, with the 11px floor of US4. `TestContainerViewAspectRatio`
  checks both limits (found during implementation).
