container "timer" {
  system      = system.tracking
  description = "Starts one run every hour"
  technology  = "EventBridge"
  docs        = "architecture/docs/tracking.md"

  uses "start" {
    target = container.fetcher
  }
}

container "fetcher" {
  system      = system.tracking
  description = "Fetches scan events from carrier APIs"
  technology  = "Lambda"
  docs        = "architecture/docs/tracking.md"

  uses "geocode" {
    target = system.geocoding_api
  }
  uses "fetch" {
    target = system.carrier_api
  }
  uses "store" {
    target = container.event_store
  }
  uses "enqueue" {
    target = system.message_bus
  }
}

container "event_store" {
  system      = system.tracking
  description = "Scan events, as received"
  technology  = "S3"
  docs        = "architecture/docs/tracking.md"

  uses "backed" {
    target = system.blob_storage
  }
}

container "decoder" {
  system      = system.tracking
  description = "Extracts parcel IDs and statuses from carrier payloads"
  technology  = "Lambda"

  uses "read" {
    target = container.event_store
  }
  uses "write" {
    target = container.shipment_store
  }
}

container "mapper" {
  system      = system.tracking
  description = "Maps carrier status codes to our status set"
  technology  = "Lambda"

  uses "update" {
    target = container.shipment_store
  }
}

container "correlator" {
  system      = system.tracking
  description = "Links scan events to our shipments"
  technology  = "Lambda"

  uses "read" {
    target = container.shipment_store
  }
  uses "orders" {
    target = system.order_system
  }
}

container "shipment_store" {
  system      = system.tracking
  description = "Shipments with their latest status"
  technology  = "DynamoDB"
}

container "tracking_api" {
  system      = system.tracking
  description = "Serves shipment status and delivery estimates"
  technology  = "API Gateway + Lambda"

  uses "read" {
    target = container.shipment_store
  }
  uses "publish" {
    target = system.order_system
  }
}

container "ops_console" {
  system      = system.tracking
  description = "Shows shipments and their scan history"
  technology  = "React"

  uses "api" {
    target = container.tracking_api
  }
  uses "login" {
    target = system.sso
  }
}

container "alerter" {
  system      = system.tracking
  description = "Emails a summary of late parcels after each run"
  technology  = "Lambda"

  uses "summary" {
    target = container.tracking_api
  }
  uses "send" {
    target = system.mail_relay
  }
}

container "settings_ui" {
  system      = system.tracking
  description = "Edits carriers and polling intervals"
  technology  = "React"

  uses "api" {
    target = container.tracking_api
  }
  uses "login" {
    target = system.sso
  }
}

container "change_log" {
  system      = system.tracking
  description = "Who changed which delivery promise"
  technology  = "DynamoDB"
}
