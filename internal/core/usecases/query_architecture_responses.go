// Package usecases — response-building variants per detail level for QueryArchitecture.
// See query_architecture.go for the use case type, request/response types, and Execute.
package usecases

import (
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// buildSummaryResponse creates a summary-level response (~200 tokens).
func buildSummaryResponse(project *entities.Project, systems []*entities.System) *QueryArchitectureResponse {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Project: %s\n", project.Name)
	if project.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", project.Description)
	}

	fmt.Fprintf(&sb, "Systems: %d\n", len(systems))

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

	fmt.Fprintf(&sb, "Total Containers: %d\n", totalContainers)
	fmt.Fprintf(&sb, "Total Components: %d\n", totalComponents)

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

	fmt.Fprintf(&sb, "Project: %s\n", project.Name)
	if project.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n\n", project.Description)
	}

	systemSummaries := make([]*SystemSummary, 0, len(systems))

	for _, sys := range systems {
		fmt.Fprintf(&sb, "## %s\n", sys.Name)
		if sys.Description != "" {
			fmt.Fprintf(&sb, "%s\n", sys.Description)
		}

		containers := sys.ListContainers()
		fmt.Fprintf(&sb, "Containers: %d\n", len(containers))

		for _, cont := range containers {
			fmt.Fprintf(&sb, "  - %s", cont.Name)
			if cont.Description != "" {
				fmt.Fprintf(&sb, " (%s)", cont.Description)
			}
			if cont.Technology != "" {
				fmt.Fprintf(&sb, " [%s]", cont.Technology)
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

	fmt.Fprintf(&sb, "# Project: %s\n", project.Name)
	if project.Description != "" {
		fmt.Fprintf(&sb, "\nDescription: %s\n", project.Description)
	}

	if project.Version != "" {
		fmt.Fprintf(&sb, "Version: %s\n", project.Version)
	}

	systemSummaries := make([]*SystemSummary, 0, len(systems))

	for _, sys := range systems {
		fmt.Fprintf(&sb, "\n## System: %s\n", sys.Name)
		if sys.Description != "" {
			fmt.Fprintf(&sb, "Description: %s\n", sys.Description)
		}

		if sys.External {
			sb.WriteString("Type: External\n")
		}

		if len(sys.Tags) > 0 {
			fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(sys.Tags, ", "))
		}

		containers := sys.ListContainers()
		fmt.Fprintf(&sb, "\nContainers (%d):\n", len(containers))

		for _, cont := range containers {
			fmt.Fprintf(&sb, "\n### %s\n", cont.Name)
			if cont.Description != "" {
				fmt.Fprintf(&sb, "Description: %s\n", cont.Description)
			}
			if cont.Technology != "" {
				fmt.Fprintf(&sb, "Technology: %s\n", cont.Technology)
			}

			components := cont.ListComponents()
			if len(components) > 0 {
				fmt.Fprintf(&sb, "Components (%d):\n", len(components))
				for _, comp := range components {
					fmt.Fprintf(&sb, "  - %s", comp.Name)
					if comp.Description != "" {
						fmt.Fprintf(&sb, " (%s)", comp.Description)
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
