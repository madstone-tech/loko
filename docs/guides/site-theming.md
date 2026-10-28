# Site theming

`loko build` and `loko serve` produce a static site with a built-in theme in Madstone branding.
It has light and dark modes, which follow the reader's system setting, and a toggle to override
it. To change the look, put override files in `<project root>/templates/`. The directory is
optional; without it the built-in theme is used.

The normative contracts are
[specs/014-viewmodel-renderers/contracts/theme.md](../../specs/014-viewmodel-renderers/contracts/theme.md)
and, for the blocks added in 1.3,
[specs/021-site-redesign/contracts/theme-blocks.md](../../specs/021-site-redesign/contracts/theme-blocks.md).
This page is the short version.

## Overridable files

| File | Kind | What it overrides |
|---|---|---|
| `custom.css` | stylesheet | Empty in the built-in theme and loaded after `style.css`. Start here |
| `style.css` | stylesheet | Replaces the built-in stylesheet, byte for byte |
| `site.js` | script | Replaces the built-in search, tree, keyboard paging, theme toggle and diagram zoom |
| `layout.gohtml` | template | Blocks `layout`, `head`, `header`, `nav`, `footer`, `breadcrumb`, `pager`, `theme-toggle` |
| `index.gohtml` | template | Blocks `index`, `stat-cards`, `row-list` |
| `element.gohtml` | template | Blocks `element`, `element-tables`, `relation-table`, `element-children`, `prose-missing`, `prose-inherited` |
| `rail.gohtml` | template | Blocks `rail`, `part-of`, `siblings`, `in-context` |
| `view.gohtml` | template | Blocks `view`, `in-this-view` |
| `partials.gohtml` | template | Blocks `tag-chips`, `link`, `diagram`, `icon-tile`, `chip` |

A `.gohtml` file replaces only the `{{define}}` blocks it contains; every other block stays
built-in. Files with any other extension are ignored. The fonts (`fonts.css`) and icons
(`icons.css`) are generated and cannot be overridden.

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

Replace the footer and nothing else:

```gohtml
{{/* templates/layout.gohtml */}}
{{define "footer"}}<footer class="footer"><p>Platform architecture. Edit the *.loko.hcl source, not this page.</p></footer>{{end}}
```

Templates receive view-model values only: the page's element, view or projection, its `Root`
(the relative link to the site root), its pruned navigation `Tree`, its breadcrumb `Crumbs`, and
its header `Label`, `Meta` and `Lead`. They cannot read files or the compiled model directly.

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

`loko serve` watches `templates/`, so theme edits show up in the browser on save.
