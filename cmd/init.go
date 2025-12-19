package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
)

// InitCommand scaffolds a new loko project.
type InitCommand struct {
	projectName string
	projectPath string
	description string
}

// NewInitCommand creates a new init command.
func NewInitCommand(projectName string) *InitCommand {
	return &InitCommand{
		projectName: projectName,
		projectPath: projectName,
	}
}

// WithDescription sets the project description.
func (ic *InitCommand) WithDescription(desc string) *InitCommand {
	ic.description = desc
	return ic
}

// WithPath sets the project path.
func (ic *InitCommand) WithPath(path string) *InitCommand {
	ic.projectPath = path
	return ic
}

// Execute runs the init command.
// Creates a new project directory with loko.toml and src/ directory.
func (ic *InitCommand) Execute(ctx context.Context) error {
	if ic.projectName == "" {
		return fmt.Errorf("project name is required")
	}

	// Validate project name
	if err := entities.ValidateName(ic.projectName); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Create project directory
	absPath, err := filepath.Abs(ic.projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project entity
	project, err := entities.NewProject(ic.projectName)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	project.Path = absPath
	if ic.description != "" {
		project.Description = ic.description
	}

	// Save project (creates loko.toml and src/)
	repo := filesystem.NewProjectRepository()
	if err := repo.SaveProject(ctx, project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}
