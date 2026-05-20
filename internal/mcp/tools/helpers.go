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

// createSystemSchema is the JSON schema for the create_system tool input.
var createSystemSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"project_root":     map[string]any{"type": "string", "description": "Root directory of the project"},
		"name":             map[string]any{"type": "string", "description": "System name (e.g., 'Payment Service')"},
		"description":      map[string]any{"type": "string", "description": "What does this system do?"},
		"responsibilities": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Key responsibilities (e.g., 'Process payments', 'Store user data')"},
		"key_users":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Primary users/actors (e.g., 'User', 'Admin', 'Payment Gateway')"},
		"dependencies":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "External dependencies (e.g., 'Database', 'Cache', 'Message Queue')"},
		"external_systems": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "External system integrations (e.g., 'Payment API', 'Email Service')"},
		"primary_language": map[string]any{"type": "string", "description": "Primary programming language (e.g., 'Go', 'Python', 'JavaScript')"},
		"framework":        map[string]any{"type": "string", "description": "Framework/library (e.g., 'Fiber', 'Django', 'React')"},
		"database":         map[string]any{"type": "string", "description": "Database technology (e.g., 'PostgreSQL', 'MongoDB', 'Redis')"},
		"tags":             map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional tags for categorization"},
	},
	"required": []string{"project_root", "name"},
}

// updateSystemSchema is the JSON schema for the update_system tool input.
var updateSystemSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"project_root":     map[string]any{"type": "string", "description": "Root directory of the project"},
		"system_name":      map[string]any{"type": "string", "description": "System name or ID to update"},
		"description":      map[string]any{"type": "string", "description": "New description (leave empty to keep current)"},
		"responsibilities": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace responsibilities list"},
		"key_users":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace key users list"},
		"dependencies":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace dependencies list"},
		"external_systems": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace external systems list"},
		"primary_language": map[string]any{"type": "string", "description": "Primary programming language"},
		"framework":        map[string]any{"type": "string", "description": "Framework/library"},
		"database":         map[string]any{"type": "string", "description": "Database technology"},
		"tags":             map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace tags list"},
	},
	"required": []string{"project_root", "system_name"},
}

// createContainerSchema is the JSON schema for the create_container tool input.
var createContainerSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"project_root": map[string]any{"type": "string", "description": "Root directory of the project"},
		"system_name":  map[string]any{"type": "string", "description": "Parent system name"},
		"name":         map[string]any{"type": "string", "description": "Container name (e.g., 'API Server', 'Web Frontend', 'Database')"},
		"description":  map[string]any{"type": "string", "description": "What does this container do? (e.g., 'Handles all REST API requests')"},
		"technology":   map[string]any{"type": "string", "description": "Technology stack (e.g., 'Go + Fiber', 'Node.js + Express', 'PostgreSQL 15')"},
		"tags":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags for categorization (e.g., 'backend', 'database', 'frontend')"},
	},
	"required": []string{"project_root", "system_name", "name"},
}

// updateContainerSchema is the JSON schema for the update_container tool input.
var updateContainerSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"project_root":   map[string]any{"type": "string", "description": "Root directory of the project"},
		"system_name":    map[string]any{"type": "string", "description": "Parent system name"},
		"container_name": map[string]any{"type": "string", "description": "Container name or ID to update"},
		"description":    map[string]any{"type": "string", "description": "New description (leave empty to keep current)"},
		"technology":     map[string]any{"type": "string", "description": "New technology stack (leave empty to keep current)"},
		"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace tags list"},
	},
	"required": []string{"project_root", "system_name", "container_name"},
}

// createComponentSchema is the JSON schema for the create_component tool input.
var createComponentSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"project_root":   map[string]any{"type": "string", "description": "Root directory of the project"},
		"system_name":    map[string]any{"type": "string", "description": "Parent system name"},
		"container_name": map[string]any{"type": "string", "description": "Parent container name"},
		"name":           map[string]any{"type": "string", "description": "Component name (e.g., 'Auth Handler', 'Product Service', 'Cache Manager')"},
		"description":    map[string]any{"type": "string", "description": "What does this component do? (e.g., 'Handles JWT authentication')"},
		"technology":     map[string]any{"type": "string", "description": "Technology/implementation details (e.g., 'Go', 'React Component', 'Python module')"},
		"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags for categorization (e.g., 'auth', 'handler', 'service')"},
		"preview":        map[string]any{"type": "boolean", "description": "Whether to include a diagram preview in the response", "default": false},
	},
	"required": []string{"project_root", "system_name", "container_name", "name"},
}

// updateComponentSchema is the JSON schema for the update_component tool input.
var updateComponentSchema = map[string]any{
	"type":     "object",
	"required": []string{"project_root", "system_name", "container_name", "component_name"},
	"properties": map[string]any{
		"project_root":   map[string]any{"type": "string", "description": "Root directory of the project"},
		"system_name":    map[string]any{"type": "string", "description": "Parent system name"},
		"container_name": map[string]any{"type": "string", "description": "Parent container name"},
		"component_name": map[string]any{"type": "string", "description": "Component name or ID to update"},
		"description":    map[string]any{"type": "string", "description": "New description (leave empty to keep current)"},
		"technology":     map[string]any{"type": "string", "description": "New technology (leave empty to keep current)"},
		"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replace tags list"},
	},
}

