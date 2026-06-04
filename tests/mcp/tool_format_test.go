package mcp

import (
	"context"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

// newTestEncoder returns a fresh TOON encoder for tests.
func newTestEncoder() *encoding.Encoder {
	return encoding.NewEncoder()
}

// mockRepo returns a minimal project repo with a single system and container.
func mockRepo() *mockProjectRepo {
	project, _ := entities.NewProject("TestProject")
	project.Description = "A test project"
	project.Version = "1.0.0"

	sys1, _ := entities.NewSystem("AuthService")
	sys1.Description = "Authentication service"
	cont1, _ := entities.NewContainer("API")
	cont1.Description = "REST API"
	cont1.Technology = "Go"
	sys1.AddContainer(cont1)

	return &mockProjectRepo{project: project, systems: []*entities.System{sys1}}
}

// mockProjectRepo is a minimal in-memory repo for tool tests.
type mockProjectRepo struct {
	project *entities.Project
	systems []*entities.System
}

func (m *mockProjectRepo) LoadProject(_ context.Context, _ string) (*entities.Project, error) {
	return m.project, nil
}
func (m *mockProjectRepo) SaveProject(_ context.Context, _ *entities.Project) error {
	return nil
}
func (m *mockProjectRepo) ListSystems(_ context.Context, _ string) ([]*entities.System, error) {
	return m.systems, nil
}
func (m *mockProjectRepo) LoadSystem(_ context.Context, _, _ string) (*entities.System, error) {
	return nil, nil
}
func (m *mockProjectRepo) SaveSystem(_ context.Context, _ string, _ *entities.System) error {
	return nil
}
func (m *mockProjectRepo) LoadContainer(_ context.Context, _, _, _ string) (*entities.Container, error) {
	return nil, nil
}
func (m *mockProjectRepo) SaveContainer(_ context.Context, _, _ string, _ *entities.Container) error {
	return nil
}
func (m *mockProjectRepo) LoadComponent(_ context.Context, _, _, _, _ string) (*entities.Component, error) {
	return nil, nil
}
func (m *mockProjectRepo) SaveComponent(_ context.Context, _, _, _ string, _ *entities.Component) error {
	return nil
}

// assertTOONWrapper checks that the response is a TOON wrapper map.
func assertTOONWrapper(t *testing.T, result any) map[string]any {
	t.Helper()
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["format"] != "toon" {
		t.Fatalf("expected format=toon, got %v", m["format"])
	}
	payload, ok := m["payload"].(string)
	if !ok || payload == "" {
		t.Fatalf("expected non-empty payload, got %v", m["payload"])
	}
	if m["token_estimate"] == nil {
		t.Fatalf("expected token_estimate, got nil")
	}
	return m
}

// assertJSONMap checks that the response is a plain JSON map (not a TOON wrapper).
func assertJSONMap(t *testing.T, result any) map[string]any {
	t.Helper()
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, hasPayload := m["payload"]; hasPayload {
		t.Fatalf("expected plain JSON map, got TOON wrapper")
	}
	return m
}

// assertError checks that the result is an error containing expected text.
func assertError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !containsString(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %q", want, err.Error())
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// US1: Default TOON format for read tools
// ─────────────────────────────────────────────────────────────────────────────

func TestQueryProject_DefaultReturnsTOON(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryProjectTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{"project_root": "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTOONWrapper(t, result)
}

func TestQueryArchitecture_DefaultReturnsTOON(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryArchitectureTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"detail":       "summary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := assertTOONWrapper(t, result)
	if m["detail"] != "summary" {
		t.Fatalf("expected detail=summary, got %v", m["detail"])
	}
}

func TestSearchElements_DefaultReturnsTOON(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewSearchElementsTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"query":        "*",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTOONWrapper(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// US2: JSON escape hatch and error handling
// ─────────────────────────────────────────────────────────────────────────────

func TestQueryProject_JSONReturnsPlainMap(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryProjectTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"format":       "json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := assertJSONMap(t, result)
	if m["project"] == nil {
		t.Fatalf("expected project key in JSON response")
	}
}

func TestQueryProject_ExplicitTOONReturnsWrapper(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryProjectTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"format":       "toon",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTOONWrapper(t, result)
}

func TestQueryProject_InvalidFormatReturnsError(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryProjectTool(repo, newTestEncoder())
	_, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"format":       "xml",
	})
	assertError(t, err, "invalid format")
}

func TestQueryArchitecture_LegacyTextMapsToTOON(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryArchitectureTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"detail":       "summary",
		"format":       "text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTOONWrapper(t, result)
}

func TestQueryArchitecture_LegacyCompactMapsToTOON(t *testing.T) {
	repo := mockRepo()
	tool := tools.NewQueryArchitectureTool(repo, newTestEncoder())
	result, err := tool.Call(context.Background(), map[string]any{
		"project_root": ".",
		"detail":       "summary",
		"format":       "compact",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTOONWrapper(t, result)
}
