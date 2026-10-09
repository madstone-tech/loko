system "notification_service" {
  title       = "Notification Service"
  description = "Handles all user notifications across email, SMS and push"
  owner       = "engagement-team"
  docs        = "./docs/notification-service.md"
  tags        = ["notifications"]
}

container "notifier" {
  system      = system.notification_service
  title       = "Notifier"
  description = "Sends notifications and manages preferences"
  technology  = "Node.js, gRPC"
  docs        = "./docs/notification-service.md"

  uses "events" {
    target      = container.event_bus
    description = "Consumes order.* and user.created"
    technology  = "Kafka"
    kind        = "trigger"
  }

  uses "store" {
    target      = container.notification_db
    description = "Reads templates and preferences"
  }

  uses "email" {
    target      = external.email_provider
    description = "Sends email"
    technology  = "HTTPS"
    kind        = "async"
  }

  uses "sms" {
    target      = external.sms_provider
    description = "Sends SMS"
    technology  = "HTTPS"
    kind        = "async"
  }

  uses "push" {
    target      = external.push_provider
    description = "Sends push notifications"
    technology  = "HTTPS"
    kind        = "async"
  }
}

container "notification_db" {
  system      = system.notification_service
  title       = "Notification DB"
  description = "Templates and user preferences"
  technology  = "MongoDB"
  shape       = "database"
  docs        = "./docs/notification-service.md"
}
