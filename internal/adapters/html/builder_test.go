package html

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestNewBuilder tests the NewBuilder factory function.
func TestNewBuilder(t *testing.T) {
	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}
	if builder == nil {
		t.Fatal("expected builder to be non-nil")
	}
	if builder.templates == nil {
		t.Error("expected templates to be initialized")
	}
}

// TestBuildSiteBasic tests basic site building with a simple project.
func TestBuildSiteBasic(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:        "Test Project",
		Description: "A test project",
		Version:     "1.0.0",
		Systems:     make(map[string]*entities.System),
	}

	systems := []*entities.System{
		{
			ID:          "payment-service",
			Name:        "Payment Service",
			Description: "Handles payment processing",
			Tags:        []string{"backend", "critical"},
			Containers:  make(map[string]*entities.Container),
		},
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify output structure
	expectedFiles := []string{
		"index.html",
		"systems/payment-service.html",
		"search.json",
		"styles/style.css",
		"js/main.js",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s not found", file)
		}
	}
}

// TestBuildSiteMultipleSystems tests site building with multiple systems.
func TestBuildSiteMultipleSystems(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Multi-System Project",
		Systems: make(map[string]*entities.System),
	}

	systems := []*entities.System{
		{
			ID:          "auth-service",
			Name:        "Auth Service",
			Description: "Authentication and authorization",
			Containers:  make(map[string]*entities.Container),
		},
		{
			ID:          "payment-service",
			Name:        "Payment Service",
			Description: "Payment processing",
			Containers:  make(map[string]*entities.Container),
		},
		{
			ID:          "notification-service",
			Name:        "Notification Service",
			Description: "Sends notifications",
			Containers:  make(map[string]*entities.Container),
		},
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify all system pages were created
	for _, sys := range systems {
		path := filepath.Join(tmpDir, "systems", sys.ID+".html")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected system page %s not found", sys.ID)
		}
	}
}

// TestBuildSiteWithContainers tests site building with containers.
func TestBuildSiteWithContainers(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Project with Containers",
		Systems: make(map[string]*entities.System),
	}

	containers := make(map[string]*entities.Container)
	containers["api"] = &entities.Container{
		ID:          "api",
		Name:        "API Server",
		Description: "REST API",
		Technology:  "Go + Fiber",
		Tags:        []string{"backend"},
		Components:  make(map[string]*entities.Component),
	}
	containers["db"] = &entities.Container{
		ID:          "db",
		Name:        "Database",
		Description: "PostgreSQL database",
		Technology:  "PostgreSQL 15",
		Tags:        []string{"data"},
		Components:  make(map[string]*entities.Component),
	}

	systems := []*entities.System{
		{
			ID:          "payment-service",
			Name:        "Payment Service",
			Description: "Handles payments",
			Containers:  containers,
		},
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify system page was created
	systemPath := filepath.Join(tmpDir, "systems", "payment-service.html")
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		t.Errorf("expected system page not found")
	}

	// Verify search index includes containers
	searchPath := filepath.Join(tmpDir, "search.json")
	content, err := os.ReadFile(searchPath)
	if err != nil {
		t.Fatalf("failed to read search.json: %v", err)
	}

	if len(content) == 0 {
		t.Error("search.json is empty")
	}
}

// TestBuildSystemPage tests building a single system page.
func TestBuildSystemPage(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	system := &entities.System{
		ID:          "test-system",
		Name:        "Test System",
		Description: "A test system",
		Tags:        []string{"test"},
		Containers:  make(map[string]*entities.Container),
	}

	containers := []*entities.Container{
		{
			ID:          "container1",
			Name:        "Container 1",
			Description: "First container",
			Technology:  "Go",
			Components:  make(map[string]*entities.Component),
		},
	}

	err = builder.BuildSystemPage(ctx, system, containers, tmpDir)
	if err != nil {
		t.Fatalf("BuildSystemPage failed: %v", err)
	}

	// Verify system page was created
	systemPath := filepath.Join(tmpDir, "systems", "test-system.html")
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		t.Errorf("expected system page not found")
	}

	// Verify content
	content, err := os.ReadFile(systemPath)
	if err != nil {
		t.Fatalf("failed to read system page: %v", err)
	}

	if len(content) == 0 {
		t.Error("system page is empty")
	}
}

// TestBuildSiteNilProject tests error handling for nil project.
func TestBuildSiteNilProject(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	err = builder.BuildSite(ctx, nil, []*entities.System{}, tmpDir)
	if err == nil {
		t.Error("expected error for nil project")
	}
}

// TestBuildSiteEmptyOutputDir tests error handling for empty output directory.
func TestBuildSiteEmptyOutputDir(t *testing.T) {
	ctx := context.Background()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Test",
		Systems: make(map[string]*entities.System),
	}

	err = builder.BuildSite(ctx, project, []*entities.System{}, "")
	if err == nil {
		t.Error("expected error for empty output directory")
	}
}

// TestBuildSystemPageNilSystem tests error handling for nil system.
func TestBuildSystemPageNilSystem(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	err = builder.BuildSystemPage(ctx, nil, []*entities.Container{}, tmpDir)
	if err == nil {
		t.Error("expected error for nil system")
	}
}

