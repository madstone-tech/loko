package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

type CreateContainerRequest struct {
	Name        string
	Description string
	Technology  string
	Tags        []string
}

type CreateContainer struct {
	repo ProjectRepository
}

func NewCreateContainer(repo ProjectRepository) *CreateContainer {
	return &CreateContainer{repo: repo}
}

func (uc *CreateContainer) Execute(ctx context.Context, req *CreateContainerRequest) (*entities.Container, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	container, err := entities.NewContainer(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	container.Description = req.Description
	container.Technology = req.Technology
	container.Tags = req.Tags
	if err := container.Validate(); err != nil {
		return nil, fmt.Errorf("container validation failed: %w", err)
	}
	return container, nil
}
