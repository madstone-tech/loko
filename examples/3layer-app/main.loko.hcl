project "ecommerce-platform" {
  description  = "A three-tier e-commerce application with frontend, backend API and database tiers"
  loko_version = "~> 1.0"
  layout       = "elk"
}

person "customer" {
  title       = "Customer"
  description = "Browses products, manages a cart and places orders"
  docs        = "./docs/customer.md"

  uses "browses" {
    target      = component.router
    description = "Uses the storefront"
    technology  = "Browser, HTTPS"
  }
}

system "ecommerce" {
  title       = "E-Commerce Platform"
  description = "Storefront, business logic and persistence for an online shop"
  docs        = "./docs/ecommerce.md"
}

# Presentation tier

container "frontend" {
  system      = system.ecommerce
  title       = "Frontend"
  description = "Single-page application for the storefront"
  technology  = "React 18, TypeScript"
  docs        = "./docs/frontend.md"
  tags        = ["frontend", "ui"]
}

component "router" {
  container   = container.frontend
  title       = "Router"
  description = "Maps URLs to pages"
  technology  = "React Router"
  docs        = "./docs/frontend.md"

  uses "renders" {
    target      = component.ui
    description = "Renders"
  }
}

component "ui" {
  container   = container.frontend
  title       = "UI components"
  description = "Catalog, cart and checkout pages"
  technology  = "React, Tailwind CSS"
  docs        = "./docs/frontend.md"

  uses "state" {
    target      = component.store
    description = "Dispatches actions and selects state"
  }
}

component "store" {
  container   = container.frontend
  title       = "Redux store"
  description = "Cart, session and catalog state; calls the API"
  technology  = "Redux Toolkit"
  docs        = "./docs/frontend.md"

  uses "api" {
    target      = component.middleware
    description = "Calls the REST API"
    technology  = "HTTPS/JSON"
  }
}

# Application tier

container "backend" {
  system      = system.ecommerce
  title       = "Backend API"
  description = "REST service implementing all business logic"
  technology  = "Go, Chi"
  docs        = "./docs/backend.md"
  tags        = ["backend", "api"]
}

component "middleware" {
  container   = container.backend
  title       = "Middleware"
  description = "JWT authentication, logging, CORS"
  technology  = "Chi middleware"
  docs        = "./docs/backend.md"
  tags        = ["api-layer"]

  uses "handlers" {
    target      = component.handlers
    description = "Passes authenticated requests"
  }
}

component "handlers" {
  container   = container.backend
  title       = "HTTP handlers"
  description = "Products, orders and auth endpoints"
  docs        = "./docs/backend.md"
  tags        = ["api-layer"]

  uses "services" {
    target      = component.services
    description = "Calls"
  }
}

component "services" {
  container   = container.backend
  title       = "Services"
  description = "Catalog, ordering and payment logic"
  docs        = "./docs/backend.md"
  tags        = ["business-layer"]

  uses "validate" {
    target      = component.validators
    description = "Validates input"
  }

  uses "repos" {
    target      = component.repositories
    description = "Uses"
  }
}

component "validators" {
  container   = container.backend
  title       = "Validators"
  description = "Business rule and input validation"
  docs        = "./docs/backend.md"
  tags        = ["business-layer"]
}

component "repositories" {
  container   = container.backend
  title       = "Repositories"
  description = "Data access over the database driver"
  technology  = "pgx"
  docs        = "./docs/backend.md"
  tags        = ["data-layer"]

  uses "models" {
    target      = component.models
    description = "Maps rows to models"
  }

  uses "writes" {
    target      = container.primary_db
    description = "Writes"
    technology  = "SQL"
  }

  uses "reads" {
    target      = container.read_replica
    description = "Reads"
    technology  = "SQL"
  }
}

component "models" {
  container   = container.backend
  title       = "Models"
  description = "User, product and order types"
  docs        = "./docs/backend.md"
  tags        = ["data-layer"]
}

# Data tier

container "primary_db" {
  system      = system.ecommerce
  title       = "Primary DB"
  description = "Users, products and orders"
  technology  = "PostgreSQL 15"
  shape       = "database"
  docs        = "./docs/database.md"
  tags        = ["database"]

  uses "replication" {
    target      = container.read_replica
    description = "Streams changes"
    technology  = "Streaming replication"
    kind        = "async"
  }

  uses "backup" {
    target      = container.backups
    description = "Daily backup"
    kind        = "async"
  }
}

container "read_replica" {
  system      = system.ecommerce
  title       = "Read replica"
  description = "Serves read traffic"
  technology  = "PostgreSQL 15"
  shape       = "database"
  docs        = "./docs/database.md"
  tags        = ["database"]
}

container "backups" {
  system      = system.ecommerce
  title       = "Backup storage"
  description = "Daily database backups"
  technology  = "Object storage"
  shape       = "bucket"
  docs        = "./docs/database.md"
  tags        = ["database"]
}

view "request-path" {
  include   = [container.frontend, container.backend]
  direction = "right"
}

view "data-tier" {
  tags = ["database"]
}
