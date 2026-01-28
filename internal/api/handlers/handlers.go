// Package handlers provides HTTP handlers for the loko API.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Handlers contains all API handlers.
type Handlers struct {
	projectRoot string
	repo        usecases.ProjectRepository

	// Build tracking
	builds     map[string]*buildStatus
	buildMutex sync.RWMutex
	buildID    int
}

// buildStatus tracks an in-progress or completed build.
type buildStatus struct {
	ID               string
	Status           string // "building", "complete", "failed"
	StartTime        time.Time
	EndTime          time.Time
	FilesGenerated   int
	DiagramsRendered int
	OutputDir        string
	Error            string
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(projectRoot string, repo usecases.ProjectRepository) *Handlers {
	return &Handlers{
		projectRoot: projectRoot,
		repo:        repo,
		builds:      make(map[string]*buildStatus),
	}
}

// GetProject handles GET /api/v1/project.
func (h *Handlers) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project, err := h.repo.LoadProject(ctx, h.projectRoot)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "project not found")
		return
	}

	systems, err := h.repo.ListSystems(ctx, h.projectRoot)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list systems")
		return
	}

	totalContainers := 0
	totalComponents := 0
	for _, sys := range systems {
		totalContainers += sys.ContainerCount()
		totalComponents += sys.ComponentCount()
	}

	resp := ProjectResponse{
		Success:        true,
		Name:           project.Name,
		Description:    project.Description,
		Version:        project.Version,
		SystemCount:    len(systems),
		ContainerCount: totalContainers,
		ComponentCount: totalComponents,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// ListSystems handles GET /api/v1/systems.
func (h *Handlers) ListSystems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project, _ := h.repo.LoadProject(ctx, h.projectRoot)

	systems, err := h.repo.ListSystems(ctx, h.projectRoot)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list systems")
		return
	}

	summaries := make([]SystemSummary, 0, len(systems))
	for _, sys := range systems {
		summaries = append(summaries, SystemSummary{
			ID:             sys.ID,
			Name:           sys.Name,
			Description:    sys.Description,
			ContainerCount: sys.ContainerCount(),
			ComponentCount: sys.ComponentCount(),
			Tags:           sys.Tags,
		})
	}

	projectName := ""
	if project != nil {
		projectName = project.Name
	}

	resp := SystemsResponse{
		Success:     true,
		Systems:     summaries,
		TotalCount:  len(summaries),
		ProjectName: projectName,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// GetSystem handles GET /api/v1/systems/{id}.
func (h *Handlers) GetSystem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	systemID := r.PathValue("id")

	if systemID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "system id required")
		return
	}

	system, err := h.repo.LoadSystem(ctx, h.projectRoot, systemID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "system not found")
		return
	}

	containers := make([]ContainerSummary, 0)
	for _, cont := range system.ListContainers() {
		containers = append(containers, ContainerSummary{
			ID:             cont.ID,
			Name:           cont.Name,
			Description:    cont.Description,
			Technology:     cont.Technology,
			ComponentCount: cont.ComponentCount(),
			Tags:           cont.Tags,
		})
	}

	resp := SystemDetailResponse{
		Success: true,
		System: &SystemSummary{
			ID:             system.ID,
			Name:           system.Name,
			Description:    system.Description,
			ContainerCount: system.ContainerCount(),
			ComponentCount: system.ComponentCount(),
			Tags:           system.Tags,
		},
		Containers: containers,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// TriggerBuild handles POST /api/v1/build.
func (h *Handlers) TriggerBuild(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req BuildRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// Set defaults
	if req.OutputDir == "" {
		req.OutputDir = "dist"
	}
	if req.Format == "" {
		req.Format = "html"
	}

	// Create build ID
	h.buildMutex.Lock()
	h.buildID++
	buildID := formatBuildID(h.buildID)
	status := &buildStatus{
		ID:        buildID,
		Status:    "building",
		StartTime: time.Now(),
		OutputDir: req.OutputDir,
	}
	h.builds[buildID] = status
	h.buildMutex.Unlock()

	// Start build in background
	go h.executeBuild(ctx, buildID, req)

	// Return immediately with build ID
	resp := BuildResponse{
		Success: true,
		BuildID: buildID,
		Status:  "building",
		Message: "Build started",
	}

	WriteJSON(w, http.StatusAccepted, resp)
}

// executeBuild runs the build process.
func (h *Handlers) executeBuild(ctx context.Context, buildID string, req BuildRequest) {
	h.buildMutex.Lock()
	status := h.builds[buildID]
	h.buildMutex.Unlock()

	// Load project and systems
	project, err := h.repo.LoadProject(ctx, h.projectRoot)
	if err != nil {
		h.failBuild(buildID, "failed to load project: "+err.Error())
		return
	}

	systems, err := h.repo.ListSystems(ctx, h.projectRoot)
	if err != nil {
		h.failBuild(buildID, "failed to list systems: "+err.Error())
		return
	}

	// Create adapters
	renderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		h.failBuild(buildID, "failed to create site builder: "+err.Error())
		return
	}

	// Create progress reporter that updates build status
	progressReporter := &buildProgressReporter{
		handler: h,
		buildID: buildID,
	}

	// Execute build
	buildDocs := usecases.NewBuildDocs(renderer, siteBuilder, progressReporter)
	err = buildDocs.Execute(ctx, project, systems, req.OutputDir)

	h.buildMutex.Lock()
	defer h.buildMutex.Unlock()

	if err != nil {
		status.Status = "failed"
		status.Error = err.Error()
	} else {
		status.Status = "complete"
	}
	status.EndTime = time.Now()
}

