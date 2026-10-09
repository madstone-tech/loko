deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "edge" {
    instance "api_gateway" {
      of = container.api_gateway

      binding "cloudformation" {
        address = "OrdersApi"
      }
    }

    instance "user_pool" {
      of = container.user_pool

      binding "cloudformation" {
        address = "UserPool"
      }
    }
  }

  node "functions" {
    instance "create_order" {
      of         = component.create_order
      attributes = { memory = 256, timeout = 30 }

      binding "cloudformation" {
        address = "CreateOrderFunction"
      }
    }

    instance "get_order" {
      of         = component.get_order
      attributes = { memory = 256, timeout = 30 }

      binding "cloudformation" {
        address = "GetOrderFunction"
      }
    }

    instance "list_orders" {
      of         = component.list_orders
      attributes = { memory = 256, timeout = 30 }

      binding "cloudformation" {
        address = "ListOrdersFunction"
      }
    }

    instance "cancel_order" {
      of         = component.cancel_order
      attributes = { memory = 256, timeout = 30 }

      binding "cloudformation" {
        address = "CancelOrderFunction"
      }
    }

    instance "process_payment" {
      of         = component.process_payment
      attributes = { memory = 512, timeout = 60, reserved_concurrency = 10 }

      binding "cloudformation" {
        address = "ProcessPaymentFunction"
      }
    }

    instance "update_inventory" {
      of         = component.update_inventory
      attributes = { memory = 512, timeout = 60 }

      binding "cloudformation" {
        address = "UpdateInventoryFunction"
      }
    }

    instance "send_notification" {
      of         = component.send_notification
      attributes = { memory = 512, timeout = 60 }

      binding "cloudformation" {
        address = "SendNotificationFunction"
      }
    }

    instance "scheduled_tasks" {
      of = container.scheduled_tasks

      binding "cloudformation" {
        addresses = ["DailyReportFunction", "CleanupExpiredFunction"]
      }
    }
  }

  node "messaging" {
    instance "order_queue" {
      of = container.order_queue

      binding "cloudformation" {
        address = "OrderProcessingQueue"
      }
    }

    instance "order_dlq" {
      of = container.order_dlq

      binding "cloudformation" {
        address = "OrderProcessingDLQ"
      }
    }

    instance "schedule" {
      of = container.schedule

      binding "cloudformation" {
        addresses = ["DailyReportSchedule", "CleanupSchedule"]
      }
    }
  }

  node "data" {
    instance "orders_table" {
      of = container.orders_table

      binding "cloudformation" {
        address = "OrdersTable"
      }
    }

    instance "secrets" {
      of = container.secrets

      binding "cloudformation" {
        addresses = ["PaymentProviderSecret", "EmailProviderSecret"]
      }
    }
  }
}
