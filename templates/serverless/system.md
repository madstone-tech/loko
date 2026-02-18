---
name: "{{SystemName}}"
description: "{{Description}}"
---

# {{SystemName}}

{{Description}}

## Overview

This serverless system provides event-driven functionality using AWS Lambda functions, triggered by various event sources including API Gateway, SQS, SNS, and scheduled events.

## Event Sources

- **API Gateway**: REST/HTTP API endpoints for synchronous requests
- **SQS Queues**: Asynchronous message processing
- **SNS Topics**: Pub/sub event distribution
- **EventBridge**: Scheduled and event-driven triggers
- **S3**: Object storage event triggers

## Lambda Functions

{{container_table}}

<!-- The system is composed of function groups organized by purpose:

- **API Handlers**: Functions triggered by API Gateway requests
- **Event Processors**: Functions triggered by queue/topic messages
- **Scheduled Tasks**: Functions triggered by EventBridge schedules -->

## External Integrations

- **DynamoDB**: NoSQL data persistence
- **S3**: Object storage for files and artifacts
- **Secrets Manager**: Secure credential storage
- **CloudWatch**: Logging and monitoring

## Technology Stack

- **Language**: {{Language}}
- **Framework**: {{Framework}}
- **Database**: {{Database}}
- **Runtime**: AWS Lambda
- **Infrastructure**: AWS SAM / CDK / Terraform

## Security

- IAM roles with least-privilege permissions
- VPC configuration for private resources
- Encryption at rest and in transit
- API Gateway authorization (Cognito/JWT/API Keys)

## Observability

- CloudWatch Logs for function logs
- X-Ray for distributed tracing
- CloudWatch Metrics for performance monitoring
- CloudWatch Alarms for alerting
