# Site theming

`loko build` and `loko serve` produce a static site with a built-in theme. To change its look, put
override files in `<project root>/templates/`. The directory is optional; without it the built-in
theme is used.

The normative contract is
[specs/014-viewmodel-renderers/contracts/theme.md](../../specs/014-viewmodel-renderers/contracts/theme.md).
This page is the short version.

## Overridable files

| File | Kind | What it overrides |
|---|---|---|
| `custom.css` | stylesheet | Empty in the built-in theme and loaded after `style.css`. Start here |
| `style.css` | stylesheet | Replaces the built-in stylesheet, byte for byte |
| `site.js` | script | Replaces built-in navigation and search |
| `layout.gohtml` | template | Blocks `layout`, `head`, `header`, `nav`, `footer` |
| `index.gohtml` | template | Block `index` |
| `element.gohtml` | template | Blocks `element`, `element-tables`, `relation-table`, `element-children`, `prose-missing` |
| `view.gohtml` | template | Block `view` |
| `partials.gohtml` | template | Blocks `tag-chips`, `link`, `diagram` |

A `.gohtml` file replaces only the `{{define}}` blocks it contains; every other block stays
built-in. Files with any other extension are ignored.

## Examples

Recolour the site without replacing the stylesheet. The built-in theme uses CSS variables such as
`--color-primary`, `--color-bg`, `--color-text` and `--font-family`:

```css
/* templates/custom.css */
:root {
  --color-primary: #0a7d6e;
  --font-family: "Inter", system-ui, sans-serif;
}
```

Replace the footer and nothing else:

```gohtml
{{/* templates/layout.gohtml */}}
{{define "footer"}}<footer class="footer"><p>Platform architecture. Edit the *.loko.hcl source, not this page.</p></footer>{{end}}
```

Templates receive view-model values only (`PageData{Project, Page, Root}`, where `Root` is the
relative link to the site root). They cannot read files or the compiled model directly.

## Errors

A theme is applied completely or not at all. The build fails with `theme_invalid`, naming the
file, when:

- a `.gohtml`, `.css` or `.js` file has a name not in the table (the message lists the valid names),
- a template does not parse,
- a template defines a block the built-in theme does not have,
- a template fails while rendering.

`loko serve` watches `templates/`, so theme edits show up in the browser on save.
