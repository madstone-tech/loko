# Contract: Theme blocks (amends 014 `contracts/theme.md`)

The resolution rules of 014 are unchanged:

- An override file replaces only the blocks it defines.
- An unknown block name is reported with its file.
- An execution error is attributed to the overrides.

This contract lists every block after 021 and what each receives.

## Compatibility rule (spec FR-042)

1. Every block name from 014 still exists and is still called by the built-in layout. An override
   that defines only 014 block names builds without changes.
2. When a block's placement or data changes, the change is listed in `CHANGELOG.md` and ADR-0020.
   Data stays field-compatible: an override that reads fields 014 documented keeps executing.
3. New blocks may be added. Overriding them is optional.

## Overridable files

| File | Blocks | Status |
|---|---|---|
| `layout.gohtml` | `layout`, `head`, `header`, `nav`, `footer`, `breadcrumb`, `pager`, `theme-toggle` | `breadcrumb`, `pager`, `theme-toggle` new |
| `index.gohtml` | `index`, `stat-cards`, `row-list` | `stat-cards`, `row-list` new |
| `element.gohtml` | `element`, `element-tables`, `relation-table`, `element-children`, `prose-missing`, `prose-inherited` | `prose-inherited` new |
| `rail.gohtml` | `rail`, `part-of`, `siblings`, `in-context` | new file |
| `view.gohtml` | `view`, `in-this-view` | `in-this-view` new |
| `partials.gohtml` | `tag-chips`, `link`, `diagram`, `icon-tile`, `chip` | `icon-tile`, `chip` new |
| `style.css` | replaces the built-in stylesheet | unchanged rule, new markup (see below) |
| `custom.css` | empty, loaded after `style.css` | unchanged |
| `fonts.css`, `icons.css` | not overridable: generated from the bundled fonts and icons and loaded before `style.css` | new, generated |
| `site.js` | replaces the built-in script | unchanged rule, more behaviour (see below) |

## Changed meaning (listed in the changelog)

| Block | 014 | 021 |
|---|---|---|
| `header` | `h1` + subtitle | accent bar, kind label, icon tile + title, metadata line, lead. Data adds `.Element`/`.View` fields only. |
| `nav` | rendered once per page depth; lists views | rendered per page; adds the pruned tree (`.Tree`), project description, footer note. `.Projection.Views` and `.Root` are still present. |
| `element-tables` | Uses and Used by tables in the main column | the rail's Talks to section. `.Element.Uses`/`.UsedBy` are still present. |
| `relation-table` | receives `{Root, Rows []RelationRow}` | receives `{Root, Rows []NeighbourRow}`. `Other`, `Description`, `Technology` and `Tags` keep their meaning. |
| `element-children` | "Contains" list of links | Children row list. Items are still `LinkRef`, now with `Technology` and `Description`. |
| `diagram` | `<img>` in a figure | inline SVG with zoom controls when `.Inline` is set, else `<img>`. Used by the view page and the index. |
| `index` | description, landscape, raw element table | stat cards, landscape figure, system and view row lists |

## Page data (`pageData`, additions only)

| Field | Type | On |
|---|---|---|
| `Tree` | []NavNode | every page (pruned, research R7) |
| `Crumbs` | []LinkRef | every page; empty on the index |
| `Inline` | template.HTML | element pages (In context) and view pages; empty when no diagram |
| `Label` | string | kind label above the title, e.g. `COMPONENT · IN GATEWAY`, `SYSTEM VIEW · AUTOMATIC` |
| `Meta` | []string | metadata line items (technology, id, `file:line`, counts) |

## Whole-file replacements

A user `style.css` or `site.js` written for 014 still loads, but it now meets new markup and class
names. Expect broken layout or missing behaviour (theme toggle, tree, pan/zoom, paging). The
changelog says so. Additive styling belongs in `custom.css`.
