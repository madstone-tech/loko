package tools

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/madstone-tech/loko/internal/core/entities"
)

// convertInterfaceSlice converts a slice of interface{} to a slice of strings.
func convertInterfaceSlice(slice []any) []string {
	if len(slice) == 0 {
		return nil
	}
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

// countDiagrams counts total diagrams in all systems and containers.
func countDiagrams(systems []*entities.System) int {
	count := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			count++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				count++
			}
		}
	}
	return count
}

// validateDiagramStructure checks for structural and C4 compliance issues.
func validateDiagramStructure(d2Source, level string) ([]string, []string) {
	var warnings []string
	var suggestions []string

	// Check for comments
	if !containsSubstring(d2Source, "#") {
		suggestions = append(suggestions, "Add comments to explain diagram structure")
	}

	// Level-specific checks
	switch level {
	case "system":
		if !containsSubstring(d2Source, "User") && !containsSubstring(d2Source, "user") {
			suggestions = append(suggestions, "C4 Level 1 typically includes 'User' - consider adding user/actor")
		}
		if countDiagramNodes(d2Source) < 2 {
			warnings = append(warnings, "System context diagram should have at least 2 nodes (User and System)")
		}

	case "container":
		if countDiagramNodes(d2Source) < 2 {
			warnings = append(warnings, "Container diagram should have at least 2 components")
		}
		if !containsSubstring(d2Source, "{\n") {
			suggestions = append(suggestions, "Consider grouping related components with container blocks { }")
		}

	case "component":
		if countDiagramNodes(d2Source) < 1 {
			warnings = append(warnings, "Component diagram should have at least 1 component")
		}
	}

	// General best practices
	if !containsSubstring(d2Source, "direction:") && !containsSubstring(d2Source, "direction ") {
		suggestions = append(suggestions, "Consider specifying diagram direction (e.g., 'direction: right') for clarity")
	}

	if !containsSubstring(d2Source, "tooltip") && !containsSubstring(d2Source, "description") {
		suggestions = append(suggestions, "Add tooltips or descriptions to nodes for better documentation")
	}

	return warnings, suggestions
}

// containsSubstring checks if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// countDiagramNodes counts the number of nodes in a D2 diagram source.
func countDiagramNodes(d2Source string) int {
	count := 0
	// Count lines with nodes (heuristic: lines with colons that aren't comments or directives)
	for line := range strings.SplitSeq(d2Source, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "direction") {
			count++
		}
	}
	return count
}

// mapToStruct converts a map[string]any to a typed struct using mapstructure.
// This replaces runtime type assertions with compile-time type safety.
// The output parameter must be a pointer to a struct.
//
// Example:
//
//	var args QueryDependenciesArgs
//	if err := mapToStruct(inputMap, &args); err != nil {
//	    return nil, err
//	}
func mapToStruct(input map[string]any, output any) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		WeaklyTypedInput: true, // Allow string-to-number conversions, etc.
		Result:           output,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(input); err != nil {
		return fmt.Errorf("failed to decode map to struct: %w", err)
	}

	return nil
}
