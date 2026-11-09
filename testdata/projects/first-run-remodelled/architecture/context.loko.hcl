person "planner" {
  description = "Reviews late shipments"
  docs        = "docs/people.md"

  uses "reviews" {
    target = container.ops_console
  }
}

person "dispatcher" {
  description = "Approves delivery-promise changes"
  docs        = "docs/people.md"

  uses "approves" {
    target = container.ops_console
  }
}

person "operator" {
  description = "Configures carriers and polling intervals"
  docs        = "docs/people.md"

  uses "configures" {
    target = container.settings_ui
  }
}

system "tracking" {
  description = "The tracking pipeline the team builds and runs: it tracks parcels from carrier scans and publishes delivery estimates"
  docs        = "docs/tracking.md"
}

external "carrier_api" {
  description = "Carrier tracking API, run by a carrier"
}

external "geocoding_api" {
  description = "Address geocoding API, a SaaS product"
}

external "mail_relay" {
  description = "Transactional email, a SaaS product"
}

external "sso" {
  description = "Single sign-on, run by the customer's IT team"
}

external "order_system" {
  description = "The customer's system of record for orders"
}

external "blob_storage" {
  description = "Cloud-managed object storage"
}

external "message_bus" {
  description = "Cloud-managed message queue"
}
