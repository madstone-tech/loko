deployment "staging" {
  provider = "aws"
  region   = "us-east-1"

  instance "web" {
    of = container.web
    binding "terraform" { address = "module.web.aws_s3_bucket.this" }
  }
  instance "api" {
    of = container.api
    binding "terraform" { address = "module.api.aws_lambda_function.this" }
  }
  instance "gateway" {
    of = container.gateway
    binding "terraform" { address = "module.gateway.aws_lambda_function.this" }
  }
}

deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    node "subnet-a" {
      instance "web" {
        of = container.web
        binding "terraform" { address = "module.web.aws_s3_bucket.prod" }
      }
      instance "api" {
        of = container.api
        binding "terraform" { address = "module.api.aws_lambda_function.prod" }
      }
    }

    instance "gateway" {
      of = container.gateway
      binding "terraform" { address = "module.gateway.aws_lambda_function.prod" }
    }
    instance "ledger" {
      of = container.ledger
      binding "terraform" { address = "module.ledger.aws_db_instance.prod" }
    }
  }
}
