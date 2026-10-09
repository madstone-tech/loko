# loko documentation

loko compiles a C4 software architecture from HCL. Start with the quick start, keep the language
reference open while writing, and reach for the guides when you wire loko into CI or an assistant.

## Using loko

| Document | What it covers |
|---|---|
| [Quick start](quickstart.md) | Install, write an architecture, validate, build, serve and query it |
| [Language reference](language.md) | Every block, attribute and function in `*.loko.hcl`, with a complete example |
| [CLI reference](cli-reference.md) | Every command, flag and exit code |
| [Configuration](configuration.md) | Where settings live, given that there is no configuration file |
| [MCP integration](mcp-integration.md) | The MCP server and its five tools, and client setup |
| [Examples](../examples/) | Complete example projects |

## Guides

| Guide | What it covers |
|---|---|
| [CI/CD integration](guides/ci-cd-integration.md) | Validating and building in GitHub Actions and GitLab, and the container image |
| [Site theming](guides/site-theming.md) | Overriding the site's layout, styles and scripts from `<project>/templates/` |
| [TOON format](guides/toon-format-guide.md) | The compact encoding used by MCP reads, `query` and `export` |

## For assistants

Background for an LLM working with a loko architecture, alongside [MCP integration](mcp-integration.md):

| Document | What it covers |
|---|---|
| [C4 model](llm/c4-model.md) | The C4 levels and how each maps to loko's HCL |
| [Patterns](llm/patterns.md) | Common architectures written in HCL |

## Project

| Document | What it covers |
|---|---|
| [Roadmap](roadmap.md) | What has shipped and what is planned |
| [Architecture decisions](adr/) | ADRs; [ADR-0012](adr/0012-hcl-source-of-truth.md) onwards describe v1 |
| [Constitution compliance](architecture/constitution-compliance.md) | The mechanically enforced architecture rules and how to read a failure |
| [Contributing](../CONTRIBUTING.md) | Development setup, workflow and checklist |
| [Changelog](../CHANGELOG.md) | Release notes |

Feature specifications, plans and research for each stage are in [`specs/`](../specs/).

## Support

- Questions and ideas: [GitHub Discussions](https://github.com/madstone-tech/loko/discussions)
- Bugs: [GitHub Issues](https://github.com/madstone-tech/loko/issues)
