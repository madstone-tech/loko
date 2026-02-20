package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
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

// validateElementPath checks that an element path (e.g. "agwe/api-lambda") uses
// only valid slug characters (lowercase alphanumeric + hyphens per segment).
// If the path contains uppercase or spaces, it returns a descriptive error with the corrected slug.
// Returns nil if the path is already valid.
func validateElementPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	normalized := entities.NormalizeName(path)
	// NormalizeName lowercases and replaces spaces/special chars with hyphens.
	// If normalization changes the path, the original was invalid.
	if normalized == path {
		return normalized, nil
	}

	return normalized, fmt.Errorf(
		"element path %q is not a valid slug — did you mean %q? (use lowercase alphanumeric + hyphens, slash separates path segments)",
		path, normalized,
	)
}

// suggestSlugID generates a "did you mean X?" suggestion for element lookups.
// Given a raw input name, it applies NormalizeName and checks if the result exists in the graph.
// Returns the suggestion string or empty string if no match.
func suggestSlugID(input string, graph *entities.ArchitectureGraph) string {
	if graph == nil {
		return ""
	}

	// Normalize the input name to get the potential slug ID
	normalized := entities.NormalizeName(input)

	// Check if the normalized ID exists in the graph
	if graph.GetNode(normalized) != nil {
		return normalized
	}

	// Try to resolve using ShortIDMap
	if qualifiedID, ok := graph.ResolveID(normalized); ok {
		return qualifiedID
	}

	return ""
}

// notFoundError formats a standard error message with a "did you mean X?" suggestion.
// entityType: the type of element (e.g., "component", "container", "system")
// input: the user-provided name that wasn't found
// suggestion: the suggested slug ID (can be empty)
func notFoundError(entityType, input, suggestion string) error {
	baseMsg := fmt.Sprintf("%s %q not found", entityType, input)

	if suggestion != "" {
		return fmt.Errorf("%s — did you mean %q?", baseMsg, suggestion)
	}

	return fmt.Errorf("%s — try running 'query_architecture' to see available elements", baseMsg)
}

// getGraphFromProject builds and returns an ArchitectureGraph from a project.
// Returns nil if building the graph fails.
func getGraphFromProject(ctx context.Context, repo usecases.ProjectRepository, projectRoot string) (*entities.ArchitectureGraph, error) {
	return getGraphFromProjectWithRel(ctx, repo, nil, projectRoot)
}

// getGraphFromProjectWithRel builds and returns an ArchitectureGraph, optionally
// loading TOML relationships when relRepo is non-nil.
func getGraphFromProjectWithRel(ctx context.Context, repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository, projectRoot string) (*entities.ArchitectureGraph, error) {
	// Load project and systems
	project, err := repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Build architecture graph (includes relationships.toml when relRepo is non-nil).
	graphBuilder := usecases.NewBuildArchitectureGraphWithRelRepo(relRepo)
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	return graph, nil
}
