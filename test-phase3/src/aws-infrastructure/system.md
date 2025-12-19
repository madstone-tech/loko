---
name: "AWS Infrastructure"
description: "AWS STS Organization managing cloud resources across non-prod and shared-services accounts"
type: "external-system"
tags:
  - "infrastructure"
  - "cloud"
---

# AWS Infrastructure

Manages the complete cloud infrastructure for the MRO Scheduler system, including compute, storage, networking, and deployment pipelines across multiple AWS accounts.

## Responsibilities

- Provide compute resources (ECS Fargate)
- Manage container registries and image repositories
- Handle database and storage operations
- Authenticate and authorize users
- Deploy and manage services
- Optimize resource allocation with ML

## Architecture

- **Non-Prod Account**: Hosts UI and API services, databases, and storage
- **Shared-Services Account**: Manages container image build pipeline
- **Cross-Account Image Replication**: Syncs images from shared to non-prod

## Key Services

- ECS Fargate for containerized workloads
- RDS for SQL Server database
- S3 for object storage
- Cognito for authentication
- SageMaker for ML-based optimization
- CodePipeline for CI/CD
