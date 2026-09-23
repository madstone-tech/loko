package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mock DiagramGenerator
// ─────────────────────────────────────────────────────────────────────────────

type mockDiagramGenerator struct {
	failSystem    bool
	failContainer bool
	failComponent bool
}

func (m *mockDiagramGenerator) GenerateSystemContextDiagram(_ *entities.System) (string, error) {
	if m.failSystem {
		return "", os.ErrPermission
	}
	return "# mock system diagram", nil
}

func (m *mockDiagramGenerator) GenerateContainerDiagram(_ *entities.System) (string, error) {
	if m.failContainer {
		return "", os.ErrPermission
	}
	return "# mock container diagram", nil
}

func (m *mockDiagramGenerator) GenerateComponentDiagram(_ *entities.Container) (string, error) {
	if m.failComponent {
		return "", os.ErrPermission
	}
	return "# mock component diagram", nil
}

var _ usecases.DiagramGenerator = (*mockDiagramGenerator)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────────────────────────────────────

// initTestProject creates a minimal project with one system on disk.
func initTestProject(t *testing.T) (projectRoot string, repo usecases.ProjectRepository) {
	t.Helper()

	tmpDir := t.TempDir()
	projectRoot = filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	repo = filesystem.NewProjectRepository()
	ctx := context.Background()

	project, err := entities.NewProject("myproject")
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	project.Path = projectRoot

	if err := repo.SaveProject(ctx, project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	system, err := entities.NewSystem("Payment Service")
	if err != nil {
		t.Fatalf("NewSystem: %v", err)
	}
	system.Path = filepath.Join(projectRoot, "src", system.ID)

	if err := project.AddSystem(system); err != nil {
		t.Fatalf("AddSystem: %v", err)
	}

	if err := repo.SaveSystem(ctx, projectRoot, system); err != nil {
		t.Fatalf("SaveSystem: %v", err)
	}

	return projectRoot, repo
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

// TestCreateContainerTool_DiagramPathInResponse validates FR-015 and US2:
// the response diagram field must be a file path, not the old fallback message.
func TestCreateContainerTool_DiagramPathInResponse(t *testing.T) {
	projectRoot, repo := initTestProject(t)

	gen := &mockDiagramGenerator{}
	tool := NewCreateContainerTool(repo, gen)

	args := map[string]any{
		"project_root": projectRoot,
		"system_name":  "Payment Service",
		"name":         "API Server",
		"description":  "Handles REST requests",
		"technology":   "Go + Fiber",
	}

	result, err := tool.Call(context.Background(), args)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	resp, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any response, got %T", result)
	}

	container, ok := resp["container"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'container' key in response, got %v", resp)
	}

	// FR-015: id field must be present
	if _, ok := container["id"]; !ok {
		t.Error("response missing 'id' field (FR-015 violation)")
	}

	// US2: diagram field must be a file path, not the old fallback message
	diagram, ok := container["diagram"].(string)
	if !ok {
		t.Fatal("response 'diagram' field is missing or not a string")
	}

	if strings.Contains(diagram, "update_diagram") {
		t.Errorf("diagram field contains old fallback message %q — expected a file path", diagram)
	}

	if !strings.HasSuffix(diagram, ".d2") {
		t.Errorf("diagram field %q does not end with .d2 — expected a path to a D2 file", diagram)
	}
}

// TestCreateContainerTool_IDAlwaysPresent validates FR-015:
// id must be present even when diagram generation is skipped.
func TestCreateContainerTool_IDAlwaysPresent(t *testing.T) {
	projectRoot, repo := initTestProject(t)

	// No diagram generator (nil) — simulates old-style constructor.
	tool := NewCreateContainerTool(repo, nil)

	args := map[string]any{
		"project_root": projectRoot,
		"system_name":  "Payment Service",
		"name":         "DB Proxy",
	}

	result, err := tool.Call(context.Background(), args)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	resp := result.(map[string]any)
	container := resp["container"].(map[string]any)

	if id, ok := container["id"]; !ok || id == "" {
		t.Error("response 'id' field must be non-empty (FR-015 violation)")
	}
}

// TestCreateContainerTool_MissingSystemName validates required field error.
func TestCreateContainerTool_MissingSystemName(t *testing.T) {
	projectRoot, repo := initTestProject(t)
	tool := NewCreateContainerTool(repo, &mockDiagramGenerator{})

	args := map[string]any{
		"project_root": projectRoot,
		"name":         "API Server",
	}

	_, err := tool.Call(context.Background(), args)
	if err == nil {
		t.Error("expected error when system_name is missing")
	}
}

// TestCreateContainerTool_MissingName validates required field error.
func TestCreateContainerTool_MissingName(t *testing.T) {
	projectRoot, repo := initTestProject(t)
	tool := NewCreateContainerTool(repo, &mockDiagramGenerator{})

	args := map[string]any{
		"project_root": projectRoot,
		"system_name":  "Payment Service",
	}

	_, err := tool.Call(context.Background(), args)
	if err == nil {
		t.Error("expected error when name is missing")
	}
}
