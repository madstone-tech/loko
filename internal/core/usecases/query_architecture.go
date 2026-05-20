package usecases

import (
	"context"
	"fmt"
)

// QueryArchitectureRequest holds parameters for querying architecture.
type QueryArchitectureRequest struct {
	Detail       string // "summary", "structure", or "full"
	Format       string // "json", "toon", or "text" (default: "text")
	TargetSystem string // optional, for targeted queries
}

// SystemSummary is a summary of a system for responses.
type SystemSummary struct {
	Name        string
	Description string
	Containers  int
	Components  int
}

// QueryArchitectureResponse is the response from querying architecture.
type QueryArchitectureResponse struct {
	Text           string
	TokenEstimate  int
	Detail         string
	Format         string // "json", "toon", or "text"
	Systems        []*SystemSummary
	ContainerCount int
	ComponentCount int
	// RawData contains structured data for JSON/TOON encoding
	RawData any `json:"-"`
}

// QueryArchitecture is the use case for token-efficient architecture queries.
type QueryArchitecture struct {
	repo ProjectRepository
}

// NewQueryArchitecture creates a new QueryArchitecture use case.
func NewQueryArchitecture(repo ProjectRepository) *QueryArchitecture {
	return &QueryArchitecture{repo: repo}
}

// Execute performs an architecture query with the specified detail level.
//
// Detail levels:
// - "summary": ~200 tokens - project overview with system counts
// - "structure": ~500 tokens - systems and their containers
// - "full": complete details - all systems, containers, components
//
// Format options:
// - "text": human-readable markdown (default)
// - "json": structured JSON
// - "toon": Token-Optimized Object Notation (30-40% fewer tokens than JSON)
//
// Returns error if detail level is invalid or project not found.
func (uc *QueryArchitecture) Execute(ctx context.Context, projectID, detail string) (*QueryArchitectureResponse, error) {
	return uc.ExecuteWithFormat(ctx, projectID, detail, "text")
}

// ExecuteWithFormat performs an architecture query with specified detail level and format.
func (uc *QueryArchitecture) ExecuteWithFormat(ctx context.Context, projectID, detail, format string) (*QueryArchitectureResponse, error) {
	// Validate detail level
	if !isValidDetailLevel(detail) {
		return nil, fmt.Errorf("invalid detail level: %s (expected summary, structure, or full)", detail)
	}

	// Validate format
	if format == "" {
		format = "text"
	}
	if !isValidFormat(format) {
		return nil, fmt.Errorf("invalid format: %s (expected text, json, or toon)", format)
	}

	// Load project
	project, err := uc.repo.LoadProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// List systems
	systems, err := uc.repo.ListSystems(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}

	// Build response based on detail level
	var resp *QueryArchitectureResponse

	switch detail {
	case "summary":
		resp = buildSummaryResponse(project, systems)
	case "structure":
		resp = buildStructureResponse(project, systems)
	case "full":
		resp = buildFullResponse(project, systems)
	}

	// Apply format transformation
	resp.Format = format
	if format != "text" {
		resp = applyFormat(resp, project, systems, detail, format)
	}

	return resp, nil
}

// isValidFormat checks if the format is valid.
func isValidFormat(format string) bool {
	return format == "text" || format == "json" || format == "toon"
}

// isValidDetailLevel checks if the detail level is valid.
func isValidDetailLevel(detail string) bool {
	return detail == "summary" || detail == "structure" || detail == "full"
}

// estimateTokens provides a rough token count estimate.
// Approximation: ~4 characters per token (average), adjusted for code/structured text.
func estimateTokens(text string) int {
	// Use a combination of character count and word count
	// Claude models typically use ~4 chars/token on average
	charTokens := len(text) / 4

	// For structured text, add a base multiplier
	words := 0
	inWord := false
	for _, ch := range text {
		isSpace := ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
		if !isSpace && !inWord {
			words++
			inWord = true
		} else if isSpace {
			inWord = false
		}
	}

	// Use the higher of the two estimates (char-based tends to be more accurate for structured text)
	wordTokens := int(float64(words) * 1.3)

	if charTokens > wordTokens {
		return charTokens
	}
	return wordTokens
}
