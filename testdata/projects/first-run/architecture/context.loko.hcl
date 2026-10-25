person "analyst" {
  description = "Reviews price recommendations"
  docs        = "architecture/docs/people.md"

  uses "reviews" {
    target = container.dashboard
  }
}

person "buyer" {
  description = "Approves price changes"
  docs        = "architecture/docs/people.md"

  uses "approves" {
    target = container.dashboard
  }
}

person "admin" {
  description = "Configures sources and schedules"
  docs        = "architecture/docs/people.md"

  uses "configures" {
    target = container.admin_ui
  }
}

system "pricing" {
  description = "The pricing pipeline the team builds and runs"
  docs        = "architecture/docs/pricing.md"
}

system "vendor_api" {
  description = "Competitor catalogue API, run by a vendor"
}

system "search_api" {
  description = "Web search API, a SaaS product"
}

system "email_service" {
  description = "Transactional email, a SaaS product"
}

system "identity_provider" {
  description = "Single sign-on, run by the customer's IT team"
}

system "erp" {
  description = "The customer's system of record for prices"
}

system "object_storage" {
  description = "Cloud-managed object storage"
}

system "queue_service" {
  description = "Cloud-managed message queue"
}
