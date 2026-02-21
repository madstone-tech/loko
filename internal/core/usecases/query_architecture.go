package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
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

// applyFormat transforms the response based on requested format.
func applyFormat(resp *QueryArchitectureResponse, project *entities.Project, systems []*entities.System, detail, format string) *QueryArchitectureResponse {
	// Build structured data for the response
	rawData := buildRawData(project, systems, detail)
	resp.RawData = rawData

	switch format {
	case "json":
		resp.Text = formatAsJSON(rawData)
	case "toon":
		resp.Text = formatAsTOON(rawData, detail)
	}

	resp.TokenEstimate = estimateTokens(resp.Text)
	return resp
}

// buildRawData creates structured data for JSON/TOON encoding.
func buildRawData(project *entities.Project, systems []*entities.System, detail string) map[string]any {
	data := map[string]any{
		"name":        project.Name,
		"description": project.Description,
	}

	if project.Version != "" {
		data["version"] = project.Version
	}

	totalContainers := 0
	totalComponents := 0

	switch detail {
	case "summary":
		systemNames := make([]string, 0, len(systems))
		for _, sys := range systems {
			systemNames = append(systemNames, sys.Name)
			totalContainers += sys.ContainerCount()
			totalComponents += sys.ComponentCount()
		}
		data["systems"] = len(systems)
		data["containers"] = totalContainers
		data["components"] = totalComponents
		data["system_names"] = systemNames

	case "structure":
		systemList := make([]map[string]any, 0, len(systems))
		for _, sys := range systems {
			containers := make([]map[string]any, 0)
			for _, cont := range sys.ListContainers() {
				containers = append(containers, map[string]any{
					"id":         cont.ID,
					"name":       cont.Name,
					"technology": cont.Technology,
				})
				totalContainers++
			}
			totalComponents += sys.ComponentCount()

			systemList = append(systemList, map[string]any{
				"id":          sys.ID,
				"name":        sys.Name,
				"description": sys.Description,
				"containers":  containers,
			})
		}
		data["systems"] = systemList
		data["total_containers"] = totalContainers
		data["total_components"] = totalComponents

	case "full":
		systemList := make([]map[string]any, 0, len(systems))
		for _, sys := range systems {
			containers := make([]map[string]any, 0)
			for _, cont := range sys.ListContainers() {
				components := make([]map[string]any, 0)
				for _, comp := range cont.ListComponents() {
					components = append(components, map[string]any{
						"id":          comp.ID,
						"name":        comp.Name,
						"description": comp.Description,
						"technology":  comp.Technology,
					})
				}

				containers = append(containers, map[string]any{
					"id":          cont.ID,
					"name":        cont.Name,
					"description": cont.Description,
					"technology":  cont.Technology,
					"components":  components,
				})
				totalContainers++
				totalComponents += len(components)
			}

			sysData := map[string]any{
				"id":          sys.ID,
				"name":        sys.Name,
				"description": sys.Description,
				"containers":  containers,
			}
			if sys.External {
				sysData["external"] = true
			}
			if len(sys.Tags) > 0 {
				sysData["tags"] = sys.Tags
			}
			systemList = append(systemList, sysData)
		}
		data["systems"] = systemList
		data["total_containers"] = totalContainers
		data["total_components"] = totalComponents
	}

	return data
}

// formatAsJSON formats data as indented JSON.
func formatAsJSON(data any) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// formatAsTOON formats data using Token-Optimized Object Notation.
func formatAsTOON(data any, detail string) string {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return "{}"
	}

	var sb strings.Builder

	// Header with project name
	name, _ := dataMap["name"].(string)
	desc, _ := dataMap["description"].(string)

	sb.WriteString(fmt.Sprintf("@%s", name))
	if desc != "" && len(desc) <= 60 {
		sb.WriteString(fmt.Sprintf(":%s", desc))
	}
	sb.WriteString("\n")

	switch detail {
	case "summary":
		// Compact stats line
		systems, _ := dataMap["systems"].(int)
		containers, _ := dataMap["containers"].(int)
		components, _ := dataMap["components"].(int)
		sb.WriteString(fmt.Sprintf("S%d/C%d/K%d\n", systems, containers, components))

		// System names
		if names, ok := dataMap["system_names"].([]string); ok && len(names) > 0 {
			sb.WriteString(strings.Join(names, ","))
		}

	case "structure":
		if systems, ok := dataMap["systems"].([]map[string]any); ok {
			for _, sys := range systems {
				sysName, _ := sys["name"].(string)
				sysDesc, _ := sys["description"].(string)

				sb.WriteString(fmt.Sprintf("S:%s", sysName))
				if sysDesc != "" && len(sysDesc) <= 40 {
					sb.WriteString(fmt.Sprintf(":%s", sysDesc))
				}
				sb.WriteString("\n")

				if containers, ok := sys["containers"].([]map[string]any); ok {
					for _, cont := range containers {
						contName, _ := cont["name"].(string)
						tech, _ := cont["technology"].(string)
						sb.WriteString(fmt.Sprintf("  C:%s", contName))
						if tech != "" {
							sb.WriteString(fmt.Sprintf("[%s]", tech))
						}
						sb.WriteString("\n")
					}
				}
			}
		}

	case "full":
		if systems, ok := dataMap["systems"].([]map[string]any); ok {
			for _, sys := range systems {
				sysName, _ := sys["name"].(string)
				sb.WriteString(fmt.Sprintf("S:%s\n", sysName))

				if containers, ok := sys["containers"].([]map[string]any); ok {
					for _, cont := range containers {
						contName, _ := cont["name"].(string)
						tech, _ := cont["technology"].(string)
						sb.WriteString(fmt.Sprintf("  C:%s", contName))
						if tech != "" {
							sb.WriteString(fmt.Sprintf("[%s]", tech))
						}
						sb.WriteString("\n")

						if components, ok := cont["components"].([]map[string]any); ok {
							for _, comp := range components {
								compName, _ := comp["name"].(string)
								sb.WriteString(fmt.Sprintf("    K:%s\n", compName))
							}
						}
					}
				}
			}
		}
	}

	return sb.String()
}

