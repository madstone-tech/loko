container "api" {
  system = system.payments
  docs   = "./docs/api.md"

  uses "gw" {
    target = container.gateway
  }
}
