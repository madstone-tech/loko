project "order-processing" {
  description  = "Serverless order processing on AWS Lambda, API Gateway, SQS and DynamoDB"
  loko_version = "~> 1.0"
}

# Lambda functions live in functions.loko.hcl and the production environment in
# deployment.loko.hcl. loko merges every *.loko.hcl under the project root.

person "customer" {
  title       = "Customer"
  description = "Places and tracks orders"
  docs        = "./docs/actors.md"

  uses "orders" {
    target      = container.api_gateway
    description = "Creates and views orders"
    technology  = "HTTPS"
  }
}

person "admin" {
  title       = "Admin"
  description = "Lists and cancels orders"
  docs        = "./docs/actors.md"

  uses "manage" {
    target      = container.api_gateway
    description = "Manages orders"
    technology  = "HTTPS"
  }
}

system "order_processing" {
  title       = "Order Processing API"
  description = "Order creation, payment processing and fulfillment"
  owner       = "orders-team"
  docs        = "./docs/order-processing.md"
  tags        = ["serverless"]
}

# Managed AWS resources

container "api_gateway" {
  system      = system.order_processing
  title       = "API Gateway"
  description = "REST API for order operations; Cognito authorizer on every route"
  technology  = "API Gateway REST API"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws"]

  uses "authorize" {
    target      = container.user_pool
    description = "Validates the bearer token"
    technology  = "JWT"
  }
}

container "user_pool" {
  system      = system.order_processing
  title       = "User pool"
  description = "Customer and admin sign-in"
  technology  = "Cognito User Pool"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws"]
}

container "order_queue" {
  system      = system.order_processing
  title       = "Order processing queue"
  description = "One message per new order"
  technology  = "Amazon SQS"
  shape       = "queue"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws", "async"]

  uses "dead_letter" {
    target      = container.order_dlq
    description = "After 3 failed receives"
    technology  = "Redrive policy"
    kind        = "async"
  }
}

container "order_dlq" {
  system      = system.order_processing
  title       = "Dead-letter queue"
  description = "Messages that failed processing"
  technology  = "Amazon SQS"
  shape       = "queue"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws", "async"]
}

container "schedule" {
  system      = system.order_processing
  title       = "Schedules"
  description = "Daily report and expired-order cleanup"
  technology  = "EventBridge rules"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws"]
}

container "orders_table" {
  system      = system.order_processing
  title       = "Orders table"
  description = "Orders and their status"
  technology  = "DynamoDB"
  shape       = "database"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws"]
}

container "secrets" {
  system      = system.order_processing
  title       = "Secrets"
  description = "Provider API keys"
  technology  = "AWS Secrets Manager"
  docs        = "./docs/aws-resources.md"
  tags        = ["aws"]
}

# External integrations

external "payment_provider" {
  title       = "Payment provider"
  description = "Card payment processing"
  docs        = "./docs/external-services.md"
}

external "email_provider" {
  title       = "Email provider"
  description = "Transactional email delivery"
  docs        = "./docs/external-services.md"
}

external "inventory_service" {
  title       = "Inventory service"
  description = "Stock levels and reservations"
  docs        = "./docs/external-services.md"
}

view "order-intake" {
  include   = [person.customer, container.api_gateway, container.api_handlers, container.order_queue]
  direction = "right"
}

view "async-processing" {
  tags      = ["async"]
  include   = [container.event_processors]
  direction = "right"
}
