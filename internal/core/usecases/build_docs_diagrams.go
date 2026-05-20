// Package usecases — parallel diagram-rendering pipeline for BuildDocs.
// See build_docs.go for the use case type, constructors, and Execute methods.
package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// diagramJob represents a single diagram rendering task.
type diagramJob struct {
	source        string // D2 source code to render
	fileName      string // Output SVG filename (e.g., "sys-id.svg")
	label         string // Human-readable label for progress (e.g., "system PaymentService")
	writeD2Source bool   // When true, write the D2 source alongside the SVG as <stem>.d2
}

// diagramResult holds the outcome of a diagram rendering job.
type diagramResult struct {
	index      int
	svgContent string
	err        error
}

// renderDiagrams renders all D2 diagrams to SVG files using a parallel worker pool.
func (uc *BuildDocs) renderDiagrams(
	ctx context.Context,
	systems []*entities.System,
	outputDir string,
) error {
	// Collect all diagram jobs
	type pathSetter func(path string)
	var jobs []diagramJob
	var setters []pathSetter

	enhancer := NewEnhanceComponentDiagram()

	for _, sys := range systems {
		if sys.Diagram != nil {
			fileName := fmt.Sprintf("%s.svg", sys.ID)
			source := sys.Diagram.Source
			jobs = append(jobs, diagramJob{
				source:   source,
				fileName: fileName,
				label:    fmt.Sprintf("system %s", sys.Name),
			})
			s := sys // capture for closure
			setters = append(setters, func(path string) { s.DiagramPath = path })
		}

		for _, container := range sys.Containers {
			if container.Diagram != nil {
				fileName := fmt.Sprintf("%s_%s.svg", sys.ID, container.ID)
				jobs = append(jobs, diagramJob{
					source:   container.Diagram.Source,
					fileName: fileName,
					label:    fmt.Sprintf("container %s/%s", sys.Name, container.Name),
				})
				c := container
				setters = append(setters, func(path string) { c.DiagramPath = path })
			}

			for _, component := range container.Components {
				if component.Diagram != nil {
					enhancedSource, err := enhancer.Execute(component, container, sys)
					if err != nil {
						return fmt.Errorf("failed to enhance diagram for component %s/%s/%s: %w",
							sys.Name, container.Name, component.Name, err)
					}
					fileName := fmt.Sprintf("%s_%s_%s.svg", sys.ID, container.ID, component.ID)
					jobs = append(jobs, diagramJob{
						source:        enhancedSource,
						fileName:      fileName,
						label:         fmt.Sprintf("component %s/%s/%s", sys.Name, container.Name, component.Name),
						writeD2Source: true,
					})
					comp := component
					setters = append(setters, func(path string) { comp.DiagramPath = path })
				}
			}
		}
	}

	if len(jobs) == 0 {
		return nil
	}

	uc.progressReporter.ReportInfo(fmt.Sprintf("Rendering %d diagrams...", len(jobs)))

	// Create diagrams directory once
	diagramsDir := filepath.Join(outputDir, "diagrams")
	if err := os.MkdirAll(diagramsDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagrams directory: %w", err)
	}

	// Determine worker count
	numWorkers := min(8, len(jobs))

	// Channel-based worker pool
	jobCh := make(chan int, len(jobs))
	resultCh := make(chan diagramResult, len(jobs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobCh {
				job := jobs[idx]
				svgContent, err := uc.diagramRenderer.RenderDiagram(ctx, job.source)
				resultCh <- diagramResult{index: idx, svgContent: svgContent, err: err}
			}
		}()
	}

	// Send all jobs
	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	completed := 0
	for result := range resultCh {
		completed++
		job := jobs[result.index]

		if result.err != nil {
			return fmt.Errorf("failed to render diagram for %s: %w", job.label, result.err)
		}

		uc.progressReporter.ReportProgress(
			fmt.Sprintf("Rendered %s", job.label),
			completed, len(jobs),
			fmt.Sprintf("Rendering diagrams (%d/%d)", completed, len(jobs)),
		)

		// Write SVG to disk
		diagramPath := filepath.Join(diagramsDir, job.fileName)
		if err := os.WriteFile(diagramPath, []byte(result.svgContent), 0644); err != nil {
			return fmt.Errorf("failed to save diagram for %s: %w", job.label, err)
		}

		// Write the enhanced D2 source alongside the SVG so it can be inspected and
		// diffed against the src/ component .d2 file.
		if job.writeD2Source {
			d2FileName := strings.TrimSuffix(job.fileName, ".svg") + ".d2"
			d2Path := filepath.Join(diagramsDir, d2FileName)
			if err := os.WriteFile(d2Path, []byte(job.source), 0644); err != nil {
				return fmt.Errorf("failed to save D2 source for %s: %w", job.label, err)
			}
		}

		// Set diagram path on entity
		setters[result.index](filepath.Join("diagrams", job.fileName))
	}

	uc.progressReporter.ReportProgress("Diagrams", len(jobs), len(jobs), "All diagrams rendered")
	return nil
}
