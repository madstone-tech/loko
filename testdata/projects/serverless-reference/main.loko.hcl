project "serverless-reference" {
  description = "Anonymized serverless AWS system (feature 016 fixture)"
}

person "client_user" {
  description = "External client reading project status"

  uses "browses" {
    target      = container.cdn
    description = "Uses the dashboard"
    technology  = "HTTPS"
  }
  title = "Client user"
}

person "staff" {
  description = "Staff curating projects and clients"

  uses "administers" {
    target      = container.cdn
    description = "Curates projects and clients"
    technology  = "HTTPS"
  }
  title = "Staff"
}

system "portal" {
  description = "Read-only dashboard over an upstream system"
  owner       = "platform"
  title       = "Project portal"
}

external "upstream" {
  description = "Upstream source system (REST)"
  title       = "Upstream system"
}

external "client_inbox" {
  description = "The client's email inbox"
  title       = "Client email inbox"
}

container "cdn" {
  system      = system.portal
  description = "Static hosting for the SPA bundle (origin access control)"
  tags        = ["aws"]
  technology  = "CloudFront + S3"

  uses "bundle" {
    target      = container.spa
    description = "Serves the bundle"
  }
  title = "CloudFront + S3"
}

container "spa" {
  system      = system.portal
  description = "Client dashboard and staff admin in one bundle; holds no secrets"
  technology  = "React (Vite)"

  uses "signin" {
    target      = container.cognito
    description = "Requests and exchanges a one-time code"
  }

  uses "calls" {
    target      = container.apigw
    description = "JSON with Bearer ID token"
    technology  = "HTTPS"
  }
  title = "Web app (SPA)"
}

container "cognito" {
  system      = system.portal
  description = "Passwordless email sign-in; ID token carries tenant and groups"
  tags        = ["aws"]
  technology  = "Cognito User Pool"

  uses "otp" {
    target      = external.client_inbox
    description = "Emails the one-time code"
    technology  = "SES"
  }

  uses "claims" {
    target      = container.pretoken
    description = "Adds tenant and group claims at token issue"
    technology  = "Lambda trigger"
  }
  title = "Cognito user pool"
}

container "pretoken" {
  system      = system.portal
  description = "Pre-token-generation trigger: adds tenant and group claims"
  tags        = ["lambda"]
  technology  = "AWS Lambda (Go)"
  shape       = "function"
  title       = "Pre-token Lambda"
}

container "apigw" {
  system      = system.portal
  description = "JWT authorizer on every route except GET /health"
  tags        = ["aws"]
  technology  = "API Gateway HTTP API"

  uses "authorize" {
    target      = container.cognito
    description = "Validates the ID token"
    technology  = "JWT"
  }

  uses "reads" {
    target      = container.read_api
    description = "Read routes"
  }

  uses "admin" {
    target      = container.admin_api
    description = "ANY /admin/*"
  }
  title = "API Gateway"
}

container "read_api" {
  system      = system.portal
  description = "Read routes; DynamoDB read-only IAM"
  tags        = ["lambda"]
  technology  = "AWS Lambda (Go, arm64)"

  uses "query" {
    target      = container.table
    description = "Query (read only)"
    tags        = ["read-only"]
  }

  uses "logo" {
    target      = container.uploads
    description = "Presigned GET for org logos"
  }
  shape = "function"
  title = "Read API Lambda"
}

container "admin_api" {
  system      = system.portal
  description = "Staff-gated admin routes; holds write IAM"
  tags        = ["lambda"]
  technology  = "AWS Lambda (Go, arm64)"

  uses "write" {
    target      = container.table
    description = "Get / Put / Update / Delete"
    tags        = ["write"]
  }

  uses "users" {
    target      = container.cognito
    description = "Creates, suspends and deletes users"
  }

  uses "logos" {
    target      = container.uploads
    description = "Uploads org logos"
  }

  uses "resync" {
    target      = container.sync_queue
    description = "Enqueues an on-demand sync"
    technology  = "SQS"
    kind        = "async"
  }
  shape = "function"
  title = "Admin API Lambda"
}

container "schedule" {
  system      = system.portal
  description = "rate(30 minutes)"
  tags        = ["aws"]
  technology  = "EventBridge rule"
  title       = "Sync schedule"
}

