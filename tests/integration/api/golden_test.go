// Package api_golden contains HTTP API golden-fixture regression tests.
//
// Golden files live at tests/golden/api/<route-slug>.golden.json. Each file is
// a JSON object with keys: method, path, status, headers (selected), body
// (normalised JSON).
//
// Regenerate all goldens:
//
//	go test ./tests/integration/api/ -update
//
// Assert existing goldens (default):
//
//	go test ./tests/integration/api/
package api_golden

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/api/handlers"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

var update = flag.Bool("update", false, "regenerate golden files")

// goldenDir is relative to the repo root; resolved at runtime via
// filepath.Join(repoRoot(), ...).
const goldenRelDir = "tests/golden/api"

// repoRoot returns the absolute path to the repository root by walking up from
// the test binary's working directory until go.mod is found.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root (go.mod not found)")
		}
		dir = parent
	}
}

// ----------------------------------------------------------------------------
// Mock repository – mirrors createTestProject() from handlers_test.go exactly.
// ----------------------------------------------------------------------------

type mockProjectRepository struct {
	project *entities.Project
	systems []*entities.System
}

func (m *mockProjectRepository) LoadProject(_ context.Context, _ string) (*entities.Project, error) {
	return m.project, nil
}
func (m *mockProjectRepository) SaveProject(_ context.Context, _ *entities.Project) error {
	return nil
}
func (m *mockProjectRepository) ListSystems(_ context.Context, _ string) ([]*entities.System, error) {
	return m.systems, nil
}
func (m *mockProjectRepository) LoadSystem(_ context.Context, _, systemID string) (*entities.System, error) {
	for _, s := range m.systems {
		if s.ID == systemID {
			return s, nil
		}
	}
	return nil, fmt.Errorf("system not found: %s", systemID)
}
func (m *mockProjectRepository) SaveSystem(_ context.Context, _ string, _ *entities.System) error {
	return nil
}
func (m *mockProjectRepository) LoadContainer(_ context.Context, _, _, _ string) (*entities.Container, error) {
	return nil, nil
}
func (m *mockProjectRepository) SaveContainer(_ context.Context, _, _ string, _ *entities.Container) error {
	return nil
}
func (m *mockProjectRepository) LoadComponent(_ context.Context, _, _, _, _ string) (*entities.Component, error) {
	return nil, nil
}
func (m *mockProjectRepository) SaveComponent(_ context.Context, _, _, _ string, _ *entities.Component) error {
	return nil
}

// Ensure mockProjectRepository satisfies the interface at compile time.
var _ usecases.ProjectRepository = (*mockProjectRepository)(nil)

// createFixtureProject mirrors createTestProject() from handlers_test.go.
func createFixtureProject() (*entities.Project, []*entities.System) {
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

// ----------------------------------------------------------------------------
// Route descriptors
// ----------------------------------------------------------------------------

type routeCase struct {
	// slug is the filename stem used for the golden file.
	slug string
	// method is the HTTP method.
	method string
	// path is the full request path.
	path string
	// pathValue sets a named path parameter via req.SetPathValue; key=value.
	pathValue string
	// body is optional request body.
	body string
	// handler is the function to invoke directly (bypasses mux).
	handler http.HandlerFunc
}

// ----------------------------------------------------------------------------
// Normalisation helpers
// ----------------------------------------------------------------------------

// buildIDRe matches build IDs of the form YYYYMMDD-NNNN.
var buildIDRe = regexp.MustCompile(`\d{8}-\d{4}`)

// normaliseBody unmarshals a JSON body into map[string]any, replaces
// volatile fields with stable placeholders, then re-marshals with sorted
// keys and consistent indentation.
func normaliseBody(raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	// First pass: replace raw timestamp-like strings before unmarshalling.
	raw = buildIDRe.ReplaceAllString(raw, "<BUILD_ID>")

	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		// Not JSON – return as-is string.
		return raw, nil
	}

	normaliseValue(v)
	return v, nil
}

// normaliseValue walks the decoded JSON tree and replaces volatile fields.
func normaliseValue(v any) {
	switch node := v.(type) {
	case map[string]any:
		for k, val := range node {
			switch strings.ToLower(k) {
			case "uptime", "timestamp", "duration", "duration_ms", "start_time", "end_time":
				node[k] = "<VOLATILE>"
			case "build_id":
				if s, ok := val.(string); ok {
					node[k] = buildIDRe.ReplaceAllString(s, "<BUILD_ID>")
				}
			default:
				normaliseValue(val)
			}
		}
	case []any:
		for _, item := range node {
			normaliseValue(item)
		}
	}
}

// ----------------------------------------------------------------------------
// Golden file structure
// ----------------------------------------------------------------------------

type goldenRecord struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

// capturedHeaders lists the response headers we record (all others are
// ignored to avoid volatility from Date, X-Request-Id, etc.).
var capturedHeaders = []string{
	"Content-Type",
}

// ----------------------------------------------------------------------------
// Test
// ----------------------------------------------------------------------------

