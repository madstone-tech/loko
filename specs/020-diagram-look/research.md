# Research: Diagram Look

Decisions made while planning feature 020, with the evidence behind each. The spike referred to
throughout was a throwaway patch of `internal/adapters/d2` rendered against
`examples/microservices` on 2026-10-27; none of its code is kept.

## R1. Icons are drawn by loko after d2 renders, not with d2's `icon:` attribute

**Decision**: The D2 source carries no `icon:`. After d2 renders the SVG, a loko pass inserts each
element's icon into its node group: one `<symbol>` per icon used in the diagram, defined once in a
`<defs>` block, and one `<use>` per node, placed in the card's top-left corner inside the padding
d2 already leaves.

**Rationale**: d2 v0.7.1 gives any node with an icon extra padding equal to the label's height, on
both axes (`d2graph.go:1616`, "give shapes with icons extra padding to fit their label"). For a
four-line label that is about 70px in width and height per card, which is why the spike's cards
looked empty and why SC-004 (no view grows more than 10%) cannot be met with `icon:`. Drawing the
icon afterwards leaves d2's sizing exactly as today, so view sizes do not grow at all. d2 pads a
rectangle by 20px on each side (`shape.defaultPadding = 40`), so a 16px icon inset 4px from the
corner sits in the padding and never overlaps the label block, which starts 20px in.

`<symbol>` plus `<use>` keeps the size cost to one definition per icon and about 100 bytes per
node (SC-005). Inline paths, rather than a base64 `data:` image, render in every SVG viewer and
inherit the card's text colour through `stroke="currentColor"` on the `<use>`.

**Alternatives considered**:
- d2 `icon:` with `icon.near: top-left` (the spike). Rejected: the padding above makes every card
  much larger; no d2 option removes it.
- d2 `icon:` plus explicit `width`/`height` attributes (d2 drops its padding when sizes are given).
  Rejected: needs every label measured before layout, which means a second compile per view and
  doubles layout time against the 10-second budget.
- Icon inline before the name, with room reserved by padding the name line. Rejected: needs every
  name line measured to place the icon, and the reserved spaces would also appear in the `.d2`
  output and under the d2 CLI.
- `data:` URI `<image>` per node. Rejected: a full base64 copy per node (SC-005), and an `<image>`
  cannot take the card's text colour.

**Spec impact**: FR-016 is amended. The `.d2` files carry the cards, palette and label text but no
icons; the d2 CLI renders them as plain cards. Earlier the spec required icons there too, which
held only for the `icon:` approach.

## R2. The label pass styles lines by role, using a plan from the emitter

