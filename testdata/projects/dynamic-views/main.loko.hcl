project "dynamic-views" {
  description = "Ordered flows over a small shop (feature 019 fixture)"
}

person "customer" {
  title       = "Customer"
  description = "Places orders"
  docs        = "docs/customer.md"

  uses "browse" {
    target     = container.web
    technology = "HTTPS"
  }
}

external "payments" {
  title       = "Payment provider"
  description = "Takes card payments"
}

system "shop" {
  title       = "Shop"
  description = "Takes and fulfils orders"
  docs        = "docs/shop.md"
}

container "web" {
  system      = system.shop
  title       = "Web app"
  description = "The storefront"
  technology  = "React"
  docs        = "docs/web.md"

  uses "orders" {
    target      = container.api
    description = "Places the order"
    technology  = "JSON/HTTPS"
  }
}

container "api" {
  system      = system.shop
  title       = "Order API"
  description = "Accepts orders"
  technology  = "Go"
  docs        = "docs/api.md"

  uses "charge" {
    target      = external.payments
    description = "Charges the card"
    technology  = "HTTPS"
  }

  uses "enqueue" {
    target      = container.queue
    description = "Queues fulfilment"
    technology  = "SQS"
    kind        = "async"
  }
}

component "handler" {
  container   = container.api
  description = "HTTP handlers"
  docs        = "docs/api.md"

  uses "save" {
    target = component.repo
  }
}

component "repo" {
  container   = container.api
  description = "Order persistence"
  docs        = "docs/api.md"

  uses "write" {
    target      = container.db
    description = "Stores the order"
    technology  = "SQL"
  }
}

container "queue" {
  system      = system.shop
  title       = "Fulfilment queue"
  description = "Orders waiting to be fulfilled"
  technology  = "SQS"
  shape       = "queue"
  docs        = "docs/queue.md"
}

container "worker" {
  system      = system.shop
  title       = "Fulfilment worker"
  description = "Ships orders"
  technology  = "AWS Lambda"
  shape       = "function"
  docs        = "docs/worker.md"

  uses "triggered" {
    target = container.queue
    kind   = "trigger"
  }

  uses "update" {
    target      = container.db
    description = "Marks the order shipped"
    technology  = "SQL"
  }
}

container "db" {
  system      = system.shop
  title       = "Orders database"
  description = "One row per order"
  technology  = "PostgreSQL"
  shape       = "database"
  docs        = "docs/db.md"
}

view "checkout" {
  title       = "Checkout"
  description = "From the storefront to a stored, paid order."

  step "browse" {
    from = person.customer
    to   = container.web
  }

  step "order" {
    from = container.web
    to   = container.api
    via  = container.web.uses.orders
  }

  step "charge" {
    from        = container.api
    to          = external.payments
    description = "Charge card"
  }

  step "store" {
    from = container.api
    to   = container.db
  }
}

view "async_order" {
  title = "Fulfilment"

  step "enqueue" {
    from = container.api
    to   = container.queue
  }

  step "trigger" {
    from        = container.queue
    to          = container.worker
    description = "Invokes the worker"
  }

  step "ship" {
    from = container.worker
    to   = container.db
  }
}

view "checkout_sequence" {
  title = "Checkout, as a sequence"
  style = "sequence"

  step "browse" {
    from = person.customer
    to   = container.web
  }

  step "order" {
    from = container.web
    to   = container.api
    via  = container.web.uses.orders
  }

  step "charge" {
    from        = container.api
    to          = external.payments
    description = "Charge card"
  }

  step "charged" {
    from        = external.payments
    to          = container.api
    reply       = true
    description = "Payment confirmed"
  }

  step "store" {
    from = container.api
    to   = container.db
  }
}
