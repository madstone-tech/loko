// Package devserver implements the PreviewServer port: a loopback-only HTTP
// server that serves the last build from memory and pushes reloads over
// Server-Sent Events. It is used only by `loko serve` and is not the HTTP API
// removed in v1.0 (research R13).
package devserver
