package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// scaffoldContainer creates a new container entity within a system (US1, T027).
func (uc *ScaffoldEntity) scaffoldContainer(ctx context.Context, req *ScaffoldEntityRequest, project *entities.Project, result *ScaffoldEntityResult) error {
	// Validate parent path
	if len(req.ParentPath) == 0 {
		return fmt.Errorf("parent path must contain system name for container")
	}

	// Load parent system
	systemID := entities.NormalizeName(req.ParentPath[0])
	system, err := uc.projectRepo.LoadSystem(ctx, req.ProjectRoot, systemID)
	if err != nil {
		return fmt.Errorf("failed to load parent system: %w", err)
	}

	// Create container entity
	container, err := entities.NewContainer(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Set optional fields
	container.Description = req.Description
	if req.Technology != "" {
		container.Technology = req.Technology
	}
	if len(req.Tags) > 0 {
		container.Tags = req.Tags
	}

	// Set path
	container.Path = filepath.Join(system.Path, container.ID)

	// Add to system
	if err := system.AddContainer(container); err != nil {
		return fmt.Errorf("failed to add container to system: %w", err)
	}

	// Save container
	if err := uc.projectRepo.SaveContainer(ctx, req.ProjectRoot, systemID, container); err != nil {
		return fmt.Errorf("failed to save container: %w", err)
	}

	result.EntityID = container.ID
	result.FilesCreated = append(result.FilesCreated, filepath.Join(container.Path, "container.toml"))

	// Generate container diagram (shows all containers in system)
	if uc.diagramGenerator != nil {
		d2Source, err := uc.diagramGenerator.GenerateContainerDiagram(system)
		if err != nil {
			return fmt.Errorf("failed to generate container diagram: %w", err)
		}

		d2Path := filepath.Join(container.Path, "container.d2")
		if err := os.MkdirAll(container.Path, 0755); err != nil {
			return fmt.Errorf("failed to create container directory: %w", err)
		}
		if err := os.WriteFile(d2Path, []byte(d2Source), 0644); err != nil {
			return fmt.Errorf("failed to write D2 diagram: %w", err)
		}

		result.DiagramPath = d2Path
		result.FilesCreated = append(result.FilesCreated, d2Path)
	}

	return nil
}
