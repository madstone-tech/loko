# Quickstart Guide

Get started with loko in 5 minutes. This guide walks you through creating your first C4 architecture documentation project.

## Prerequisites

- Go 1.27 or later. Nothing else: diagrams render inside loko, so no `d2` install is needed.

## Installation

```bash
go install github.com/madstone-tech/loko@latest
loko --help
```

## Create Your First Project

### 1. Describe the architecture

Create a directory and an `arch.loko.hcl` file in it:

```hcl
project "payments" {
  description = "Payment processing platform"
}

person "customer" {
  uses "pays" {
    target      = container.api
    description = "Pays for an order"
  }
}

system "payments" {
  description = "Authorization, capture, settlement"
  docs        = "./docs/payments.md"
}

container "api" {
  system     = system.payments
  technology = "Go"

  uses "store" {
    target = container.db
  }
}

container "db" {
  system     = system.payments
  technology = "PostgreSQL"
}

component "handler" {
  container   = container.api
  description = "HTTP entry point"
}
```

Every `*.loko.hcl` file beneath the directory is part of the architecture. The full grammar is in
[the language reference](language.md).

### 2. Check it

```bash
loko validate
```

Every problem is reported in one run with file, line and column.

### 3. Format it

```bash
loko fmt
```

### 4. Export it (optional)

```bash
loko export --format json > ir.json
```

### 5. Build documentation

```bash
# Diagrams (D2 and SVG), markdown and a site, into ./dist — no configuration
loko build

# Only the site (SVG is added because the site embeds it)
loko build --format html

# Build into another directory
loko build --out public
```

No separate `d2` installation is needed: diagrams render inside loko.

### 6. Preview your documentation

```bash
# Build in memory, serve on loopback, reload the browser on every save
loko serve

# Open http://127.0.0.1:8080 in your browser
```

## Output

`loko build` writes:

```
dist/
├── .loko-manifest          # the files loko owns; only these are ever pruned
├── index.html              # site entry: the landscape
├── view/<view>.html
├── element/<kind>/<name>.html
├── assets/
├── diagrams/<view>.d2
├── diagrams/<view>.svg
└── md/                     # the same content as markdown
```

Every file opens with a notice naming the source it came from. Rebuilding an unchanged
architecture changes nothing, so `dist/` can be committed.

## Watch Mode

`loko serve` watches for you: it rebuilds whenever an architecture file, a prose file it references,
or a theme override changes, and shows compile errors in the browser until you fix them. There is
no separate `loko watch`.

## Validation

Check your architecture for issues:

```bash
loko validate
```

It reports unresolved references, wrong parents, containment cycles and unknown blocks as errors,
and missing prose, empty systems and elements with no relationships as warnings. Add `--strict`
to make warnings fail.

## Using with Claude (MCP)

loko includes an MCP server for AI-assisted architecture design:

```bash
# Start MCP server (for Claude Desktop integration)
loko mcp
```

See the [MCP Integration Guide](mcp-integration.md) for setup instructions.

## Next Steps

- Read the [language reference](language.md) for every block and attribute
- Read the [Configuration Reference](configuration.md): there is no config file in v1
- Explore [example projects](../examples/) for common architecture patterns
- Learn about [MCP integration](mcp-integration.md) for AI-assisted design

## Common Commands

| Command | Description |
|---------|-------------|
| `loko validate` | Compile and report diagnostics |
| `loko fmt` | Canonical formatting (`--check` for CI) |
| `loko export` | Compiled architecture as JSON or TOON |
| `loko build` | Render diagrams, markdown and a site into `dist/` |
| `loko build --format md` | Markdown only (adds the SVG it embeds) |
| `loko serve` | Preview server with live reload |
| `loko mcp` | Start MCP server |

## Getting Help

```bash
# General help
loko --help

# Command-specific help
loko build --help
loko serve --help
```

For issues and feature requests, visit: https://github.com/madstone-tech/loko/issues