func TestAPIGolden(t *testing.T) {
	root := repoRoot(t)
	goldenDir := filepath.Join(root, goldenRelDir)

	project, systems := createFixtureProject()
	repo := &mockProjectRepository{project: project, systems: systems}
	h := handlers.NewHandlers(".", repo)

	// NOTE: GET /health is intentionally excluded.
	// handleHealth is a method on the unexported *api.Server type, which is
	// not accessible from this package. Constructing a full api.Server and
	// binding it to a real listener would introduce non-determinism (port
	// allocation, goroutines). The /health route is covered by the unit test
	// in package api. The routes below cover all /api/v1/* handlers, which
	// are the FR-008-relevant entity routes per T063.
	cases := []routeCase{
		{
			slug:    "get-project",
			method:  http.MethodGet,
			path:    "/api/v1/project",
			handler: h.GetProject,
		},
		{
			slug:    "list-systems",
			method:  http.MethodGet,
			path:    "/api/v1/systems",
			handler: h.ListSystems,
		},
		{
			slug:      "get-system",
			method:    http.MethodGet,
			path:      "/api/v1/systems/authservice",
			pathValue: "id=authservice",
			handler:   h.GetSystem,
		},
		{
			slug:    "post-build",
			method:  http.MethodPost,
			path:    "/api/v1/build",
			body:    `{"format":"html","output_dir":"dist"}`,
			handler: h.TriggerBuild,
		},
		{
			slug:    "get-validate",
			method:  http.MethodGet,
			path:    "/api/v1/validate",
			handler: h.Validate,
		},
	}

	// NOTE: GET /api/v1/build/{id} is covered separately below with a
	// special fixture that first triggers a build and then queries the
	// resulting ID, so both sides are deterministic within one test run.

	for _, tc := range cases {
		tc := tc
		t.Run(tc.slug, func(t *testing.T) {
			runGoldenCase(t, goldenDir, tc)
		})
	}

	// Special case: GET /api/v1/build/{id} — trigger a build first, capture
	// the returned ID, then query it immediately (build may still be
	// in-flight, but status + success are stable at query time).
	t.Run("get-build-status", func(t *testing.T) {
		runBuildStatusGolden(t, goldenDir, h)
	})
}

func runGoldenCase(t *testing.T, goldenDir string, tc routeCase) {
	t.Helper()

	var bodyReader *strings.Reader
	if tc.body != "" {
		bodyReader = strings.NewReader(tc.body)
	} else {
		bodyReader = strings.NewReader("")
	}

	req := httptest.NewRequest(tc.method, tc.path, bodyReader)
	if tc.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set named path values for routes like /systems/{id}.
	if tc.pathValue != "" {
		kv := strings.SplitN(tc.pathValue, "=", 2)
		if len(kv) == 2 {
			req.SetPathValue(kv[0], kv[1])
		}
	}

	w := httptest.NewRecorder()
	tc.handler(w, req)

	rec := buildRecord(t, tc.method, tc.path, w)
	assertOrUpdateGolden(t, goldenDir, tc.slug, rec)
}

func runBuildStatusGolden(t *testing.T, goldenDir string, h *handlers.Handlers) {
	t.Helper()

	// Step 1: trigger a build and capture the build ID.
	triggerReq := httptest.NewRequest(http.MethodPost, "/api/v1/build",
		strings.NewReader(`{"format":"html","output_dir":"dist"}`))
	triggerReq.Header.Set("Content-Type", "application/json")
	triggerW := httptest.NewRecorder()
	h.TriggerBuild(triggerW, triggerReq)

	var triggerResp map[string]any
	if err := json.Unmarshal(triggerW.Body.Bytes(), &triggerResp); err != nil {
		t.Fatalf("failed to decode TriggerBuild response: %v", err)
	}
	buildID, ok := triggerResp["build_id"].(string)
	if !ok || buildID == "" {
		t.Fatal("TriggerBuild response missing build_id")
	}

	// Step 2: query build status.
	statusReq := httptest.NewRequest(http.MethodGet,
		"/api/v1/build/"+buildID, nil)
	statusReq.SetPathValue("id", buildID)
	statusW := httptest.NewRecorder()
	h.GetBuildStatus(statusW, statusReq)

	rec := buildRecord(t, http.MethodGet, "/api/v1/build/<BUILD_ID>", statusW)
	assertOrUpdateGolden(t, goldenDir, "get-build-status", rec)
}

// buildRecord constructs a goldenRecord from a recorder response.
func buildRecord(t *testing.T, method, path string, w *httptest.ResponseRecorder) goldenRecord {
	t.Helper()

	headers := make(map[string]string)
	for _, h := range capturedHeaders {
		if v := w.Result().Header.Get(h); v != "" {
			headers[h] = v
		}
	}

	body, err := normaliseBody(w.Body.String())
	if err != nil {
		t.Fatalf("normaliseBody: %v", err)
	}

	return goldenRecord{
		Method:  method,
		Path:    path,
		Status:  w.Code,
		Headers: headers,
		Body:    body,
	}
}

// assertOrUpdateGolden either writes the golden file (when -update is set)
// or reads it and compares byte-for-byte after re-serialising the current
// response through the same marshal path.
func assertOrUpdateGolden(t *testing.T, goldenDir, slug string, rec goldenRecord) {
	t.Helper()

	goldenPath := filepath.Join(goldenDir, slug+".golden.json")

	// Serialise the current record deterministically.
	current, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatalf("marshal current record: %v", err)
	}
	current = append(current, '\n') // trailing newline for VCS friendliness

	if *update {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, current, 0644); err != nil {
			t.Fatalf("write golden %s: %v", goldenPath, err)
		}
		t.Logf("updated golden: %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v\n  hint: run with -update to generate it", goldenPath, err)
	}

	if string(current) != string(want) {
		t.Errorf("golden mismatch for %s:\n--- want ---\n%s\n--- got ---\n%s\n--- diff ---\n%s",
			slug, want, current, diffLines(string(want), string(current)))
	}
}

// diffLines produces a simple unified-style diff between two multi-line strings.
func diffLines(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	var sb strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}
	for i := 0; i < max; i++ {
		var wl, gl string
		if i < len(wantLines) {
			wl = wantLines[i]
		}
		if i < len(gotLines) {
			gl = gotLines[i]
		}
		if wl != gl {
			fmt.Fprintf(&sb, "line %d:\n  want: %q\n   got: %q\n", i+1, wl, gl)
		}
	}
	return sb.String()
}