// buildSummaryResponse creates a summary-level response (~200 tokens).
func buildSummaryResponse(project *entities.Project, systems []*entities.System) *QueryArchitectureResponse {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Project: %s\n", project.Name))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", project.Description))
	}

	sb.WriteString(fmt.Sprintf("Systems: %d\n", len(systems)))

	totalContainers := 0
	totalComponents := 0
	systemSummaries := make([]*SystemSummary, 0, len(systems))
	for _, sys := range systems {
		cc := sys.ContainerCount()
		kc := sys.ComponentCount()
		totalContainers += cc
		totalComponents += kc
		systemSummaries = append(systemSummaries, &SystemSummary{
			Name:        sys.Name,
			Description: sys.Description,
			Containers:  cc,
			Components:  kc,
		})
	}

	sb.WriteString(fmt.Sprintf("Total Containers: %d\n", totalContainers))
	sb.WriteString(fmt.Sprintf("Total Components: %d\n", totalComponents))

	return &QueryArchitectureResponse{
		Text:           sb.String(),
		TokenEstimate:  estimateTokens(sb.String()),
		Detail:         "summary",
		Systems:        systemSummaries,
		ContainerCount: totalContainers,
		ComponentCount: totalComponents,
	}
}

// buildStructureResponse creates a structure-level response (~500 tokens).
func buildStructureResponse(project *entities.Project, systems []*entities.System) *QueryArchitectureResponse {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Project: %s\n", project.Name))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n\n", project.Description))
	}

	systemSummaries := make([]*SystemSummary, 0, len(systems))

	for _, sys := range systems {
		sb.WriteString(fmt.Sprintf("## %s\n", sys.Name))
		if sys.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n", sys.Description))
		}

		containers := sys.ListContainers()
		sb.WriteString(fmt.Sprintf("Containers: %d\n", len(containers)))

		for _, cont := range containers {
			sb.WriteString(fmt.Sprintf("  - %s", cont.Name))
			if cont.Description != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", cont.Description))
			}
			if cont.Technology != "" {
				sb.WriteString(fmt.Sprintf(" [%s]", cont.Technology))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")

		systemSummaries = append(systemSummaries, &SystemSummary{
			Name:        sys.Name,
			Description: sys.Description,
			Containers:  len(containers),
			Components:  sys.ComponentCount(),
		})
	}

	return &QueryArchitectureResponse{
		Text:          sb.String(),
		TokenEstimate: estimateTokens(sb.String()),
		Detail:        "structure",
		Systems:       systemSummaries,
	}
}

// buildFullResponse creates a full-detail response (all tokens).
func buildFullResponse(project *entities.Project, systems []*entities.System) *QueryArchitectureResponse {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Project: %s\n", project.Name))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription: %s\n", project.Description))
	}

	if project.Version != "" {
		sb.WriteString(fmt.Sprintf("Version: %s\n", project.Version))
	}

	systemSummaries := make([]*SystemSummary, 0, len(systems))

	for _, sys := range systems {
		sb.WriteString(fmt.Sprintf("\n## System: %s\n", sys.Name))
		if sys.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", sys.Description))
		}

		if sys.External {
			sb.WriteString("Type: External\n")
		}

		if len(sys.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(sys.Tags, ", ")))
		}

		containers := sys.ListContainers()
		sb.WriteString(fmt.Sprintf("\nContainers (%d):\n", len(containers)))

		for _, cont := range containers {
			sb.WriteString(fmt.Sprintf("\n### %s\n", cont.Name))
			if cont.Description != "" {
				sb.WriteString(fmt.Sprintf("Description: %s\n", cont.Description))
			}
			if cont.Technology != "" {
				sb.WriteString(fmt.Sprintf("Technology: %s\n", cont.Technology))
			}

			components := cont.ListComponents()
			if len(components) > 0 {
				sb.WriteString(fmt.Sprintf("Components (%d):\n", len(components)))
				for _, comp := range components {
					sb.WriteString(fmt.Sprintf("  - %s", comp.Name))
					if comp.Description != "" {
						sb.WriteString(fmt.Sprintf(" (%s)", comp.Description))
					}
					sb.WriteString("\n")
				}
			}
		}

		systemSummaries = append(systemSummaries, &SystemSummary{
			Name:        sys.Name,
			Description: sys.Description,
			Containers:  len(containers),
			Components:  sys.ComponentCount(),
		})
	}

	return &QueryArchitectureResponse{
		Text:          sb.String(),
		TokenEstimate: estimateTokens(sb.String()),
		Detail:        "full",
		Systems:       systemSummaries,
	}
}

// Helper functions

// isValidDetailLevel checks if the detail level is valid.
func isValidDetailLevel(detail string) bool {
	return detail == "summary" || detail == "structure" || detail == "full"
}

// estimateTokens provides a rough token count estimate.
// Approximation: ~4 characters per token (average), adjusted for code/structured text
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