// failBuild marks a build as failed.
func (h *Handlers) failBuild(buildID, errMsg string) {
	h.buildMutex.Lock()
	defer h.buildMutex.Unlock()

	if status, ok := h.builds[buildID]; ok {
		status.Status = "failed"
		status.Error = errMsg
		status.EndTime = time.Now()
	}
}

// GetBuildStatus handles GET /api/v1/build/{id}.
func (h *Handlers) GetBuildStatus(w http.ResponseWriter, r *http.Request) {
	buildID := r.PathValue("id")

	if buildID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "build id required")
		return
	}

	h.buildMutex.RLock()
	status, ok := h.builds[buildID]
	h.buildMutex.RUnlock()

	if !ok {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	var durationMS int64
	if !status.EndTime.IsZero() {
		durationMS = status.EndTime.Sub(status.StartTime).Milliseconds()
	} else {
		durationMS = time.Since(status.StartTime).Milliseconds()
	}

	resp := BuildResponse{
		Success:          status.Status != "failed",
		BuildID:          status.ID,
		Status:           status.Status,
		DurationMS:       durationMS,
		OutputDir:        status.OutputDir,
		FilesGenerated:   status.FilesGenerated,
		DiagramsRendered: status.DiagramsRendered,
		Error:            status.Error,
	}

	if status.Status == "complete" {
		resp.Message = "Build completed successfully"
	} else if status.Status == "failed" {
		resp.Message = "Build failed"
	} else {
		resp.Message = "Build in progress"
	}

	WriteJSON(w, http.StatusOK, resp)
}

// Validate handles GET /api/v1/validate.
func (h *Handlers) Validate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	systems, err := h.repo.ListSystems(ctx, h.projectRoot)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list systems")
		return
	}

	issues := make([]ValidationIssue, 0)
	errorCount := 0
	warningCount := 0

	// Check for systems without containers
	for _, sys := range systems {
		if sys.ContainerCount() == 0 {
			issues = append(issues, ValidationIssue{
				Code:     "EMPTY_SYSTEM",
				Severity: "warning",
				Message:  "System has no containers",
				Location: "systems/" + sys.ID,
			})
			warningCount++
		}

		// Check for containers without descriptions
		for _, cont := range sys.ListContainers() {
			if cont.Description == "" {
				issues = append(issues, ValidationIssue{
					Code:     "MISSING_DESCRIPTION",
					Severity: "warning",
					Message:  "Container missing description",
					Location: "systems/" + sys.ID + "/containers/" + cont.ID,
				})
				warningCount++
			}
		}
	}

	valid := errorCount == 0
	message := "Validation passed"
	if errorCount > 0 {
		message = "Validation failed with errors"
	} else if warningCount > 0 {
		message = "Validation passed with warnings"
	}

	resp := ValidateResponse{
		Success:      true,
		Valid:        valid,
		ErrorCount:   errorCount,
		WarningCount: warningCount,
		Issues:       issues,
		Message:      message,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// buildProgressReporter implements usecases.ProgressReporter for build tracking.
type buildProgressReporter struct {
	handler *Handlers
	buildID string
}

func (r *buildProgressReporter) ReportProgress(step string, current, total int, message string) {
	// Update build status with progress
}

func (r *buildProgressReporter) ReportError(err error) {
	r.handler.failBuild(r.buildID, err.Error())
}

func (r *buildProgressReporter) ReportSuccess(message string) {
	// Build success is handled by executeBuild
}

func (r *buildProgressReporter) ReportInfo(message string) {
	// Log info if needed
}

// formatBuildID generates a build ID.
func formatBuildID(n int) string {
	return time.Now().Format("20060102") + "-" + padInt(n, 4)
}

// padInt pads an integer with leading zeros.
func padInt(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// Response types

// BuildRequest is the request body for POST /api/v1/build.
type BuildRequest struct {
	Format      string `json:"format,omitempty"`
	Incremental bool   `json:"incremental,omitempty"`
	OutputDir   string `json:"output_dir,omitempty"`
}

// BuildResponse is the response for POST /api/v1/build.
type BuildResponse struct {
	Success          bool   `json:"success"`
	BuildID          string `json:"build_id,omitempty"`
	Status           string `json:"status"`
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
	Severity string `json:"severity"`
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
	Success        bool   `json:"success"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Version        string `json:"version,omitempty"`
	SystemCount    int    `json:"system_count"`
	ContainerCount int    `json:"container_count"`
	ComponentCount int    `json:"component_count"`
}

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// HealthResponse is the response for GET /health.
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	D2Version string `json:"d2_version,omitempty"`
}
