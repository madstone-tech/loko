# Declared BEFORE the system it points at, in a file that sorts LAST.
# Resolution must not depend on declaration or file order (FR-020).
container "gateway" {
  system = system.payments
  docs   = "./docs/gateway.md"

  uses "api" {
    target = container.api
  }
}
