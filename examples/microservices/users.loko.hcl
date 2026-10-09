system "user_service" {
  title       = "User Service"
  description = "Manages user accounts and authentication"
  owner       = "identity-team"
  docs        = "./docs/user-service.md"
  tags        = ["users"]
}

container "user_api" {
  system      = system.user_service
  title       = "User API"
  description = "Registration, login, profiles and tokens"
  technology  = "Go, gRPC"
  docs        = "./docs/user-service.md"

  uses "store" {
    target      = container.user_db
    description = "Reads and writes users"
    technology  = "SQL"
  }

  uses "sessions" {
    target      = container.session_cache
    description = "Caches sessions"
  }

  uses "events" {
    target      = container.event_bus
    description = "Publishes user.created, user.updated, user.deleted"
    technology  = "Kafka"
    kind        = "async"
  }
}

container "user_db" {
  system      = system.user_service
  title       = "User DB"
  description = "User accounts and profiles"
  technology  = "PostgreSQL"
  shape       = "database"
  docs        = "./docs/user-service.md"
}

container "session_cache" {
  system      = system.user_service
  title       = "Session cache"
  description = "Active sessions"
  technology  = "Redis"
  shape       = "database"
  docs        = "./docs/user-service.md"
}
