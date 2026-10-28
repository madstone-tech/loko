project "p" {}

system "shop" {}

container "web" {
  system = system.shop

  uses "orders" {
    target = container.api
  }
}

container "api" {
  system = system.shop
}

view "checkout" {
  title = "Checkout"
  style = "sequence"

  step "order" {
    from        = container.web
    to          = container.api
    via         = container.web.uses.orders
    description = "Places the order"
  }

  step "ack" {
    from  = container.api
    to    = container.web
    reply = true
  }

  step "again" {
    from = container.web
    to   = container.api
  }
}

view "selected" {
  include = [system.shop]
  exclude = [container.api]
  tags    = ["pci"]
}