// createComponentsSchema is the JSON schema for the create_components tool input.
// Kept here (a data file) to stay within the 100-line MCP handler limit.
var createComponentsSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"project_root":   map[string]any{"type": "string", "description": "Root directory of the project"},
		"system_name":    map[string]any{"type": "string", "description": "Parent system name"},
		"container_name": map[string]any{"type": "string", "description": "Parent container name"},
		"components": map[string]any{
			"type": "array", "minItems": 1,
			"description": "Array of component definitions to create",
			"items": map[string]any{
				"type":     "object",
				"required": []string{"name"},
				"properties": map[string]any{
					"name":        map[string]any{"type": "string", "description": "Component name"},
					"description": map[string]any{"type": "string", "description": "What does this component do?"},
					"technology":  map[string]any{"type": "string", "description": "Technology/implementation details"},
					"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags for categorization"},
				},
			},
		},
	},
	"required": []string{"project_root", "system_name", "container_name", "components"},
}

// relationshipToMap converts a Relationship entity to a JSON-friendly map.
func relationshipToMap(rel *entities.Relationship) map[string]any {
	m := map[string]any{
		"id":     rel.ID,
		"source": rel.Source,
		"target": rel.Target,
		"label":  rel.Label,
	}
	if rel.Type != "" {
		m["type"] = rel.Type
	}
	if rel.Technology != "" {
		m["technology"] = rel.Technology
	}
	if rel.Direction != "" {
		m["direction"] = rel.Direction
	}
	return m
}

// diagramMessageFor returns a human-readable diagram status message for tool responses.
// When diagramPath is empty it instructs the user to use the update_diagram tool.
func diagramMessageFor(diagramPath string) string {
	if diagramPath != "" {
		return "D2 template created at " + diagramPath
	}
	return "Use 'update_diagram' tool to add D2 diagram"
}

// getComponentString safely extracts a string from a component map.
func getComponentString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// scaffoldOneComponent creates a single component. Returns the result map and entity ID (empty on error).
func scaffoldOneComponent(
	ctx context.Context,
	repo usecases.ProjectRepository,
	projectRoot, systemName, containerName string,
	compMap map[string]any,
) (map[string]any, string) {
	name := getComponentString(compMap, "name")
	if name == "" {
		return map[string]any{"status": "error", "error": "name is required"}, ""
	}

	var tags []string
	if tagsIface, ok := compMap["tags"].([]any); ok {
		tags = convertInterfaceSlice(tagsIface)
	}

	scaffoldUC := usecases.NewScaffoldEntity(repo)
	scaffoldResult, err := scaffoldUC.Execute(ctx, &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot,
		EntityType:  "component",
		ParentPath:  []string{systemName, containerName},
		Name:        name,
		Description: getComponentString(compMap, "description"),
		Technology:  getComponentString(compMap, "technology"),
		Tags:        tags,
	})
	if err != nil {
		return map[string]any{
			"name": name, "status": "error",
			"error": fmt.Sprintf("failed to scaffold component: %v", err),
		}, ""
	}
	return map[string]any{"name": name, "status": "created"}, scaffoldResult.EntityID
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

// graphNodesToMapList converts a slice of graph nodes to JSON-friendly maps.
func graphNodesToMapList(nodes []*entities.GraphNode) []map[string]any {
	list := make([]map[string]any, len(nodes))
	for i, n := range nodes {
		list[i] = map[string]any{"id": n.ID, "name": n.Name, "type": n.Type, "level": n.Level}
	}
	return list
}

// appendPathToTarget adds path_to_target to the result map when a target is specified.
func appendPathToTarget(result map[string]any, graph *entities.ArchitectureGraph, compQID, compID, targetID string) {
	targetQID := targetID
	if qid, ok := graph.ResolveID(targetID); ok {
		targetQID = qid
	}
	if path := graph.GetPath(compQID, targetQID); path != nil {
		result["path_to_target"] = graphNodesToMapList(path)
	} else {
		result["path_to_target"] = nil
		result["note"] = fmt.Sprintf("No path found from %s to %s", compID, targetID)
	}
}

// queryComponentDependencies returns dependency data for a single component.
func queryComponentDependencies(args QueryDependenciesArgs, container *entities.Container, graph *entities.ArchitectureGraph) (any, error) {
	comp, exists := container.Components[args.ComponentID]
	if !exists {
		return nil, notFoundError("component", args.ComponentID, suggestSlugID(args.ComponentID, graph))
	}
	compQID, ok := graph.ResolveID(args.ComponentID)
	if !ok && graph.GetNode(args.ComponentID) != nil {
		compQID = args.ComponentID
	} else if !ok {
		return nil, notFoundError("component", args.ComponentID, "")
	}
	deps := graph.GetDependencies(compQID)
	result := map[string]any{
		"component":          map[string]any{"id": comp.ID, "name": comp.Name, "type": "component", "level": 3},
		"dependencies":       graphNodesToMapList(deps),
		"relationship_count": len(deps),
	}
	if args.TargetComponentID != "" {
		appendPathToTarget(result, graph, compQID, args.ComponentID, args.TargetComponentID)
	}
	return result, nil
}

// queryContainerDependencies returns the union of all dependencies from all
// components in the container.
func queryContainerDependencies(container *entities.Container, graph *entities.ArchitectureGraph) map[string]any {
	seen := make(map[string]bool)
	var allDeps []map[string]any
	for shortCompID := range container.Components {
		qualifiedID, ok := graph.ResolveID(shortCompID)
		if !ok {
			continue
		}
		for _, dep := range graph.GetDependencies(qualifiedID) {
			if !seen[dep.ID] {
				seen[dep.ID] = true
				allDeps = append(allDeps, map[string]any{"id": dep.ID, "name": dep.Name, "type": dep.Type, "level": dep.Level})
			}
		}
	}
	if allDeps == nil {
		allDeps = []map[string]any{}
	}
	return map[string]any{
		"container":        map[string]any{"id": container.ID, "name": container.Name, "type": "container", "level": 2},
		"dependencies":     allDeps,
		"dependency_count": len(allDeps),
		"component_count":  len(container.Components),
	}
}
