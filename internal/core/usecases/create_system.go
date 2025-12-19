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

	// Set optional fields
	system.Description = req.Description
	system.Tags = req.Tags
	system.External = req.External

	// Validate the complete system
	if err := system.Validate(); err != nil {
		return nil, fmt.Errorf("system validation failed: %w", err)
	}

	return system, nil
}
