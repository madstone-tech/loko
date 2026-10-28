# ADR-0020: Site redesign: upstream artefacts, inline diagrams and the Madstone theme

**Status:** Accepted
**Date:** 2026-10-28
**Feature:** [021-site-redesign](../../specs/021-site-redesign/spec.md)
**Related:** [ADR-0013](0013-viewmodel-renderers.md) (backends translate, the projection decides),
[ADR-0019](0019-diagram-look.md) (diagram look)

## Context

The generated site's container and component pages gave no context outside the diagram. The
sidebar listed only views. "Part of" went up one level. Children were not listed. Uses and Used by
repeated a neighbour once per relationship. Lists showed raw ids. The dark theme's links were hard
to read, and wide diagrams shrank their labels to about 6px.

The owner wanted an element page to say how the element relates to its parent and siblings, and
what it does. Approved mockups set the layout: prose first, with a context rail. They also set the
Madstone branding.

## Decision

1. **Upstream artefacts** (research R1; amends 014 FR-012).
   - A backend whose format requires another receives `RenderOptions.Upstream`, a function that
     blocks until the backends it depends on have finished and returns their artefacts.
   - All backends still start together. The HTML backend renders its pages with a placeholder
     where a diagram goes, then asks for the SVGs and fills the placeholders. A first version
     made HTML wait for SVG before starting, which doubled `loko serve`'s rebuild time on the
     1,020-element benchmark and broke its 2 s budget on CI.
   - The HTML site needs the SVG bytes to highlight an element and to zoom. An `<img>` cannot be
     styled, and a script cannot reach a separate SVG from `file://`.
   - `Format.Requires` already made HTML depend on SVG. Backends still never import each other.
   - Rejected: a renderer port for the HTML backend, which risks rendering a view twice. Also
     rejected: placeholders filled in by the use case, which would put HTML knowledge in core.
2. **`data-loko-address` on SVG cards** (research R2; contracts/svg-address.md).
   - The SVG pass writes each card's model address after its class.
   - The site marks the page's own card with `loko-focus` at build time, so highlighting needs no
     script.
   - Rejected: 020's `g.<base64 path>` selector, which would couple the site to d2's internal
     naming.
3. **One inline SVG per page** (research R3).
   - An element page inlines its In context view, and a view page inlines its own diagram. System
     and container pages show their own view as an image.
   - With one SVG per page, d2's ids and the icon symbols cannot collide.
   - The cost is about 24–48 KB per element page.
4. **Brand typography exception** (spec FR-030, research R6).
   - The brand prescribes monospace body text. Documentation pages use Source Sans 3 for prose,
     for long-form legibility and to match the diagrams' Source Sans.
   - Orbitron is used for headings, JetBrains Mono for labels, metadata, ids and code, and
     Audiowide for the wordmark and stat numbers.
5. **Fonts delivered as a generated `fonts.css`** (research R6, 014 FR-020).
   - The 8 Latin-subset WOFF2 files (75 KB) travel as base64 inside a generated stylesheet, and
     the icons inside a generated `icons.css`.
   - 014 requires every generated file to open with a notice, and a binary WOFF2 cannot carry one.
   - The licences are published as one `LICENSES.txt` under the notice. The fonts are subset at
     development time by `tools/fonts/subset.sh`, never by `loko build`.
6. **Theme block compatibility** (research R10).
   - Every 014 block name still exists and is still called. New blocks are added.
   - Blocks whose content changed (`header`, `nav`, `element-tables`, `relation-table`,
     `element-children`, `diagram`, `index`) keep data an old override can read. For example, a
     neighbour row keeps `Other`, `Description`, `Technology` and `Tags`, and the built-in
     `relation-table` accepts both row types.
   - Whole-file `style.css` and `site.js` replacements meet new markup. The changelog says so.
7. **Wider page dependencies** (research R9).
   - An element page shows its parent, siblings, neighbours and the parent's diagram. Its notice
     therefore names their files (`ElementPage.SiteSources`), and a change to one of them rewrites
     the page.
   - Renaming a top-level element rewrites every page, because the tree lists it everywhere. View
     titles in the sidebar already behaved this way.
   - On the small `two-systems` fixture, changing one description now rewrites 8 of 69 files. That
     is 88% unchanged against 014 SC-003's 90%, and the test now expects at least 85%.
   - Markdown pages keep `Sources` as before and do not change.
8. **WCAG overrides stay in loko** (spec FR-032).
   - In light mode, links use `#2D66AE`, and accent text uses `#128200`.
   - The faintest text tone is decoration only.
   - These live in loko's stylesheet and are checked by `contrast_test.go`. The brand tokens are
     unchanged.

The Lucide icon table moved from the d2 adapter to `viewmodel` (`IconFor`), so the diagram and the
site draw the same icon. The site embeds copies of the icon files, and a test keeps them identical.

## Consequences

**Positive:**
- An element page answers what it is part of, what sits beside it, what it talks to and what it
  does, without the reader studying a diagram.
- The site works offline and from disk, prints cleanly, works on a phone, and scores 100 in
  Lighthouse accessibility on the pages checked.
- Old theme overrides keep building.

**Negative:**
- The site is larger. Assets grow about 145 KB per site, and HTML grows about 24–48 KB per element
  page. On the 1,020-element benchmark the output grows from 42 MB to 90 MB, and the build time
  grows from 3.0 s to about 3.3 s. A one-description rebuild through `loko serve` goes from
  0.37 s to 0.46 s.
- A one-element change rewrites more site pages than before.
- The HTML backend depends on the SVG backend's output. It still starts at the same time, but
  cannot finish first.

**Mitigations:**
- The font budget (500 KB) is enforced by a test.
- The performance test still guards the 10 s build and the 2 s edit-to-browser budget.
- The sidebar's views list is rendered once per page depth and current view, and the tree is
  built in Go rather than by a recursive template, so per-page navigation stays cheap.
- Markdown output is unchanged, byte for byte.
