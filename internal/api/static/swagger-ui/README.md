# Swagger UI Assets

This directory contains the minimal Swagger UI build for serving OpenAPI documentation at `/api/docs`.

## Setup Instructions

Download the minimal Swagger UI dist build from:
- https://github.com/swagger-api/swagger-ui/releases/latest

Extract only the following files:
- swagger-ui.css
- swagger-ui-bundle.js
- swagger-ui-standalone-preset.js
- favicon-16x16.png
- favicon-32x32.png

Target size: < 400KB gzipped

## Embedding

These assets are embedded in the Go binary using `go:embed` directives in `internal/api/static/swagger.go`.

## Usage

The Swagger UI is served at `/api/docs` and configured to load the OpenAPI spec from `/api/v1/openapi.json`.

## Offline Operation

All assets are embedded in the binary with no CDN dependencies, ensuring the Swagger UI works offline.
