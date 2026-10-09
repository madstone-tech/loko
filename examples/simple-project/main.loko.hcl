project "simple-api" {
  description  = "A simple REST API architecture example"
  loko_version = "~> 1.0"
}

person "client" {
  title       = "Client"
  description = "An application or developer calling the REST API"
  docs        = "./docs/client.md"

  uses "requests" {
    target      = component.handlers
    description = "Calls the REST API"
    technology  = "HTTPS/JSON"
  }
}

system "api_service" {
  title       = "API Service"
  description = "RESTful API service for the application"
  docs        = "./docs/api-service.md"
  tags        = ["api", "backend", "rest"]
}

container "api" {
  system      = system.api_service
  title       = "API"
  description = "Go HTTP service: handlers, business services and repositories"
  technology  = "Go, net/http"
  docs        = "./docs/api.md"
}

container "database" {
  system      = system.api_service
  title       = "PostgreSQL"
  description = "Primary relational store"
  technology  = "PostgreSQL 15"
  shape       = "database"
  docs        = "./docs/data-stores.md"
}

container "cache" {
  system      = system.api_service
  title       = "Redis cache"
  description = "Read-through cache for hot lookups"
  technology  = "Redis 7"
  shape       = "database"
  docs        = "./docs/data-stores.md"
}

component "handlers" {
  container   = container.api
  title       = "HTTP handlers"
  description = "Request and response handling, input validation"
  docs        = "./docs/api.md"

  uses "services" {
    target      = component.services
    description = "Calls"
  }
}

component "services" {
  container   = container.api
  title       = "Business services"
  description = "Business logic"
  docs        = "./docs/api.md"

  uses "repos" {
    target      = component.repos
    description = "Uses"
  }

  uses "cache" {
    target      = container.cache
    description = "Caches lookups"
    technology  = "Redis protocol"
  }
}

component "repos" {
  container   = container.api
  title       = "Repositories"
  description = "Data access"
  docs        = "./docs/api.md"

  uses "database" {
    target      = container.database
    description = "Reads and writes"
    technology  = "SQL"
  }
}

view "request-path" {
  include   = [container.api, container.database, container.cache]
  direction = "right"
}
