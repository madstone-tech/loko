package usecases

import (
	"context"
	"testing"
)

// TestNewUpdateDiagram tests creating an UpdateDiagram use case.
func TestNewUpdateDiagram(t *testing.T) {
	uc := NewUpdateDiagram()

	if uc == nil {
		t.Error("NewUpdateDiagram() returned nil")
	}
}

// TestUpdateDiagramExecute tests the Execute method.
func TestUpdateDiagramExecute(t *testing.T) {
	uc := NewUpdateDiagram()
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		request *UpdateDiagramRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: &UpdateDiagramRequest{
				ProjectRoot: tempDir,
				DiagramPath: "test.d2",
				D2Source:    "test -> diagram",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		{
			name: "empty D2 source",
			request: &UpdateDiagramRequest{
				ProjectRoot: tempDir,
				DiagramPath: "test.d2",
				D2Source:    "",
			},
			wantErr: true,
			errMsg:  "D2 source code cannot be empty",
		},
		{
			name: "empty D2 source with whitespace",
			request: &UpdateDiagramRequest{
				ProjectRoot: tempDir,
				DiagramPath: "test.d2",
				D2Source:    "   \n\t  ",
			},
			wantErr: true,
			errMsg:  "D2 source code cannot be empty",
		},
		{
			name: "invalid extension",
			request: &UpdateDiagramRequest{
				ProjectRoot: tempDir,
				DiagramPath: "test.txt",
				D2Source:    "test -> diagram",
			},
			wantErr: true,
		},
		{
			name: "absolute path",
			request: &UpdateDiagramRequest{
				ProjectRoot: tempDir,
				DiagramPath: "/absolute/path/test.d2",
				D2Source:    "test -> diagram",
			},
			wantErr: true,
		},
		{
			name: "valid with subdirectory",
			request: &UpdateDiagramRequest{
				ProjectRoot: tempDir,
				DiagramPath: "sub/dir/test.d2",
				D2Source:    "test -> diagram",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uc.Execute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errMsg != "" && err != nil && err.Error() != tt.errMsg {
					t.Errorf("Execute() error message = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}

			if result == nil {
				t.Fatal("Execute() returned nil result")
			}

			if result.FilePath == "" {
				t.Error("Execute() returned empty FilePath")
			}
		})
	}
}

// TestUpdateDiagramRequestValidation tests various request validation scenarios.
func TestUpdateDiagramRequestValidation(t *testing.T) {
	uc := NewUpdateDiagram()
	tempDir := t.TempDir()

	// Test valid request
	req := &UpdateDiagramRequest{
		ProjectRoot: tempDir,
		DiagramPath: "valid.d2",
		D2Source:    "x -> y: calls",
	}

	result, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.FilePath == "" {
		t.Error("Expected non-empty FilePath")
	}
}
