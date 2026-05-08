package usecases

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// initProjectRepoMock captures SaveProject calls for InitProject tests.
type initProjectRepoMock struct {
	saved   *entities.Project
	saveErr error
}

func (m *initProjectRepoMock) SaveProject(_ context.Context, project *entities.Project) error {
	m.saved = project
	return m.saveErr
}

func (m *initProjectRepoMock) LoadProject(_ context.Context, _ string) (*entities.Project, error) {
	return nil, nil
}

func (m *initProjectRepoMock) ListSystems(_ context.Context, _ string) ([]*entities.System, error) {
	return nil, nil
}

func (m *initProjectRepoMock) SaveSystem(_ context.Context, _ string, _ *entities.System) error {
	return nil
}

func (m *initProjectRepoMock) LoadSystem(_ context.Context, _, _ string) (*entities.System, error) {
	return nil, nil
}

func (m *initProjectRepoMock) DeleteSystem(_ context.Context, _, _ string) error {
	return nil
}

func (m *initProjectRepoMock) SaveContainer(_ context.Context, _, _ string, _ *entities.Container) error {
	return nil
}

func (m *initProjectRepoMock) LoadContainer(_ context.Context, _, _, _ string) (*entities.Container, error) {
	return nil, nil
}

func (m *initProjectRepoMock) DeleteContainer(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *initProjectRepoMock) SaveComponent(_ context.Context, _, _, _ string, _ *entities.Component) error {
	return nil
}

func (m *initProjectRepoMock) LoadComponent(_ context.Context, _, _, _, _ string) (*entities.Component, error) {
	return nil, nil
}

func (m *initProjectRepoMock) DeleteComponent(_ context.Context, _, _, _, _ string) error {
	return nil
}

func TestInitProjectExecuteRejectsNilRequest(t *testing.T) {
	uc := NewInitProject(&initProjectRepoMock{})
	if err := uc.Execute(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

func TestInitProjectExecuteRejectsEmptyName(t *testing.T) {
	uc := NewInitProject(&initProjectRepoMock{})
	err := uc.Execute(context.Background(), &InitProjectRequest{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "project name is required") {
		t.Fatalf("error message %q does not mention required name", err)
	}
}

func TestInitProjectExecuteRejectsInvalidName(t *testing.T) {
	uc := NewInitProject(&initProjectRepoMock{})
	err := uc.Execute(context.Background(), &InitProjectRequest{Name: "Has Spaces!"})
	if err == nil {
		t.Fatal("expected validation error for invalid name")
	}
}

func TestInitProjectExecuteCreatesProject(t *testing.T) {
	tmp := t.TempDir()
	repo := &initProjectRepoMock{}
	uc := NewInitProject(repo)

	projectPath := filepath.Join(tmp, "myproj")
	err := uc.Execute(context.Background(), &InitProjectRequest{
		Name:        "myproj",
		Path:        projectPath,
		Description: "demo project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected SaveProject to be called")
	}
	if repo.saved.Name != "myproj" {
		t.Errorf("project name = %q, want %q", repo.saved.Name, "myproj")
	}
	abs, _ := filepath.Abs(projectPath)
	if repo.saved.Path != abs {
		t.Errorf("project path = %q, want %q", repo.saved.Path, abs)
	}
	if repo.saved.Description != "demo project" {
		t.Errorf("project description = %q, want %q", repo.saved.Description, "demo project")
	}
	if _, err := os.Stat(projectPath); err != nil {
		t.Errorf("project directory not created: %v", err)
	}
}

func TestInitProjectExecuteDefaultsPathToName(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	repo := &initProjectRepoMock{}
	uc := NewInitProject(repo)
	err := uc.Execute(context.Background(), &InitProjectRequest{Name: "defaultpath"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("SaveProject not called")
	}
	expected, _ := filepath.Abs("defaultpath")
	if repo.saved.Path != expected {
		t.Errorf("default path = %q, want %q", repo.saved.Path, expected)
	}
}
