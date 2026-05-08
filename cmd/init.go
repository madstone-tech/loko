package cmd

import (
	"context"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/usecases"
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
	uc := usecases.NewInitProject(filesystem.NewProjectRepository())
	return uc.Execute(ctx, &usecases.InitProjectRequest{
		Name:        ic.projectName,
		Path:        ic.projectPath,
		Description: ic.description,
	})
}
