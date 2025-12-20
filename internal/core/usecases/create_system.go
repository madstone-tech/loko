package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// CreateSystemRequest holds the input for creating a system.
type CreateSystemRequest struct {
	// Name is the display name of the system (e.g., "Payment Service")
	Name string

	// Description explains what the system does
	Description string

	// Tags for categorization
	Tags []string

	// External indicates if this is an external system
	External bool

	// Responsibilities lists key responsibilities of this system
	Responsibilities []string

	// Dependencies lists external systems or services this system depends on
	Dependencies []string

	// PrimaryLanguage is the main programming language used
	PrimaryLanguage string

	// Framework is the primary framework/library
	Framework string

	// Database is the primary data storage technology
	Database string

	// KeyUsers lists the primary users or actors
	KeyUsers []string

	// ExternalSystems lists external systems this integrates with
	ExternalSystems []string
}

// CreateSystem is the use case for creating a new system.
type CreateSystem struct {
	repo ProjectRepository
}

// NewCreateSystem creates a new CreateSystem use case.
func NewCreateSystem(repo ProjectRepository) *CreateSystem {
	return &CreateSystem{repo: repo}
}

// Execute creates a new system with the given request.
// Returns the created system or an error if validation fails.
func (uc *CreateSystem) Execute(ctx context.Context, req *CreateSystemRequest) (*entities.System, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create the system (validates name)
	system, err := entities.NewSystem(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create system: %w", err)
	}

	// Set basic fields
	system.Description = req.Description
	system.Tags = req.Tags
	system.External = req.External

	// Set detailed fields
	system.Responsibilities = req.Responsibilities
	system.Dependencies = req.Dependencies
	system.PrimaryLanguage = req.PrimaryLanguage
	system.Framework = req.Framework
	system.Database = req.Database
	system.KeyUsers = req.KeyUsers
	system.ExternalSystems = req.ExternalSystems

	// Validate the complete system
	if err := system.Validate(); err != nil {
		return nil, fmt.Errorf("system validation failed: %w", err)
	}

	return system, nil
}
