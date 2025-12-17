package entities

import (
	"path/filepath"
	"time"
)

// DiagramFormat represents the output format of a rendered diagram.
type DiagramFormat string

const (
	DiagramFormatSVG DiagramFormat = "svg"
	DiagramFormatPNG DiagramFormat = "png"
)

// Diagram represents a D2 diagram with source and rendered output.
type Diagram struct {
	// ID is the unique identifier (derived from source path)
	ID string
	
	// SourcePath is the path to the .d2 source file
	SourcePath string
	
	// Source is the D2 source code
	Source string
	
	// OutputPath is the path to the rendered output (may be empty if not rendered)
	OutputPath string
	
	// Format is the output format (svg or png)
	Format DiagramFormat
	
	// Hash is the content hash for cache invalidation
	Hash string
	
	// RenderedAt is when the diagram was last rendered
	RenderedAt time.Time
	
	// Error contains any rendering error
	Error string
}

// NewDiagram creates a new diagram from a source file path.
func NewDiagram(sourcePath string) (*Diagram, error) {
	if err := ValidatePath(sourcePath); err != nil {
		return nil, NewValidationError("Diagram", "SourcePath", sourcePath, "invalid source path", err)
	}
	
	// Derive ID from filename without extension
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	id := base[:len(base)-len(ext)]
	
	return &Diagram{
		ID:         id,
		SourcePath: sourcePath,
		Format:     DiagramFormatSVG, // Default to SVG
	}, nil
}

// Validate checks if the diagram is valid.
func (d *Diagram) Validate() error {
	var errs ValidationErrors
	
	if d.ID == "" {
		errs.Add("Diagram", "ID", "", "id is required", ErrEmptyID)
	}
	
	if d.SourcePath == "" {
		errs.Add("Diagram", "SourcePath", "", "source path is required", ErrEmptyPath)
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// IsRendered returns true if the diagram has been rendered.
func (d *Diagram) IsRendered() bool {
	return d.OutputPath != "" && d.Error == ""
}

// NeedsRender returns true if the diagram needs to be (re)rendered.
// This is determined by comparing the source hash with the cached hash.
func (d *Diagram) NeedsRender(currentHash string) bool {
	return d.Hash != currentHash || !d.IsRendered()
}

// SetSource updates the source code and clears the render state.
func (d *Diagram) SetSource(source string) {
	d.Source = source
	d.OutputPath = ""
	d.Hash = ""
	d.RenderedAt = time.Time{}
	d.Error = ""
}

// SetRendered marks the diagram as successfully rendered.
func (d *Diagram) SetRendered(outputPath, hash string) {
	d.OutputPath = outputPath
	d.Hash = hash
	d.RenderedAt = time.Now()
	d.Error = ""
}

// SetError marks the diagram as failed to render.
func (d *Diagram) SetError(err string) {
	d.Error = err
	d.OutputPath = ""
}
