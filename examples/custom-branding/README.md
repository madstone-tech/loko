# Custom-branded loko example

`simple-project` with a custom brand theme for a fictional "Example Co." in
`templates/`. `loko build` and `loko serve` pick the directory up by
convention. The theme uses the theming API of loko 1.6: it overrides one block,
sets one font token, and ships its own mark and fonts.

## Try it

```bash
loko serve -p examples/custom-branding
```

Open http://127.0.0.1:8080. `loko serve` watches `templates/`, so theme edits
show up in the browser on save.

## What the theme changes

| File | Change |
|---|---|
| `templates/custom.css` | Brand colour tokens (forest `#2F6B4F`, deep forest `#183A2B`, amber `#D98E2B`, warm paper `#F7F5EF`) on loko's CSS variables, for light and dark mode. `--font-ui` and `--font-prose` set Atkinson Hyperlegible for interface text and prose. Headings use a transitional serif (Charter) from the system. Square corners and a solid amber rule under the header. |
| `templates/layout.gohtml` | Only the `brand` block: the sidebar shows the Example Co. mark. Search, tree, views list and footers stay built-in. |
| `templates/fonts.css` | Replaces the built-in font faces. The site loads only Atkinson Hyperlegible. |
| `templates/assets/` | The mark (`mark.svg`) and the fonts with their licence. `loko build` publishes them at `assets/theme/`. |
| `main.loko.hcl` | `footer` on the project block sets the page footer text. |

The diagrams keep loko's light paper: the d2 palette is fixed by convention
(see `docs/configuration.md`, "Removed settings") and cannot be themed through
`templates/`.

Atkinson Hyperlegible is © Braille Institute of America, Inc., under the SIL
Open Font License 1.1 (`templates/assets/fonts/OFL-atkinson-hyperlegible.txt`).

To copy the theme into another project:

```bash
cp -R examples/custom-branding/templates path/to/your-project/
```
