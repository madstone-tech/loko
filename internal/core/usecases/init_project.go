package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// InitProjectRequest carries the inputs for InitProject.Execute.
type InitProjectRequest struct {
	Name        string
	Path        string
	Description string
}

// InitProject scaffolds a new loko project on disk: validates the name,
// constructs the project entity, ensures the project directory exists, and
// persists the project (writes loko.toml + src/) via the repository.
type InitProject struct {
	repo ProjectRepository
}

// NewInitProject creates an InitProject use case backed by repo.
func NewInitProject(repo ProjectRepository) *InitProject {
	return &InitProject{repo: repo}
}

// Execute runs the init flow. Returns nil on success or a wrapped error on
// any validation, filesystem, or persistence failure.
func (uc *InitProject) Execute(ctx context.Context, req *InitProjectRequest) error {
	if req == nil {
		return fmt.Errorf("init project request cannot be nil")
	}
	if req.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if err := entities.ValidateName(req.Name); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	path := req.Path
	if path == "" {
		path = req.Name
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	project, err := entities.NewProject(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	project.Path = absPath
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := uc.repo.SaveProject(ctx, project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}
	return nil
}
