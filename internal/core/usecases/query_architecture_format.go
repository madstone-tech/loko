// Package usecases — output formatting layer for QueryArchitecture.
// See query_architecture.go for the use case type, request/response types, and Execute.
package usecases

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

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

	fmt.Fprintf(&sb, "@%s", name)
	if desc != "" && len(desc) <= 60 {
		fmt.Fprintf(&sb, ":%s", desc)
	}
	sb.WriteString("\n")

	switch detail {
	case "summary":
		// Compact stats line
		systems, _ := dataMap["systems"].(int)
		containers, _ := dataMap["containers"].(int)
		components, _ := dataMap["components"].(int)
		fmt.Fprintf(&sb, "S%d/C%d/K%d\n", systems, containers, components)

		// System names
		if names, ok := dataMap["system_names"].([]string); ok && len(names) > 0 {
			sb.WriteString(strings.Join(names, ","))
		}

	case "structure":
		if systems, ok := dataMap["systems"].([]map[string]any); ok {
			for _, sys := range systems {
				sysName, _ := sys["name"].(string)
				sysDesc, _ := sys["description"].(string)

				fmt.Fprintf(&sb, "S:%s", sysName)
				if sysDesc != "" && len(sysDesc) <= 40 {
					fmt.Fprintf(&sb, ":%s", sysDesc)
				}
				sb.WriteString("\n")

				if containers, ok := sys["containers"].([]map[string]any); ok {
					for _, cont := range containers {
						contName, _ := cont["name"].(string)
						tech, _ := cont["technology"].(string)
						fmt.Fprintf(&sb, "  C:%s", contName)
						if tech != "" {
							fmt.Fprintf(&sb, "[%s]", tech)
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
				fmt.Fprintf(&sb, "S:%s\n", sysName)

				if containers, ok := sys["containers"].([]map[string]any); ok {
					for _, cont := range containers {
						contName, _ := cont["name"].(string)
						tech, _ := cont["technology"].(string)
						fmt.Fprintf(&sb, "  C:%s", contName)
						if tech != "" {
							fmt.Fprintf(&sb, "[%s]", tech)
						}
						sb.WriteString("\n")

						if components, ok := cont["components"].([]map[string]any); ok {
							for _, comp := range components {
								compName, _ := comp["name"].(string)
								fmt.Fprintf(&sb, "    K:%s\n", compName)
							}
						}
					}
				}
			}
		}
	}

	return sb.String()
}
