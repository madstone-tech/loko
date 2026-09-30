# Contract: theme overrides

**Feature**: `014-viewmodel-renderers` | **Status**: normative

**Directory**: `<project root>/templates/`. It is optional. If it is absent, the built-in theme is
used and no warning is issued (US7/AC2).

## Overridable files

| File | Kind | Defines (`{{define}}` blocks) |
|---|---|---|
| `layout.gohtml` | html/template | `layout`, `head`, `header`, `nav`, `footer` |
| `index.gohtml` | html/template | `index` |
| `element.gohtml` | html/template | `element`, `element-tables`, `relation-table`, `element-children`, `prose-missing` |
| `view.gohtml` | html/template | `view` |
| `partials.gohtml` | html/template | `tag-chips`, `link`, `diagram` |
| `style.css` | stylesheet | Replaces the built-in stylesheet |
| `custom.css` | stylesheet | Empty in the built-in theme and loaded after `style.css`, so additive styling can go here |
| `site.js` | script | Replaces built-in navigation and search |

## Resolution

1. Built-in templates are parsed first. Then each `.gohtml` override is parsed in name order, and
   any block it defines replaces the built-in block of the same name. An override may therefore
   contain only `{{define "header"}}…{{end}}`, and everything else stays built-in (FR-034).
2. A `.css` or `.js` override replaces the built-in file of the same name, byte for byte.
3. Files in `templates/` with any other extension are ignored.

## Template data

Templates receive only `viewmodel` values: `Projection`, `ViewModel` or `ElementPage`, wrapped in a
`PageData{Project, Page, Root}` value, where `Root` is the relative href to the site root. Templates
cannot reach the file system or the IR.

## Errors (FR-035): `theme_invalid`, which fails the build

| Case | Message names |
|---|---|
| A `.gohtml`, `.css` or `.js` file whose name is not in the table | the file and the list of overridable names |
| A template parse error | the file, line, and parser message |
| A `{{define}}` of a block name not defined by any built-in template | the file and the block name |
| A template execution error while rendering | the file that defined the failing block |

The build never falls back silently. A theme that half-applies is refused.
