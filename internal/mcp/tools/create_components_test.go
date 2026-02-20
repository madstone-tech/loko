package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
)

// initTestProjectWithContainer creates a project, system, and container on disk for testing.
func initTestProjectWithContainer(t *testing.T) (projectRoot string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectRoot = filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	repo := filesystem.NewProjectRepository()
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

	container, err := entities.NewContainer("API Server")
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	container.Path = filepath.Join(system.Path, container.ID)
	if err := system.AddContainer(container); err != nil {
		t.Fatalf("AddContainer: %v", err)
	}
	if err := repo.SaveContainer(ctx, projectRoot, system.ID, container); err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	return projectRoot
}

// TestCreateComponentsTool_AllSucceed validates happy-path batch creation.
func TestCreateComponentsTool_AllSucceed(t *testing.T) {
	projectRoot := initTestProjectWithContainer(t)
	repo := filesystem.NewProjectRepository()
	tool := NewCreateComponentsTool(repo)

	args := map[string]any{
		"project_root":   projectRoot,
		"system_name":    "Payment Service",
		"container_name": "API Server",
		"components": []any{
			map[string]any{"name": "Auth Handler", "description": "Handles JWT auth"},
			map[string]any{"name": "Rate Limiter", "technology": "Go"},
			map[string]any{"name": "Request Validator"},
		},
	}

	result, err := tool.Call(context.Background(), args)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	resp, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}

	created, _ := resp["created"].(int)
	failed, _ := resp["failed"].(int)

	if created != 3 {
		t.Errorf("expected created=3, got %d", created)
	}
	if failed != 0 {
		t.Errorf("expected failed=0, got %d", failed)
	}

	results, ok := resp["results"].([]map[string]any)
	if !ok {
		t.Fatalf("expected results to be []map[string]any, got %T", resp["results"])
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Each result must have id field (FR-015)
	for i, item := range results {
		if _, ok := item["id"]; !ok {
			t.Errorf("result[%d] missing 'id' field (FR-015 violation)", i)
		}
		if status, _ := item["status"].(string); status != "created" {
			t.Errorf("result[%d] expected status='created', got %q", i, status)
		}
	}
}

// TestCreateComponentsTool_PartialFailureContinuesBatch validates that per-item
// errors do not abort remaining components.
func TestCreateComponentsTool_PartialFailureContinuesBatch(t *testing.T) {
	projectRoot := initTestProjectWithContainer(t)
	repo := filesystem.NewProjectRepository()
	tool := NewCreateComponentsTool(repo)

	args := map[string]any{
		"project_root":   projectRoot,
		"system_name":    "Payment Service",
		"container_name": "API Server",
		"components": []any{
			map[string]any{"name": "Good Component"},
			map[string]any{}, // missing name — per-item error
			map[string]any{"name": "Another Good Component"},
		},
	}

	result, err := tool.Call(context.Background(), args)
	if err != nil {
		t.Fatalf("Call() should not return top-level error on per-item failure, got %v", err)
	}

	resp := result.(map[string]any)
	created, _ := resp["created"].(int)
	failed, _ := resp["failed"].(int)

	if created != 2 {
		t.Errorf("expected created=2, got %d", created)
	}
	if failed != 1 {
		t.Errorf("expected failed=1, got %d", failed)
	}

	results := resp["results"].([]map[string]any)
	if len(results) != 3 {
		t.Fatalf("expected 3 results (all processed), got %d", len(results))
	}

	// Middle item should be error, not abort
	if status, _ := results[1]["status"].(string); status != "error" {
		t.Errorf("expected results[1].status='error', got %q", status)
	}
}

// TestCreateComponentsTool_EmptyArrayReturnsError validates the empty-array guard.
func TestCreateComponentsTool_EmptyArrayReturnsError(t *testing.T) {
	projectRoot := initTestProjectWithContainer(t)
	repo := filesystem.NewProjectRepository()
	tool := NewCreateComponentsTool(repo)

	args := map[string]any{
		"project_root":   projectRoot,
		"system_name":    "Payment Service",
		"container_name": "API Server",
		"components":     []any{},
	}

	_, err := tool.Call(context.Background(), args)
	if err == nil {
		t.Error("expected error for empty components array")
	}
}

// TestCreateComponentsTool_MissingSystemNameReturnsError validates required field.
func TestCreateComponentsTool_MissingSystemNameReturnsError(t *testing.T) {
	projectRoot := initTestProjectWithContainer(t)
	repo := filesystem.NewProjectRepository()
	tool := NewCreateComponentsTool(repo)

	args := map[string]any{
		"project_root":   projectRoot,
		"container_name": "API Server",
		"components":     []any{map[string]any{"name": "Component A"}},
	}

	_, err := tool.Call(context.Background(), args)
	if err == nil {
		t.Error("expected error when system_name is missing")
	}
}

// TestCreateComponentsTool_MissingContainerNameReturnsError validates required field.
func TestCreateComponentsTool_MissingContainerNameReturnsError(t *testing.T) {
	projectRoot := initTestProjectWithContainer(t)
	repo := filesystem.NewProjectRepository()
	tool := NewCreateComponentsTool(repo)

	args := map[string]any{
		"project_root": projectRoot,
		"system_name":  "Payment Service",
		"components":   []any{map[string]any{"name": "Component A"}},
	}

	_, err := tool.Call(context.Background(), args)
	if err == nil {
		t.Error("expected error when container_name is missing")
	}
}
