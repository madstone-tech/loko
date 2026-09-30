project "deployment-nested" {}

system "core" {}

container "edge" {
  system = system.core
  uses "svc" { target = container.svc }
}

container "svc" {
  system = system.core
  uses "store" { target = container.store }
  uses "report" { target = container.reporting }
}

container "store" { system = system.core }

container "reporting" {
  system      = system.core
  description = "Deployed elsewhere; boundary edge in prod"
}

deployment "prod" {
  node "l1" {
    instance "edge" {
      of = container.edge
      binding "terraform" { address = "module.edge.this" }
    }
    node "l2" {
      node "l3" {
        instance "svc" {
          of = container.svc
          binding "terraform" { address = "module.svc.this" }
        }
        node "l4" {
          node "l5" {
            instance "store" {
              of = container.store
              binding "terraform" { address = "module.store.this" }
            }
          }
        }
      }
    }
  }
}