container "scheduler" {
  system      = system.portal
  description = "Enqueues one SQS message per project"
  tags        = ["lambda"]
  technology  = "AWS Lambda (Go)"

  uses "trigger" {
    target      = container.schedule
    description = "Invoked on schedule"
    technology  = "EventBridge"
    kind        = "trigger"
  }

  uses "list" {
    target      = container.table
    description = "Scans for project META items"
  }

  uses "enqueue" {
    target      = container.sync_queue
    description = "One message per project"
    technology  = "SQS"
    kind        = "async"
  }
  shape = "function"
  title = "Scheduler Lambda"
}

container "sync_queue" {
  system      = system.portal
  description = "One message per project; redrive after 3 receives"
  tags        = ["sqs"]
  technology  = "Amazon SQS"

  uses "dead_letter" {
    target      = container.sync_dlq
    description = "After 3 failed receives"
    technology  = "Redrive policy"
    kind        = "async"
  }
  shape = "queue"
  title = "Sync queue"
}

container "sync_dlq" {
  system      = system.portal
  description = "Dead letters; CloudWatch alarm on depth"
  tags        = ["sqs"]
  technology  = "Amazon SQS"

  uses "alarm" {
    target      = container.alerts
    description = "Alarm on queue depth"
    technology  = "CloudWatch"
    kind        = "async"
  }
  shape = "queue"
  title = "Sync dead-letter queue"
}

container "sync" {
  system      = system.portal
  description = "One project per invocation: pull upstream, score, write"
  tags        = ["lambda"]
  technology  = "AWS Lambda (Go)"

  uses "consume" {
    target      = container.sync_queue
    description = "Event source mapping, batch size 1"
    technology  = "SQS"
    kind        = "trigger"
  }

  uses "pull" {
    target      = external.upstream
    description = "Pulls source records"
    technology  = "HTTPS"
  }

  uses "snapshot" {
    target      = container.table
    description = "Writes the snapshot, reconciles removals"
    tags        = ["write"]
  }

  uses "fail" {
    target      = container.alerts
    description = "Publishes on pull failure"
    technology  = "SNS"
    kind        = "async"
  }
  shape = "function"
  title = "Sync Lambda"
}

container "alerts" {
  system      = system.portal
  description = "Sync failure notifications"
  tags        = ["sns"]
  technology  = "Amazon SNS"
  shape       = "topic"
  title       = "Alert topic"
}

container "table" {
  system      = system.portal
  description = "Single table for all application data"
  tags        = ["aws"]
  technology  = "DynamoDB"
  shape       = "database"
  title       = "Application table"
}

container "uploads" {
  system      = system.portal
  description = "Client org logos; presigned GET"
  tags        = ["aws"]
  technology  = "Amazon S3"
  shape       = "bucket"
  title       = "Uploads bucket"
}

component "api_main" {
  container   = container.read_api
  description = "cmd/api: Lambda entry point, route table"
  technology  = "Go package"

  uses "platform" {
    target      = component.platform
    description = "Builds platform requests"
  }

  uses "projects" {
    target      = component.projects
    description = "Registers List, Dashboard"
  }

  uses "me" {
    target      = component.me
    description = "Registers Handler"
  }

  uses "orgs" {
    target      = component.orgs
    description = "Registers GetLogo"
  }

  uses "health" {
    target      = component.health
    description = "Registers Check"
  }
}

component "platform" {
  container   = container.read_api
  description = "internal/platform: request, response, JSON"
  technology  = "Go package"
}

component "auth" {
  container   = container.read_api
  description = "internal/auth: parses claims; tenant ownership check"
  technology  = "Go package"
}

component "projects" {
  container   = container.read_api
  description = "internal/projects: list and dashboard"
  technology  = "Go package"

  uses "platform" {
    target      = component.platform
    description = "Request / Response / JSON"
  }

  uses "claims" {
    target      = component.auth
    description = "Checks the tenant owns the project"
  }

  uses "query" {
    target      = container.table
    description = "Queries one project"
  }
}

component "me" {
  container   = container.read_api
  description = "internal/me: the signed-in user"
  technology  = "Go package"

  uses "platform" {
    target      = component.platform
    description = "Request / Response / JSON"
  }

  uses "claims" {
    target      = component.auth
    description = "Parses claims"
  }
}

component "orgs" {
  container   = container.read_api
  description = "internal/orgs: org logo"
  technology  = "Go package"

  uses "platform" {
    target      = component.platform
    description = "Request / Response / JSON"
  }

  uses "presign" {
    target      = container.uploads
    description = "Presigned GET"
  }
}

