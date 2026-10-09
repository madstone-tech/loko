component "api" {
  container = container.core
}

container "core" {
  system = system.shop
}

system "shop" {}

moved {
  from = container.api
  to   = component.api
}

moved {
  from = system.store
  to   = system.shop
}
