# Contract: SVG Diagram Output

What a rendered `diagrams/<view>.svg` guarantees after feature 020. Consumers: people opening or
embedding the file, the HTML site (014, and 021 which inlines SVGs and highlights elements), the
markdown output, and user CSS (014 FR-018).

## Unchanged

- The file starts with the XML declaration, then the generated-file notice comment (014 R7).
- Every element and instance node group keeps its classes: the base64 d2 path first, then
  `kind-<kind>` and one `tag-<tag>` per tag. External nodes keep `external`.
- All label text is SVG `<text>`/`<tspan>`; no `<foreignObject>`, no script, no external
  reference (no `http:`, `https:` or file `href`).
- The font is the Source Sans Pro subset d2 embeds. Its regular style is now embedded whenever
  the diagram has at least one description line, because descriptions use it (research R3).
- Output is byte-identical for the same input on any machine.

## New

1. **Icon definitions.** When the diagram has at least one element card, the SVG contains one
   `<defs>` block holding a `<symbol id="loko-icon-<name>" viewBox="0 0 24 24">` for each icon it
   uses, and no others. `<name>` is one of: `user`, `boxes`, `container`, `blocks`, `globe`,
   `database`, `list-ordered`, `megaphone`, `square-function`, `archive`. Each symbol wraps the
   Lucide drawing in a `<g>` with Lucide's stroke attributes and `stroke="currentColor"`.
2. **One icon per card.** Each element and instance node group contains exactly one
   `<use class="loko-icon" href="#loko-icon-<name>" …>` positioned 4px inside the card's top-left
   corner, 16×16, with `color` set to the card's text colour. Boundary and outside node groups
   contain none.
3. **Label line styles.** In an element card's `<text>`, each `<tspan>` carries an inline `style`
   for its role (data-model, *Label line roles*): the kind line is smaller and lighter than the
   bold title, and the description uses the regular face (`font-family:<hash>-font-regular`),
   smaller and lighter. No line sets `font-weight`. Empty spacer lines are untouched.
4. **Edge label styles.** In an edge label, the technology and tag lines carry inline smaller,
   lighter styles; the description line is untouched.
5. **Colours.** Card fills, outlines and text use the kind palette (research R5); edges and edge
   labels use `#737373` / `#3D3D3D`; boundaries `#737373` / `#3D3D3D`. Nodes outside the view keep
   `#999999` / `#666666`. `#18C800` (the brand accent) never appears.
6. **Rounded cards.** Element and instance cards have rounded corners (`rx="6"` on the card's
   `<rect>`, from d2's `border-radius: 6`). Boundaries and outside nodes keep their corners.
7. **Shape class.** An element or instance node with a declared `shape` carries a `shape-<shape>`
   class beside its `kind-*` class (research R12).

## Selectors other features may rely on

| Selector | Matches |
|---|---|
| `g.kind-<kind>` | an element or instance card of that kind |
| `g.tag-<tag>` | a card carrying that tag |
| `g.<base64 path>` | one specific node (021 highlights an element this way) |
| `use.loko-icon` | every card icon |
| `g.shape-<shape>` | a card whose element declares that shape |

## D2 source (`diagrams/<view>.d2`)

- Valid input for the d2 CLI (checked with v0.9.0 during planning).
- Same cards (rounded, `border-radius: 6`), palette, edge colours, label text and classes (including
  `shape-<shape>`) as the SVG; no icons and no line styles, which only loko's SVG pass adds (FR-016
  as amended).
