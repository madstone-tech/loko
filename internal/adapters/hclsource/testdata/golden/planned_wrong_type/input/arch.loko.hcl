project "p" {}

deployment "prod" {
  instance "api" {
    of      = container.api
    planned = "yes"
  }
}
