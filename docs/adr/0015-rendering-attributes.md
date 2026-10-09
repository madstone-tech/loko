# ADR-0015: Rendering attributes for titles, shapes, relationship kinds and direction

**Status:** Accepted
**Date:** 2026-10-09
**Feature:** [016-rendering-fidelity](../../specs/016-rendering-fidelity/spec.md)
**Related:** [ADR-0013](0013-viewmodel-renderers.md) (outputs are projections),
[ADR-0014](0014-hcl-authoring.md) (assistants edit HCL)

## Context

Two architectures were modelled entirely through the MCP tools. Both compiled and answered queries
well, but their diagrams were worse than the hand-drawn ones they should replace:

- every container was the same rectangle;
- boxes showed raw identifiers;
- container views were too wide to read;
- triggers had to be drawn against the data flow to keep dependency queries right.

## Decision

1. **Plain optional attributes, no style language.** The language gains `title` (any element),
   `shape` (containers and externals), `kind` and `tags` (relationships), and `direction` (views).
   They live in the existing blocks, so there is nothing new to learn. Each is a closed set or a
   string, and colours and fonts stay the C4 palette.

2. **Rendering only.** None of the attributes changes references, resolution or queries. In
   particular, `kind = "trigger"` is written on the invoked element and targets its trigger, the
   same rule as every other relationship: the declaring element depends on its target. Only the
   drawn arrow reverses, because the swap happens during edge lifting, which only rendering uses.
   There is exactly one way to write a trigger.

3. **Async is not crossing.** Edges leaving a view were already dashed (`stroke-dash: 4`). Async
   edges use long dashes (`stroke-dash: 8`), so existing diagrams are unchanged and the two
   meanings stay distinct. View pages carry a legend.

4. **Top-down by default** for container, component and declared views. This is a one-time,
   deliberate change to existing output, recorded in the changelog. The landscape and deployment
   views keep left-to-right. A view can set `direction`.

5. **Titles use plain labels**: the title, then `[Kind: Technology] · name`, so the name stays
   visible for people who search the source. D2 markdown labels could make the name smaller, but
   they are embedded HTML (`<foreignObject>`) and render as empty boxes in PNG and PDF converters
   and many SVG tools; a diagram must render everywhere. Pages and markdown show the name smaller.
   Untitled nodes keep their labels byte for byte.

6. **ELK is opt-in** (added during housekeeping). `layout = "elk"` on the project sets the engine
   for every view, and a view's own `layout` overrides it. Dagre stays the default: d2 v0.7.1
   starts a fresh JS runtime per layout, and ELK's engine is the heavier, about 5× slower (0.2 s
   against 1.1 s on the serverless reference; 2.8 s against 13.7 s on 1,020 elements, over the
   10 s budget). The engine is written into the D2 source as `d2-config.layout-engine`, so the
   `.d2` artefacts lay out the same under the d2 CLI and the render cache keys on it for free.

## Consequences

- Projects that use none of the attributes export byte-identically, and render byte-identically
  apart from the `direction` line.
- Three new diagnostics: `invalid_attribute_value`, `shape_not_allowed`, `empty_title`.
- If the default layout of a dense view cannot meet the ratio limit (FR-006), a project can
  choose ELK. Projects that never set `layout` render exactly as before.

## Alternatives rejected

- **Shapes from tags.** Tags select elements for views, so a tag rename would change drawings.
- **A `style` block.** It invites per-element colours and fonts, which undermine a consistent C4
  look, and it is more than any test architecture needed.
- **Triggers declared on the trigger**, with queries inverting them. That gives two dependency
  directions to learn, and a special case in every query.
- **Restyling crossing edges.** That would change every existing diagram.
