# Contract: CLI

New flags only. Existing flags and outputs are unchanged (SC-007).

## `loko validate`

| Flag | Notes |
|---|---|
| `--summary` | group diagnostics by code: count, then up to 20 addresses |
| `--severity error` | show errors only; the exit code is unchanged |
| `--ignore <code>` | repeatable or comma-separated; warning codes only |

Text summary:

```
2 errors, 31 warnings (4 ignored)

error    unresolved_reference   2   container.api, container.web
warning  missing_docs          25   container.a, container.b, … (+5 more)
warning  unbound_instance       6   deployment.prod.node.web.instance.api, …
```

`--format json` with `--summary` emits the MCP `summary` shape with full detail.

Exit codes are unchanged: `0` clean, `1` errors, `2` warnings with `--strict`. Ignored warnings do
not count towards `--strict`. `--ignore` with an error code or an unknown code exits `1` and lists
the warning codes.

## `loko build` and `loko serve`

| Flag | Notes |
|---|---|
| `--ignore <code>` | as above, so strict builds and strict validation agree |

## Project-level

`ignore_warnings` on the `project` block applies to every command, without flags. Per-run
`--ignore` adds to it.
