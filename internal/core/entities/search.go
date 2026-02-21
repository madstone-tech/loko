package entities

// SearchElementsRequest represents a request to search architecture elements.
// Used by the search_elements MCP tool to filter and query elements.
type SearchElementsRequest struct {
	// ProjectRoot is the root directory of the project to search.
	ProjectRoot string

	// Query is the search pattern (supports glob wildcards: *, ?).
	// Examples: "payment*", "api-*", "*-service"
	Query string

	// Type filters by element type (system, container, component).
	// Empty string means no type filter.
	Type string

	// Technology filters by technology stack (e.g., "Go", "Python", "TypeScript").
	// Empty string means no technology filter.
	Technology string

	// Tag filters by element tag (e.g., "critical", "production", "experimental").
	// Empty string means no tag filter.
	Tag string

	// Limit sets the maximum number of results to return.
	// Default: 20, Maximum: 100
	Limit int
}

// SearchElementsResponse represents the response from searching architecture elements.
type SearchElementsResponse struct {
	// Results contains the matching elements.
	Results []SearchElement

	// TotalMatched is the total number of elements that matched the query
	// (before limit was applied).
	TotalMatched int

	// Message provides context about the results (e.g., "No elements found matching query").
	Message string
}

// SearchElement represents a single element in search results.
type SearchElement struct {
	// ID is the unique qualified identifier (e.g., "payment-service/api/auth-handler").
	ID string

	// Name is the element's display name.
	Name string

	// Type is the element type (system, container, component).
	Type string

	// Description provides a brief explanation of the element.
	Description string

	// Technology is the technology stack (e.g., "Go", "Python", "TypeScript").
	// Only populated for containers and components.
	Technology string

	// Tags are labels assigned to the element (e.g., ["critical", "production"]).
	Tags []string

	// ParentID is the qualified ID of the parent element (if any).
	// Components have a container parent, containers have a system parent.
	ParentID string
}

// Validate checks if the search request is valid.
func (r *SearchElementsRequest) Validate() error {
	if r.ProjectRoot == "" {
		return NewValidationError("SearchElementsRequest", "project_root", "", "project root is required", nil)
	}

	if r.Query == "" {
		return NewValidationError("SearchElementsRequest", "query", "", "query pattern is required", nil)
	}

	// Validate type filter
	if r.Type != "" {
		validTypes := map[string]bool{
			"system":    true,
			"container": true,
			"component": true,
		}
		if !validTypes[r.Type] {
			return NewValidationError("SearchElementsRequest", "type", r.Type, "invalid type filter (must be: system, container, component)", nil)
		}
	}

	// Validate and apply default/max limits
	if r.Limit <= 0 {
		r.Limit = 20 // Default limit
	}
	if r.Limit > 100 {
		r.Limit = 100 // Maximum limit
	}

	return nil
}

// FindRelationshipsRequest represents a request to find relationships between elements.
// Used by the find_relationships MCP tool to query architecture graph edges.
type FindRelationshipsRequest struct {
	// ProjectRoot is the root directory of the project to search.
	ProjectRoot string

	// SourcePattern is the glob pattern to match source element IDs.
	// Examples: "api-handler", "backend-*", "*-service"
	SourcePattern string

	// TargetPattern is the glob pattern to match target element IDs.
	// Empty string means no target filter.
	TargetPattern string

	// RelationshipType filters by relationship type (e.g., "depends-on", "uses", "calls").
	// Empty string means no type filter.
	RelationshipType string

	// Limit sets the maximum number of results to return.
	// Default: 20, Maximum: 100
	Limit int
}

// FindRelationshipsResponse represents the response from finding relationships.
type FindRelationshipsResponse struct {
	// Relationships contains the matching relationships.
	Relationships []GraphRelationship

	// TotalMatched is the total number of relationships that matched the query
	// (before limit was applied).
	TotalMatched int

	// Message provides context about the results (e.g., "No relationships found").
	Message string
}

// GraphRelationship represents a connection between two architecture elements
// as read from the in-memory architecture graph. It is a read-only result DTO
// returned by the find_relationships use case.
//
// For the persistent C4 model relationship entity (stored in relationships.toml),
// see entities.Relationship in relationship.go.
type GraphRelationship struct {
	// SourceID is the qualified ID of the source element.
	SourceID string

	// TargetID is the qualified ID of the target element.
	TargetID string

	// Type describes the relationship type (e.g., "depends-on", "uses", "calls", "contains").
	Type string

	// Description provides additional context about the relationship.
	Description string
}

// Validate checks if the find relationships request is valid.
func (r *FindRelationshipsRequest) Validate() error {
	if r.ProjectRoot == "" {
		return NewValidationError("FindRelationshipsRequest", "project_root", "", "project root is required", nil)
	}

	// At least one pattern must be specified
	if r.SourcePattern == "" && r.TargetPattern == "" {
		return NewValidationError("FindRelationshipsRequest", "source_pattern", "", "at least one of source_pattern or target_pattern is required", nil)
	}

	// Validate and apply default/max limits
	if r.Limit <= 0 {
		r.Limit = 20 // Default limit
	}
	if r.Limit > 100 {
		r.Limit = 100 // Maximum limit
	}

	return nil
}
