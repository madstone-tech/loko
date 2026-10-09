# Architecture Patterns in loko

How common architecture patterns look in loko's HCL. Each snippet compiles when placed in a
project with a `project` block: `loko validate` reports no errors. The snippets omit `docs` and
some relationships for brevity, so expect `missing_docs` and `orphan_element` warnings. The grammar is in the [language reference](../language.md)
and the C4 mapping in [c4-model.md](c4-model.md).

General rules for all of them:

- Data stores, queues, topics and buckets are **containers** with a `shape`.
- A relationship is a `uses` block in the element that **depends on** the other.
- `kind = "async"` draws a dashed edge; `kind = "trigger"` marks a target that invokes the
  declaring element. Neither changes queries.

## Application patterns

### Three-tier monolith

```hcl
person "user" {
  uses "app" {
    target     = container.web
    technology = "HTTPS"
  }
}

system "shop" {
  description = "Storefront"
}

container "web" {
  system     = system.shop
  title      = "Web app"
  technology = "React"

  uses "api" {
    target     = container.app
    technology = "REST/JSON"
  }
}

container "app" {
  system     = system.shop
  title      = "Application server"
  technology = "Java (Spring Boot)"

  uses "db" {
    target     = container.db
    technology = "JDBC"
  }
}

container "db" {
  system     = system.shop
  title      = "Database"
  technology = "PostgreSQL"
  shape      = "database"
}
```

### Microservices

One `system` per independently owned service, each with its own data store. Cross-service calls
go between containers in different systems, so they show on the landscape.

```hcl
system "orders" {
  owner = "orders-team"
}

system "payments" {
  owner = "payments-team"
}

container "orders_api" {
  system     = system.orders
  technology = "Go"

  uses "store" {
    target = container.orders_db
  }

  uses "charge" {
    target      = container.payments_api
    description = "Charges the order"
    technology  = "gRPC"
  }
}

container "orders_db" {
  system = system.orders
  shape  = "database"
}

container "payments_api" {
  system     = system.payments
  technology = "Go"

  uses "store" {
    target = container.payments_db
  }
}

container "payments_db" {
  system = system.payments
  shape  = "database"
}
```

If the services are one team's units and are deployed together, model them as containers of one
system instead.

### Serverless

```hcl
system "intake" {
  description = "Order intake"
}

container "gateway" {
  system     = system.intake
  title      = "API Gateway"
  technology = "Amazon API Gateway"

  uses "invoke" {
    target = container.create_order
  }
}

container "create_order" {
  system     = system.intake
  title      = "Create order"
  technology = "AWS Lambda (Go)"
  shape      = "function"

  uses "save" {
    target = container.orders_table
  }

  uses "enqueue" {
    target = container.order_queue
    kind   = "async"
  }
}

container "orders_table" {
  system     = system.intake
  technology = "DynamoDB"
  shape      = "database"
}

container "order_queue" {
  system     = system.intake
  technology = "SQS"
  shape      = "queue"
}

container "fulfil" {
  system     = system.intake
  title      = "Fulfil order"
  technology = "AWS Lambda (Go)"
  shape      = "function"

  uses "consume" {
    target = container.order_queue
    kind   = "trigger" # drawn order_queue -> fulfil
  }
}
```

## Communication patterns

### Synchronous request/response

The default: `kind` omitted means `sync`.

```hcl
system "s" {}

container "client" {
  system = system.s

  uses "call" {
    target     = container.server
    technology = "HTTPS"
  }
}

container "server" {
  system = system.s
}
```

### Asynchronous messaging (queue)

Producer and consumer both depend on the queue. The producer's edge is `async`; the consumer's is
a `trigger` if the queue invokes it, or `async` if it polls.

```hcl
system "s" {}

container "producer" {
  system = system.s

  uses "send" {
    target = container.jobs
    kind   = "async"
  }
}

container "jobs" {
  system = system.s
  shape  = "queue"
}

container "worker" {
  system = system.s

  uses "poll" {
    target = container.jobs
    kind   = "async"
  }
}
```

### Publish/subscribe (topic)

