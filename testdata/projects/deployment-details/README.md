# deployment-details

The cloud layout from [the first-run report](../../../specs/017-mcp-validation-ergonomics/first-run-report.md)
(item #6): an AWS account, a region and a VPC holding a Lambda function, a DynamoDB table and an S3
bucket, plus a `dev` deployment running the same function locally.

Feature 018 uses it for the deployment node attributes (`title`, `technology`, `description`,
`tags`), the instance `technology` override and a titled, described view.
