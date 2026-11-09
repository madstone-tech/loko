deployment "prod" {
  provider = "aws"
  region   = "us-east-1"

  node "aws" {
    instance "timer" {
      of      = container.timer
      planned = true
    }
    instance "fetcher" {
      of      = container.fetcher
      planned = true
    }
    instance "decoder" {
      of      = container.decoder
      planned = true
    }
    instance "mapper" {
      of      = container.mapper
      planned = true
    }
    instance "correlator" {
      of      = container.correlator
      planned = true
    }
    instance "tracking_api" {
      of      = container.tracking_api
      planned = true
    }
    instance "event_store" {
      of      = container.event_store
      planned = true
    }
    instance "shipment_store" {
      of      = container.shipment_store
      planned = true
    }
  }
}
