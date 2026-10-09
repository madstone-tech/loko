# The loko language

**Status**: normative for v1.0 · **Syntax**: HCL v2 native

You describe an architecture in `*.loko.hcl` files. This page is the complete grammar — every block,
every attribute, every function. There is nothing else.

Everything here is permanent surface for v1.x. Your files are the source of truth and loko has no
migration command, so a construct cannot be removed once it ships. That is why the language is
smaller than you might expect.

> **New here?** Write the example under [`project`](#project) into `arch.loko.hcl`, run
> `loko validate`, and add blocks from this page one at a time. Every error names a file, a line,
> and a column, and most name the thing you probably meant.

Discovery: every file matching `*.loko.hcl` beneath the project root, recursively, merged into one
architecture (FR-001). Anything else under the root is ignored (FR-002). JSON syntax is not accepted.

---

## Top-level blocks

| Block | Labels | Cardinality |
|---|---|---|
| `project` | name | exactly 1 across all files |
| `person` | name | 0..n |
| `system` | name | 0..n |
| `container` | name | 0..n |
| `component` | name | 0..n |
| `external` | name | 0..n |
| `deployment` | name | 0..n |
| `view` | name | 0..n |
| `reconcile` | — | 0..1 |
| `locals` | — | 0..n |
| `moved` | — | 0..n |

Any other top-level block is `unknown_block` (FR-018, FR-028). In particular `for_each`, `dynamic`,
`variable`, and `module` are rejected by name with a message stating they are deliberately excluded
from v1.0, so the error teaches rather than merely refuses.

---

## `project`

```hcl
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}
```

| Attribute | Type | Required |
|---|---|---|
| `description` | string | no |
| `layout` | `"dagre"` or `"elk"` | no |
| `loko_version` | string constraint | no |

`layout` picks the layout engine for every diagram, generated or declared; a view can override it.
`dagre` (the default) is fast. `elk` routes edges at right angles and packs dense views more
tightly, but lays out roughly five times slower, so it suits small and medium projects.

`loko_version` accepts `=`, `!=`, `>`, `>=`, `<`, `<=`, `~>`, and comma-separated conjunctions
(`">= 1.0, < 2.0"`). Unsatisfied → `version_unsatisfied` error (FR-028).

---

## Logical elements

```hcl
system "payments" {
  description = "Authorization, capture, settlement"
  owner       = "platform-team"
  docs        = "./docs/payments.md"
  tags        = ["pci"]
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"
  tags       = ["public", "pci"]

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "PostgreSQL wire protocol"
  }
}

component "handler" {
  container   = container.api
  description = "HTTP entry point"
}
```

| Attribute | Applies to | Type | Notes |
|---|---|---|---|
| `description` | all | string | |
| `owner` | all | string | |
| `technology` | all | string | |
| `tags` | all | list(string) | sorted and de-duplicated in the IR |
| `docs` | all | string | project-relative path; missing file → `docs_not_found` warning |
| `title` | all | string | display title in diagrams, pages and markdown; the address keeps the name. Empty → `empty_title` |
| `shape` | `container`, `external` | string | `database`, `queue`, `topic`, `function` or `bucket`; elsewhere → `shape_not_allowed`, other values → `invalid_attribute_value` |
| `system` | `container` | reference | **required**; must resolve to a `system` |
| `container` | `component` | reference | **required**; must resolve to a `container` |

`person` and `external` take the common attributes only and have no parent.

### `uses` — relationships

Nested inside the element they originate from (FR-008). The label is the relationship's local name
and yields the address `container.api.uses.orders` (FR-024).

| Attribute | Type | Required |
|---|---|---|
| `target` | reference | **yes** |
| `description` | string | no |
| `technology` | string | no |
| `kind` | string | no: `sync` (default), `async` or `trigger` |
| `tags` | list(string) | no |

`kind` and `tags` change only how the relationship is drawn, never queries:

- `async` is drawn with long dashes. Short dashes mean "leaves this view".
- `trigger` says the declaring element is *invoked by* its target, such as a Lambda fed by a queue.
  Write it on the invoked element, targeting its trigger, like any other relationship: the
  declaring element still depends on its target. Diagrams draw the arrow from the trigger, the way
  the data flows.

```hcl
container "sync" {
  system = system.portal
  title  = "Sync Lambda"
  shape  = "function"

  uses "consume" {
    target = container.sync_queue
    kind   = "trigger"      # drawn sync_queue → sync
  }
}
```

A `uses` block targeting its own element is permitted and produces a `self_relationship` warning.
Cycles across elements are legal and are **not** diagnosed (FR-032).

---

## References

A reference is a bare traversal, never a quoted string:

```hcl
system = system.payments     # correct
system = "system.payments"   # a string; rejected as the wrong type
```

References are extracted structurally and never evaluated (research R2), so they cannot be computed
or interpolated. Resolution is order-independent and crosses file boundaries (FR-019, FR-020).

---

## Deployment plane

```hcl
deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    node "subnet-a" {
      instance "api" {
        of         = container.api
        attributes = { memory = 1024, timeout = 30 }

        binding "terraform" {
          address = "module.api.aws_lambda_function.this"
        }
      }
    }
  }
}
```

`node` blocks nest to arbitrary depth (FR-012). **A `node` does not contribute a segment to the
addresses of the instances inside it** — the instance above is `deployment.prod.instance.api`, while
the node is `deployment.prod.node.vpc-main.subnet-a`. Instance names are therefore unique per
`deployment` (FR-012a); a collision across two different nodes is a `duplicate_instance_name` error.

| Block | Attribute | Type | Required |
|---|---|---|---|
| `deployment` | `provider`, `account`, `region` | string | no |
| `instance` | `of` | reference to a logical element | **yes** |
| `instance` | `attributes` | map of scalars | no |
| `binding` | `address` | string, exact or glob (`module.api.*`) | one of the three |
| `binding` | `addresses` | list(string) | one of the three |
| `binding` | `tags` | map(string) | one of the three |

`binding` takes one label, its kind — `terraform` or `cloudformation` in v1.0. Two bindings selecting
the same exact identifier → `duplicate_claim` error. An instance with no binding → `unbound_instance`
warning (FR-029), not an error.

---

## `reconcile`

```hcl
reconcile {
  ignore = [
    "aws_iam_role_policy_attachment.*",
    "aws_cloudwatch_log_group.*",
  ]
}
```

Parsed, validated, and carried into the IR unchanged (FR-015). Consumed by the reconciliation stage;
it has no effect on this release's behaviour.

---

## `view`

```hcl
view "payment-path" {
  include = [system.payments, container.api]
  exclude = [container.orders_db]
  tags    = ["pci"]
}
```

`direction` (`down` or `right`) sets the layout; declared views default to `down`. Generated
container and component views are laid out `down`, and the landscape and deployment views `right`.
`layout` (`dagre` or `elk`) overrides the project's layout engine for this view.

Every reference must resolve. `loko build` renders each view as `diagrams/<label>.{d2,svg}` and a
site page, alongside the views it derives automatically.

**Selection.** The view shows:

- every `include` element **and all its descendants**,
- plus every element carrying **any** of the `tags`,
- minus every `exclude` element **and all its descendants**.

With neither `include` nor `tags`, the view starts from every element. Elements nest under their
nearest shown ancestor. A connection to something not shown is drawn reaching the view's boundary
(an `outside` marker) rather than dropped; that includes connections to excluded elements. A view
that selects nothing produces a `view_empty` warning and no file.

**Automatic views** need no declaration:

| View | Produced when | Name |
|---|---|---|
| Landscape of every top-level element | anything is declared | `landscape` |
| One system's containers, and what they talk to | the system has containers | `system-<name>` |
| One container's components | the container has components | `container-<name>` |
| One environment's placement and instances | the environment has instances | `deployment-<name>` |

A declared `view` whose label equals an automatic view's name replaces it, and a `view_shadowed`
warning says so. The output layout is described in the
[CLI reference](cli-reference.md#loko-build).

---

## `locals` and functions

```hcl
locals {
  team = "platform"
  tier = join("-", ["public", "edge"])
}

system "payments" {
  owner = local.team
}
```

Exactly five functions are available: `join`, `split`, `lower`, `upper`, `replace` (FR-017). Native
string interpolation (`"${local.team}-api"`) is supported. Any other call is an `unknown_function`
error whose detail lists the five (FR-017a).

Locals may reference other locals but not elements; a local is a value, and references are not values
(research R2).

---

## `moved`

```hcl
moved {
  from = container.api
  to   = component.api
}
```

A `moved` block records a rename, so the history of an element survives it. The element used to
be declared at `from` and is now declared at `to`. The kind may change as well as the name. You
rarely write one by hand: the MCP `move` tool, and a `rename` edit through `apply_edit`, rename
the element, rewrite every reference to it, and append the block for you.

| Rule | Diagnostic |
|---|---|
| The block is top-level and has no label; `from` and `to` are both required | `unknown_block` / `unknown_attribute` |
| `from` and `to` are bare element addresses (`kind.name`), never strings or expressions | `wrong_reference_kind` |
| `from` must not name a declared element (it no longer exists) | `moved_from_declared` |
| `to` must name a declared element, directly or along a chain of moves | `moved_to_unresolved` |
| An address is the `from` of at most one `moved` block | `moved_duplicate_from` (related range: the first block) |

- **Chains** are valid: `a → b` and `b → c` with only `c` declared. Renaming through the tools
  keeps chains short, because a rename also rewrites earlier blocks' `to`.
- **Renaming back** to an address the element once had removes the block that recorded the
  move away from it.
- **Removing** an element removes the `moved` blocks that lead to it: a removed element has no
  address for its history to point at.
- **A kind change** drops the parent attribute the new kind does not take (`system` on a
  container, `container` on a component). If the new kind needs a different parent, set it in the
  same `apply_edit` batch: the rename, then an update.
- Moves appear in the IR as `moves`, sorted by `from` and omitted when there are none, so a
  project without renames exports exactly as before. Rendering and queries ignore them; the diff
  stage uses them to recognise a renamed element as the same element.

---

## Compatibility commitments

1. Every block, attribute, and function above is permanent for v1.x. Removal requires a major version.
2. Additions are minor-version changes; a file using a newer construct fails on an older tool with
   `unknown_block` or `unknown_attribute`, which is why `loko_version` exists.
3. Address forms are frozen (see [data-model.md](../data-model.md) §1). The diff stage depends on
   their stability, so changing one silently would make every historical comparison wrong.

---

## A complete example

Everything in this file compiles. `loko validate` reports no errors on it.

```hcl
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}

locals {
  team = "platform"
}

person "customer" {
  description = "Buys things"
  docs        = "./docs/customer.md"

  uses "checkout" {
    target      = container.api
    description = "Places an order"
  }
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = local.team
  docs        = "./docs/payments.md"
  tags        = ["pci"]
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"
  docs       = "./docs/api.md"
  tags       = ["public", "pci"]

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "PostgreSQL wire protocol"
  }

  uses "cards" {
    target      = external.stripe
    description = "Authorizes and captures card payments"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
  docs       = "./docs/db.md"
  tags       = ["pci"]
}

component "handler" {
  container   = container.api
  description = "HTTP entry point"
  docs        = "./docs/handler.md"

  uses "store" {
    target      = container.orders_db
    description = "Persists the order"
  }
}

external "stripe" {
  description = "Card processing"
  docs        = "./docs/stripe.md"
}

deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    node "subnet-a" {
      instance "api" {
        of         = container.api
        attributes = { memory = 1024, timeout = 30 }

        binding "terraform" {
          address = "module.api.aws_lambda_function.this"
        }
      }
    }
  }
}

view "payment-path" {
  include = [system.payments, container.api]
  exclude = [container.orders_db]
}

reconcile {
  ignore = ["aws_iam_role_policy_attachment.*"]
}
```

## Commands

| Command | What it does |
|---|---|
| `loko validate` | Compile and report every problem found, with file/line/column |
| `loko validate --strict` | Warnings fail the run too |
| `loko validate --format json` | Machine-readable diagnostics for CI |
| `loko fmt` | Rewrite source in canonical form |
| `loko fmt --check` | List non-canonical files, write nothing, exit 1 |
| `loko export --format json\|toon` | Write the compiled architecture |

Exit codes are `0` clean, `1` errors, `2` warnings under `--strict`. There are exactly three.

## Migrating from v0.2

There is no `loko migrate`. Writing a D2 arrow parser, a frontmatter reader, TOML handling, and v0's
union-merge conflict semantics was not worth it for the install base — and half of that work
duplicates the diagram-import adapter planned for v1.2.

The capability exists anyway, through MCP:

1. Keep your v0.2 checkout. The `v0.2.x` tag stays installable.
2. Point an agent at the v0 tree over MCP and ask it to write `*.loko.hcl`, reading the markdown
   frontmatter and `.d2` files as it goes.
3. Run `loko validate` after each file. Every unresolved reference is an error with a position, so
   the loop converges.
4. Delete the v0 tree once validate is clean.

What changes, concretely:

| v0.2 | v1.0 |
|---|---|
| `loko.toml` | the `project` block |
| Markdown frontmatter relationships | `uses` blocks inside the source element |
| D2 arrows as a second relationship source | removed — there is one source |
| `loko validate --check-drift` | removed. Drift cannot occur with one source of truth |
| Prose *containing* the model | prose *referenced* by `docs = "./…"` |

Prose files do not change. Point `docs` at them and they are carried through untouched.
