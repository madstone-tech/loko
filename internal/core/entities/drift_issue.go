package entities

// DriftSeverity represents the severity of a drift issue.
type DriftSeverity int

const (
	DriftWarning DriftSeverity = iota // Cosmetic inconsistencies
	DriftError                        // Broken references, data integrity issues
)

// DriftType categorizes the kind of drift detected.
type DriftType int

const (
	DriftDescriptionMismatch  DriftType = iota // D2 tooltip != frontmatter description
	DriftMissingComponent                      // D2 references non-existent component
	DriftOrphanedRelationship                  // Frontmatter relationship to deleted component
)

// DriftIssue represents a detected inconsistency between data sources.
type DriftIssue struct {
	ComponentID string        // Component where drift detected
	Type        DriftType     // Category of drift
	Severity    DriftSeverity // Warning or Error
	Message     string        // Human-readable description
	Context     string        // Additional context (e.g., expected vs actual)
}

// NewDriftIssue creates a validated DriftIssue.
func NewDriftIssue(componentID string, driftType DriftType, message string, context string) *DriftIssue {
	severity := DriftWarning
	if driftType == DriftMissingComponent || driftType == DriftOrphanedRelationship {
		severity = DriftError
	}

	return &DriftIssue{
		ComponentID: componentID,
		Type:        driftType,
		Severity:    severity,
		Message:     message,
		Context:     context,
	}
}
