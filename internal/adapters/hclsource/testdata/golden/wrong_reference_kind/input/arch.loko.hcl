project "p" {}

system "payments" {}

container "api" {
  system = "system.payments"
}
