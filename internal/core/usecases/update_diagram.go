package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateDiagramRequest contains the input parameters for updating a D2 diagram.
type UpdateDiagramRequest struct {
	ProjectRoot string // filesystem path to project
	DiagramPath string // relative path to .d2 file within project
	D2Source    string // D2 source code to write
}

// UpdateDiagramResult contains the output of the update diagram operation.
type UpdateDiagramResult struct {
	FilePath string // absolute path of written file
}

// UpdateDiagram is a use case for validating and writing D2 diagram source code.
type UpdateDiagram struct{}

// NewUpdateDiagram creates a new UpdateDiagram use case instance.
func NewUpdateDiagram() *UpdateDiagram {
	return &UpdateDiagram{}
}

// Execute validates the request and writes the D2 source code to the specified file.
//
// Preconditions:
//   - Project exists at ProjectRoot
//   - DiagramPath is within project source directory
//   - D2Source is non-empty
//
// Postconditions:
//   - D2 file written at DiagramPath
//   - File contains provided D2Source content
//
// Error Cases:
//   - ErrInvalidPath: path outside project or invalid extension
//   - ErrEmptyContent: D2Source is empty
func (uc *UpdateDiagram) Execute(ctx context.Context, req *UpdateDiagramRequest) (*UpdateDiagramResult, error) {
	// Validate request is not nil
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate D2Source is non-empty
	trimmedSource := strings.TrimSpace(req.D2Source)
	if trimmedSource == "" {
		return nil, fmt.Errorf("D2 source code cannot be empty")
	}

	// Validate DiagramPath has .d2 extension
	if filepath.Ext(req.DiagramPath) != ".d2" {
		return nil, fmt.Errorf("diagram path must have .d2 extension: %s", req.DiagramPath)
	}

	// Validate DiagramPath is relative (not absolute)
	if filepath.IsAbs(req.DiagramPath) {
		return nil, fmt.Errorf("diagram path must be relative, not absolute: %s", req.DiagramPath)
	}

	// Construct absolute path
	absPath := filepath.Join(req.ProjectRoot, req.DiagramPath)

	// Ensure parent directory exists
	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(absPath, []byte(req.D2Source), 0644); err != nil {
		return nil, fmt.Errorf("failed to write D2 file: %w", err)
	}

	return &UpdateDiagramResult{
		FilePath: absPath,
	}, nil
}