**Decision**: The emitter writes each label as plain lines (title or name, kind line, blank,
description) and, for the SVG backend, also returns a *label plan*: for every node and edge, its
d2 path, its icon, and the role of each label line. The SVG pass finds each node and edge group
by its class (d2 writes the base64 of the object's path as the first class), looks up the plan,
and gives each `<tspan>` the style for its role.

**Rationale**: d2 writes one `<tspan>` per label line (`d2svg.RenderText`), so styling by line is
exact. The pass cannot infer roles from the text: a title may wrap over several lines (016) and a
description may start with `[`. The emitter already knows which line is which, so it records it.
Styles are inline `style` attributes on each `<tspan>`, not a `<style>` block, so they survive
being inlined into another document (021) and work in renderers that ignore CSS.

**Alternatives considered**: matching lines by content (the spike: first line bold, lines starting
with `[` as the kind line). Rejected for the reasons above. Markdown labels. Rejected in 016 R2:
they render through `<foreignObject>`, which many SVG viewers do not support.

**Checked during implementation (T022, T023)** on `examples/microservices` `container-gateway.svg`:

| Viewer | Icons | Palette | Line sizes and opacity | Embedded fonts |
|---|---|---|---|---|
| Chromium (offline, file served locally) | yes | yes | yes | used: name and kind line bold, description regular; only `data:` font requests |
| `rsvg-convert` 2.x | yes | yes | yes | ignored: system sans-serif, so the name is not bold |
| Inkscape 1.x | yes | yes | yes | ignored ("font face rule limited support"); blank spacer lines collapse |

`href` alone is enough for all three; no `xlink:href` is needed. Inkscape fills boundaries
(`fill: transparent`) black. That was already true before this feature.

## R3. Bold only where d2 already measured bold, and weight changes by font family

**Decision**: Node labels: the first line(s) (the title or name) and the kind line keep d2's bold
face. The description switches to d2's regular face. Both are smaller than the name. Edge labels:
the relationship description is left exactly as d2 drew it; only the technology and tag lines
are made smaller.

A lighter weight is selected by **font family**, not by `font-weight`. d2 v0.7.1 embeds each
style as its own single-face family: `.<hash> .text-bold { font-family: "<hash>-font-bold" }` and
`.<hash> .text { font-family: "<hash>-font-regular" }` (`d2svg.go` 2440–2462 and 2556–2575). A
`font-weight:400` inside a bold text keeps the bold face, and `font-weight:600` can make a
renderer synthesise extra bold on a face that is already bold. So the description tspan sets
`font-family:<hash>-font-regular` inline, and no role sets `font-weight`. `<hash>` is read from
the class d2 gives its inner `<svg>` (`d2-<n>`). That is `diagram.HashID(nil)`, the value d2
prefixes its font families with (`d2svg.go` 2796, our `RenderOpts` sets no salt), read from the
output so that a mutation inside `d2svg.Render` cannot make the two differ.

d2 adds the regular `@font-face` only when its output already contains `class="text"`. A view
with no regular-weight text would have no regular face. After rendering, `compile` therefore adds
the regular `@font-face` when it is missing:
`d2fonts.SourceSansPro.Font(0, d2fonts.FONT_STYLE_REGULAR).GetEncodedSubset(diagram.GetCorpus())`.
The corpus is every label in the diagram, so the subset holds every description glyph. This is
the font d2 already embeds (FR-007), only in its regular style.

**Rationale**: d2 measures a node label in bold (elements have `class="text-bold"`) and an edge
label in its regular weight. Making any line *wider* than d2 measured overflows the space it
reserved: the spike bolded edge labels and they ran into neighbouring edges. Making lines
smaller, or switching them from bold to regular, never overflows.

## R4. Text sizes, spacing and opacity

**Decision**:

| Role | Face | Size | Letter-spacing | Opacity |
|---|---|---|---|---|
| Name / title | bold (d2's) | 1em (d2's 16px) | normal | 1 |
| Kind line | bold (d2's) | 0.7em | 0.06em | 0.85 |
| Description | regular (`font-family:<hash>-font-regular`) | 0.9em | normal | 0.9 |
| Edge description | as drawn by d2 | 1em | normal | 1 |
| Edge technology | as drawn by d2 | 0.75em | 0.04em | 0.8 |
| Edge tags | as drawn by d2 | 0.75em | normal | 0.72 |

The smallest size, 0.7em of d2's 16px, is 11.2px at the diagram's natural size. No role may go
below 11px (US4).

**Rationale**: these are the spike's values, adjusted for contrast (R5) and for d2's font
families (R3). d2 spaces lines evenly (`dy` is the label height divided by the line count), so
smaller lines leave a little air below them; the visual weight still reads name → kind →
description, which is what the owner approved.

## R5. Palette and contrast

**Decision**:

| Kind | Fill | Outline | Text |
|---|---|---|---|
| person | `#1A4275` | `#14365F` | `#FFFFFF` |
| system | `#1E4F8F` | `#183F73` | `#FFFFFF` |
| container | `#2D66AE` (Signal 700) | `#245390` | `#FFFFFF` |
| component | `#93C5FD` | `#60A5FA` | `#0F1117` |
| external | `#525252`, dashed | `#3D3D3D` | `#FFFFFF` |
| boundary (subject, group) | transparent, dashed | `#737373` | `#3D3D3D` |
| outside node | unchanged (`#999999` / `#666666`) | | |
| edges | | `#737373` | `#3D3D3D` |

Contrast of every line against its fill (WCAG AA needs 4.5 for normal text):

| Fill | Name | Kind line (0.85) | Description (0.9) |
|---|---|---|---|
| `#1A4275` | 10.1 | 7.8 | 8.5 |
| `#1E4F8F` | 8.2 | 6.4 | 7.0 |
| `#2D66AE` | 5.8 | 4.7 | 5.1 |
| `#93C5FD` (dark text) | 10.5 | 7.8 | 8.6 (approx.) |
| `#525252` | 7.8 | 6.2 | 6.7 |

Edge text `#3D3D3D` on the white diagram background (d2 theme 0, `N7` = `#FFFFFF`) is 10.9. The
technology line at 0.8 opacity is 5.9. The tag line at 0.72 opacity is 4.7. At the spike's 0.7 it
was 4.47, which fails, so the tag opacity is 0.72.

Nodes outside the view keep their colours (`#999999` stroke, `#666666` text). They are not cards,
and the spec's clarification of 2026-10-27 (analysis) keeps them as they are.

**Rationale**: the spike's container fill `#3B7BC8` gave 4.33 for the name and 3.6 for the kind
line at 0.85, both failing. Signal 700 is the lightest ramp step that passes every line at these
opacities, and it keeps the C4 order (person and system darker than container, component
lightest). The brand's green accent is not used (FR-009).

## R6. Where the three tables live (FR-011)

**Decision**:
- **Palette**: stays in `internal/core/entities/viewmodel/style.go` (`elementPalette` and the
  boundary style). The projection keeps deciding colours; the d2 adapter only translates them.
  Edge colours join it as a small `EdgePalette` value so the adapter does not hard-code them.
- **Icon table**: `internal/adapters/d2/icons.go`: a map from `vm.Shape` and element kind to an
  icon name, and the vendored Lucide files embedded with `embed`.
- **Label text styles**: `internal/adapters/d2/labelstyle.go`: a map from line role to its style
  (R4).

**Rationale**: each is a single table a later customisation feature can replace. Icons and text
styles are presentation details of the SVG backend; other backends (markdown, HTML) never draw
them. The palette is already a core decision (FR-008 of 014: backends decide nothing).

The icon table is a translation table, not a decision. The projection already decides *what* an
element is (`Style.Shape` and its kind). The SVG backend maps that value to its own glyph, as the
emitter maps `ShapeOval` to d2's `oval`. The comment on `viewmodel.Style` says so, and ADR-0019
records it.

## R7. Vendored Lucide files

**Decision**: Copy the 10 icons from `lucide-static` (pinned version, recorded in the notice) into
`internal/adapters/d2/icons/`, unmodified, with Lucide's `LICENSE` (ISC). At load time each file's
inner elements become a `<symbol id="loko-icon-<name>" viewBox="0 0 24 24">`, wrapped in a `<g>`
with Lucide's own root attributes (`fill="none" stroke="currentColor" stroke-width="2"
stroke-linecap="round" stroke-linejoin="round"`). The `<use>` sets only `color`, the card's text
colour. Putting the stroke attributes in the symbol, once per diagram, saves about 95 bytes per
card against setting them on every `<use>` (found while measuring SC-005).

**Rationale**: FR-004 asks for the upstream files with their notice; embedding keeps the binary
self-contained. The symbol ids are fixed strings, so output stays byte-identical (FR-015). d2
prefixes its own ids with a diagram hash; `loko-icon-` cannot collide with them. When two
diagrams are inlined into one HTML page (021), identical ids define identical symbols, so a
duplicate is harmless.

## R8. Kind line text

**Decision**: The emitter writes the kind line itself, without brackets:
`CONTAINER · Go, gRPC`. The kind is upper-cased by the emitter (the `kindTitles` table gains
upper-case values); the technology is written as the author wrote it. For a titled element the
line ends with ` · <name>`. An edge's technology loses its brackets too: it is the bare
technology (`gRPC`) on its own line, styled by role.

**Rationale**: clarification Q1 and Q4. Upper-casing in the emitter, not with CSS
`text-transform`, keeps it correct in viewers that ignore CSS and in the `.d2` output.

## R9. Sequence diagrams (FR-018)

**Decision**: Sequence participants go through the same pass. d2 draws a participant as an
ordinary rectangle with the same padding and label, so the corner icon and line roles apply
unchanged. Lifelines and step arrows are edges: step labels get the edge roles.

**Verification task**: render view `checkout_sequence` of `testdata/projects/dynamic-views` (the
019 sequence golden) and check the icon sits inside the participant card without touching the
name or the lifeline. If it does not fit, FR-018's fallback applies and the plan records why.
(`examples/microservices` has no sequence view: `order-flow` is a static view.)

## R10. Performance and size

**Decision**: The pass is a single scan over the SVG bytes using the plan's lookup; no XML parser.
The render cache stores the processed bytes, so a hit skips both compile and pass. Its key must
change from the D2 source alone to the D2 source plus the serialised label plan: the pass output
depends on the plan, and two views with the same D2 source can have different line roles (see
below).

**Rationale**: d2 compile dominates (2.8 s for 1,020 elements under dagre); a linear byte pass is
milliseconds. Size: one `<defs>` of the symbols used, about 110 bytes per card for its `<use>`,
and the inline styles per label. Both are measured in the existing `cmd/build_perf_test.go` and
`TestSVGArtifactSize`.

**Measured during implementation (T037)**: the 1,020-element benchmark's SVGs grew 16.6%
(4,788,265 → 5,583,569 bytes), over the 15% the spec first set. About 5.8 points come from the
upper-case kind line: it adds capital glyphs to all three font subsets d2 embeds in every
diagram. Most of the rest is a fixed cost per card (the `<use>` and the line styles), which is
large next to the benchmark's four-character labels. The example projects grew 9.6–12.7%. Moving
Lucide's stroke attributes from every `<use>` into the symbols (R7) saved about 2 points. The
owner chose to amend SC-005 to 20% and keep the design (spec clarification of 2026-10-27,
implementation).

The regular `@font-face` that `compile` adds (R3) is a large item where it is added: a base64 subset of
Source Sans Pro for the whole diagram corpus. `compile` adds it only when the plan has at least
one description line and d2 did not already embed the face. A diagram with any regular-weight
text (most edge labels) already has it, so it costs nothing there. T037 measures the total.

Line roles cannot be recovered from the source text: an edge whose description holds a newline
reads the same as a description followed by a technology line. The `shape-<shape>` class (R12)
puts the shape into the D2 source, so icons alone would not need the plan in the key.

## R11. Deployment views

**Decision**: No special case. Deployment environments and nodes are `RoleSubject` and
`RoleGroup` boundaries (FR-013, new neutral colours only); instances are `RoleInstance` with an
element kind, so they get the element card, icon and label roles.

## R12. The declared shape stays in the `.d2` output

**Decision**: Every element and instance node with a declared `shape` carries a `shape-<shape>`
class (for example `shape-database`) in the `.d2` output, beside its `kind-*` and `tag-*` classes.
The emitter adds it when it translates `Style.Shape`, and declares it in the `classes` block like
every other class. d2 writes it onto the node group in the SVG, so user CSS can target it too.

**Rationale**: constitution Principle VI says a new default that removes information from an
existing output needs an ADR, a changelog entry and the previous shape on request. Drawing every
element as `rectangle` would remove the declared shape from the `.d2` file. With the class, the
information stays in the source: a reader or a tool still sees `shape-database`. Only the glyph
changes, from a cylinder to a card with an icon. That is a presentation change, recorded in
ADR-0019 and the changelog. FR-014 is kept: `kind-*` and `tag-*` stay on the same nodes.

**Alternatives considered**: a `--d2-shapes` option to restore the old shapes. Rejected: it adds a
second rendering path for no information gain (Principle VII). A `# shape: database` comment.
Rejected: d2 drops comments, so the SVG would lose it, and a class is already the mechanism for
kinds and tags.

## R13. Rounded cards

**Decision**: Every element and instance card has `style.border-radius: 6` in the D2 source.
Boundaries and outside nodes keep their current corners (FR-013).

**Rationale**: FR-001 asks for a rounded card. d2 draws a `rectangle` with square corners unless
`border-radius` is set, and the emitter set none before this feature. Six pixels is visibly round
and stays inside the 20px padding, so the icon at 4px from the corner stays clear of the curve's
widest point. The value is in the D2 source, so the d2 CLI draws the same corners (FR-016).
