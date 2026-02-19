package usecases

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// DetectDriftRequest specifies what to check.
type DetectDriftRequest struct {
	ProjectRoot string
	Systems     []*entities.System // If nil, use repo to load them
}

// DetectDriftResult holds all drift issues found.
type DetectDriftResult struct {
	Issues            []entities.DriftIssue
	HasErrors         bool // any DriftError
	HasWarnings       bool // any DriftWarning
	ComponentsChecked int
}

// DetectDrift detects inconsistencies between D2 diagram source and frontmatter.
type DetectDrift struct {
	repo     ProjectRepository
	d2Parser D2Parser // Optional: if nil, skip D2-based checks
}

// NewDetectDrift creates a new DetectDrift use case without D2 parsing.
func NewDetectDrift(repo ProjectRepository) *DetectDrift {
	return &DetectDrift{
		repo:     repo,
		d2Parser: nil,
	}
}

// NewDetectDriftWithD2 creates a new DetectDrift use case with D2 parsing enabled.
func NewDetectDriftWithD2(repo ProjectRepository, parser D2Parser) *DetectDrift {
	return &DetectDrift{
		repo:     repo,
		d2Parser: parser,
	}
}

// Execute runs all drift checks on the project.
func (uc *DetectDrift) Execute(ctx context.Context, req *DetectDriftRequest) (*DetectDriftResult, error) {
	var systems []*entities.System
	var err error

	// Load systems if not provided
	if req.Systems != nil {
		systems = req.Systems
	} else {
		systems, err = uc.repo.ListSystems(ctx, req.ProjectRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to list systems: %w", err)
		}
	}

	// Collect all issues
	var issues []entities.DriftIssue
	componentsChecked := 0

	// Check each system, container, and component
	for _, system := range systems {
		for _, container := range system.Containers {
			for _, component := range container.Components {
				componentsChecked++

				// Check orphaned relationships
				orphanedIssues := uc.checkOrphanedRelationships(component, systems)
				issues = append(issues, orphanedIssues...)

				// Check D2 drift if parser is available and component has path
				if uc.d2Parser != nil && component.Path != "" {
					d2Issues, err := uc.checkD2Drift(ctx, component, uc.d2Parser)
					if err != nil {
						// Log error but continue processing
						continue
					}
					issues = append(issues, d2Issues...)
				}
			}
		}
	}

	// Determine if there are errors or warnings
	hasErrors := false
	hasWarnings := false
	for _, issue := range issues {
		switch issue.Severity {
		case entities.DriftError:
			hasErrors = true
		case entities.DriftWarning:
			hasWarnings = true
		}
	}

	return &DetectDriftResult{
		Issues:            issues,
		HasErrors:         hasErrors,
		HasWarnings:       hasWarnings,
		ComponentsChecked: componentsChecked,
	}, nil
}

// checkOrphanedRelationships checks if component relationships point to non-existent components.
func (uc *DetectDrift) checkOrphanedRelationships(component *entities.Component, systems []*entities.System) []entities.DriftIssue {
	var issues []entities.DriftIssue

	// Create a map of all component IDs for quick lookup
	allComponentIDs := make(map[string]bool)
	for _, system := range systems {
		for _, container := range system.Containers {
			for _, comp := range container.Components {
				allComponentIDs[comp.ID] = true
			}
		}
	}

	// Check each relationship
	for targetID := range component.Relationships {
		if !allComponentIDs[targetID] {
			issue := entities.NewDriftIssue(
				component.ID,
				entities.DriftOrphanedRelationship,
				fmt.Sprintf("Orphaned relationship target: component '%s' no longer exists", targetID),
				fmt.Sprintf("Relationship: '%s → %s', Target '%s' not found", component.ID, targetID, targetID),
			)
			issues = append(issues, *issue)
		}
	}

	return issues
}

// checkD2Drift checks for drift in D2 diagram files.
func (uc *DetectDrift) checkD2Drift(ctx context.Context, component *entities.Component, d2Parser D2Parser) ([]entities.DriftIssue, error) {
	var issues []entities.DriftIssue

	// Try to parse D2 files in component directory
	d2Rels, err := uc.parseComponentD2(ctx, component.Path, d2Parser)
	if err != nil {
		// Skip D2 parsing errors - graceful degradation
		return issues, nil
	}

	// Check for missing components referenced in D2
	missingComponentIssues := uc.checkMissingComponents(component, d2Rels)
	issues = append(issues, missingComponentIssues...)

	// TODO: Check for description mismatches between D2 tooltips and frontmatter descriptions
	// This would require parsing D2 source to extract tooltips and comparing with component.Description

	return issues, nil
}

// parseComponentD2 reads and parses D2 diagram files for a component.
func (uc *DetectDrift) parseComponentD2(ctx context.Context, componentPath string, d2Parser D2Parser) ([]entities.D2Relationship, error) {
	// Look for any .d2 file inside the component directory
	entries, err := os.ReadDir(componentPath)
	if err != nil {
		// Directory not accessible — treat as no D2 file (graceful degradation)
		return nil, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".d2") {
			d2Path := componentPath + "/" + name
			data, err := os.ReadFile(d2Path)
			if err != nil {
				return nil, err
			}
			return d2Parser.ParseRelationships(ctx, string(data))
		}
	}

	return nil, nil // no D2 file found — valid state
}

// checkMissingComponents checks if D2 relationships reference non-existent components.
func (uc *DetectDrift) checkMissingComponents(component *entities.Component, d2Rels []entities.D2Relationship) []entities.DriftIssue {
	var issues []entities.DriftIssue

	// This would check if D2 relationships reference components that don't exist
	// For now, return empty slice as placeholder
	return issues
}
