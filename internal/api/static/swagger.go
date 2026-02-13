package static

import (
	"embed"
)

// SwaggerUI contains the embedded Swagger UI static assets
//
// These assets are served at /api/docs to provide interactive API documentation.
// The Swagger UI is configured to load the OpenAPI spec from /api/v1/openapi.json.
//
// Assets are embedded at build time using go:embed, ensuring offline operation
// without CDN dependencies.
//
//go:embed swagger-ui/*
var SwaggerUI embed.FS
