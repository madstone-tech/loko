package entities

import (
	"testing"
	"time"
)

func TestNewDiagram(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		wantID     string
		wantErr    bool
	}{
		{"valid simple", "system.d2", "system", false},
		{"valid with path", "/path/to/diagram.d2", "diagram", false},
		{"valid nested", "src/payment/container.d2", "container", false},
		{"empty", "", "", true},
		{"path traversal", "../../../etc/passwd", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDiagram(tt.sourcePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDiagram(%q) error = %v, wantErr %v", tt.sourcePath, err, tt.wantErr)
				return
			}
			if err == nil {
				if d.ID != tt.wantID {
					t.Errorf("NewDiagram(%q).ID = %q, want %q", tt.sourcePath, d.ID, tt.wantID)
				}
				if d.Format != DiagramFormatSVG {
					t.Errorf("NewDiagram should default to SVG format")
				}
			}
		})
	}
}

func TestDiagram_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		d, _ := NewDiagram("system.d2")
		if err := d.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("invalid - empty id", func(t *testing.T) {
		d := &Diagram{ID: "", SourcePath: "test.d2"}
		if err := d.Validate(); err == nil {
			t.Error("Validate() should fail for empty ID")
		}
	})

	t.Run("invalid - empty source path", func(t *testing.T) {
		d := &Diagram{ID: "test", SourcePath: ""}
		if err := d.Validate(); err == nil {
			t.Error("Validate() should fail for empty SourcePath")
		}
	})
}

func TestDiagram_RenderState(t *testing.T) {
	d, _ := NewDiagram("system.d2")

	// Initially not rendered
	if d.IsRendered() {
		t.Error("New diagram should not be rendered")
	}

	// Needs render with any hash
	if !d.NeedsRender("abc123") {
		t.Error("New diagram should need render")
	}

	// Set source
	d.SetSource("User -> API")
	if d.Source != "User -> API" {
		t.Error("SetSource failed")
	}

	// Mark as rendered
	d.SetRendered("/output/system.svg", "abc123")
	if !d.IsRendered() {
		t.Error("Diagram should be rendered after SetRendered")
	}
	if d.OutputPath != "/output/system.svg" {
		t.Error("SetRendered should set OutputPath")
	}
	if d.Hash != "abc123" {
		t.Error("SetRendered should set Hash")
	}
	if d.RenderedAt.IsZero() {
		t.Error("SetRendered should set RenderedAt")
	}
	if d.Error != "" {
		t.Error("SetRendered should clear Error")
	}

	// Same hash - no render needed
	if d.NeedsRender("abc123") {
		t.Error("Same hash should not need render")
	}

	// Different hash - needs render
	if !d.NeedsRender("different") {
		t.Error("Different hash should need render")
	}

	// Set error
	d.SetError("d2 syntax error")
	if d.IsRendered() {
		t.Error("Diagram should not be rendered after SetError")
	}
	if d.Error != "d2 syntax error" {
		t.Error("SetError failed")
	}
	if d.OutputPath != "" {
		t.Error("SetError should clear OutputPath")
	}
}

func TestDiagram_SetSourceClearsState(t *testing.T) {
	d, _ := NewDiagram("system.d2")
	d.SetRendered("/output/system.svg", "abc123")
	d.RenderedAt = time.Now()

	// SetSource should clear render state
	d.SetSource("new content")

	if d.OutputPath != "" {
		t.Error("SetSource should clear OutputPath")
	}
	if d.Hash != "" {
		t.Error("SetSource should clear Hash")
	}
	if !d.RenderedAt.IsZero() {
		t.Error("SetSource should clear RenderedAt")
	}
	if d.Error != "" {
		t.Error("SetSource should clear Error")
	}
}
