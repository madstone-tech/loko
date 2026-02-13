package entities

// C4Entity is the common interface for all C4 model entities.
// This enables type-safe graph operations without runtime type assertions.
type C4Entity interface {
	// GetID returns the entity's unique identifier
	GetID() string

	// GetName returns the entity's display name
	GetName() string

	// GetEntityType returns the C4 entity type: "system", "container", or "component"
	GetEntityType() string
}
