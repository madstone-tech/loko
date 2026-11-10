# Contract: Rendering defaults (1.6.0)

**Feature**: `023-v160-theming-mcp` | **Status**: normative. Recorded in ADR-0022.

## R-1 Default layout engine

- The projection always resolves a view's layout: the view's `layout`, else the
  project's `layout`, else `elk`. The view model holds `"elk"` or `"dagre"`, never
  an empty value.
- The d2 emitter always writes `vars.d2-config.layout-engine`, so a `.d2` file
  renders the same under the plain d2 CLI, whose own default is dagre.
- `layout = "dagre"` on the project or the view gives the 1.5 drawing.
- Sequence-style dynamic views ignore the layout, as in 1.5.
- `projection.json` and `describe` show the resolved layout of each view.

## R-2 Wide-view hint text

The hint fires under the same thresholds as 1.5 (at least 3,000 px wide and at
least 2.5 times wider than tall), and only a view's own `direction` or `layout`
suppresses it.

| Case | Advice |
|---|---|
| Generated view, no project layout | `The project uses the default layout, elk. Set layout = "dagre" on the project to try the other engine, or declare a view of these elements and give it a direction.` |
| Generated view, project layout set | Unchanged from 1.5: `The project already sets layout = "<engine>". Declare a view of these elements and give it a direction.` |
| Declared view, no view layout | `Try direction = "right" or layout = "dagre" on the view.` |

## R-3 Outside endpoints

- The `project` block and each `deployment` block accept `outside`, with the
  values `expanded` and `collapsed`. A `deployment` value wins over the project
  value. The default is `expanded`. The `view` block does not take `outside`:
  declared views are never deployment views.
- `expanded`: each element outside the view that an edge reaches becomes its own
  node. A component endpoint lifts to its container, to match instance
  granularity. The node id is `outside__<node id of the element>`, the label is
  the element's title (else its name), and the role and style are the 1.5
  outside ones. The node has no address, so the SVG carries no
  `data-loko-address` for it (021 svg-address contract). Edges merge per inside
  node, outside element and direction.
- `collapsed`: all such edges end at one node with the 1.5 id (`outside`), label
  and style.
- In 1.6.0 the setting applies to deployment views only. System, container and
  declared views keep the 1.5 single outside node.
- The html crossing list and the markdown "Leaves this view" list find the
  direction by the endpoint's role. The collapsed text stays byte-identical to
  1.5. Expanded text reads `api → Ledger (outside): Records payments`.

## R-4 Performance

- elk costs about 4 times the dagre build time. Measured on the
  `cmd/build_perf_test.go` project (1,020 elements, 3 runs each): dagre 3.2–3.7 s,
  elk 14.7 s.
- `TestPerformanceBuild` sets `layout = "dagre"` on its generated project. It
  keeps the 10 s full-build and 2 s edit-to-browser budgets for that path.
- The language reference and the CHANGELOG tell authors of very large models to
  set `layout = "dagre"` for faster builds.
