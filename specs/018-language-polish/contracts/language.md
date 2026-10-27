# Contract: Language additions

Seven permanent v1.x attributes and one resolution rule. Every attribute is optional; a project
that uses none of them compiles, exports and renders exactly as in 1.1.0.

## `node`: `title`, `technology`, `description`, `tags`

```hcl
deployment "prod" {
  node "aws" {
    title      = "AWS account 1234"
    technology = "AWS"

    node "vpc" {
      title       = "Production VPC"
      technology  = "Amazon VPC"
      description = "Private subnets in eu-west-1a and 1b"
      tags        = ["network"]

      instance "api" {
        of         = container.api
        technology = "AWS Lambda"
      }
    }
  }
}
```

| Rule | Result |
|---|---|
| `title = ""` | compile error, as for an element's empty title |
| `tags` with a duplicate | the element tag rule applies |
| A nested node | carries its own attributes; nothing is inherited |
| Drawn on the deployment diagram | `<title>` / `[<technology>] · <name>`; without a title, `<name>` / `[<technology>]`; never the description |
| Site page of the deployment | a "Nodes" table when any node has one of the four attributes |

## `instance.technology`

| Rule | Result |
|---|---|
| Set | that instance's box on its deployment diagram shows it instead of the element's technology |
| Unset | the element's technology, as today |
| The element has no technology | allowed |

## `view`: `title`, `description`

```hcl
view "phase_1" {
  title       = "Phase 1: ingestion only"
  description = "What runs before the pricing engine exists. Scoring and publishing are left out."
  include     = [system.pricing]
}
```

| Rule | Result |
|---|---|
| `title = ""` | compile error |
| `title` set | site heading, `<title>`, navigation entry and image `alt` use it; the D2 and SVG draw it as a heading at the top of the canvas |
| `description` set | shown under the page heading; not drawn on the canvas |
| Neither | output unchanged from 1.1.0 |
| Identity | the name stays the address (`view.phase_1`), the file name and the link |

## `docs` resolution

A path resolves from the directory of the file that declares the element.

| Situation | Result |
|---|---|
| Found next to the declaring file | used; no diagnostic |
| Not there; found from the project root | used; `docs_root_relative` (info) gives the file-relative path to write |
| Found at both, different files | the file-relative one is used; `docs_ambiguous` (warning) names both |
| Found at neither | `docs_not_found` (warning) names both locations checked |
| Leaves the project root (`../` too far) | treated as not found, as today |
| Declared in a file at the root | both rules give the same file; no diagnostic |

The export's `docs` is the resolved path, relative to the project root.

## Diagnostic codes

| Code | Severity | When |
|---|---|---|
| `docs_root_relative` | info | a docs path found only by the 1.1.0 rule |
| `docs_ambiguous` | warning | a docs path that names different files under the two rules |
| `wide_view` | info | a rendered view ≥ 3,000 px wide and ≥ 2.5 times wider than tall, with no `direction` or `layout` chosen (see [cli-mcp.md](cli-mcp.md)) |

All three can be listed in `ignore_warnings` and passed to `--ignore`.
