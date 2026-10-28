# Manual Verification: Site Redesign

**Purpose**: the browser checks from quickstart §5–§6 that automated tests do not cover.
**Run**: 2026-10-28, headless Chromium through Chrome DevTools, against `loko serve` on
`examples/microservices` (live-reload check on a scratch copy).

## Checked

- [x] Lighthouse accessibility 100 on a component page (dark) and a declared view page (light).
  The only failure was a missing meta description, since added (`head` block).
- [x] Dark and light follow `prefers-color-scheme`. The toggle switches the theme, and the choice
  persists across pages and reloads. Toggling back to the system theme removes the stored choice.
- [x] In context: one inline SVG per element page, exactly one `loko-focus` card, other nodes
  faded (all 23 component and container pages in microservices).
- [x] Zoom: four wheel steps on the 7,026-px-wide landscape take a label from 11px to 23px
  (SC-009). The toolbar appears with scripts.
- [x] 375px width: no horizontal scroll, rail above the prose, menu button shown, sidebar hidden
  (SC-008, FR-038).
- [x] Arrow keys move to the next sibling, and do nothing while the search box has focus (FR-004).
- [x] Reduced motion: the cursor animation is `none` (FR-033).
- [x] No external requests on a view page (SC-007, `performance.getEntriesByType("resource")`).
- [x] Live reload: editing a component's description in `*.loko.hcl` reloads the open page with
  the new lead and the new text in the highlighted card (FR-041).

## Checked by the project owner (2026-10-28)

- [x] Print preview of an element page and a view page (FR-037), alongside `TestPrintStyles`.
- [x] Scripts disabled: system theme or light, no toggle, tree links work (FR-024, US4-4).
- [x] Opening the built site from `file://` with the network off.
- [x] Firefox and Safari.
- [x] No-preference colour scheme falls back to light, alongside `TestThemeDefaultsLight`.
