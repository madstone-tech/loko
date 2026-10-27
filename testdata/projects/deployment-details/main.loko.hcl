project "deployment-details" {
  description = "A serverless system deployed as account, region and VPC (feature 018 fixture)"
}

person "customer" {
  title       = "Customer"
  description = "Places and tracks orders"
  docs        = "docs/customer.md"

  uses "orders" {
    target     = container.api
    technology = "HTTPS"
  }
}

system "shop" {
  title       = "Shop"
  description = "Takes and fulfils orders"
  docs        = "docs/shop.md"
}

container "api" {
  system      = system.shop
  title       = "Order API"
  description = "Accepts and lists orders"
  technology  = "AWS Lambda (Go)"
  docs        = "docs/api.md"

  uses "store" {
    target     = container.orders
    technology = "AWS SDK"
  }

  uses "receipts" {
    target     = container.receipts
    technology = "AWS SDK"
  }
}

container "orders" {
  system      = system.shop
  title       = "Orders table"
  description = "One item per order"
  technology  = "DynamoDB"
  shape       = "database"
  docs        = "docs/orders.md"
}

container "receipts" {
  system      = system.shop
  title       = "Receipts bucket"
  description = "PDF receipts"
  technology  = "S3"
  shape       = "bucket"
  docs        = "docs/receipts.md"
}

deployment "prod" {
  node "aws" {
    title      = "AWS account 1234"
    technology = "AWS"

    node "eu_west_1" {
      node "vpc" {
        title       = "Production VPC"
        technology  = "Amazon VPC"
        description = "Private subnets in eu-west-1a and eu-west-1b"
        tags        = ["network", "private"]

        instance "api" {
          of = container.api
          binding "terraform" { address = "module.api.aws_lambda_function.this" }
        }
        instance "orders" {
          of = container.orders
          binding "terraform" { address = "module.orders.aws_dynamodb_table.this" }
        }
        instance "receipts" {
          of = container.receipts
          binding "terraform" { address = "module.receipts.aws_s3_bucket.this" }
        }
      }
    }
  }
}

deployment "dev" {
  node "laptop" {
    instance "api" {
      of         = container.api
      technology = "local process"
      binding "terraform" { address = "module.api_dev.null_resource.this" }
    }
  }
}

view "request_path" {
  title       = "How an order is stored"
  description = "The order path only: receipts are left out."
  include     = [person.customer, container.api, container.orders]
}
