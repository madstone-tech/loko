deployment "prod" {
  provider = "aws"
  region   = "us-east-1"

  node "aws" {
    instance "timer" {
      of = container.timer
    }
    instance "fetcher" {
      of = container.fetcher
    }
    instance "decoder" {
      of = container.decoder
    }
    instance "mapper" {
      of = container.mapper
    }
    instance "correlator" {
      of = container.correlator
    }
    instance "tracking_api" {
      of = container.tracking_api
    }
    instance "event_store" {
      of = container.event_store
    }
    instance "shipment_store" {
      of = container.shipment_store
    }
  }
}
