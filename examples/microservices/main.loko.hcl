project "order-management" {
  description  = "A microservices-based order management system with event-driven communication"
  loko_version = "~> 1.0"
}

# Services are split across files, one per service:
#   gateway.loko.hcl, users.loko.hcl, orders.loko.hcl, notifications.loko.hcl, platform.loko.hcl,
#   and deployment.loko.hcl for the production environment.
# loko merges every *.loko.hcl under the project root into one architecture.

person "customer" {
  title       = "Customer"
  description = "Uses the web and mobile apps to manage an account and place orders"
  docs        = "./docs/actors.md"

  uses "apps" {
    target      = container.gateway
    description = "Web and mobile app requests"
    technology  = "HTTPS"
  }
}

external "partner" {
  title       = "Third-party integrator"
  description = "Partner systems calling the public API"
  docs        = "./docs/actors.md"

  uses "api" {
    target      = container.gateway
    description = "Public API calls"
    technology  = "HTTPS"
  }
}

external "email_provider" {
  title       = "Email provider"
  description = "Transactional and marketing email delivery"
  docs        = "./docs/providers.md"
}

external "sms_provider" {
  title       = "SMS provider"
  description = "Text message delivery"
  docs        = "./docs/providers.md"
}

external "push_provider" {
  title       = "Push provider"
  description = "Mobile push notification delivery"
  docs        = "./docs/providers.md"
}

view "order-flow" {
  include   = [container.order_api, container.event_bus, container.notifier]
  direction = "right"
}

view "edge" {
  include = [system.api_gateway]
}
