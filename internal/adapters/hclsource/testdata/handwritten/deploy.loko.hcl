// Deployment, hand-written.
deployment "prod" {
  region   = "us-east-1"

  node "vpc" {
    # inner comment
    node "subnet-a" {
      instance "api" {
        of = container.api
        binding "terraform" { address = "module.api.this" }
      }
    }
  }

  instance "gateway" {
    of = container.gateway
    binding "terraform" {
      address = "module.gateway.this"   // trailing
    }
  }
}
