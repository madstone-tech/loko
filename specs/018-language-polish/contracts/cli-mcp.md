# Contract: Information diagnostics in the CLI and MCP

## Severity

`info` joins `error` and `warning` in every machine-readable diagnostic (JSON, TOON, MCP), and in
`specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`.

| Behaviour | Error | Warning | Info |
|---|---|---|---|
| Stops artefacts being written | yes | no | no |
| Exit code | 1 | 2 under `--strict` | never changes it |
| Counted in warning totals | | yes | no |
| `ignore_warnings`, `--ignore`, MCP `ignore` | refused | removed | removed |

## Text output

```text
Info: Docs path resolved from the project root
  on architecture/containers.loko.hcl line 5:
   5:   docs        = "architecture/docs/pricing.md"
container.collector references "architecture/docs/pricing.md", found from the project root. Paths
resolve from the declaring file first; write "docs/pricing.md" instead.

0 errors, 0 warnings, 7 info
```

The summary line adds `, N info` only when N > 0, so 1.1.0 output of a model with no information
diagnostics is unchanged.

## `--severity` / MCP `severity`

| Value | Shows |
|---|---|
| omitted | everything |
| `error` | errors only (unchanged) |
| `warning` | errors and warnings, no information *(new)* |

The exit code is the same whatever is shown.

## `--summary` / MCP `summary`

Groups are ordered errors, warnings, then information; each group line carries its severity, as
today.

## Where `wide_view` appears

| Command | Gives `wide_view`? |
|---|---|
| `loko build` with SVG or HTML output | yes |
| `loko build --format d2` (no layout) | no |
| MCP `build` | yes, in `diagnostics` |
| `loko validate`, MCP `validate` | no: nothing is laid out |
| `loko serve` | no: serve shows errors only (017) |

The `wide_view` detail names the view, its size and ratio, and what to try:

- declared view: `Try direction = "right"` (or `"down"`, the other one) `or layout = "elk" on the view.`
- generated view: `Set layout = "elk" on the project, or declare a view of these elements.`
