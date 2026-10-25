deployment "prod" {
  provider = "aws"
  region   = "us-east-1"

  node "aws" {
    instance "scheduler" {
      of      = container.scheduler
      planned = true
    }
    instance "collector" {
      of      = container.collector
      planned = true
    }
    instance "parser" {
      of      = container.parser
      planned = true
    }
    instance "normalizer" {
      of      = container.normalizer
      planned = true
    }
    instance "matcher" {
      of      = container.matcher
      planned = true
    }
    instance "pricing_api" {
      of      = container.pricing_api
      planned = true
    }
    instance "raw_store" {
      of      = container.raw_store
      planned = true
    }
    instance "clean_store" {
      of      = container.clean_store
      planned = true
    }
  }
}