// TestBuildSiteWithDiagrams tests site building with diagrams.
func TestBuildSiteWithDiagrams(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Project with Diagrams",
		Systems: make(map[string]*entities.System),
	}

	systems := []*entities.System{
		{
			ID:          "system-with-diagram",
			Name:        "System with Diagram",
			Description: "Has a diagram",
			Containers:  make(map[string]*entities.Container),
			Diagram: &entities.Diagram{
				ID:     "system-diagram",
				Source: "shape: rect",
			},
		},
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify system page was created
	systemPath := filepath.Join(tmpDir, "systems", "system-with-diagram.html")
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		t.Errorf("expected system page not found")
	}
}

// TestBuildSiteEmptySystems tests site building with empty systems list.
func TestBuildSiteEmptySystems(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Empty Project",
		Systems: make(map[string]*entities.System),
	}

	err = builder.BuildSite(ctx, project, []*entities.System{}, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify index page was created
	indexPath := filepath.Join(tmpDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("expected index page not found")
	}
}

// TestBuildSiteWithNilSystems tests site building with nil systems in slice.
func TestBuildSiteWithNilSystems(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Project with nil systems",
		Systems: make(map[string]*entities.System),
	}

	systems := []*entities.System{
		{
			ID:         "valid-system",
			Name:       "Valid System",
			Containers: make(map[string]*entities.Container),
		},
		nil,
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify valid system page was created
	systemPath := filepath.Join(tmpDir, "systems", "valid-system.html")
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		t.Errorf("expected system page not found")
	}
}

// TestSearchIndexContent tests the content of the generated search index.
func TestSearchIndexContent(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Search Test",
		Systems: make(map[string]*entities.System),
	}

	containers := make(map[string]*entities.Container)
	containers["api"] = &entities.Container{
		ID:          "api",
		Name:        "API",
		Description: "REST API",
		Components:  make(map[string]*entities.Component),
	}

	systems := []*entities.System{
		{
			ID:          "payment",
			Name:        "Payment Service",
			Description: "Handles payments",
			Containers:  containers,
		},
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Read and verify search index
	searchPath := filepath.Join(tmpDir, "search.json")
	content, err := os.ReadFile(searchPath)
	if err != nil {
		t.Fatalf("failed to read search.json: %v", err)
	}

	// Verify it contains expected entries
	contentStr := string(content)
	if !contains(contentStr, "Payment Service") {
		t.Error("search index missing system name")
	}
	if !contains(contentStr, "API") {
		t.Error("search index missing container name")
	}
}

// TestAssetGeneration tests that CSS and JS assets are generated.
func TestAssetGeneration(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:    "Asset Test",
		Systems: make(map[string]*entities.System),
	}

	err = builder.BuildSite(ctx, project, []*entities.System{}, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify CSS file
	cssPath := filepath.Join(tmpDir, "styles", "style.css")
	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("failed to read CSS: %v", err)
	}
	if len(cssContent) == 0 {
		t.Error("CSS file is empty")
	}

	// Verify JS file
	jsPath := filepath.Join(tmpDir, "js", "main.js")
	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("failed to read JS: %v", err)
	}
	if len(jsContent) == 0 {
		t.Error("JS file is empty")
	}
}

// TestBuildSiteWithProjectMetadata tests site building with project metadata.
func TestBuildSiteWithProjectMetadata(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	project := &entities.Project{
		Name:        "Metadata Project",
		Description: "Project with metadata",
		Version:     "2.0.0",
		Systems:     make(map[string]*entities.System),
		Metadata: map[string]any{
			"author": "Test Author",
			"date":   time.Now(),
		},
	}

	systems := []*entities.System{
		{
			ID:         "test",
			Name:       "Test System",
			Containers: make(map[string]*entities.Container),
		},
	}

	err = builder.BuildSite(ctx, project, systems, tmpDir)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Verify index page was created with metadata
	indexPath := filepath.Join(tmpDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "Metadata Project") {
		t.Error("index missing project name")
	}
	if !contains(contentStr, "2.0.0") {
		t.Error("index missing version")
	}
}

// TestBuildSystemPageWithComponents tests building system page with components.
func TestBuildSystemPageWithComponents(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	builder, err := NewBuilder()
	if err != nil {
		t.Fatalf("NewBuilder failed: %v", err)
	}

	components := make(map[string]*entities.Component)
	components["handler"] = &entities.Component{
		ID:          "handler",
		Name:        "Request Handler",
		Description: "Handles HTTP requests",
	}

	containers := []*entities.Container{
		{
			ID:          "api",
			Name:        "API Server",
			Description: "REST API",
			Technology:  "Go",
			Components:  components,
		},
	}

	system := &entities.System{
		ID:         "payment",
		Name:       "Payment Service",
		Containers: make(map[string]*entities.Container),
	}

	err = builder.BuildSystemPage(ctx, system, containers, tmpDir)
	if err != nil {
		t.Fatalf("BuildSystemPage failed: %v", err)
	}

	// Verify system page was created
	systemPath := filepath.Join(tmpDir, "systems", "payment.html")
	content, err := os.ReadFile(systemPath)
	if err != nil {
		t.Fatalf("failed to read system page: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "Request Handler") {
		t.Error("system page missing component name")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
