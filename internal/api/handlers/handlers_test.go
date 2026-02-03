package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// MockProjectRepository implements usecases.ProjectRepository for testing.
type MockProjectRepository struct {
	project *entities.Project
	systems []*entities.System
}

func (m *MockProjectRepository) LoadProject(ctx context.Context, projectRoot string) (*entities.Project, error) {
	return m.project, nil
}

func (m *MockProjectRepository) SaveProject(ctx context.Context, project *entities.Project) error {
	return nil
}

func (m *MockProjectRepository) ListSystems(ctx context.Context, projectRoot string) ([]*entities.System, error) {
	return m.systems, nil
}

func (m *MockProjectRepository) LoadSystem(ctx context.Context, projectRoot, systemID string) (*entities.System, error) {
	for _, sys := range m.systems {
		if sys.ID == systemID {
			return sys, nil
		}
	}
	return nil, nil
}

func (m *MockProjectRepository) SaveSystem(ctx context.Context, projectRoot string, system *entities.System) error {
	return nil
}

func (m *MockProjectRepository) LoadContainer(ctx context.Context, projectRoot, systemID, containerID string) (*entities.Container, error) {
	return nil, nil
}

func (m *MockProjectRepository) SaveContainer(ctx context.Context, projectRoot, systemID string, container *entities.Container) error {
	return nil
}

func (m *MockProjectRepository) LoadComponent(ctx context.Context, projectRoot, systemID, containerID, componentID string) (*entities.Component, error) {
	return nil, nil
}

func (m *MockProjectRepository) SaveComponent(ctx context.Context, projectRoot, systemID, containerID string, component *entities.Component) error {
	return nil
}

func createTestProject() (*entities.Project, []*entities.System) {
	project, _ := entities.NewProject("TestProject")
	project.Description = "A test project"
	project.Version = "1.0.0"

	sys1, _ := entities.NewSystem("AuthService")
	sys1.Description = "Authentication service"
	cont1, _ := entities.NewContainer("API")
	cont1.Description = "REST API"
	cont1.Technology = "Go"
	sys1.AddContainer(cont1)

	sys2, _ := entities.NewSystem("UserService")
	sys2.Description = "User management"

	return project, []*entities.System{sys1, sys2}
}

func TestGetProject(t *testing.T) {
	project, systems := createTestProject()
	repo := &MockProjectRepository{project: project, systems: systems}
	h := NewHandlers(".", repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/project", nil)
	w := httptest.NewRecorder()

	h.GetProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ProjectResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Name != "TestProject" {
		t.Errorf("expected name=TestProject, got %s", resp.Name)
	}
	if resp.SystemCount != 2 {
		t.Errorf("expected 2 systems, got %d", resp.SystemCount)
	}
}

func TestListSystems(t *testing.T) {
	project, systems := createTestProject()
	repo := &MockProjectRepository{project: project, systems: systems}
	h := NewHandlers(".", repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/systems", nil)
	w := httptest.NewRecorder()

	h.ListSystems(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SystemsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected 2 systems, got %d", resp.TotalCount)
	}
	if resp.ProjectName != "TestProject" {
		t.Errorf("expected project name=TestProject, got %s", resp.ProjectName)
	}
}

func TestGetSystem(t *testing.T) {
	project, systems := createTestProject()
	repo := &MockProjectRepository{project: project, systems: systems}
	h := NewHandlers(".", repo)

	// Create request with path value - ID is normalized (lowercase)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/systems/authservice", nil)
	req.SetPathValue("id", "authservice")
	w := httptest.NewRecorder()

	h.GetSystem(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SystemDetailResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.System.Name != "AuthService" {
		t.Errorf("expected system name=AuthService, got %s", resp.System.Name)
	}
	if len(resp.Containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(resp.Containers))
	}
}

func TestValidate(t *testing.T) {
	project, systems := createTestProject()
	repo := &MockProjectRepository{project: project, systems: systems}
	h := NewHandlers(".", repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/validate", nil)
	w := httptest.NewRecorder()

	h.Validate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ValidateResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected success=true")
	}
	// Should have warnings for empty system and missing descriptions
	if resp.WarningCount < 1 {
		t.Error("expected at least 1 warning")
	}
}

func TestTriggerBuild(t *testing.T) {
	project, systems := createTestProject()
	repo := &MockProjectRepository{project: project, systems: systems}
	h := NewHandlers(".", repo)

	body := strings.NewReader(`{"format":"html","output_dir":"dist"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/build", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.TriggerBuild(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var resp BuildResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Status != "building" {
		t.Errorf("expected status=building, got %s", resp.Status)
	}
	if resp.BuildID == "" {
		t.Error("expected build ID to be set")
	}
}
