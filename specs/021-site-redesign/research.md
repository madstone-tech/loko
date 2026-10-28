# Research: Site Redesign

Measurements were taken on 2026-10-28 on the developer machine. The benchmark is the 1,020-element
project from `cmd/build_perf_test.go`: 20 systems × 10 containers × 4 components.

| Measure | Today |
|---|---|
| `loko build --format html` wall time | 2.8 s (21.8 s CPU) |
| Artefacts / output size | 1,467 / 37 MB |
| SVG per view (221 views) | min 22 KB, median 24 KB, max 38 KB |
| HTML pages | 1,242, average 24.9 KB |

## R1. How the HTML backend gets SVG bytes

**Decision**: upstream artefacts, delivered on demand.

- `renderAll` starts every backend at once, as before.
- A backend whose `Format.Requires()` is not empty (`html`, `md`) gets
  `RenderOptions.Upstream`: a function that blocks until the backends that require nothing
  (`svg`, `d2`) have finished, then returns the required formats' artefacts keyed by path.
- The HTML backend renders every page with a placeholder comment where a diagram goes, then calls
  `Upstream` and replaces each placeholder with the inlined SVG, or with an `<img>` when the SVG
  is missing. The markdown backend never calls it.

**Revision (after CI):** the first version ran the stages one after the other, so HTML waited
for SVG. On the 1,020-element benchmark that took `loko serve`'s one-description rebuild from
0.37 s to 0.88 s, and over its 2 s budget on CI runners. With the overlap restored, plus a cached
views list and a Go-built tree, it is 0.46 s.

**Rationale**:
- The spec needs CSS to highlight nodes inside the diagram (FR-014) and pan/zoom (FR-021). Both
  need the SVG in the page's DOM.
- An `<img>` cannot be styled from the page.
- `<object>` plus a script cannot reach the SVG from `file://` in Chromium, because each file is
  its own origin.
- Fetching the SVG is blocked on `file://` for the same reason.
- `Format.Requires` already makes HTML depend on SVG. This only lets the data follow that
  dependency.
- Backends still never import each other.
- Cost: the placeholder fill is one byte replacement per page, a few milliseconds per build.

**Alternatives considered**:
- *The HTML backend calls the SVG renderer through a new port.* This adds a port and risks two
  concurrent renders of the same view before the cache fills.
- *Placeholders substituted by the use case after render.* This puts HTML knowledge in core.
- *A script-loaded SVG.* It fails offline from disk.

Amends 014 FR-012, recorded in ADR-0020.

## R2. Finding this element's node in the inlined SVG

**Decision**:
- The d2 SVG pass (020) adds `data-loko-address="<address>"` to every element and instance node
  group, next to its classes.
- When the HTML adapter inlines a diagram for an element page, it adds the class `loko-focus` to
  the one group whose attribute matches the page's address. This is a byte edit on that group's
  start tag.
- CSS fades every other `g[data-loko-address]` and every edge group, and outlines `.loko-focus`.

**Rationale**:
- 020's `g.<base64 path>` selector encodes d2's internal object path. Recomputing it in the HTML
  adapter would couple it to d2's naming.
- An address attribute is stable, readable and additive to the 020 contract
  (`contracts/svg-address.md`).
- Marking the group at build time works without scripts and needs no per-page `<style>`.

**Alternatives considered**:
- *A per-page `<style>` with an attribute selector.* It works too, but adds a style element per
  page and needs CSS-string escaping.
- *A `:target` rule inside the SVG.* It needs ids and URL fragments, and `<img>` support is
  inconsistent.

## R3. One inline SVG per page

**Decision**:
- Element pages inline their In context diagram.
- View pages inline the view.
- The index keeps the landscape as an `<img>`, because it needs no highlighting or zoom.
- A system or container page shows its own view as an `<img>` figure with an "Open view" link
  beside its Children list.
- At most one SVG is inlined per page, so d2's ids and `loko-icon-*` symbols never collide.
- When inlining, the adapter drops the XML declaration and the generated-file comment. The SVG gets
  `role="img"` and `aria-labelledby` pointing at the figure caption.

**Rationale**: d2 scopes its CSS under a per-diagram class, but symbol ids are global. One SVG per
page is the simplest way to rule out a collision. It also caps the size cost at one diagram per
page, about +24 KB.

## R4. Pan and zoom

**Decision**:
- The script changes the inline SVG's `viewBox`:
  - The wheel zooms around the cursor.
  - A click zooms in by 2× at the point clicked, unless the click lands on a link.
  - Dragging pans.
  - A reset button restores the original `viewBox`.
  - `+`/`-`/`0` work when the figure has focus.
- Without scripts, the SVG shows at its fitted width (`width:100%`, `height:auto`), and the
  "Open SVG" link opens it full size.

**Rationale**:
- It needs no library.
- Changing the `viewBox` keeps the vector sharp.
- It meets SC-009: a 6px label reaches 12px at 2× zoom.

