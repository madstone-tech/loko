package entities

import (
	"errors"
	"fmt"
)

// D2Relationship represents a relationship extracted from D2 diagram syntax.
type D2Relationship struct {
	Source string // Source component ID (extracted from D2 node)
	Target string // Target component ID (extracted from D2 node)
	Label  string // Arrow label (relationship type)
}

// NewD2Relationship creates a validated D2Relationship.
func NewD2Relationship(source, target, label string) (*D2Relationship, error) {
	if source == "" {
		return nil, errors.New("source cannot be empty")
	}
	if target == "" {
		return nil, errors.New("target cannot be empty")
	}
	return &D2Relationship{
		Source: source,
		Target: target,
		Label:  label, // Label can be empty (unlabeled arrow)
	}, nil
}

// Key returns a unique identifier for deduplication (source+target+label).
func (r *D2Relationship) Key() string {
	return fmt.Sprintf("%s->%s:%s", r.Source, r.Target, r.Label)
}
