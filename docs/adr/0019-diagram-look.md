# ADR-0019: Diagram look: element cards with icons, drawn by a post-render SVG pass

**Status:** Accepted
**Date:** 2026-10-27
**Feature:** [020-diagram-look](../../specs/020-diagram-look/spec.md)
**Related:** [ADR-0013](0013-viewmodel-renderers.md) (backends translate, the projection decides),
[ADR-0015](0015-rendering-attributes.md) (titles and shapes)

## Context

The rendered SVG diagrams worked, but looked crude. The stock d2 shapes for people, databases,
queues, topics and functions were plain outlines. Every line of an element's label (name,
`[Kind: Technology]`, description) was drawn in the same font, size and colour, so nothing stood
out. A spike on `examples/microservices` showed a cleaner default without a new dependency: each
element as a card with a small line icon, and the label lines styled as a hierarchy.

## Decision

1. **Every element is a rounded card with an icon.** The emitter draws every element and instance
   as a `rectangle` with `border-radius: 6`. A Lucide icon in the card's top-left corner says what
   the element is: its declared `shape` first, its kind second. The ten icons are vendored
   unmodified from `lucide-static` (ISC) and ship as `LICENSE-lucide`. Boundaries and the outside
   node keep their shape.

2. **Icons are added after d2 renders, not with d2's `icon:`.** d2 pads any node that has an icon
   by its label's height on both axes, which made the spike's cards oversized and broke the 10%
   size limit. The emitter returns a *label plan* beside the D2 source: each card's icon and
   text colour, and the role of every label line. One linear pass over the rendered SVG then adds
   a `<defs>` of `<symbol>`s, one `<use>` per card, and an inline style per `<tspan>`. No XML
   parser is used. The pass finds each group by its first class, the base64 of d2's ID.

3. **Line roles, not text matching.** The emitter knows which line is the title, the kind line or
   the description, so it records the role. The pass cannot infer it: titles wrap over several
   lines, and an edge description may contain a newline. The plan is therefore part of the
   render-cache key, beside the D2 source.

4. **A lighter line changes font family, not weight.** d2 embeds each style as its own
   single-face family. `font-weight` cannot select a lighter face, and it can make a renderer
   synthesise extra bold. The description therefore sets d2's regular family inline, and no role
   sets `font-weight`. When d2 drew no regular text, `compile` adds d2's regular `@font-face`.
   Lines are only ever made smaller or lighter than d2 measured, so they never overflow.

5. **The kind line is written by the emitter.** `CONTAINER · Go, gRPC`: the kind in capitals, the
   technology as written, and a titled element's name at the end. This replaces ADR-0015's
   `[Kind: Technology] · name`. Upper-casing in the emitter keeps the `.d2` output and viewers that
   ignore CSS correct.

6. **The palette moves to the Madstone Signal Blue ramp.** The C4 order stays: person and system
   darkest, container in between, component lightest; externals are dark neutral and dashed.
   Edges and boundaries use brand neutrals (`#737373`, `#3D3D3D`), and the green accent is never
   used. Every text line on a card passes WCAG AA at its opacity, and a test checks it.

7. **The declared shape stays in the `.d2` output as a class.** A shaped node carries
   `shape-<shape>` beside `kind-*` and `tag-*`. Constitution Principle VI asks that a new default
   that removes information keep the previous shape available on request. The class keeps the
   information, so nothing is removed; only the glyph changes, from a cylinder to a card with the
   database icon. The `.d2` output carries no icons and no line styles: those are SVG-only.

8. **The icon table is a translation.** The projection decides what an element is (its `Shape`
   and kind). The SVG backend maps that to its own glyph, as it maps `ShapeOval` to d2's `oval`.
   This keeps ADR-0013's rule that backends decide nothing. The palette, the icon table and the
   line styles are each one table, so a later customisation feature can replace them.

## Consequences

**Positive:**
- Diagrams read at a glance: name, then kind and technology, then description, with an icon.
- View sizes do not grow: d2 sizes every card as before (largest change on the examples: +2.8%).
- SVGs stay self-contained: no network, no scripts, no `foreignObject`; text stays selectable.
  Chromium, `rsvg-convert` and Inkscape all draw the icons.

**Negative:**
- Every diagram and every golden file changes.
- The pass depends on how d2 v0.7.1 structures its SVG (flat groups, base64 IDs, one `<tspan>` per
  line). A d2 upgrade can break it.
- SVG files grow by 9.6–12.7% on the examples and by 16.6% on the 1,020-element benchmark, whose
  four-character labels make the per-card cost large. About a third of it comes from the capital
  letters of the kind line, which enlarge every embedded font subset. SC-005 allows 20%.
- Under the d2 CLI, a `.d2` file draws plain cards: the database is no longer a cylinder there.

**Mitigations:**
- Tests find every planned group in real d2 output and check one icon per card and one style per
  line, so a d2 change that breaks the pass fails the build.
- The `shape-<shape>` class keeps the shape in the `.d2` source and the SVG for tools and CSS.
