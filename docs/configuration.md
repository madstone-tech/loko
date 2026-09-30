# Configuration Reference

loko v1 has **no configuration file**. `loko.toml`, `~/.loko/config.toml`, and the settings they
held were removed in v1.0 (ADR-0012). Everything is configured in one of three places.

## The `project` block

Project-wide settings live in the architecture source itself:

```hcl
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}
```

See [the language reference](language.md#project).

## Command-line flags

| Need | Flag |
|---|---|
| Choose output formats | `loko build --format d2,svg,md,html` |
| Choose the output directory | `loko build --out public` |
| Fail on warnings | `loko build --strict`, `loko validate --strict` |
| Preview port | `loko serve --port 3000` |
| Project root | `--project` / `-p` on every command |

See [the CLI reference](cli-reference.md).

## Theme overrides

To change the site's look, put replacement files in `<project root>/templates/`:

- **Stylesheets and scripts**: `style.css`, `custom.css` and `site.js` replace the built-in file.
- **Templates**: a `.gohtml` file replaces the individual `{{define}}` blocks it contains.

Anything malformed, or any unknown file name, fails the build and names the file. The contract is
in [specs/014-viewmodel-renderers/contracts/theme.md](../specs/014-viewmodel-renderers/contracts/theme.md).

## Removed settings

| v0 setting | v1 |
|---|---|
| `[outputs]`, `html`, `markdown`, `pdf` | `--format`. PDF output was removed. |
| `[build] parallel`, `max_workers` | Always parallel, one worker per CPU |
| `[server] serve_port`, `hot_reload` | `loko serve --port`; reload is always on |
| `api_port`, `LOKO_API_KEY` | The HTTP API was removed |
| `d2.cache`, `--clean` | No disk cache. `serve` caches renders in memory only |
| `D2_LAYOUT`, `D2_THEME` | The layout and palette are fixed by convention; restyle the site with a theme |
