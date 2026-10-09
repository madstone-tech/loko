# C4 Model Guide for LLMs

The C4 knowledge needed to model an architecture in loko. The grammar itself is in the
[language reference](../language.md); this page maps C4 onto it.

## What is the C4 model?

A hierarchical way to describe software architecture, created by Simon Brown, with four levels of
zoom: context, containers, components and code. loko models the first three and the deployment
view. You declare elements once in `*.loko.hcl` files; loko derives the diagrams.

## Element kinds

| C4 concept | loko block | Parent attribute | Example |
|---|---|---|---|
| Person (user, actor) | `person` | none | a customer, an operator |
| Software system (yours) | `system` | none | "Payments" |
| External system | `external` | none | a card processor, an email provider |
| Container (deployable/runnable unit, data store) | `container` | `system = system.<name>` (required) | API, SPA, database, queue |
| Component (grouping inside a container) | `component` | `container = container.<name>` (required) | handler, repository |
| Deployment | `deployment` with nested `node` and `instance` | `of = <element>` on an instance | prod environment |

Level 4 (code) is out of scope: leave it to the IDE.

Addresses are `kind.name` (`container.api`). Names are unique per kind across the whole project,
not per parent, so `container.api` in two systems is a `duplicate_declaration` error: name it
`orders_api` and `billing_api`, and use `title` for the display name.

## The levels and the views loko derives

| Level | Question | Generated view |
|---|---|---|
| 1. System context | Who uses the system, and what does it talk to? | `landscape` (every top-level element) |
| 2. Container | What are the runnable units and their technologies? | `system-<name>`, for each system with containers |
| 3. Component | How is one container divided? | `container-<name>`, for each container with components |
| Deployment | Where do containers run? | `deployment-<name>`, for each environment with instances |

Add a `view` block only for a cut the generated views do not give you (a request path, a
compliance scope by tag).

## Relationships

A relationship is a `uses` block inside the element it starts from. The label is its local name.

```hcl
container "api" {
  system     = system.orders
  technology = "Go"

  uses "db" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "SQL"
  }
}
```

An element stands for itself and everything inside it. In the example below, the relationship to
the database is declared on a component, yet `loko query dependencies container.api` reports
`container.orders_db`, and the container-level diagram draws the edge from `api`. So declare
relationships at the most specific level you know.

`kind = "async"` (dashed) and `kind = "trigger"` (a target that invokes the declaring element,
such as a queue feeding a function) change only the drawing.

## Example: e-commerce

```hcl
project "shop" {
  description = "Online store"
}

person "customer" {
  description = "Buys things"

  uses "browse" {
    target     = container.web
    technology = "HTTPS"
  }
}

system "orders" {
  description = "Order lifecycle"
}

external "payment_provider" {
  description = "Card processing"
}

container "web" {
  system     = system.orders
  title      = "Web app"
  technology = "React"

  uses "api" {
    target      = container.api
    description = "Places and tracks orders"
    technology  = "REST/JSON"
  }
}

container "api" {
  system     = system.orders
  title      = "Order API"
  technology = "Go"
}

container "orders_db" {
  system     = system.orders
  title      = "Orders database"
  technology = "PostgreSQL"
  shape      = "database"
}

component "order_handler" {
  container   = container.api
  description = "HTTP handlers for orders"

  uses "repo" {
    target = component.order_repository
  }

  uses "charge" {
    target      = external.payment_provider
    description = "Authorizes payment"
  }
}

component "order_repository" {
  container   = container.api
  description = "Data access"

  uses "db" {
    target     = container.orders_db
    technology = "SQL"
  }
}
```

This yields `landscape`, `system-orders` and `container-api` with no further declarations.

## Rules of thumb

1. **Start at the top.** Declare people, systems and externals, then containers, then components
   only where a container's inside matters.
2. **Databases, queues and buckets are containers**, not systems. Give them a `shape`.
3. **Don't over-decompose.** A container with one component needs none.
4. **Describe responsibility, not implementation.** One or two sentences; put technology in
   `technology`, longer prose in a markdown file referenced by `docs = "./docs/<name>.md"`. An
   element without one gets a `missing_docs` warning (the example above omits them for brevity).
5. **Every element should have a relationship.** An element with none is an `orphan_element`
   warning.
6. **References are bare traversals**: `system = system.orders`, never `"system.orders"`.

## Working through MCP

Read with `describe` (start at `level: summary`), ask with `query` (`dependents`,
`dependencies`, `path`, `orphans`, `coupling`), and change the source with `apply_edit` and `move`,
which compile every change before writing. See [MCP Integration](../mcp-integration.md).

## References

- [C4 model](https://c4model.com/)
- [loko language reference](../language.md)
- [Architecture patterns in HCL](patterns.md)
