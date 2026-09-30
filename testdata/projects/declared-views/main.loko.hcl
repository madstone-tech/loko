project "declared-views" {
  description = "Renderer fixture: declared views"
}

person "customer" {
  description = "Buys things"

  uses "browses" {
    target      = container.web
    description = "Browses the catalogue"
    technology  = "HTTPS"
  }
}

external "bank" {
  description = "Card network and acquiring bank"
}

system "shop" {
  description = "Storefront and order management"
  owner       = "commerce-team"
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = "platform-team"
  tags        = ["pci"]
}

system "archive" {
  description = "Cold storage, nothing modelled yet"
}

container "web" {
  system      = system.shop
  description = "Single-page storefront"
  technology  = "React"
  tags        = ["edge"]

  uses "api" {
    target      = container.api
    description = "Calls the storefront API"
    technology  = "JSON/HTTPS"
  }
}

container "api" {
  system      = system.shop
  description = "Storefront API"
  technology  = "Go"

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "PostgreSQL wire protocol"
  }

  uses "charge" {
    target      = container.gateway
    description = "Charges the customer"
    technology  = "gRPC"
  }
}

container "orders_db" {
  system      = system.shop
  description = "Order store"
  technology  = "PostgreSQL"
}

container "gateway" {
  system      = system.payments
  description = "Payment gateway"
  technology  = "Go"
  tags        = ["pci"]

  uses "acquire" {
    target      = external.bank
    description = "Authorizes and captures"
    technology  = "ISO 8583"
  }

  uses "record" {
    target      = container.ledger
    description = "Records every movement"
  }
}

container "ledger" {
  system      = system.payments
  description = "Double-entry ledger"
  technology  = "PostgreSQL"
}

container "fraud" {
  system      = system.payments
  description = "Fraud scoring"
  technology  = "Python"

  uses "read" {
    target      = container.ledger
    description = "Reads recent movements"
  }
}

component "handler" {
  container   = container.api
  description = "HTTP entry point"

  uses "repo" {
    target      = component.repo
    description = "Loads and saves orders"
  }
}

component "repo" {
  container   = container.api
  description = "Order repository"

  uses "db" {
    target      = container.orders_db
    description = "SQL queries"
  }
}

component "authorizer" {
  container   = container.gateway
  description = "Card authorization"

  uses "capture" {
    target      = component.capture
    description = "Hands off approved authorizations"
  }
}

component "capture" {
  container   = container.gateway
  description = "Settlement capture"
}

view "payment-path" {
  include = [system.shop, system.payments]
  exclude = [container.ledger]
}

view "pci-only" {
  tags = ["pci"]
}

view "nothing" {
  include = [system.archive]
  exclude = [system.archive]
}

view "landscape" {
  include = [person.customer]
}
