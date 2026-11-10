# Site theming

`loko build` and `loko serve` produce a static site with a built-in theme in Madstone branding.
It has light and dark modes, which follow the reader's system setting, and a toggle to override
it. To change the look, put override files in `<project root>/templates/`. The directory is
optional; without it the built-in theme is used.

The normative contracts are
[specs/014-viewmodel-renderers/contracts/theme.md](../../specs/014-viewmodel-renderers/contracts/theme.md)
and, for the blocks added in 1.3,
[specs/021-site-redesign/contracts/theme-blocks.md](../../specs/021-site-redesign/contracts/theme-blocks.md).
The 1.6 additions are in
[specs/023-v160-theming-mcp/contracts/theming.md](../../specs/023-v160-theming-mcp/contracts/theming.md).
This page is the short version.

## Overridable files

| File | Kind | What it overrides |
|---|---|---|
| `custom.css` | stylesheet | Empty in the built-in theme and loaded after `style.css`. Start here |
| `style.css` | stylesheet | Replaces the built-in stylesheet, byte for byte |
| `site.js` | script | Replaces the built-in search, tree, keyboard paging, theme toggle and diagram zoom |
| `layout.gohtml` | template | Blocks `layout`, `head`, `header`, `nav`, `brand`, `footer`, `breadcrumb`, `pager`, `theme-toggle` |
| `index.gohtml` | template | Blocks `index`, `stat-cards`, `row-list` |
| `element.gohtml` | template | Blocks `element`, `element-tables`, `relation-table`, `element-children`, `prose-missing`, `prose-inherited` |
| `rail.gohtml` | template | Blocks `rail`, `part-of`, `siblings`, `in-context` |
| `view.gohtml` | template | Blocks `view`, `in-this-view` |
| `partials.gohtml` | template | Blocks `tag-chips`, `link`, `diagram`, `icon-tile`, `chip` |

A `.gohtml` file replaces only the `{{define}}` blocks it contains; every other block stays
built-in. Files with any other extension are ignored. The fonts (`fonts.css`) and icons
(`icons.css`) are generated and cannot be overridden.

## Stable classes, variables, and blocks

