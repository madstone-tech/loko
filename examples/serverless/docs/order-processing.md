# Order Processing API

Serverless order processing system handling order creation, payment processing and fulfillment.

## Overview

The system processes e-commerce orders through Lambda functions triggered by API Gateway requests,
SQS messages and EventBridge schedules. The architecture gives high availability, automatic
scaling and pay-per-use pricing.

## Event sources

- **API Gateway**: REST API for order operations (`POST /orders`, `GET /orders/{id}`)
- **SQS**: order processing queue for asynchronous payment and fulfillment
- **EventBridge**: scheduled tasks for order cleanup and reporting
- **DynamoDB Streams**: real-time order status change notifications

## Lambda functions

Functions are grouped by purpose:

- **API handlers**: Create Order, Get Order, List Orders, Cancel Order
- **Event processors**: Process Payment, Update Inventory, Send Notification
- **Scheduled tasks**: Generate Daily Report, Clean Up Expired Orders

## External integrations

- **Payment provider**: payment processing
- **Email provider**: email notifications
- **Inventory service**: stock management
- **Shipping provider**: fulfillment tracking

## Technology stack

- **Language**: Go
- **Runtime**: AWS Lambda (`provided.al2`, arm64)
- **Database**: DynamoDB
- **Infrastructure**: AWS SAM

## Security

- IAM roles with least-privilege permissions per function
- API Gateway with Cognito User Pool authorization
- Encryption at rest (DynamoDB) and in transit (TLS)

## Observability

- CloudWatch Logs with structured JSON logging
- X-Ray distributed tracing across all functions
- CloudWatch Metrics with custom business KPIs
- CloudWatch Alarms for error rates and latency
