# TOON Format Guide

[TOON](https://github.com/toon-format/spec) (Token-Oriented Object Notation, v3) is a text
encoding of the JSON data model designed to cost fewer tokens when an LLM reads it. loko encodes
it with `github.com/toon-format/toon-go`.

## Where loko uses TOON

| Surface | How | Default |
|---|---|---|
| MCP reads (`describe`, `query`, `validate`) | `format` argument: `toon` or `json` | `toon` ([ADR-0011](../adr/0011-toon-mcp-default.md)) |
| `loko query` | `--format text\|json\|toon` | `text` |
| `loko export` | `--format json\|toon` | `json` |

`loko build` does not produce TOON; it renders `d2`, `svg`, `md` and `html`.

The TOON and JSON forms carry the same data with the same field names. `loko query --format toon`
prints exactly what the MCP `query` tool returns.

## What it looks like

Uniform arrays become tables: field names once in the header, one row per item. The `[#N]`
marker gives the row count.

```text
$ loko query coupling --format toon
ok: true
kind: coupling
coupling[#14]{address,fanIn,fanOut}:
  container.api,1,2
  container.gateway,1,2
  component.repo,1,1
  ...
revision: r1-55dce38f29f9c653
```

The same answer as JSON is one object per row, with every key repeated:

```json
{"ok":true,"kind":"coupling","coupling":[{"address":"container.api","fanIn":1,"fanOut":2},{"address":"container.gateway","fanIn":1,"fanOut":2}, ...],"revision":"r1-55dce38f29f9c653"}
```

Nested objects use indentation, like YAML:

```text
$ loko export --format toon
schemaVersion: 1
project:
  name: two-systems
  description: "Renderer fixture: two systems, six containers, two environments"
elements[#15]:
  - address: component.authorizer
    kind: component
    name: authorizer
    parent: container.gateway
    description: Card authorization
    range:
      file: main.loko.hcl
      startLine: 139
      ...
relationships[#10]:
  ...
```

## How much it saves

Measured on `testdata/projects/two-systems`, in bytes:

| Output | JSON | TOON |
|---|---|---|
| `loko query coupling` (tabular) | 795 (compact) | 404 |
| `loko export` (nested) | 16,284 (indented) / 10,134 (compact) | 12,135 |

Tabular answers, which is most of what the MCP tools return, shrink by about half. The full
export is deeply nested, so TOON beats indented JSON but not minified JSON. For an LLM reading the
whole model, prefer `describe` at the level you need over pasting an export.

## Practical advice

- Leave MCP reads on the `toon` default. Ask for `json` only when a client parses the result.
- Use `describe` with `level: summary`, then `structure`, then `full` scoped by `address`, rather
  than starting at `full`.
- Use `loko export --format json` for programs and CI artifacts; it is the more widely parsed
  form and is byte-stable across runs.
- Both forms are byte-stable: same source, same bytes.

## References

- [TOON specification](https://github.com/toon-format/spec)
- [MCP Integration](../mcp-integration.md): the `format` argument on each tool
- [CLI reference](../cli-reference.md): `loko query` and `loko export`
