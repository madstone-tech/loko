# Quickstart: validating feature 018

Each scenario proves one user story end to end. Language shapes are in
[contracts/language.md](contracts/language.md); diagnostic behaviour in
[contracts/cli-mcp.md](contracts/cli-mcp.md).

**Prerequisites**: `make build`, so `./loko` is this branch. Work on a copy of each fixture.

## 1. Describe deployment infrastructure (Story 1)

Fixture: a copy of `testdata/projects/serverless-reference`.

1. Give `node "aws"` a `title` and `technology`, and a nested node all four attributes. Give one
   `dev` instance `technology = "local process"`.
2. `loko build --format d2,svg,html`: the `deployment-prod` diagram labels those nodes
   `<title>` / `[<technology>] · <name>`; no description is drawn.
3. `deployment-dev`: the instance shows `local process`; in `deployment-prod` the same element shows
   its own technology.
4. The `deployment-prod` page has a "Nodes" table with the description and tags; a deployment page
   whose nodes have no attributes has none.
5. `loko export`: groups carry the new fields; the unmodified fixture exports byte-identically to
   1.1.0.
6. MCP `schema` with `block: "node"` lists the four attributes; `apply_edit` sets and removes `title`
   on a group and `technology` on an instance, with preview.
7. `title = ""` on a node: compile error.

## 2. Views that say what they show (Story 2)

1. Give one declared view a `title` and a `description`; build HTML.
2. The view's page heading, browser title and navigation entry show the title; its URL and file
   name still use the name.
3. The description is under the heading.
4. Open `diagrams/<view>.svg` on its own: the title is drawn at the top; the description is not.
   Same with `layout = "elk"`.
5. A view without a title: its `.d2` and `.svg` are byte-identical to 1.1.0.

## 3. Docs links that survive moving files (Story 3)

Fixture: `testdata/projects/first-run` (root-relative paths from `architecture/`).

1. `loko validate`: 7 `docs_root_relative` info lines, each with the file-relative path to write;
   `--strict` exits `0`; warning count unchanged.
2. Rewrite one path in the suggested form: its hint disappears, the site still links the document.
3. Move `architecture/containers.loko.hcl` and `architecture/docs/` together into
   `architecture/pricing/`: the rewritten path still resolves.
4. Put a different file at the root-relative location too: `docs_ambiguous` names both; the
   file-relative one is shown on the site.
5. Point a path at nothing: `docs_not_found` names both locations checked.
6. `loko validate --ignore docs_root_relative`: the hints are gone.

## 4. Hint when a view renders too wide (Story 4)

1. `loko build --format svg` on `testdata/projects/first-run`: one `wide_view` info for
   `system-pricing` naming 3875 × 1441 and suggesting `layout = "elk"` on the project.
2. Add `layout = "elk"` to its project: no hint, and the diagram is narrower.
3. `loko build --format d2`: no hint. `loko validate`: no hint.
4. `--strict` exit codes match the run without the hint; `--ignore wide_view` removes it.
5. MCP `build` returns the hint in `diagnostics`; MCP `validate` with `severity: "warning"` hides
   information.

## Regression

`make test`, `make lint`, `make audit-constitution`. The determinism job, the export goldens and
the D2 and HTML goldens of fixtures that do not use the new attributes pass unchanged (SC-006).