**Alternatives considered**: CSS `transform: scale()`. It blurs on some engines and changes the
layout box.

## R5. Theme selection

**Decision**:
- Tokens are CSS custom properties.
- The light values sit on `:root`.
- The dark values apply under `@media (prefers-color-scheme: dark) { :root:not([data-theme=light]) … }`
  and under `:root[data-theme=dark]`.
- A five-line inline script in `<head>` reads `localStorage["loko-theme"]` (in `try`/`catch`) and
  sets `data-theme` before first paint.
- The toggle cycles between light and dark. It writes the choice, and it removes it when the
  choice matches the system.

**Rationale**:
- With no preference, scripts off, print or an old browser, the media query does not match and
  `:root` light applies (FR-024).
- The inline script avoids a flash of the wrong theme. A blocked `localStorage` is harmless.

## R6. Fonts

**Decision**: bundle 8 static, Latin-subset WOFF2 files.

| Family | Weights | Use | Source |
|---|---|---|---|
| Orbitron | 600, 700 | headings, title | upstream (OFL); instanced from the variable font |
| Source Sans 3 | 400, 400 italic, 600 | prose, leads, descriptions | upstream (OFL) |
| JetBrains Mono | 400, 600 | sidebar, labels, ids, chips, code | upstream (OFL); instanced |
| Audiowide | 400 | wordmark, stat numbers | `mdstn-branding/fonts` (OFL) |

- Subset: Basic Latin, Latin-1, Latin Extended-A common letters, general punctuation, `€`, `™`,
  and the brand glyphs `▸ ▾ → ›` (U+25B8, U+25BE, U+2192, U+203A).
- `tools/fonts/subset.sh` runs `fontTools` (instancer + `pyftsubset --flavor=woff2`) at development
  time. Its outputs and each family's `OFL.txt` are committed under `theme/fonts/`.
- Estimate: about 150–220 KB in total, against the 500 KB cap.
- Every stack ends in a system family (`system-ui`, `ui-monospace`) for characters outside the
  subset.
- `font-display: swap`.
- **Delivery.** The WOFF2 files are not published as separate files. The HTML backend generates one
  `assets/fonts.css`: the generated-file notice, then one `@font-face` per font with a
  `data:font/woff2;base64,…` source. Base64 adds about 33%, roughly 200–290 KB, still well under
  the 500 KB cap on the WOFF2 bytes.
  - **Why:** 014 FR-020 and `TestBuildNotice` require every generated file to open with the notice,
    and a binary WOFF2 cannot.
  - **Also:** the licences are published in one `assets/LICENSES.txt` under the notice, and the Eye
    SVG gets the notice as an XML comment.
  - **Icons:** the same applies to the site's icons. They are emitted as one generated
    `assets/icons.css`, with each Lucide file as a base64 CSS mask. The icon table moved to
    `viewmodel.IconFor`, so the diagram and the site share it, and a test keeps the site's copied
    icon files identical to the diagram's.
  - **Rejected:** exempting binary files from the notice rule. That would amend 014 and add an
    exemption list to a test that today has none.

**Rationale**:
- Static instances are smaller than variable fonts once subset.
- Keeping the build free of font tooling keeps it deterministic and dependency-free.
- `mdstn-branding` holds only Audiowide locally. Its token file references Orbitron and JetBrains
  Mono TTFs that are absent from the checkout, and Source Sans 3 is not in the brand repo at all.
  So three families are taken from upstream releases, with their OFL licences.

**Brand exception** (spec FR-030): the brand prescribes monospace body text. The site uses Source
Sans 3 for prose, for long-form legibility and to match the diagrams' Source Sans. This is recorded
in ADR-0020.

## R7. Navigation tree and pruning

**Decision**:
- The projection carries one full `Nav` tree: people, systems with containers with components, and
  externals, in address order within each kind.
- `viewmodel.NavFor(address)` returns the pruned copy for a page: the top level, the children of
  each ancestor of the page's element, and the element's own children.
- The index and view pages get the top level only.
- Collapsed nodes keep their link.
- The HTML adapter renders `nav` per page instead of once per depth.

**Rationale**:
- Pruning is model logic. It belongs in core and can be tested without HTML.
- Per-page rendering costs about 1,242 small template runs, which is negligible.
- Old `nav` overrides still get `.Projection.Views` and `.Root`.

## R8. Element page context in the projection

**Decision**: these new `ElementPage` fields are computed once per build in
`usecases/project_context.go` (details in `data-model.md`):

- `Ancestors`
- `Siblings` (with `Calls` and `CalledBy` flags)
- `Neighbours` (one per other element, with `Direction`, the union of the relationships'
  technology, the first description in relationship-address order, and `Sibling`)
- `Prev` and `Next`
- `ContextDiagram`
- `ProseInherited`
- `Source` (file and line)

`LinkRef` gains optional `Technology` and `Description` so that `Children` and `Parent` rows can
show them.

`Uses` and `UsedBy` stay unchanged for the markdown backend (FR-043).

The siblings of a top-level element are the other top-level elements of the same kind.

