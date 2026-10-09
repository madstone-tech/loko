deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "eu-west-1"

  node "cluster" {
    node "edge" {
      instance "gateway" {
        of         = container.gateway
        attributes = { replicas = 3 }

        binding "terraform" {
          address = "module.gateway.helm_release.this"
        }
      }
    }

    node "services" {
      instance "user_api" {
        of         = container.user_api
        attributes = { replicas = 2 }

        binding "terraform" {
          address = "module.user_service.helm_release.this"
        }
      }

      instance "order_api" {
        of         = container.order_api
        attributes = { replicas = 3 }

        binding "terraform" {
          address = "module.order_service.helm_release.this"
        }
      }

      instance "notifier" {
        of         = container.notifier
        attributes = { replicas = 2 }

        binding "terraform" {
          address = "module.notification_service.helm_release.this"
        }
      }
    }
  }

  node "data" {
    instance "event_bus" {
      of = container.event_bus

      binding "terraform" {
        address = "module.kafka.aws_msk_cluster.this"
      }
    }

    instance "user_db" {
      of = container.user_db

      binding "terraform" {
        address = "module.user_db.aws_db_instance.this"
      }
    }

    instance "order_db" {
      of = container.order_db

      binding "terraform" {
        address = "module.order_db.aws_db_instance.this"
      }
    }

    instance "redis" {
      of = container.session_cache

      binding "terraform" {
        address = "module.redis.aws_elasticache_replication_group.this"
      }
    }
  }
}
