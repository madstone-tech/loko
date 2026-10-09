person "user" {
  title = "End user"
}

system "shop" {
  title = "Shop"
}

container "db" {
  system = system.shop
  title  = "Orders database"
  shape  = "database"
}

container "worker" {
  system = system.shop
  shape  = "function"

  uses "consume" {
    target = container.db
    kind   = "trigger"
    tags   = ["read"]
  }
}

external "psp" {
  shape = "queue"
}

view "flow" {
  include   = [system.shop]
  direction = "right"
}
