# ADR-0024: A theming API for rebranding the site

**Status:** Accepted
**Date:** 2026-11-10
**Feature:** [023-v160-theming-mcp](../../specs/023-v160-theming-mcp/spec.md)
**Contract:** [contracts/theming.md](../../specs/023-v160-theming-mcp/contracts/theming.md)
**Related:** [ADR-0013](0013-viewmodel-renderers.md) (theme overrides),
[ADR-0020](0020-site-redesign.md) (site blocks)

## Context

A team that publishes a site for each project it delivers rebranded the 1.5 site and reported
four problems (feedback items 7, 11, 12 and 13):

- The interface text and the code text used one variable, `--font-mono`. To change the
  interface font, the team redefined the font of each interface selector in `custom.css`.
- The logo was inside the `nav` block. To change the logo, the team copied the whole `nav` block.
  The copy goes out of date when loko changes the sidebar.
- The footer text was in the `footer` block. To change one sentence, the team overrode a
  template.
- The build did not say that it applied a theme. A theme in the wrong directory failed silently.

## Decision

1. **T-1 Interface font variable.** `style.css` declares `--font-ui` next to `--font-mono`. The
   15 interface rules use `var(--font-ui)`. The rule `code, kbd, pre` keeps `var(--font-mono)`.
   The default value of `--font-ui` is the 1.5 value of `--font-mono`, so a site with no theme
   uses the same fonts as 1.5.

2. **T-2 Brand block.** `layout.gohtml` defines `{{define "brand"}}` with the 1.5 brand markup,
   byte for byte. The built-in `nav` calls it in the same place. A theme overrides `brand` to
   change the logo, and keeps the built-in search, tree and sidebar footer. A theme that
   overrides `nav` keeps working, because it replaces the whole block as before.

3. **T-3 Project footer.** The `project` block accepts `footer`, a string. The built-in `footer`
   block shows it as escaped plain text. When `footer` is not set, the block shows the 1.5
   markup. The value goes through the declaration, the IR, the projection header and the
   authoring edit schema, so the MCP `schema` tool and `apply_edit` know it.

4. **T-6 Theme report.** When `templates/` holds a theme file, `loko build` prints one line after
   the `built` line, for example
   `theme: templates/ (custom.css; layout.gohtml: brand, footer)`. The build use case lists the
   files. The html backend adds the blocks that each template file defines, because only that
   backend parses templates. `RenderOptions` carries a pointer to the report, because it is
   passed by value. The MCP `build` result has the same data in its `theme` field.

We chose a block and a variable over new language attributes (for example a project logo). A
block and a variable reuse the theme mechanism that exists, and they cover any markup and any
font.

## Consequences

**Positive:**
- A theme changes the logo, the interface font and the footer without a copy of a built-in block
  other than `brand`.
- A site with no theme renders the same pages as 1.5. Only `assets/style.css` changes, by the
  new variable.
- The build output shows which theme files and blocks were applied.

**Negative:**
- `brand` and `--font-ui` are new stable names for all of 1.x.
- A build without the `html` format lists the theme files with no blocks.

**Mitigations:**
- The theming guide lists `brand` and `--font-ui` with the other stable names.
- The 014 and 021 block guards (`TestAll014BlocksStillDefined`, `Test014OverridesStillBuild`)
  keep passing unchanged, so no 1.5 block name is lost.
