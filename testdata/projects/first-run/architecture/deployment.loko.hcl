deployment "prod" {
  provider = "aws"
  region   = "us-east-1"

  node "aws" {
    instance "scheduler" {
      of = container.scheduler
    }
    instance "collector" {
      of = container.collector
    }
    instance "parser" {
      of = container.parser
    }
    instance "normalizer" {
      of = container.normalizer
    }
    instance "matcher" {
      of = container.matcher
    }
    instance "pricing_api" {
      of = container.pricing_api
    }
    instance "raw_store" {
      of = container.raw_store
    }
    instance "clean_store" {
      of = container.clean_store
    }
  }
}