A theme depends on three surfaces: the block names in the table above, the CSS classes the
templates emit, and the CSS custom properties. The build checks block names (see
[Errors](#errors)). It does not check CSS, so this section is the contract for classes and
variables.

- A listed class or variable keeps its name for all of 1.x. A rename is a breaking change, and the
  changelog names the old and the new name.
- A new class or variable is an addition. The changelog mentions it.
- An unlisted class has no stability promise.

1.6 adds the `brand` block and the `--font-ui` variable. Both are stable.

Classes the built-in templates emit:

| Template | Classes |
|---|---|
| `layout.gohtml` (page shell) | `skip-link`, `sidebar`, `sidebar-header`, `brand`, `brand-link`, `wordmark`, `accent-bar`, `project-name`, `project-description`, `search-box`, `search-input`, `sidebar-nav`, `nav-section`, `system-list`, `system-link`, `element-list`, `container`, `active`, `sidebar-footer`, `theme-toggle`, `main-content`, `topbar`, `menu-toggle`, `breadcrumb`, `sep`, `cursor`, `page-header`, `kind-label`, `title`, `lead`, `meta`, `pager`, `pager-prev`, `pager-next`, `pager-dir`, `pager-name`, `footer` |
| `layout.gohtml` (`<body>`) | `page-index`, `page-view` with `view-<kind>`, `page-element` with the element's kind and tag classes |
| `index.gohtml` | `landscape`, `stats`, `stat-card`, `stat-number`, `stat-label`, `stat-sub`, `systems`, `views`, `environments`, `rows`, `row`, `row-body`, `row-head`, `row-kind`, `row-desc`, `element-list` |
| `element.gohtml` | `element-layout`, `element-main`, `markdown-content`, `prose-missing`, `prose-inherited`, `source-note`, `children`, `talks-to`, `neighbours`, `dir`, `dir-<direction>`, `rail-card`, `visually-hidden`, and the `row` classes |
| `rail.gohtml` | `rail`, `rail-card`, `compact`, `part-of`, `part-of-up`, `siblings`, `in-context`, `context-figure`, `current`, `row-name`, and the `row` and `diagram` classes |
| `view.gohtml` | `in-this-view`, `nodes`, `steps`, `diagram-legend`, `legend-async`, `legend-crossing`, `row-name`, and the `row` classes |
| `partials.gohtml` | `diagram`, `diagram-panel`, `diagram-figure`, `diagram-link`, `diagram-toolbar`, `zoomable`, `chip`, `tags`, `tag`, `tag-<tag>`, `icon`, `icon-tile`, `kind-<kind>` |

Custom properties in `style.css`. Set them in `custom.css` on `:root`, and again for dark mode as
described in [Examples](#examples):

| Group | Variables |
|---|---|
| Page and text | `--bg`, `--card`, `--border`, `--shadow`, `--txt-1`, `--txt-2`, `--txt-3`, `--txt-4` |
| Links and accent | `--link`, `--accent`, `--accent-text`, `--accent-tint` |
| Chips | `--chip-bg`, `--chip-text` |
| Sidebar | `--side-bg`, `--side-hover`, `--side-link`, `--side-muted`, `--side-txt` |
| Diagram paper | `--paper`, `--paper-border` |
| Element tiles | `--kind-<kind>-bg`, `--kind-<kind>-fg` for `person`, `system`, `container`, `component` and `external` |
| Shape | `--radius-button`, `--radius-card`, `--radius-tag` |
| Type | `--font-display`, `--font-heading`, `--font-prose`, `--font-ui`, `--font-mono`. `--font-ui` is the interface font: the sidebar, search, breadcrumb, labels, table headings, chips and buttons. `--font-mono` is for code and HCL only. Both default to JetBrains Mono |

## Examples

Change the colours without replacing the stylesheet. The built-in theme keeps its colours in CSS
variables: `--bg`, `--card`, `--txt-1` to `--txt-3` (text), `--link`, `--accent` and the
`--kind-<kind>-bg`/`-fg` pairs for element tiles. Light values sit on `:root`, and dark values
under `:root[data-theme="dark"]` and the dark `prefers-color-scheme` query. Check that your text
colours keep at least 4.5:1 contrast against their backgrounds.

```css
/* templates/custom.css */
:root {
  --link: #0a6b5f;
  --accent: #0a7d6e;
}
```

Change the interface font and keep the code font:

```css
/* templates/custom.css */
:root {
  --font-ui: "Inter", system-ui, sans-serif;
}
```

Replace the logo and keep the rest of the sidebar. The `brand` block holds only the logo link,
so the search box, the tree and the sidebar footer stay built-in:

```gohtml
{{/* templates/layout.gohtml */}}
{{define "brand"}}<div class="brand"><a class="brand-link" href="{{.Root}}index.html"><span class="wordmark">Example Co.</span></a></div>{{end}}
```

Set the footer text without a template. The `project` block accepts `footer`. The built-in
footer shows it as plain text, and shows the default text when it is not set:

```hcl
project "shop" {
  footer = "© Example Co."
}
```

Replace the footer markup and nothing else:

```gohtml
{{/* templates/layout.gohtml */}}
{{define "footer"}}<footer class="footer"><p>Platform architecture. Edit the *.loko.hcl source, not this page.</p></footer>{{end}}
```

Templates receive view-model values only: the page's element, view or projection, its `Root`
(the relative link to the site root), its pruned navigation `Tree`, its breadcrumb `Crumbs`, and
its header `Label`, `Meta` and `Lead`. They cannot read files or the compiled model directly.

## Build output

When `templates/` holds a theme file, `loko build` prints one line after the `built` line. The
line names each theme file, sorted by name. A template file also lists the blocks it defines:

```
theme: templates/ (custom.css; layout.gohtml: brand, footer)
```

The blocks come from the html backend. A build without the `html` format lists the files with no
blocks. A project with no theme prints no line. The MCP `build` result has the same data in its
`theme` field.

## Upgrading a theme from 1.2

Every 1.2 block name still exists, so a 1.2 override keeps building. Some blocks now render in a
different place:

- `element-tables` renders the rail's **Talks to** section: one row per neighbour, not a Uses and
  a Used by table.
- `relation-table` receives one row per neighbour. Each row still has `Other`, `Description`,
  `Technology` and `Tags`, and adds `Direction` (`out`, `in` or `both`) and `Sibling`.
- `nav` renders once per page, with the pruned Architecture tree in `.Tree`.

A replacement `style.css` or `site.js` written for 1.2 still loads, but it meets the new markup.
The layout will look wrong, and the theme toggle, tree, keyboard paging and diagram zoom will be
missing. Keep additions in `custom.css` instead.

## Reading the site

- **Navigation.** The sidebar's Architecture tree shows the top level and the branch of the
  current page. A collapsed branch opens its element's page. ← and → move between siblings,
  except while typing in search.
- **Theme.** The site follows the system's light or dark setting, and uses light when the setting
  is unknown. The toggle's choice is remembered in the browser.
- **Diagrams.** On a view page you can zoom with the wheel, a click or the toolbar, pan by
  dragging, and reset. An element page shows its parent's diagram with the element outlined.
- **Offline.** Fonts, icons and scripts are bundled, so the site works offline and from disk.

## Errors

A theme is applied completely or not at all. The build fails with `theme_invalid`, naming the
file, when:

- a `.gohtml`, `.css` or `.js` file has a name not in the table (the message lists the valid names),
- a template does not parse,
- a template defines a block the built-in theme does not have,
- a template fails while rendering.

These checks cover files and block names only. The build does not read your CSS, so a selector
for a class that the templates no longer emit fails silently. Use the classes in
[Stable classes, variables, and blocks](#stable-classes-variables-and-blocks).

`loko serve` watches `templates/`, so theme edits show up in the browser on save.
