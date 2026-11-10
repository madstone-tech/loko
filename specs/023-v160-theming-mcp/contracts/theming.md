# Contract: Theming API additions (1.6.0)

**Feature**: `023-v160-theming-mcp` | **Status**: normative. Amends
[014 theme.md](../../014-viewmodel-renderers/contracts/theme.md),
[021 theme-blocks.md](../../021-site-redesign/contracts/theme-blocks.md) and the
[site theming guide](../../../docs/guides/site-theming.md).

Every rule here is additive. A theme written for 1.5 builds unchanged.

## T-1 Interface font variable

- `style.css` declares `--font-ui` in the light `:root` token block. Its default
  value is the 1.5 value of `--font-mono` (`"JetBrains Mono", ui-monospace,
  monospace`), so a site with no theme renders with the same fonts as 1.5.
- The 15 interface rules that used `var(--font-mono)` in 1.5 use `var(--font-ui)`:
  `th`, `.sidebar`, `.nav-section h2`, `.search-input`, `.breadcrumb ol`,
  `.kind-label`, `.meta`, `.pager-dir`, `.theme-toggle`,
  `.diagram-toolbar button`, `.menu-toggle`, `.row-kind`, `.row-tags`, `.dir`,
  `.chip`, `.tag`, `.stat-label`, `figcaption`.
- `code, kbd, pre` keep `var(--font-mono)`. That rule is the only one that
  styles code and HCL snippets.
- `--font-ui` joins the stable variables of the theming guide ("Type" row).

## T-2 Brand block

- `layout.gohtml` defines `{{define "brand"}}`. Its markup is byte-identical to
  the 1.5 brand `<div class="brand">…</div>`. The built-in `nav` block calls
  `{{template "brand" .}}` in the same place.
- `brand` receives the page data that `nav` receives (`.Root` and the rest).
- A theme may override `brand`, `nav`, or both. When both are overridden, the
  theme's `nav` renders and the theme's `brand` is used only if that `nav` calls
  it.
- `brand` joins the `layout.gohtml` row of the overridable-blocks table.

## T-3 Project footer

- The `project` block accepts `footer` (string, optional).
- The built-in `footer` block shows `footer` as escaped plain text inside
  `<footer class="footer"><p>…</p></footer>` when it is set, and the 1.5 footer
  markup when it is not set.
- The sidebar footer (`.sidebar-footer`) is unchanged.

## T-4 Theme assets

- Every regular file under `<project root>/templates/assets/` (any depth, any
  extension) is published at `assets/theme/<path below templates/assets/>`.
  Bytes are copied unchanged. No generated-file notice is added.
- A stylesheet in `assets/` reaches an asset with a relative URL, for example
  `url("theme/logo.svg")` from `custom.css` or `fonts.css`.
- An asset path that fails `SafePath` (absolute, `..`, backslash, drive letter,
  control character), a symbolic link, or two asset paths that differ only in
  letter case fail the build with `theme_invalid` naming `templates/assets/<path>`.
  Nothing is written.
- Assets are owned like every artifact: listed in the manifest and pruned when
  removed.
- Assets are loaded and validated with the rest of the theme, before any
  backend renders. They are published only when the `html` format is built.

## T-5 Font stylesheet override

- `fonts.css` joins the overridable files. When `templates/fonts.css` exists,
  `assets/fonts.css` is the theme file (with the generated-file notice, like
  every stylesheet override), and the built-in font faces are not written.
- `assets/brand/madstone-eye.svg` and `assets/LICENSES.txt` are still written.
  They load no fonts.
- `icons.css` stays generated and not overridable.
- The `head` block keeps linking `assets/fonts.css`. No template change.

## T-6 Theme report

- When `templates/` exists, `loko build` prints one line after the `built` line:

  ```
  theme: templates/ (custom.css; fonts.css; layout.gohtml: brand, footer)
  ```

  Files are sorted by name. A template file lists the blocks it defines, sorted.
  Other files are listed bare. Entries are separated by `"; "`. When the theme
  has assets, a final entry says `assets/: <N> files`, or `assets/: 1 file` for
  one (assets are counted, not listed). Blocks come from the html backend: a
  build without the `html` format lists the template files with no blocks.
- With no `templates/` directory, no theme line is printed.
- The MCP `build` result carries the same content in a `theme` field
  (`dir`, `files[]` with `name` and `blocks`, `assets`), omitted when there is
  no theme.

## Compatibility

- The 014 and 021 block-name guards (`TestAll014BlocksStillDefined`,
  `Test014OverridesStillBuild`) keep passing unchanged.
- With no theme, the only golden file that changes is `assets/style.css`
  (T-1). Page goldens do not change.
