// Package api provides HTTP API server for loko.
package api

import "time"

// BuildRequest is the request body for POST /api/v1/build.
type BuildRequest struct {
	Format      string `json:"format,omitempty"`      // "html", "markdown", "pdf", or "all"
	Incremental bool   `json:"incremental,omitempty"` // Only build changed files
	OutputDir   string `json:"output_dir,omitempty"`  // Custom output directory
}

// BuildResponse is the response for POST /api/v1/build.
type BuildResponse struct {
	Success          bool   `json:"success"`
	BuildID          string `json:"build_id,omitempty"`
	Status           string `json:"status"` // "building", "complete", "failed"
	DurationMS       int64  `json:"duration_ms,omitempty"`
	OutputDir        string `json:"output_dir,omitempty"`
	FilesGenerated   int    `json:"files_generated,omitempty"`
	DiagramsRendered int    `json:"diagrams_rendered,omitempty"`
	Message          string `json:"message,omitempty"`
	Error            string `json:"error,omitempty"`
}

// SystemSummary is a summary of a system for API responses.
type SystemSummary struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	ContainerCount int      `json:"container_count"`
	ComponentCount int      `json:"component_count"`
	Tags           []string `json:"tags,omitempty"`
}

// SystemsResponse is the response for GET /api/v1/systems.
type SystemsResponse struct {
	Success     bool            `json:"success"`
	Systems     []SystemSummary `json:"systems"`
	TotalCount  int             `json:"total_count"`
	ProjectName string          `json:"project_name,omitempty"`
}

// ContainerSummary is a summary of a container for API responses.
type ContainerSummary struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Technology     string   `json:"technology,omitempty"`
	ComponentCount int      `json:"component_count"`
	Tags           []string `json:"tags,omitempty"`
}

// SystemDetailResponse is the response for GET /api/v1/systems/:id.
type SystemDetailResponse struct {
	Success    bool               `json:"success"`
	System     *SystemSummary     `json:"system"`
	Containers []ContainerSummary `json:"containers"`
}

// ValidationIssue represents a validation error or warning.
type ValidationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // "error" or "warning"
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
}

// ValidateResponse is the response for GET /api/v1/validate.
type ValidateResponse struct {
	Success      bool              `json:"success"`
	Valid        bool              `json:"valid"`
	ErrorCount   int               `json:"error_count"`
	WarningCount int               `json:"warning_count"`
	Issues       []ValidationIssue `json:"issues,omitempty"`
	Message      string            `json:"message,omitempty"`
}

// ProjectResponse is the response for GET /api/v1/project.
type ProjectResponse struct {
	Success        bool      `json:"success"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Version        string    `json:"version,omitempty"`
	SystemCount    int       `json:"system_count"`
	ContainerCount int       `json:"container_count"`
	ComponentCount int       `json:"component_count"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// HealthResponse is the response for GET /health.
type HealthResponse struct {
	Status    string `json:"status"` // "ok", "degraded"
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	D2Version string `json:"d2_version,omitempty"`
}