```hcl
system "s" {}

container "orders" {
  system = system.s

  uses "publish" {
    target      = container.order_events
    description = "OrderPlaced"
    kind        = "async"
  }
}

container "order_events" {
  system     = system.s
  technology = "SNS"
  shape      = "topic"
}

container "email" {
  system = system.s

  uses "subscribe" {
    target = container.order_events
    kind   = "trigger"
  }
}

container "analytics" {
  system = system.s

  uses "subscribe" {
    target = container.order_events
    kind   = "trigger"
  }
}
```

`loko query dependents container.order_events` then lists every publisher and subscriber.

## Data patterns

### CQRS

Separate write and read containers; a projector keeps the read model current.

```hcl
system "catalog" {}

container "commands" {
  system = system.catalog
  title  = "Command API"

  uses "write" {
    target = container.write_db
  }
}

container "write_db" {
  system = system.catalog
  shape  = "database"
}

container "projector" {
  system = system.catalog

  uses "changes" {
    target = container.write_db
    kind   = "trigger"
  }

  uses "update" {
    target = container.read_db
  }
}

container "read_db" {
  system     = system.catalog
  technology = "Elasticsearch"
  shape      = "database"
}

container "queries" {
  system = system.catalog
  title  = "Query API"

  uses "read" {
    target = container.read_db
  }
}
```

### Saga (orchestrated)

An orchestrator depends on each participant. Put compensation steps in `description`; loko does
not model step order.

```hcl
system "checkout" {}

container "orchestrator" {
  system     = system.checkout
  technology = "AWS Step Functions"

  uses "reserve" {
    target      = container.inventory
    description = "Reserve stock; release on failure"
  }

  uses "charge" {
    target      = container.billing
    description = "Charge card; refund on failure"
  }
}

container "inventory" {
  system = system.checkout
}

container "billing" {
  system = system.checkout
}
```

## Infrastructure patterns

### API gateway in front of several systems

```hcl
person "client" {
  uses "call" {
    target = container.edge
  }
}

system "edge" {}

system "accounts" {}

system "orders" {}

container "edge" {
  system = system.edge
  title  = "API gateway"

  uses "accounts" {
    target = container.accounts_api
  }

  uses "orders" {
    target = container.orders_api
  }
}

container "accounts_api" {
  system = system.accounts
}

container "orders_api" {
  system = system.orders
}
```

### Deployment

Where containers run, and which infrastructure resources implement them, is the deployment plane:

```hcl
system "s" {}

container "api" {
  system = system.s
}

deployment "prod" {
  provider = "aws"
  region   = "us-east-1"

  node "vpc" {
    instance "api" {
      of = container.api

      binding "terraform" {
        address = "module.api.aws_lambda_function.this"
      }
    }
  }
}
```

## Focused views

Generated views cover each level. Add a `view` for a cross-cutting cut, such as everything tagged
for a compliance scope or one request path:

```hcl
system "s" {}

container "api" {
  system = system.s
  tags   = ["pci"]
}

view "pci-scope" {
  tags      = ["pci"]
  direction = "right"
}
```

## Checking a design with queries

| Question | Command (or MCP `query` kind) |
|---|---|
| What breaks if this changes? | `loko query dependents container.orders_db --transitive` |
| What does this service rely on? | `loko query dependencies system.orders` |
| How does a request reach this? | `loko query path person.client container.orders_api` |
| What is unconnected? | `loko query orphans` |
| Where is coupling concentrated? | `loko query coupling` |

Signs of trouble:

- **Distributed monolith**: services in different systems with high fan-in and fan-out to each
  other in `coupling`.
- **Hidden async**: a queue or topic drawn with solid edges; set `kind`.
- **Missing boundary**: components in one system using components in another directly; route
  through an API container.

## Changing a design through MCP

1. `describe` with `level: structure` to see the current tree and get a `revision`.
2. `apply_edit` with the whole pattern in one batch (adds of elements, then relationships), with
   `preview: true` first to see the diffs.
3. `validate`, then `query` to check the result.

See [MCP Integration](../mcp-integration.md) for the edit format.
