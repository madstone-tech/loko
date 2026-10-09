# API handlers: synchronous, invoked by API Gateway

container "api_handlers" {
  system      = system.order_processing
  title       = "API handlers"
  description = "Lambda functions handling synchronous API Gateway requests"
  technology  = "AWS Lambda (Go, arm64)"
  shape       = "function"
  docs        = "./docs/api-handlers.md"
  tags        = ["lambda"]
}

component "create_order" {
  container   = container.api_handlers
  title       = "Create order"
  description = "POST /orders: validates, stores and enqueues a new order"
  technology  = "Go"
  docs        = "./docs/create-order.md"

  uses "invoke" {
    target      = container.api_gateway
    description = "POST /orders"
    kind        = "trigger"
  }

  uses "store" {
    target      = container.orders_table
    description = "PutItem"
  }

  uses "enqueue" {
    target      = container.order_queue
    description = "SendMessage"
    technology  = "SQS"
    kind        = "async"
  }
}

component "get_order" {
  container   = container.api_handlers
  title       = "Get order"
  description = "GET /orders/{id}"
  technology  = "Go"
  docs        = "./docs/api-handlers.md"

  uses "invoke" {
    target      = container.api_gateway
    description = "GET /orders/{id}"
    kind        = "trigger"
  }

  uses "read" {
    target      = container.orders_table
    description = "GetItem"
  }
}

component "list_orders" {
  container   = container.api_handlers
  title       = "List orders"
  description = "GET /orders with pagination"
  technology  = "Go"
  docs        = "./docs/api-handlers.md"

  uses "invoke" {
    target      = container.api_gateway
    description = "GET /orders"
    kind        = "trigger"
  }

  uses "query" {
    target      = container.orders_table
    description = "Query"
  }
}

component "cancel_order" {
  container   = container.api_handlers
  title       = "Cancel order"
  description = "DELETE /orders/{id}"
  technology  = "Go"
  docs        = "./docs/api-handlers.md"

  uses "invoke" {
    target      = container.api_gateway
    description = "DELETE /orders/{id}"
    kind        = "trigger"
  }

  uses "update" {
    target      = container.orders_table
    description = "UpdateItem (conditional)"
  }
}

# Event processors: asynchronous, fed by the order queue

container "event_processors" {
  system      = system.order_processing
  title       = "Event processors"
  description = "Lambda functions processing order messages from SQS"
  technology  = "AWS Lambda (Go, arm64)"
  shape       = "function"
  docs        = "./docs/event-processors.md"
  tags        = ["lambda"]
}

component "process_payment" {
  container   = container.event_processors
  title       = "Process payment"
  description = "Charges the order through the payment provider"
  technology  = "Go"
  docs        = "./docs/process-payment.md"

  uses "consume" {
    target      = container.order_queue
    description = "Batch size 10, partial batch response"
    technology  = "SQS"
    kind        = "trigger"
  }

  uses "key" {
    target      = container.secrets
    description = "Reads the provider API key"
  }

  uses "charge" {
    target      = external.payment_provider
    description = "Creates a payment intent"
    technology  = "HTTPS"
  }

  uses "status" {
    target      = container.orders_table
    description = "UpdateItem"
  }
}

component "update_inventory" {
  container   = container.event_processors
  title       = "Update inventory"
  description = "Reserves stock for the order"
  technology  = "Go"
  docs        = "./docs/event-processors.md"

  uses "consume" {
    target      = container.order_queue
    description = "Order messages"
    technology  = "SQS"
    kind        = "trigger"
  }

  uses "reserve" {
    target      = external.inventory_service
    description = "Reserves stock"
    technology  = "HTTPS"
  }

  uses "status" {
    target      = container.orders_table
    description = "UpdateItem"
  }
}

component "send_notification" {
  container   = container.event_processors
  title       = "Send notification"
  description = "Emails the customer about their order"
  technology  = "Go"
  docs        = "./docs/event-processors.md"

  uses "consume" {
    target      = container.order_queue
    description = "Order messages"
    technology  = "SQS"
    kind        = "trigger"
  }

  uses "key" {
    target      = container.secrets
    description = "Reads the provider API key"
  }

  uses "email" {
    target      = external.email_provider
    description = "Sends email"
    technology  = "HTTPS"
    kind        = "async"
  }
}

# Scheduled tasks: invoked by EventBridge

container "scheduled_tasks" {
  system      = system.order_processing
  title       = "Scheduled tasks"
  description = "Lambda functions run on a schedule"
  technology  = "AWS Lambda (Go, arm64)"
  shape       = "function"
  docs        = "./docs/scheduled-tasks.md"
  tags        = ["lambda"]
}

component "daily_report" {
  container   = container.scheduled_tasks
  title       = "Generate daily report"
  description = "Summarizes the previous day's orders"
  technology  = "Go"
  docs        = "./docs/scheduled-tasks.md"

  uses "schedule" {
    target      = container.schedule
    description = "Daily"
    technology  = "EventBridge"
    kind        = "trigger"
  }

  uses "query" {
    target      = container.orders_table
    description = "Queries yesterday's orders"
  }
}

component "cleanup_expired" {
  container   = container.scheduled_tasks
  title       = "Clean up expired orders"
  description = "Cancels orders left unpaid past their expiry"
  technology  = "Go"
  docs        = "./docs/scheduled-tasks.md"

  uses "schedule" {
    target      = container.schedule
    description = "Hourly"
    technology  = "EventBridge"
    kind        = "trigger"
  }

  uses "expire" {
    target      = container.orders_table
    description = "Marks expired orders cancelled"
  }
}
