# Managed AWS resources

| Resource | Purpose |
|----------|---------|
| API Gateway (REST) | Public order API; Cognito authorizer on every route |
| Cognito User Pool | Customer and admin sign-in |
| SQS order processing queue | One message per new order; redrive to the DLQ after 3 receives |
| SQS dead-letter queue | Failed messages, alarmed on depth |
| EventBridge rules | Daily report and hourly cleanup schedules |
| DynamoDB orders table | Orders and status; encrypted at rest; streams enabled |
| Secrets Manager | Payment and email provider API keys |

All resources are defined in one AWS SAM template. Logical IDs match the bindings in
`deployment.loko.hcl`.