component "health" {
  container   = container.read_api
  description = "internal/health: liveness check"
  technology  = "Go package"

  uses "platform" {
    target      = component.platform
    description = "Request / Response / JSON"
  }
}

deployment "dev" {
  provider = "aws"
  region   = "us-east-1"

  instance "cdn" {
    of = container.cdn

    binding "terraform" {
      address = "module.web.aws_cloudfront_distribution.site"
    }
  }

  instance "cognito" {
    of = container.cognito

    binding "terraform" {
      address = "module.identity.aws_cognito_user_pool.this"
    }
  }

  instance "pretoken" {
    of = container.pretoken

    binding "terraform" {
      address = "module.pretoken.aws_lambda_function.pretoken"
    }
  }

  instance "apigw" {
    of = container.apigw

    binding "terraform" {
      address = "module.api.aws_apigatewayv2_api.http"
    }
  }

  instance "read_api" {
    of = container.read_api

    binding "terraform" {
      address = "module.api.aws_lambda_function.api"
    }
  }

  instance "admin_api" {
    of = container.admin_api

    binding "terraform" {
      address = "module.api.aws_lambda_function.admin"
    }
  }

  instance "schedule" {
    of = container.schedule

    binding "terraform" {
      address = "module.sync.aws_cloudwatch_event_rule.schedule"
    }
  }

  instance "scheduler" {
    of = container.scheduler

    binding "terraform" {
      address = "module.sync.aws_lambda_function.scheduler"
    }
  }

  instance "sync_queue" {
    of = container.sync_queue

    binding "terraform" {
      address = "module.sync.aws_sqs_queue.main"
    }
  }

  instance "sync_dlq" {
    of = container.sync_dlq

    binding "terraform" {
      address = "module.sync.aws_sqs_queue.dlq"
    }
  }

  instance "sync" {
    of = container.sync

    binding "terraform" {
      address = "module.sync.aws_lambda_function.sync"
    }
  }

  instance "alerts" {
    of = container.alerts

    binding "terraform" {
      address = "module.sync.aws_sns_topic.alerts"
    }
  }

  instance "table" {
    of = container.table

    binding "terraform" {
      address = "module.data.aws_dynamodb_table.this"
    }
  }

  instance "uploads" {
    of = container.uploads

    binding "terraform" {
      address = "module.uploads.aws_s3_bucket.uploads"
    }
  }
}

deployment "prod" {
  provider = "aws"
  region   = "us-east-1"

  instance "cdn" {
    of = container.cdn

    binding "terraform" {
      address = "module.web.aws_cloudfront_distribution.site"
    }
  }

  instance "cognito" {
    of = container.cognito

    binding "terraform" {
      address = "module.identity.aws_cognito_user_pool.this"
    }
  }

  instance "pretoken" {
    of = container.pretoken

    binding "terraform" {
      address = "module.pretoken.aws_lambda_function.pretoken"
    }
  }

  instance "apigw" {
    of = container.apigw

    binding "terraform" {
      address = "module.api.aws_apigatewayv2_api.http"
    }
  }

  instance "read_api" {
    of = container.read_api

    binding "terraform" {
      address = "module.api.aws_lambda_function.api"
    }
  }

  instance "admin_api" {
    of = container.admin_api

    binding "terraform" {
      address = "module.api.aws_lambda_function.admin"
    }
  }

  instance "schedule" {
    of = container.schedule

    binding "terraform" {
      address = "module.sync.aws_cloudwatch_event_rule.schedule"
    }
  }

  instance "scheduler" {
    of = container.scheduler

    binding "terraform" {
      address = "module.sync.aws_lambda_function.scheduler"
    }
  }

  instance "sync_queue" {
    of = container.sync_queue

    binding "terraform" {
      address = "module.sync.aws_sqs_queue.main"
    }
  }

  instance "sync_dlq" {
    of = container.sync_dlq

    binding "terraform" {
      address = "module.sync.aws_sqs_queue.dlq"
    }
  }

  instance "sync" {
    of = container.sync

    binding "terraform" {
      address = "module.sync.aws_lambda_function.sync"
    }
  }

  instance "alerts" {
    of = container.alerts

    binding "terraform" {
      address = "module.sync.aws_sns_topic.alerts"
    }
  }

  instance "table" {
    of = container.table

    binding "terraform" {
      address = "module.data.aws_dynamodb_table.this"
    }
  }

  instance "uploads" {
    of = container.uploads

    binding "terraform" {
      address = "module.uploads.aws_s3_bucket.uploads"
    }
  }
}