**`ContextDiagram`**:
1. The parent's own view, if one was produced.
2. Otherwise the first view, in projection order, that contains the element as a node.
3. Otherwise the landscape for top-level elements.
4. Otherwise empty.

This differs from `Diagram`, which prefers the element's own view and stays as it is.

**`ProseInherited`**: the element's resolved `docs` equals its parent's. Such a page shows no
prose, shows a link to the parent's page, and counts as having no prose (FR-008, FR-009).

## R9. Determinism and page dependencies

**Decision**:
- All new lists are sorted by address. `Prev` and `Next` follow the same order.
- Neighbour technologies are de-duplicated and sorted.
- A page's `Sources` grow to include the parent, siblings, neighbours and the context view's
  sources. The generated-file notice then names every file the page depicts, and incremental
  output marks the page changed when any of them changes.

**Accepted effect**: renaming or retitling a top-level element rewrites every page, because the
pruned tree lists it everywhere. View titles in the nav already behave this way. Recorded in
ADR-0020.

## R10. Theme block compatibility

**Decision**: keep every 014 block name and add new ones (`contracts/theme-blocks.md`).

- **Same data shape, new placement.** `element-tables` now renders the rail's Talks to section.
  `relation-table` renders `NeighbourRow` lists. `NeighbourRow` keeps the `RelationRow` field names
  an override can use (`Other`, `Description`, `Technology`, `Tags`), so old overrides still
  execute.
- **Kept with extended data.** `element-children` still receives `LinkRef` values, now with
  `Technology` and `Description`.
- **New behaviour inside the built-in block only.** The built-in `diagram` renders inline SVG when
  `.Inline` is set. An override that renders `<img>` keeps working, but loses highlighting and
  zoom.
- **Breaking for whole-file replacements.** An override of `style.css` or `site.js` now pairs with
  new markup. The changelog says so.
- **New file.** `rail.gohtml` is added to the overridable list.

**Rationale**: this is the clarified policy (spec Clarifications, FR-042). Overrides keep building,
and meaning changes are documented.

## R11. Contrast verification

**Decision**:
- `style.css` declares its colour tokens in one block per theme.
- A Go test parses `--token: #RRGGBB` declarations from the light and dark blocks. It checks a
  table of (text token, background token) pairs at 4.5:1 or better: body, muted, links, accent
  text, chip text and sidebar text.
- `--txt-4` must not appear in any `color:` declaration used for text. It may only be used for
  borders and decoration. The test greps for it.
- The overrides are `#128200` for accent text in light mode and `#2D66AE` for light links. They
  are defined in loko's `style.css`, not in the brand tokens.

**Rationale**: this makes SC-004 mechanical for colour. Focus rings and the manual Lighthouse pass
cover the rest (quickstart).

## R12. Overview counts and diagnostics

**Decision**: `Projection.Overview` holds:

- **Systems:** internal and external.
- **Containers:** total, and those with shape `database`, `queue` or `topic`.
- **Relationships:** total, and those of kind `async`.
- **Diagnostics:** open warnings and infos.

`Project` cannot see compile diagnostics. So `BuildArtifacts` sets `Overview.Warnings` and
`Overview.Infos` from the compiled diagnostics after `project.ignore_warnings` filtering, and
before rendering.

## R13. Print and narrow screens

**Decision**: one `@media print` block and one `@media (max-width: 760px)` block in `style.css`.

- **Print:**
  - It forces the light tokens.
  - It hides the sidebar, toggles, pager and zoom controls.
  - It flows the rail after the main column.
  - It sets `break-inside: avoid` on cards, rows and figures.
  - It drops the dot grid and shadows.
  - It prints the inline SVG at full width at its original `viewBox`. A `beforeprint` handler
    resets any zoom, and CSS sets the reset size without scripts.
- **Narrow:**
  - The grid collapses to one column.
  - The sidebar becomes an off-canvas panel behind a menu button.
  - With scripts off, the sidebar renders above the content.
  - Long ids wrap (`overflow-wrap: anywhere`).

## R14. Motion

**Decision**: the wordmark's terminal cursor runs a CSS animation with
`animation-iteration-count: 3`, about 3 seconds, and none under `prefers-reduced-motion: reduce`.
Nothing else animates. Hover and focus transitions are 0 under reduced motion.

## R15. Asset weight reporting

**Decision**: a test sums `theme/fonts/*.woff2` and fails above 500 KB. It also logs the size of
the generated `fonts.css`. Release notes report the total added weight per site: `fonts.css` +
glyph + `LICENSES.txt` + CSS + JS, measured on `examples/microservices`.

## R16. What stays unchanged

- `search-index.js` and its format. The search box moves into the new sidebar.
- The markdown backend and its goldens. The new fields are additive, and the markdown backend
  ignores `Upstream`.
- MCP and CLI outputs.
- `loko serve` live reload. It rebuilds through `BuildArtifacts` and so gets the inlined diagrams.
  The reload script injection is unchanged.
