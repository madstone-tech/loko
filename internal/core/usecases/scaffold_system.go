package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// scaffoldSystem creates a new system entity within the project (US1, T026).
func (uc *ScaffoldEntity) scaffoldSystem(ctx context.Context, req *ScaffoldEntityRequest, project *entities.Project, result *ScaffoldEntityResult) error {
	// Create system entity
	system, err := entities.NewSystem(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	// Set optional fields
	system.Description = req.Description
	if len(req.Tags) > 0 {
		system.Tags = req.Tags
	}

	// Set path
	system.Path = filepath.Join(req.ProjectRoot, project.Config.SourceDir, system.ID)

	// Add to project
	if err := project.AddSystem(system); err != nil {
		return fmt.Errorf("failed to add system to project: %w", err)
	}

	// Save system
	if err := uc.projectRepo.SaveSystem(ctx, req.ProjectRoot, system); err != nil {
		return fmt.Errorf("failed to save system: %w", err)
	}

	result.EntityID = system.ID
	result.FilesCreated = append(result.FilesCreated, filepath.Join(system.Path, "system.toml"))

	// Generate system context diagram
	if uc.diagramGenerator != nil {
		d2Source, err := uc.diagramGenerator.GenerateSystemContextDiagram(system)
		if err != nil {
			return fmt.Errorf("failed to generate system context diagram: %w", err)
		}

		d2Path := filepath.Join(system.Path, "system.d2")
		if err := os.MkdirAll(system.Path, 0755); err != nil {
			return fmt.Errorf("failed to create system directory: %w", err)
		}
		if err := os.WriteFile(d2Path, []byte(d2Source), 0644); err != nil {
			return fmt.Errorf("failed to write D2 diagram: %w", err)
		}

		result.DiagramPath = d2Path
		result.FilesCreated = append(result.FilesCreated, d2Path)
	}

	return nil
}
