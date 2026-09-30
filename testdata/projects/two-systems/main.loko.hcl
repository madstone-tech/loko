project "two-systems" {
  description = "Renderer fixture: two systems, six containers, two environments"
}

person "customer" {
  description = "Buys things"
  docs        = "./docs/customer.md"

  uses "browses" {
    target      = container.web
    description = "Browses the catalogue"
    technology  = "HTTPS"
  }
}

external "bank" {
  description = "Card network and acquiring bank"
  docs        = "./docs/bank.md"
}

system "shop" {
  description = "Storefront and order management"
  owner       = "commerce-team"
  docs        = "./docs/shop.md"
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = "platform-team"
  docs        = "./docs/payments.md"
  tags        = ["pci"]
}

system "archive" {
  description = "Cold storage, nothing modelled yet"
  docs        = "./docs/archive.md"
}

container "web" {
  system      = system.shop
  description = "Single-page storefront"
  technology  = "React"
  tags        = ["edge"]
  docs        = "./docs/web.md"

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
  docs        = "./docs/api.md"

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
  docs        = "./docs/orders_db.md"
}

container "gateway" {
  system      = system.payments
  description = "Payment gateway"
  technology  = "Go"
  tags        = ["pci"]
  docs        = "./docs/gateway.md"

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
  docs        = "./docs/ledger.md"
}

container "fraud" {
  system      = system.payments
  description = "Fraud scoring"
  technology  = "Python"
  docs        = "./docs/missing-fraud.md"

  uses "read" {
    target      = container.ledger
    description = "Reads recent movements"
  }
}

component "handler" {
  container   = container.api
  description = "HTTP entry point"
  docs        = "./docs/handler.md"

  uses "repo" {
    target      = component.repo
    description = "Loads and saves orders"
  }
}

component "repo" {
  container   = container.api
  description = "Order repository"
  docs        = "./docs/repo.md"

  uses "db" {
    target      = container.orders_db
    description = "SQL queries"
  }
}

component "authorizer" {
  container   = container.gateway
  description = "Card authorization"
  docs        = "./docs/authorizer.md"

  uses "capture" {
    target      = component.capture
    description = "Hands off approved authorizations"
  }
}

component "capture" {
  container   = container.gateway
  description = "Settlement capture"
  docs        = "./docs/capture.md"
}
