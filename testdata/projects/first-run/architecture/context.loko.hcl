person "planner" {
  description = "Reviews late shipments"
  docs        = "architecture/docs/people.md"

  uses "reviews" {
    target = container.ops_console
  }
}

person "dispatcher" {
  description = "Approves delivery-promise changes"
  docs        = "architecture/docs/people.md"

  uses "approves" {
    target = container.ops_console
  }
}

person "operator" {
  description = "Configures carriers and polling intervals"
  docs        = "architecture/docs/people.md"

  uses "configures" {
    target = container.settings_ui
  }
}

system "tracking" {
  description = "The tracking pipeline the team builds and runs: it tracks parcels from carrier scans and publishes delivery estimates"
  docs        = "architecture/docs/tracking.md"
}

system "carrier_api" {
  description = "Carrier tracking API, run by a carrier"
}

system "geocoding_api" {
  description = "Address geocoding API, a SaaS product"
}

system "mail_relay" {
  description = "Transactional email, a SaaS product"
}

system "sso" {
  description = "Single sign-on, run by the customer's IT team"
}

system "order_system" {
  description = "The customer's system of record for orders"
}

system "blob_storage" {
  description = "Cloud-managed object storage"
}

system "message_bus" {
  description = "Cloud-managed message queue"
}
