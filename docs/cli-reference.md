# CLI Reference

Complete reference for all `loko` commands and flags.

The commands are `validate`, `fmt`, `query`, `export`, `build`, `serve`, `mcp`, `version` and
`completion`. Live rebuilding is part of `serve`; there is no separate watch command.

**Exit codes**, for every command: `0` success; `1` errors; `2` warnings when `--strict` is
given.

## Global Flags

| Flag | Description |
|------|-------------|
| `--project, -p` | Project root to discover `*.loko.hcl` files beneath (default `.`) |
| `--verbose, -v` | Verbose output |
| `--help, -h` | Show help for any command |
| `--version` | Show the loko version |

---

## loko validate

Compile the architecture and report every problem found in one run, each with file, line and
column.

```bash
loko validate [--strict] [--format text|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--strict` | bool | `false` | Exit `2` when warnings are present |
| `--format` | string | `text` | `text` renders source snippets; `json` emits the diagnostics contract |

```bash
loko validate
loko validate --strict
loko validate --format json | jq .
```

There is no `--check-drift`: the architecture is authored once, so there is nothing to drift.

---

## loko fmt

Rewrite every `*.loko.hcl` file in canonical form, preserving comments and declaration order.

```bash
loko fmt [--check]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check` | bool | `false` | Write nothing; list non-canonical files and exit `1` (for CI) |

---

## loko build

Render diagrams, markdown and a browsable site from the architecture.

```bash
loko build [--format d2,svg,md,html] [--out DIR] [--strict]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | list | `d2,svg,md,html` | Formats to produce, comma-separated or repeated. `html` and `md` embed diagrams, so they add `svg`, and the summary says so |
| `--out` | string | `<project>/dist` | Output directory. A relative path resolves against the working directory, as `loko export --out` does |
| `--strict` | bool | `false` | Exit `2` when warnings are present |

Views need no configuration: a landscape, one per system that has containers, one per container
that has components, and one per environment. Declared `view` blocks are rendered alongside them;
see [the language reference](language.md#view).

**Guarantees**:
- **Byte-identical output** across runs, machines and file-discovery orders, so the output can be
  committed and reviewed.
- **A generated-file notice on every file**, naming the `*.loko.hcl` files it came from.
- **Only files the tool wrote are ever removed.** They are listed in `dist/.loko-manifest`; anything
  else in the directory is left alone, and unchanged files are not rewritten.
- **Nothing is written when compilation fails.** The previous output stays exactly as it was.
- **No external program is run.** Diagrams render in-process.

**Exit codes**: `0` success; `1` errors (compile errors, unknown format, colliding output names,
invalid theme); `2` warnings with `--strict`.

**Output layout**: `index.html`, `view/<id>.html`, `element/<kind>/<name>.html`, `assets/`,
`diagrams/<id>.{d2,svg}`, `md/…`.

**Themes**: files in `<project>/templates/` override the site's look. See
[contracts/theme.md](../specs/014-viewmodel-renderers/contracts/theme.md).

**Examples**:
```bash
loko build
loko build --format svg,html --out public
loko build --strict
```

---
## loko serve

Preview the site locally, rebuilding and reloading the browser as the architecture changes.

```bash
loko serve [--host 127.0.0.1] [--port 8080]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--host` | string | `127.0.0.1` | Interface to listen on. Use `0.0.0.0` inside a container; loko warns whenever the address is not loopback |
| `--port` | int | `8080` | Port to listen on |

In Docker, publish the port on the host's loopback so the site stays private:

```bash
docker run --rm -v "$PWD:/workspace" -p 127.0.0.1:8080:8080 \
  ghcr.io/madstone-tech/loko serve --host 0.0.0.0
```

It watches the `*.loko.hcl` files, the prose files they reference, and `templates/`. A burst of
saves causes one rebuild. When a save does not compile, every page shows the diagnostics, with
file, line and column, instead of stale output; saving a fix recovers without a restart. `serve`
builds in memory and never writes the output directory.

---
## loko mcp

Start the MCP (Model Context Protocol) server for AI assistant integration.

```bash
loko mcp [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--env` | string | `""` | Set an environment variable (`KEY=VALUE`) for the server |

The server registers five tools: `describe`, `query` and `validate` read the compiled architecture;
`apply_edit` and `move` edit the HCL source, compiling every change before anything is written. Tools
write `*.loko.hcl` files and nothing else. See [MCP Integration](./mcp-integration.md) for setup and
every tool's arguments.

---

## loko query

Ask the compiled architecture a question. These are the answers the MCP `query` tool gives:
`--format json` prints exactly what the tool returns.

```bash
loko query dependents   <address> [--transitive]
loko query dependencies <address> [--transitive]
loko query path         <from> <to>
loko query orphans
loko query coupling     [--limit 20]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `text` | `text` prints a table; `json` and `toon` are for tools and pipes |
| `--transitive` | bool | `false` | `dependents` and `dependencies`: follow relationships to any depth, reporting each element once with its distance |
| `--limit` | int | `20` | `coupling`: rows to show |

An element stands for itself and everything inside it. The dependents of `container.orders_db`
include a component in another container that calls it, reported at that component's address.

| Query | Answers |
|---|---|
| `dependents` | Who has a relationship into this element (or into anything inside it)? |
| `dependencies` | What does this element (or anything inside it) have a relationship to? |
| `path` | The shortest chain of relationships from one element to another, or `no path` |
| `orphans` | Elements with no relationship touching them or anything inside them |
| `coupling` | Elements ranked by distinct fan-in plus fan-out |

```bash
loko query dependents container.orders_db
loko query dependencies container.web --transitive
loko query path person.customer external.bank
loko query coupling --limit 10 --format json | jq .
```

An unknown address prints `unknown address "container.ap": did you mean container.api?` and exits
`1`. A project that does not compile prints its diagnostics, as `validate` does, and exits `1`.

---

## loko export

Write the compiled architecture as a byte-stable, machine-readable artifact.

```bash
loko export [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `json` | Export format: `json`, `toon` |
| `--out` | string | stdout | Output file. Nothing is written when compilation reports errors |

---

## loko completion

Generate shell completion scripts.

```bash
loko completion [bash|zsh|fish|powershell]
```

**Examples**:
```bash
# Bash
loko completion bash > /etc/bash_completion.d/loko

# Zsh
loko completion zsh > "${fpath[1]}/_loko"

# Fish
loko completion fish > ~/.config/fish/completions/loko.fish
```

---

## loko version

Print the current version.

```bash
loko version
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | Disable coloured diagnostics |
| `LOKO_VERBOSE` | Same as `--verbose` |

There is no configuration file. Project settings live in the `project` block of the HCL (see
[the language reference](language.md#project)).
