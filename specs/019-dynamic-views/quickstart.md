# Quickstart: validating feature 019

Each scenario proves one user story. Language rules are in
[contracts/language.md](contracts/language.md); tools in [contracts/cli-mcp.md](contracts/cli-mcp.md).

**Prerequisites**: `make build`, so `./loko` is this branch. Work on copies of fixtures.

## 1. Describe an ordered flow (Story 1)

On a copy of `testdata/projects/first-run`, add the report's flow in a new file
`architecture/flows.loko.hcl`: `view "hourly_run"` with steps `start` (timer → fetcher,
`via` the declared relationship), `fetch` (fetcher → the carrier API system, description
"Fetch scans") and `store` (fetcher → event store).

1. `loko validate`: no new error.
2. `loko build`: `diagrams/hourly_run.d2` has three edges labelled `1. …`, `2. Fetch scans`,
   `3. …`; the fetcher and timer sit inside the `tracking` system boundary.
3. `view/hourly_run.html` has a "Steps" table, numbered 1–3, linking each element.
4. `loko export`: `views[].steps` lists the three steps in order with resolved addresses.
5. The unmodified fixture builds byte-identically to 1.2.0.

## 2. Steps the compiler checks (Story 2)

1. Change a step's `to` to `container.collectr`: `unresolved_reference` at the step, suggesting
   `container.fetcher`.
2. Point `via` at a relationship between other elements: `step_mismatch` naming both pairs.
3. Add a step between two unrelated elements: `step_unmatched` warning; `build` still draws it;
   `--strict` exits `2`; `--ignore step_unmatched` removes it.
4. Add `include = [system.tracking]` to the view: `dynamic_view_selection`.
5. Rename `container.fetcher` with MCP `move`: every step naming it is rewritten.

## 3. A sequence diagram (Story 3)

1. Set `style = "sequence"` and build: the SVG shows lifelines timer, fetcher, carrier API,
   event store, in that order, and three numbered arrows top to bottom.
2. Add a `reply = true` step: drawn dashed in both styles.
3. The page's Steps table is unchanged from the edges style.

## 4. Through MCP (Story 4)

1. `schema` with `block: "step"`: attributes, `from`/`to` required, example.
2. `apply_edit`: add a view, add three steps, insert one with `before`, remove the last, each
   previewed first. `build`: the numbering matches.
3. `describe`: the view's `steps`. `query kind=flows` on the fetcher: its steps.

## Regression

`make test`, `make lint`, `make audit-constitution`. Determinism job, export goldens and the D2 and
HTML goldens of fixtures without steps unchanged (SC-005). A 40-step flow builds within the budget
(SC-006).
