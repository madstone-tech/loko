project "p" {}

system "shop" {}

container "api" {
  system     = system.shop
  technology = "AWS Lambda"
}

deployment "prod" {
  node "aws" {
    title      = "AWS account 1234"
    technology = "AWS"

    node "vpc" {
      title       = "Production VPC"
      technology  = "Amazon VPC"
      description = "Private subnets in eu-west-1a and 1b"
      tags        = ["network", "private"]

      instance "api" {
        of         = container.api
        technology = "AWS Lambda (arm64)"
      }
    }
  }
}

deployment "dev" {
  instance "api" {
    of         = container.api
    technology = "local process"
  }
}

view "phase_1" {
  title       = "Phase 1: ingestion only"
  description = "What runs before the pricing engine exists."
  include     = [system.shop]
}
