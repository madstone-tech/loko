package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// scaffoldComponent creates a new component entity within a container (US1, T028).
func (uc *ScaffoldEntity) scaffoldComponent(ctx context.Context, req *ScaffoldEntityRequest, project *entities.Project, result *ScaffoldEntityResult) error {
	// Validate parent path
	if len(req.ParentPath) < 2 {
		return fmt.Errorf("parent path must contain system and container names for component")
	}

	// Load parent system
	systemID := entities.NormalizeName(req.ParentPath[0])
	system, err := uc.projectRepo.LoadSystem(ctx, req.ProjectRoot, systemID)
	if err != nil {
		return fmt.Errorf("failed to load parent system: %w", err)
	}

	// Find container in system
	containerID := entities.NormalizeName(req.ParentPath[1])
	container, ok := system.Containers[containerID]
	if !ok {
		return fmt.Errorf("container %s not found in system %s", containerID, systemID)
	}

	// Create component entity
	component, err := entities.NewComponent(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	// Set optional fields
	component.Description = req.Description
	if req.Technology != "" {
		component.Technology = req.Technology
	}
	if len(req.Tags) > 0 {
		component.Tags = req.Tags
	}
	// T055: propagate technology-specific content template
	if req.ContentTemplate != "" {
		component.ContentTemplate = req.ContentTemplate
	}

	// Set path
	component.Path = filepath.Join(container.Path, component.ID)

	// Add to container
	if err := container.AddComponent(component); err != nil {
		return fmt.Errorf("failed to add component to container: %w", err)
	}

	// Save component
	if err := uc.projectRepo.SaveComponent(ctx, req.ProjectRoot, systemID, containerID, component); err != nil {
		return fmt.Errorf("failed to save component: %w", err)
	}

	result.EntityID = component.ID
	result.FilesCreated = append(result.FilesCreated, filepath.Join(component.Path, "component.toml"))

	// Generate component diagram
	if uc.diagramGenerator != nil {
		d2Source, err := uc.diagramGenerator.GenerateComponentDiagram(container)
		if err != nil {
			return fmt.Errorf("failed to generate component diagram: %w", err)
		}

		d2Path := filepath.Join(component.Path, "component.d2")
		if err := os.MkdirAll(component.Path, 0755); err != nil {
			return fmt.Errorf("failed to create component directory: %w", err)
		}
		if err := os.WriteFile(d2Path, []byte(d2Source), 0644); err != nil {
			return fmt.Errorf("failed to write D2 diagram: %w", err)
		}

		result.DiagramPath = d2Path
		result.FilesCreated = append(result.FilesCreated, d2Path)
	}

	return nil
}
