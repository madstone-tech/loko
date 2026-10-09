system "platform" {
  title       = "Shared platform"
  description = "Infrastructure shared by every service"
  owner       = "platform-team"
  docs        = "./docs/platform.md"
}

container "event_bus" {
  system      = system.platform
  title       = "Event bus"
  description = "Domain events between services"
  technology  = "Apache Kafka"
  shape       = "topic"
  docs        = "./docs/platform.md"
}
