# API Reference

loko provides a REST API for CI/CD integration and automation.

## Quick Start

```bash
# Start the API server
loko api --port 8081

# Start with authentication
loko api --port 8081 --api-key "your-secret-key"
```

## Authentication

When started with `--api-key`, all endpoints (except `/health`) require a Bearer token:

```bash
curl -H "Authorization: Bearer your-secret-key" http://localhost:8081/api/v1/project
```

Without `--api-key`, authentication is disabled (suitable for local development).

## Base URL

```
http://localhost:8081
```

## Endpoints

### Health Check

Check server status. No authentication required.

```
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime": "1h23m45s"
}
```

---

### Get Project

Get project overview and statistics.

```
GET /api/v1/project
```

**Response:**
```json
{
  "success": true,
  "name": "My Architecture",
  "description": "Architecture documentation",
  "version": "1.0.0",
  "system_count": 5,
  "container_count": 12,
  "component_count": 34
}
```

---

### List Systems

Get all systems in the project.

```
GET /api/v1/systems
```

**Response:**
```json
{
  "success": true,
  "systems": [
    {
      "id": "auth-service",
      "name": "Auth Service",
      "description": "Handles authentication",
      "container_count": 3,
      "component_count": 8,
      "tags": ["security", "identity"]
    }
  ],
  "total_count": 5,
  "project_name": "My Architecture"
}
```

---

### Get System Details

Get detailed information about a specific system.

```
GET /api/v1/systems/{id}
```

**Parameters:**
- `id` - System ID (normalized name, e.g., "auth-service")

**Response:**
```json
{
  "success": true,
  "system": {
    "id": "auth-service",
    "name": "Auth Service",
    "description": "Handles authentication",
    "container_count": 3,
    "component_count": 8,
    "tags": ["security"]
  },
  "containers": [
    {
      "id": "api",
      "name": "API",
      "description": "REST API",
      "technology": "Go",
      "component_count": 4,
      "tags": ["http"]
    }
  ]
}
```

---

### Trigger Build

Start a documentation build. Returns immediately with a build ID.

```
POST /api/v1/build
```

**Request Body (optional):**
```json
{
  "format": "html",
  "incremental": false,
  "output_dir": "dist"
}
```

**Parameters:**
- `format` - Output format: `html`, `markdown`, or `pdf` (default: `html`)
- `incremental` - Skip unchanged files (default: `false`)
- `output_dir` - Output directory (default: `dist`)

**Response:**
```json
{
  "success": true,
  "build_id": "20240115-0001",
  "status": "building",
  "message": "Build started"
}
```

---

### Get Build Status

Check the status of a build.

```
GET /api/v1/build/{id}
```

**Parameters:**
- `id` - Build ID from trigger build response

**Response (in progress):**
```json
{
  "success": true,
  "build_id": "20240115-0001",
  "status": "building",
  "duration_ms": 1500,
  "message": "Build in progress"
}
```

**Response (complete):**
```json
{
  "success": true,
  "build_id": "20240115-0001",
  "status": "complete",
  "duration_ms": 3200,
  "output_dir": "dist",
  "files_generated": 15,
  "diagrams_rendered": 8,
  "message": "Build completed successfully"
}
```

**Response (failed):**
```json
{
  "success": false,
  "build_id": "20240115-0001",
  "status": "failed",
  "duration_ms": 500,
  "message": "Build failed",
  "error": "failed to render diagram: d2 not found"
}
```

---

### Validate Architecture

Check architecture for issues.

```
GET /api/v1/validate
```

**Response:**
```json
{
  "success": true,
  "valid": true,
  "error_count": 0,
  "warning_count": 2,
  "issues": [
    {
      "code": "EMPTY_SYSTEM",
      "severity": "warning",
      "message": "System has no containers",
      "location": "systems/legacy-service"
    },
    {
      "code": "MISSING_DESCRIPTION",
      "severity": "warning",
      "message": "Container missing description",
      "location": "systems/auth-service/containers/api"
    }
  ],
  "message": "Validation passed with warnings"
}
```

## Error Responses

All endpoints return errors in a consistent format:

```json
{
  "error": "error message",
  "code": "ERROR_CODE"
}
```

**Common Error Codes:**
- `UNAUTHORIZED` - Missing or invalid API key
- `NOT_FOUND` - Resource not found
- `INVALID_INPUT` - Invalid request parameters
- `INTERNAL_ERROR` - Server error

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build Docs

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Start loko API
        run: |
          loko api --port 8081 &
          sleep 2

      - name: Trigger build
        run: |
          curl -X POST http://localhost:8081/api/v1/build \
            -H "Content-Type: application/json" \
            -d '{"format": "html"}'

      - name: Wait for build
        run: |
          # Poll build status
          for i in {1..30}; do
            STATUS=$(curl -s http://localhost:8081/api/v1/build/latest | jq -r '.status')
            if [ "$STATUS" = "complete" ]; then
              echo "Build complete!"
              exit 0
            elif [ "$STATUS" = "failed" ]; then
              echo "Build failed!"
              exit 1
            fi
            sleep 2
          done
          echo "Build timeout"
          exit 1

      - name: Upload docs
        uses: actions/upload-artifact@v4
        with:
          name: docs
          path: dist/
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any

    stages {
        stage('Build Docs') {
            steps {
                sh 'loko api --port 8081 &'
                sh 'sleep 2'

                script {
                    def response = httpRequest(
                        httpMode: 'POST',
                        url: 'http://localhost:8081/api/v1/build',
                        contentType: 'APPLICATION_JSON',
                        requestBody: '{"format": "html"}'
                    )
                    def buildId = readJSON(text: response.content).build_id

                    // Poll for completion
                    timeout(time: 5, unit: 'MINUTES') {
                        waitUntil {
                            def status = httpRequest(
                                url: "http://localhost:8081/api/v1/build/${buildId}"
                            )
                            def statusJson = readJSON(text: status.content)
                            return statusJson.status == 'complete' || statusJson.status == 'failed'
                        }
                    }
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'dist/**/*', fingerprint: true
        }
    }
}
```

## OpenAPI Specification

The full OpenAPI 3.0 specification is available at:

- File: `internal/api/openapi.yaml`
- Swagger UI: Coming soon

## Rate Limiting

Currently, no rate limiting is implemented. For production deployments, consider using a reverse proxy (nginx, Traefik) to add rate limiting.

## CORS

The API allows cross-origin requests from any origin (`Access-Control-Allow-Origin: *`). For production, configure this based on your needs.
