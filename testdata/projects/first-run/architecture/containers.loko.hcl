container "scheduler" {
  system      = system.pricing
  description = "Starts one run per day"
  technology  = "EventBridge"
  docs        = "architecture/docs/pricing.md"

  uses "start" {
    target = container.collector
  }
}

container "collector" {
  system      = system.pricing
  description = "Searches for and fetches competitor listings"
  technology  = "Lambda"
  docs        = "architecture/docs/pricing.md"

  uses "search" {
    target = system.search_api
  }
  uses "fetch" {
    target = system.vendor_api
  }
  uses "store" {
    target = container.raw_store
  }
  uses "enqueue" {
    target = system.queue_service
  }
}

container "raw_store" {
  system      = system.pricing
  description = "Fetched pages, as received"
  technology  = "S3"
  docs        = "architecture/docs/pricing.md"

  uses "backed" {
    target = system.object_storage
  }
}

container "parser" {
  system      = system.pricing
  description = "Extracts products and prices from fetched pages"
  technology  = "Lambda"

  uses "read" {
    target = container.raw_store
  }
  uses "write" {
    target = container.clean_store
  }
}

container "normalizer" {
  system      = system.pricing
  description = "Converts units, currencies and pack sizes"
  technology  = "Lambda"

  uses "update" {
    target = container.clean_store
  }
}

container "matcher" {
  system      = system.pricing
  description = "Matches competitor products to our catalogue"
  technology  = "Lambda"

  uses "read" {
    target = container.clean_store
  }
  uses "catalogue" {
    target = system.erp
  }
}

container "clean_store" {
  system      = system.pricing
  description = "Normalised products and prices"
  technology  = "DynamoDB"
}

container "pricing_api" {
  system      = system.pricing
  description = "Serves recommendations"
  technology  = "API Gateway + Lambda"

  uses "read" {
    target = container.clean_store
  }
  uses "publish" {
    target = system.erp
  }
}

container "dashboard" {
  system      = system.pricing
  description = "Shows recommendations and their evidence"
  technology  = "React"

  uses "api" {
    target = container.pricing_api
  }
  uses "login" {
    target = system.identity_provider
  }
}

container "notifier" {
  system      = system.pricing
  description = "Emails a summary after each run"
  technology  = "Lambda"

  uses "summary" {
    target = container.pricing_api
  }
  uses "send" {
    target = system.email_service
  }
}

container "admin_ui" {
  system      = system.pricing
  description = "Edits sources and schedules"
  technology  = "React"

  uses "api" {
    target = container.pricing_api
  }
  uses "login" {
    target = system.identity_provider
  }
}

container "audit_log" {
  system      = system.pricing
  description = "Who approved which price change"
  technology  = "DynamoDB"
}
