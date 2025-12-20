package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// BuildDocs orchestrates the process of rendering diagrams and building documentation.
//
// This use case:
// 1. Iterates through all systems, containers, and components
// 2. Renders D2 diagrams to SVG using the DiagramRenderer (C4 levels 1-3)
// 3. Calls the SiteBuilder to generate HTML documentation
// 4. Reports progress via ProgressReporter
type BuildDocs struct {
	diagramRenderer  DiagramRenderer
	siteBuilder      SiteBuilder
	progressReporter ProgressReporter
}

// NewBuildDocs creates a new BuildDocs use case with the given adapters.
func NewBuildDocs(
	diagramRenderer DiagramRenderer,
	siteBuilder SiteBuilder,
	progressReporter ProgressReporter,
) *BuildDocs {
	return &BuildDocs{
		diagramRenderer:  diagramRenderer,
		siteBuilder:      siteBuilder,
		progressReporter: progressReporter,
	}
}

// Execute performs a complete documentation build.
//
// It:
// 1. Renders all diagrams in the project (systems, containers, and components)
// 2. Calls BuildSite to generate HTML documentation
// 3. Reports progress and errors
// 4. Returns error if any rendering fails
func (uc *BuildDocs) Execute(
	ctx context.Context,
	project *entities.Project,
	systems []*entities.System,
	outputDir string,
) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}
	if len(systems) == 0 {
		uc.progressReporter.ReportInfo("No systems found to build")
		return nil
	}

	uc.progressReporter.ReportInfo("Starting documentation build...")

	// Count total diagrams for progress reporting
	totalDiagrams := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			totalDiagrams++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				totalDiagrams++
			}
			for _, component := range container.Components {
				if component.Diagram != nil {
					totalDiagrams++
				}
			}
		}
	}

	// Render all diagrams and save to files
	diagramCount := 0
	for _, sys := range systems {
		// Render system diagram
		if sys.Diagram != nil {
			diagramCount++
			uc.progressReporter.ReportProgress(
				fmt.Sprintf("Rendering system diagram: %s", sys.Name),
				diagramCount,
				totalDiagrams,
				fmt.Sprintf("Rendering %s diagram", sys.Name),
			)

			// Create unique filename for diagram
			diagramFileName := fmt.Sprintf("%s.svg", sys.ID)
			diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

			svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, sys.Diagram.Source)
			if err != nil {
				uc.progressReporter.ReportError(fmt.Errorf("failed to render diagram for system %s: %w", sys.Name, err))
				return fmt.Errorf("failed to render diagram for system %s: %w", sys.Name, err)
			}

			// Save SVG to file
			if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
				return fmt.Errorf("failed to create diagrams directory: %w", err)
			}
			if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
				return fmt.Errorf("failed to save diagram for system %s: %w", sys.Name, err)
			}

			// Store diagram path in entity for use by HTML builder
			sys.DiagramPath = filepath.Join("diagrams", diagramFileName)
		}

		// Render container diagrams
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				diagramCount++
				uc.progressReporter.ReportProgress(
					fmt.Sprintf("Rendering container diagram: %s/%s", sys.Name, container.Name),
					diagramCount,
					totalDiagrams,
					fmt.Sprintf("Rendering %s diagram in %s", container.Name, sys.Name),
				)

				// Create unique filename for diagram
				diagramFileName := fmt.Sprintf("%s_%s.svg", sys.ID, container.ID)
				diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

				svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, container.Diagram.Source)
				if err != nil {
					uc.progressReporter.ReportError(fmt.Errorf("failed to render diagram for container %s/%s: %w", sys.Name, container.Name, err))
					return fmt.Errorf("failed to render diagram for container %s/%s: %w", sys.Name, container.Name, err)
				}

				// Save SVG to file
				if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
					return fmt.Errorf("failed to create diagrams directory: %w", err)
				}
				if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
					return fmt.Errorf("failed to save diagram for container %s/%s: %w", sys.Name, container.Name, err)
				}

				// Store diagram path in entity for use by HTML builder
				container.DiagramPath = filepath.Join("diagrams", diagramFileName)
			}

			// Render component diagrams
			enhancer := NewEnhanceComponentDiagram()
			for _, component := range container.Components {
				if component.Diagram != nil {
					diagramCount++
					uc.progressReporter.ReportProgress(
						fmt.Sprintf("Rendering component diagram: %s/%s/%s", sys.Name, container.Name, component.Name),
						diagramCount,
						totalDiagrams,
						fmt.Sprintf("Rendering %s diagram in %s", component.Name, container.Name),
					)

					// Enhance component diagram with relationships and metadata
					enhancedD2Source, err := enhancer.Execute(component, container, sys)
					if err != nil {
						uc.progressReporter.ReportError(fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err))
						return fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					// Create unique filename for diagram
					diagramFileName := fmt.Sprintf("%s_%s_%s.svg", sys.ID, container.ID, component.ID)
					diagramPath := filepath.Join(outputDir, "diagrams", diagramFileName)

					svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, enhancedD2Source)
					if err != nil {
						uc.progressReporter.ReportError(fmt.Errorf("failed to render diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err))
						return fmt.Errorf("failed to render diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					// Save SVG to file
					if err := os.MkdirAll(filepath.Dir(diagramPath), 0755); err != nil {
						return fmt.Errorf("failed to create diagrams directory: %w", err)
					}
					if err := os.WriteFile(diagramPath, []byte(svgContent), 0644); err != nil {
						return fmt.Errorf("failed to save diagram for component %s/%s/%s: %w", sys.Name, container.Name, component.Name, err)
					}

					// Store diagram path in entity for use by HTML builder
					component.DiagramPath = filepath.Join("diagrams", diagramFileName)
				}
			}
		}
	}

	// Build the site
	uc.progressReporter.ReportProgress("Building site", len(systems), len(systems), "Generating HTML documentation...")
	err := uc.siteBuilder.BuildSite(ctx, project, systems, outputDir)
	if err != nil {
		uc.progressReporter.ReportError(fmt.Errorf("failed to build site: %w", err))
		return fmt.Errorf("failed to build site: %w", err)
	}

	uc.progressReporter.ReportSuccess(fmt.Sprintf("Documentation built successfully in %s", outputDir))
	return nil
}
