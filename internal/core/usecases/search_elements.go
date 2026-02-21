package usecases

import (
	"context"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// SearchElements searches for architecture elements matching the given criteria.
// This use case supports filtering by name pattern (glob), type, technology, and tags.
type SearchElements struct {
	repo       ProjectRepository
	buildGraph *BuildArchitectureGraph
}

// NewSearchElements creates a new SearchElements use case.
func NewSearchElements(repo ProjectRepository) *SearchElements {
	return &SearchElements{
		repo:       repo,
		buildGraph: NewBuildArchitectureGraph(),
	}
}

// Execute searches for elements matching the request criteria.
func (uc *SearchElements) Execute(ctx context.Context, req entities.SearchElementsRequest) (*entities.SearchElementsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Load systems from project
	systems, err := uc.repo.ListSystems(ctx, req.ProjectRoot)
	if err != nil {
		return nil, err
	}

	// Search and filter elements directly from systems.
	// Lowercase the pattern so matching is case-insensitive — e.g. "*email*"
	// matches "Email Sender" and "ses-email-service" equally.
	matcher := entities.NewGlobMatcher(strings.ToLower(req.Query))
	var results []entities.SearchElement
	totalMatched := 0

	// Search systems
	if req.Type == "" || req.Type == "system" {
		for _, sys := range systems {
			qualifiedID := sys.Name // Systems use their name as ID
			if uc.matchesElement(matcher, qualifiedID, sys.Name, "system", sys.Description, "", sys.Tags, req) {
				totalMatched++
				if len(results) < req.Limit {
					results = append(results, entities.SearchElement{
						ID:          qualifiedID,
						Name:        sys.Name,
						Type:        "system",
						Description: sys.Description,
						Technology:  "",
						Tags:        sys.Tags,
						ParentID:    "",
					})
				}
			}
		}
	}

	// Search containers
	if req.Type == "" || req.Type == "container" {
		for _, sys := range systems {
			for _, cont := range sys.Containers {
				qualifiedID := sys.Name + "/" + cont.Name
				if uc.matchesElement(matcher, qualifiedID, cont.Name, "container", cont.Description, cont.Technology, cont.Tags, req) {
					totalMatched++
					if len(results) < req.Limit {
						results = append(results, entities.SearchElement{
							ID:          qualifiedID,
							Name:        cont.Name,
							Type:        "container",
							Description: cont.Description,
							Technology:  cont.Technology,
							Tags:        cont.Tags,
							ParentID:    sys.Name,
						})
					}
				}
			}
		}
	}

	// Search components
	if req.Type == "" || req.Type == "component" {
		for _, sys := range systems {
			for _, cont := range sys.Containers {
				for _, comp := range cont.Components {
					qualifiedID := sys.Name + "/" + cont.Name + "/" + comp.Name
					if uc.matchesElement(matcher, qualifiedID, comp.Name, "component", comp.Description, comp.Technology, comp.Tags, req) {
						totalMatched++
						if len(results) < req.Limit {
							results = append(results, entities.SearchElement{
								ID:          qualifiedID,
								Name:        comp.Name,
								Type:        "component",
								Description: comp.Description,
								Technology:  comp.Technology,
								Tags:        comp.Tags,
								ParentID:    sys.Name + "/" + cont.Name,
							})
						}
					}
				}
			}
		}
	}

	// Build response message
	message := uc.buildMessage(totalMatched, len(results), req)

	return &entities.SearchElementsResponse{
		Results:      results,
		TotalMatched: totalMatched,
		Message:      message,
	}, nil
}

// matchesElement checks if an element matches all filter criteria.
func (uc *SearchElements) matchesElement(
	matcher *entities.GlobMatcher,
	id, name, elemType, description, technology string,
	tags []string,
	req entities.SearchElementsRequest,
) bool {
	// Check name pattern (glob match on both ID and name, case-insensitive).
	// Lowercasing both the candidates and the pattern normalises queries like
	// "*email*" to match "Email Sender" or "ses-email-service".
	idLower := strings.ToLower(id)
	nameLower := strings.ToLower(name)
	if !matcher.Match(idLower) && !matcher.Match(nameLower) {
		return false
	}

	// Check technology filter (if specified)
	if req.Technology != "" && technology != req.Technology {
		return false
	}

	// Check tag filter (if specified)
	if req.Tag != "" {
		hasTag := false
		for _, tag := range tags {
			if tag == req.Tag {
				hasTag = true
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// buildMessage creates a helpful message about the search results.
func (uc *SearchElements) buildMessage(totalMatched, returned int, req entities.SearchElementsRequest) string {
	if totalMatched == 0 {
		return "No elements found matching query"
	}

	if returned < totalMatched {
		return formatMessage("Showing %d of %d matching elements (use limit parameter to adjust)", returned, totalMatched)
	}

	filters := []string{}
	if req.Type != "" {
		filters = append(filters, "type="+req.Type)
	}
	if req.Technology != "" {
		filters = append(filters, "technology="+req.Technology)
	}
	if req.Tag != "" {
		filters = append(filters, "tag="+req.Tag)
	}

	if len(filters) > 0 {
		return formatMessage("Found %d elements matching '%s' with filters: %s", totalMatched, req.Query, strings.Join(filters, ", "))
	}

	return formatMessage("Found %d elements matching '%s'", totalMatched, req.Query)
}

// formatMessage is a helper to format messages consistently.
func formatMessage(format string, args ...interface{}) string {
	// This could be replaced with fmt.Sprintf, but having a helper
	// allows for future message formatting enhancements
	s := format
	for _, arg := range args {
		idx := strings.Index(s, "%d")
		if idx == -1 {
			idx = strings.Index(s, "%s")
		}
		if idx == -1 {
			break
		}

		var val string
		switch v := arg.(type) {
		case int:
			val = intToString(v)
		case string:
			val = v
		default:
			val = ""
		}

		s = s[:idx] + val + s[idx+2:]
	}
	return s
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var result []byte
	for i > 0 {
		result = append([]byte{byte('0' + i%10)}, result...)
		i /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}
