system "api_gateway" {
  title       = "API Gateway"
  description = "Central entry point for all client requests"
  owner       = "edge-team"
  docs        = "./docs/api-gateway.md"
  tags        = ["gateway"]
}

container "gateway" {
  system      = system.api_gateway
  title       = "Gateway"
  description = "Routes, authenticates and rate-limits requests"
  technology  = "Kong Gateway, Lua plugins"
  docs        = "./docs/api-gateway.md"
}

container "rate_limit_store" {
  system      = system.api_gateway
  title       = "Rate-limit store"
  description = "Per-client rate-limit counters"
  technology  = "Redis"
  shape       = "database"
  docs        = "./docs/api-gateway.md"
}

component "auth" {
  container   = container.gateway
  title       = "Auth middleware"
  description = "Validates JWT tokens"
  docs        = "./docs/api-gateway.md"

  uses "router" {
    target      = component.router
    description = "Validated requests"
  }
}

component "router" {
  container   = container.gateway
  title       = "Router"
  description = "Routes on path and headers"
  docs        = "./docs/api-gateway.md"

  uses "limiter" {
    target      = component.limiter
    description = "Checks the client's quota"
  }
}

component "limiter" {
  container   = container.gateway
  title       = "Rate limiter"
  description = "Per-client limits and circuit breaking"
  docs        = "./docs/api-gateway.md"

  uses "counters" {
    target      = container.rate_limit_store
    description = "Reads and increments counters"
  }

  uses "users" {
    target      = container.user_api
    description = "/api/users/*"
    technology  = "gRPC"
  }

  uses "orders" {
    target      = container.order_api
    description = "/api/orders/*"
    technology  = "gRPC"
  }

  uses "notifications" {
    target      = container.notifier
    description = "/api/notifications/*"
    technology  = "gRPC"
  }
}
