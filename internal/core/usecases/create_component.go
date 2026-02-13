package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// CreateComponentRequest holds the input for creating a component.
type CreateComponentRequest struct {
	// Name is the display name of the component (e.g., "Authentication Handler")
	Name string

	// Description explains what the component does
	Description string

	// Technology describes the implementation (e.g., "Go package", "React component")
	Technology string

	// Tags for categorization
	Tags []string
}

// CreateComponent is the use case for creating a new component.
type CreateComponent struct {
	repo ProjectRepository
}

// NewCreateComponent creates a new CreateComponent use case.
func NewCreateComponent(repo ProjectRepository) *CreateComponent {
	return &CreateComponent{repo: repo}
}

// Execute creates a new component with the given request.
// Returns the created component or an error if validation fails.
func (uc *CreateComponent) Execute(ctx context.Context, req *CreateComponentRequest) (*entities.Component, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create the component (validates name)
	component, err := entities.NewComponent(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	// Set basic fields
	component.Description = req.Description
	component.Technology = req.Technology
	component.Tags = req.Tags

	// Validate the complete component
	if err := component.Validate(); err != nil {
		return nil, fmt.Errorf("component validation failed: %w", err)
	}

	return component, nil
}
