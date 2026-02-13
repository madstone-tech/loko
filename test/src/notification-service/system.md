---
name: "Notification Service"
description: "Serverless notification system for sending email and SMS alerts with full delivery tracking, retry logic, and notification history on AWS"
---

# Notification Service

Serverless notification system for sending email and SMS alerts with full delivery tracking, retry logic, and notification history on AWS

## Context

This is a **C4 Level 1 - System Context Diagram** showing this system in the broader architecture.

## Key Users

The following actors interact with this system:

- **Internal Services**
- **Operations Team**
- **End Users**

## System Responsibilities

This system is responsible for:

- Accept notification requests via REST API and events
- Route notifications to email and SMS channels
- Send emails via Amazon SES
- Send SMS via Amazon SNS
- Track delivery status with DynamoDB
- Retry failed notifications with exponential backoff
- Provide notification status and history queries

## Internal Dependencies

This system depends on the following internal systems or services:

- Amazon SES
- Amazon SNS
- Amazon SQS
- Amazon DynamoDB
- Amazon EventBridge
- Amazon API Gateway

## External System Integrations

This system integrates with:

- Amazon SES
- Amazon SNS

## Containers

The system is composed of the following containers:

| Container | Description | Technology |
|-----------|-------------|------------|
| (Add your containers here) | | |

## Technology Stack

- **Primary Language**: Go
- **Framework/Library**: AWS Lambda + SAM
- **Database**: (To be determined)

