# ADR-0021: Wrapped description lines and the wide-view hint scope

**Status:** Accepted
**Date:** 2026-10-28
**Related:** [ADR-0019](0019-diagram-look.md) (diagram look, line roles, title wrapping),
[ADR-0020](0020-site-redesign.md) (site, diagram panel)

## Context

A real change-heavy modelling session (the PCR competitive-pricing architecture, 8 elements,
13 relationships, `layout = "elk"` on the project) produced diagrams a reader could not read:
text rendered at 1.4–2.8 px inside the site's diagram panel. Measuring the produced SVGs showed
why:

- A card's description is emitted as one d2 label line per authored line (only titles are
  wrapped, at 40 characters). d2 does not wrap plain labels, so one 282-character description
  drew a ~1,960 px card. Every description-bearing card was among the widest objects in its
  canvas.
- The canvases were ribbons: the landscape rendered 6,366 × 669 px (ratio 9.5); every view of
  this project was at least 3,000 px wide with a ratio above 2.5.
- The site's diagram panel already stretches the SVG to its full width, so the CSS was not the
  cause. At panel width ~1,120 px, the scale factor was 0.09–0.18, and text rendered at
  16 px × scale.

The built-in `wide_view` information diagnostic never fired for this project, because
`WideViewHints` returned nothing whenever the *project* set a layout ("the author chose"). A
project layout, however, only sets the engine for views that never declared one; it says nothing
about any view's shape. The session's author never learned the canvas was a ribbon.

## Decision

1. **Descriptions wrap at 50 characters** (`descWidth`, beside `titleWidth` = 40). The break is
   the same word-boundary rule titles use, per authored paragraph: an existing line break is an
   author's formatting and is kept. New lines are emitted in the same `roleDescription`, so the
   SVG pass styles them identically and no line-role or contrast behaviour changes.

2. **Only a view's own choice suppresses its wide-view hint.** A declared view with a `direction`
   or `layout` gets no hint (unchanged). The project's layout no longer suppresses hints, for
   generated or declared views.

3. **The hint does not suggest an engine the project already sets.** With no project layout, a
   generated view's hint still names `layout = "elk" on the project`. With one set, it says the
   project already sets it and names the remaining lever: declare a view of the elements and give
   it a direction.

## Consequences

**Positive:**
- On the model that motivated this ADR, every view's canvas shrank and squared. Page-width text
  went from 1.4–4.6 px to 3.0–11.4 px: the landscape 6,366 × 669 → 2,724 × 855 (text 2.8 → 6.6 px),
  container-matcher 3,919 × 1,902 → 1,766 × 2,176 (4.6 → 10.1 px), phase_2 4,283 × 1,317 →
  1,567 × 1,491 (4.2 → 11.4 px).
- The wide-view hint now reaches the projects that are most at risk of wide views — those that
  already pinned an engine project-wide — and its advice is always actionable.

**Negative:**
- Cards with a long description are taller: line count grows from 1 to ceil(len/50) per
  description. Text still never overflows (lines are only ever made no wider than d2 measured for
  the same size class, per ADR-0019's rule).
- Descriptions wrap, so a canvas grows in height where it shrank in width; total canvas area is
  similar or smaller.
- A project that relied on the project layout suppressing `wide_view` sees the hint again. The
  code is information and never changes an exit code; `--ignore wide_view` restores the old
  output.

## Verification

- `internal/adapters/d2`: `TestWrapDescription` (budget, no truncation, kept breaks) and the
  existing line-role tests pass unchanged; the d2 goldens did not change (their fixtures carry
  short descriptions).
- `internal/core/usecases`: `TestWideViewHints` pins the new suppression rule and wording.
- The motivating model rebuilds with the same diagnostics (`0 errors`), and the measured
  before/after table above was produced by rendering it before and after this change.