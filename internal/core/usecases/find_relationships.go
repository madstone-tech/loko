package usecases

import (
	"context"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// FindRelationships finds relationships between architecture elements.
// This use case supports filtering by source/target patterns and relationship type.
type FindRelationships struct {
	repo       ProjectRepository
	buildGraph *BuildArchitectureGraph
}

// NewFindRelationships creates a new FindRelationships use case.
func NewFindRelationships(repo ProjectRepository) *FindRelationships {
	return &FindRelationships{
		repo:       repo,
		buildGraph: NewBuildArchitectureGraph(),
	}
}

// Execute finds relationships matching the request criteria.
func (uc *FindRelationships) Execute(ctx context.Context, req entities.FindRelationshipsRequest) (*entities.FindRelationshipsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Load project and systems
	project, err := uc.repo.LoadProject(ctx, req.ProjectRoot)
	if err != nil {
		return nil, err
	}

	systems, err := uc.repo.ListSystems(ctx, req.ProjectRoot)
	if err != nil {
		return nil, err
	}

	// Build architecture graph
	graph, err := uc.buildGraph.Execute(ctx, project, systems)
	if err != nil {
		return nil, err
	}

	// Search relationships
	var results []entities.Relationship
	totalMatched := 0

	// Create matchers for source and target patterns
	var sourceMatcher, targetMatcher *entities.GlobMatcher
	if req.SourcePattern != "" {
		sourceMatcher = entities.NewGlobMatcher(req.SourcePattern)
	}
	if req.TargetPattern != "" {
		targetMatcher = entities.NewGlobMatcher(req.TargetPattern)
	}

	// Iterate through all edges in the graph
	for sourceID, edges := range graph.Edges {
		// Check source pattern
		if sourceMatcher != nil && !sourceMatcher.Match(sourceID) {
			continue
		}

		for _, edge := range edges {
			// Check target pattern
			if targetMatcher != nil && !targetMatcher.Match(edge.Target) {
				continue
			}

			// Check relationship type filter
			if req.RelationshipType != "" && edge.Type != req.RelationshipType {
				continue
			}

			// Match found
			totalMatched++
			if len(results) < req.Limit {
				results = append(results, entities.Relationship{
					SourceID:    sourceID,
					TargetID:    edge.Target,
					Type:        edge.Type,
					Description: edge.Description,
				})
			}
		}
	}

	// Build response message
	message := uc.buildMessage(totalMatched, len(results), req)

	return &entities.FindRelationshipsResponse{
		Relationships: results,
		TotalMatched:  totalMatched,
		Message:       message,
	}, nil
}

// buildMessage creates a helpful message about the relationship results.
func (uc *FindRelationships) buildMessage(totalMatched, returned int, req entities.FindRelationshipsRequest) string {
	if totalMatched == 0 {
		return "No relationships found"
	}

	if returned < totalMatched {
		return formatMessage("Showing %d of %d matching relationships (use limit parameter to adjust)", returned, totalMatched)
	}

	filters := []string{}
	if req.SourcePattern != "" {
		filters = append(filters, "source="+req.SourcePattern)
	}
	if req.TargetPattern != "" {
		filters = append(filters, "target="+req.TargetPattern)
	}
	if req.RelationshipType != "" {
		filters = append(filters, "type="+req.RelationshipType)
	}

	if len(filters) > 0 {
		return formatMessage("Found %d relationships with filters: %s", totalMatched, strings.Join(filters, ", "))
	}

	return formatMessage("Found %d relationships", totalMatched)
}
