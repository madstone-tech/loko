project "acme-payments" {
  description  = "The happy path: this fixture must produce no diagnostics."
  loko_version = "~> 1.0"
}

locals {
  team = "platform"
  tier = join("-", ["public", "edge"])
}

person "customer" {
  description = "Buys things"
  docs        = "./docs/customer.md"

  uses "checkout" {
    target = container.api
  }
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = local.team
  docs        = "./docs/payments.md"
  tags        = ["pci"]
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"
  docs       = "./docs/api.md"
  tags       = [local.tier, "pci"]

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "PostgreSQL wire protocol"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
  docs       = "./docs/db.md"
}

component "handler" {
  container = container.api
  docs      = "./docs/handler.md"

  uses "store" {
    target = container.orders_db
  }
}

external "stripe" {
  description = "Card processing"
  docs        = "./docs/stripe.md"
}

deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    node "subnet-a" {
      instance "api" {
        of         = container.api
        attributes = { memory = 1024, timeout = 30 }

        binding "terraform" {
          address = "module.api.aws_lambda_function.this"
        }
      }
    }
  }
}

view "payment-path" {
  include = [system.payments, container.api]
  exclude = [container.orders_db]
  tags    = ["pci"]
}

reconcile {
  ignore = ["aws_iam_role_policy_attachment.*"]
}
