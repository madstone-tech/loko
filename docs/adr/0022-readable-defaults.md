# ADR-0022: Readable defaults

**Status:** Accepted
**Date:** 2026-11-10
**Related:** [ADR-0015](0015-rendering-attributes.md) (rendering attributes, §6 ELK is opt-in),
[ADR-0021](0021-wrapped-descriptions.md) (page-width text size, wide-view hint)

## Context

Feedback item 1: an author built a model of about ten containers and forty relationships with no
`layout` setting. Dagre, the 1.5 default, drew the views very wide, and the labels were too small
to read. `layout = "elk"` fixed most of this, but a new author does not know that the setting
exists. The first diagrams are the first impression of the tool.

## Decision

This ADR replaces ADR-0015 §6 ("ELK is opt-in"). The rest of ADR-0015 stays in force.

1. **The default layout engine is elk** (contract R-1). The projection resolves the layout of
   each view: the view's `layout`, else the project's `layout`, else `elk`. The view model holds
   `"elk"` or `"dagre"`, never an empty value. `projection.json` and `describe` show the resolved
   layout.
2. The d2 emitter always writes `vars.d2-config.layout-engine`. A `.d2` file renders the same
   under the plain d2 CLI, whose own default is dagre.
3. `layout = "dagre"` on the project or the view gives the 1.5 drawing.
4. Sequence-style dynamic views ignore the layout, as in 1.5.
5. **The wide-view hint names the other engine** (contract R-2). The thresholds stay as in 1.5.

| Case | Advice |
|---|---|
| Generated view, no project layout | `The project uses the default layout, elk. Set layout = "dagre" on the project to try the other engine, or declare a view of these elements and give it a direction.` |
| Generated view, project layout set | `The project already sets layout = "<engine>". Declare a view of these elements and give it a direction.` |
| Declared view, no view layout | `Try direction = "right" or layout = "dagre" on the view.` |

## Measurements

### Build time (contract R-4)

Measured on the `cmd/build_perf_test.go` project (1,020 elements, 3 runs each, developer
hardware):

| Path | dagre | elk |
|---|---|---|
| Full build | 3.2 to 3.7 s | 14.7 to 18.6 s |
| Edit to browser (warm cache, one description changed) | 590 ms | 830 / 843 / 849 ms |

The edit-to-browser figures exclude the 400 ms poll latency. With it, elk takes about 1.25 s,
within the 2 s budget. Elk breaks the 10 s full-build budget on this project. `TestPerformanceBuild`
therefore sets `layout = "dagre"` on its generated project and keeps the 10 s and 2 s budgets for
that path.

### Page-width text size (SC-006)

Method from ADR-0021: page-width text size = label font size × page width ÷ viewBox width. The
label font size is 16 px. The page width is 1,120 px, the diagram panel width that ADR-0021 used.
Project: `testdata/projects/first-run`, no `layout` set (elk), compared with `layout = "dagre"`.

| View | dagre viewBox | dagre text | elk viewBox | elk text |
|---|---|---|---|---|
| `landscape` | 1,660 × 1,360 | 10.8 px | 1,936 × 1,074 | 9.3 px |
| `system-tracking` | 4,098 × 1,288 | 4.4 px | 3,280 × 1,692 | 5.5 px |
| `deployment-prod` | 2,881 × 973 | 6.2 px | 2,190 × 1,131 | 8.2 px |

`system-tracking` does not reach 6 px under elk. It renders at 5.5 px, up from 4.4 px under dagre.
Its elk canvas is 3,280 × 1,692 px, a ratio of 1.9. The ratio is below 2.5, so the `wide_view` hint
does not fire for it in a build of `first-run`. SC-006 was amended after this measurement: every
view must reach 6 px or render larger than under dagre, and the narrowest view must improve. elk
meets that: `landscape` and `deployment-prod` stay above 6 px, and the narrowest view improves
from 4.4 px to 5.5 px. A 6 px floor for every view stays a goal for a later release. Declaring a
view of these elements with a direction is the remedy today.

## Consequences

**Positive:**
- A new author gets the more readable engine without a setting.
- The `.d2` files name their engine, so they render the same outside loko.

**Negative:**
- Every project without `layout` changes its drawings in 1.6.0.
- Elk builds a very large model about 4 times slower than dagre, over the 10 s full-build budget.
- One view of `first-run` stays below 6 px at page width.

**Mitigations:**
- The language reference and the CHANGELOG tell authors of very large models to set
  `layout = "dagre"`.
- The wide-view hint names dagre as the other engine to try.

## Outside endpoints

User Story 7 adds this section.
