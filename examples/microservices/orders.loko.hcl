system "order_service" {
  title       = "Order Service"
  description = "Manages the order lifecycle and processing"
  owner       = "orders-team"
  docs        = "./docs/order-service.md"
  tags        = ["orders"]
}

container "order_api" {
  system      = system.order_service
  title       = "Order API"
  description = "Order creation, status, cancellation and history"
  technology  = "Go, gRPC"
  docs        = "./docs/order-service.md"

  uses "store" {
    target      = container.order_db
    description = "Reads and writes orders"
    technology  = "SQL"
  }

  uses "customer" {
    target      = container.user_api
    description = "Looks up the ordering user"
    technology  = "gRPC"
  }

  uses "notify" {
    target      = container.notifier
    description = "Sends immediate notifications"
    technology  = "gRPC"
  }

  uses "publish" {
    target      = container.event_bus
    description = "Publishes order.created, order.paid, order.shipped, order.delivered"
    technology  = "Kafka"
    kind        = "async"
  }

  uses "consume" {
    target      = container.event_bus
    description = "Consumes user.deleted to anonymize orders"
    technology  = "Kafka"
    kind        = "trigger"
  }
}

container "order_db" {
  system      = system.order_service
  title       = "Order DB"
  description = "Orders and order status history"
  technology  = "PostgreSQL"
  shape       = "database"
  docs        = "./docs/order-service.md"
}
