# loko Guides

Task-oriented guides. The references they build on are the
[language reference](../language.md), the [CLI reference](../cli-reference.md) and
[MCP Integration](../mcp-integration.md).

| Guide | Covers |
|---|---|
| [CI/CD Integration](ci-cd-integration.md) | `fmt --check`, `validate --strict`, JSON diagnostics, building the site, GitHub Actions and GitLab CI |
| [Site theming](site-theming.md) | Overriding the site's stylesheet, scripts and templates from `templates/` |
| [TOON Format](toon-format-guide.md) | The token-efficient output format used by MCP reads, `query` and `export` |

Notes for assistants and agents working on a loko project are in [`docs/llm/`](../llm/):
[C4 model](../llm/c4-model.md) and [architecture patterns](../llm/patterns.md), both in HCL.

## Writing a guide

- Start from the reader's goal, then give the steps.
- Run every command you document against a real project (`testdata/projects/` has several) and
  paste real output.
- Link to the references instead of restating them.
