system "shop" {
  title = ["not", "a", "string"]
}

container "db" {
  system = system.shop

  uses "self" {
    target = container.db
    tags   = "read"
  }
}
