# Quickstart: validating the site redesign

## Prerequisites

- `go build -o /tmp/loko .` from the repo root.
- A current Chromium-based browser and Firefox.
- For the font step only: Python 3 with `fontTools` and `brotli`.

## 1. Automated checks

```sh
go test ./...                                  # projection, theme compat, contrast, font budget, goldens
go test ./cmd -run TestPerformanceBuild -v     # 1,020 elements < 10 s (SC-005)
make lint                                      # depguard, archcheck budgets
```

Expected:
- All tests pass.
- The performance log prints a build time under 10 s.
- The font-budget test logs the total in KB (500 or less).

## 2. Element page context (US1, SC-001, SC-003)

```sh
cd examples/microservices && /tmp/loko build --format html && open dist/index.html
```

Open a component page under the gateway and check:
- The header label reads `COMPONENT · IN <PARENT>`, with the icon tile in the kind colour.
- The metadata line reads technology · id · `file:line`.
- The rail order is Part of → Siblings (current marked, calls / called by) → Talks to → In context.
- Talks to lists each neighbour once.
- In context fades every node except this one, which is outlined.
- Because this component's `docs` is its parent's file, the main column shows In context and a link
  to the parent's prose (FR-009).

Open the container page and check that it lists its children as rows.

## 3. Navigation (US2, SC-002)

- The tree highlights the current element with its branch open. View source: only the top level
  and the current path's children are present.
- The breadcrumb reads `<project> › <system> › <container> › <component>`.
- `→` and `←` move between siblings and stop at the ends. Inside the search box they move the
  cursor only.

## 4. Overview and view pages (US3, SC-009)

- The index shows four stat cards, the landscape figure with "Open view", and system and view
  rows. No raw ids appear except in metadata.
- On the widest view, wheel-zoom until labels measure at least 12px in devtools. Then drag, then
  reset.
- The "In this view" list names every element.
- A dynamic view (`examples/…` with steps) still shows its Steps table.

## 5. Theme (US4)

| Setup | Expected |
|---|---|
| OS dark | dark theme |
| OS light | light theme |
| devtools: emulate `prefers-color-scheme` none, or scripts disabled with no preference | light |
| toggle → reload → other page | choice kept |
| scripts disabled | page renders in the system theme; no toggle; all links work |

## 6. Accessibility, print, narrow, offline (US5, SC-004, SC-007, SC-008)

- Run a Lighthouse accessibility audit on the index, a component page and a view page in both
  themes. Expect no contrast or focus failures.
- Tab through a page and check that every control shows a focus ring.
- With reduced motion emulated, the cursor does not blink. Without it, it stops after about 3 s.
- Print preview an element page and a view page. Expect a light palette, no sidebar, rail under
  the prose, the diagram whole and full width, and no split cards.
- At 375px width, nothing scrolls sideways, the menu button opens the sidebar, and the rail stacks
  under the description.
- Disable the network and reload. The Network panel shows zero external requests, and the fonts
  render.

## 7. Theme compatibility (SC-010)

```sh
go test ./internal/adapters/html -run 'Theme|Override' -v
```

Every 014 override fixture builds unchanged. Then add `templates/custom.css` to an example and
check that it still loads after `style.css`.

## 8. Determinism

```sh
/tmp/loko build --format html --out /tmp/a && /tmp/loko build --format html --out /tmp/b && diff -r /tmp/a /tmp/b
```

Expect no differences.

## 9. Regenerating fonts (maintainers only)

```sh
tools/fonts/subset.sh   # writes internal/adapters/html/theme/fonts/*.woff2; commit the result
```
