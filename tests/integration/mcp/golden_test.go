// Package mcpgolden holds MCP smoke-fixture regression tests for feature
// 010-constitution-compliance (T005 capture + T030 replay).
//
// Each covered read-only MCP tool has a request/response pair under
// tests/golden/mcp/<tool>.{request,response}.json. The test drives the real
// mcp.Server over its exported Run loop (the same stdio JSON-RPC path used in
// production), backed by deterministic in-memory repositories, and asserts the
// (normalised) response is byte-equivalent to the golden.
//
// A failure means a refactor changed an MCP tool's observable response — fix the
// refactor, not the golden. To re-baseline after a deliberate, reviewed change:
//
//	go test ./tests/integration/mcp/ -update
//
// Only read tools with deterministic output are covered (no timestamps in their
// responses). Mutating tools and tools requiring a live diagram renderer are out
// of scope for this smoke set; see the gaps note at the bottom of this file.
package mcpgolden

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	adaptersenc "github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

var update = flag.Bool("update", false, "regenerate golden files instead of asserting against them")

const goldenRelDir = "tests/golden/mcp"

// volatileKeys are response object keys whose values are non-deterministic. None
// of the covered read tools currently emit them, but the scrubber is kept so the
// fixtures stay stable if a field like this is added later.
var volatileKeys = map[string]string{
	"timestamp":    "<TS>",
	"generated_at": "<TS>",
	"duration_ms":  "<DURATION>",
	"took_ms":      "<DURATION>",
	"elapsed":      "<DURATION>",
}

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
			t.Fatalf("repo root (go.mod) not found")
		}
		dir = parent
	}
}

// --- deterministic in-memory repositories ---

type mockProjectRepo struct {
	project *entities.Project
	systems []*entities.System
}

func (m *mockProjectRepo) LoadProject(_ context.Context, _ string) (*entities.Project, error) {
	return m.project, nil
}
func (m *mockProjectRepo) SaveProject(_ context.Context, _ *entities.Project) error { return nil }
func (m *mockProjectRepo) ListSystems(_ context.Context, _ string) ([]*entities.System, error) {
	return m.systems, nil
}
func (m *mockProjectRepo) LoadSystem(_ context.Context, _, systemID string) (*entities.System, error) {
	for _, s := range m.systems {
		if s.ID == systemID {
			return s, nil
		}
	}
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

var _ usecases.ProjectRepository = (*mockProjectRepo)(nil)

type mockRelRepo struct{}

func (mockRelRepo) LoadRelationships(_ context.Context, _, _ string) ([]entities.Relationship, error) {
	return []entities.Relationship{}, nil
}
func (mockRelRepo) SaveRelationships(_ context.Context, _, _ string, _ []entities.Relationship) error {
	return nil
}
func (mockRelRepo) DeleteElement(_ context.Context, _, _, _ string) error { return nil }

var _ usecases.RelationshipRepository = (*mockRelRepo)(nil)

func newTestRepos() (*mockProjectRepo, *mockRelRepo) {
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

	return &mockProjectRepo{project: project, systems: []*entities.System{sys1, sys2}}, &mockRelRepo{}
}

// --- normalisation ---

func scrub(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if ph, ok := volatileKeys[k]; ok {
				t[k] = ph
				continue
			}
			t[k] = scrub(val)
		}
		return t
	case []any:
		for i, e := range t {
			t[i] = scrub(e)
		}
		return t
	default:
		return v
	}
}

// normaliseResponse parses a JSON-RPC response, and where the result is the MCP
// content wrapper ({"content":[{"type":"text","text":"<json>"}]}), parses the
// inner text payload so the golden stores structured, diffable data rather than
// an opaque escaped string. Volatile keys are scrubbed throughout.
func normaliseResponse(raw []byte) (json.RawMessage, error) {
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if result, ok := resp["result"].(map[string]any); ok {
		if content, ok := result["content"].([]any); ok {
			for _, item := range content {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				if text, ok := m["text"].(string); ok {
					var inner any
					if json.Unmarshal([]byte(text), &inner) == nil {
						m["text"] = scrub(inner)
					}
				}
			}
		}
	}
	scrub(resp)
	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// --- cases ---

type mcpCase struct {
	tool string
	args map[string]any
}

func cases() []mcpCase {
	// project_root is "." for every case; the in-memory repos ignore the value
	// (always returning the fixture) but the tools' request validation requires a
	// non-empty project_root, so passing it yields representative success
	// responses rather than validation errors.
	return []mcpCase{
		{tool: "query_project", args: map[string]any{"project_root": ".", "format": "json"}},
		{tool: "query_architecture", args: map[string]any{"project_root": ".", "format": "json"}},
		{tool: "search_elements", args: map[string]any{"project_root": ".", "query": "*", "format": "json"}},
		{tool: "find_relationships", args: map[string]any{"project_root": ".", "source_pattern": "*"}},
		{tool: "list_relationships", args: map[string]any{"project_root": ".", "system_name": "authservice", "format": "json"}},
	}
}

func request(id int, tool string, args map[string]any) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  "tools/call",
		"params":  map[string]any{"name": tool, "arguments": args},
	}
}

func TestMCPGolden(t *testing.T) {
	goldenDir := filepath.Join(repoRoot(t), goldenRelDir)
	tcs := cases()

	// Feed all requests through one Run pass; responses come back in order.
	var in bytes.Buffer
	enc := json.NewEncoder(&in)
	for i, tc := range tcs {
		if err := enc.Encode(request(i+1, tc.tool, tc.args)); err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}

	var out bytes.Buffer
	srv := mcp.NewServer(".", &in, &out)
	repo, relRepo := newTestRepos()
	encoder := adaptersenc.NewEncoder()
	for _, tl := range []mcp.Tool{
		tools.NewQueryProjectTool(repo, encoder),
		tools.NewQueryArchitectureTool(repo, encoder),
		tools.NewSearchElementsTool(repo, encoder),
		tools.NewFindRelationshipsTool(repo),
		tools.NewListRelationshipsTool(relRepo, repo, encoder),
	} {
		if err := srv.RegisterTool(tl); err != nil {
			t.Fatalf("register %s: %v", tl.Name(), err)
		}
	}
	if err := srv.Run(context.Background()); err != nil {
		t.Fatalf("server Run: %v", err)
	}

	// Split the output buffer into successive JSON-RPC responses (Run writes one
	// encoded object per request, in request order).
	dec := json.NewDecoder(&out)
	for _, tc := range tcs {
		var resp json.RawMessage
		if err := dec.Decode(&resp); err != nil {
			t.Fatalf("decode response for %s: %v", tc.tool, err)
		}
		t.Run(tc.tool, func(t *testing.T) {
			norm, err := normaliseResponse(resp)
			if err != nil {
				t.Fatalf("normalise response: %v", err)
			}

			respPath := filepath.Join(goldenDir, tc.tool+".response.json")
			reqPath := filepath.Join(goldenDir, tc.tool+".request.json")

			if *update {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				reqBytes, _ := json.MarshalIndent(request(0, tc.tool, tc.args), "", "  ")
				if err := os.WriteFile(reqPath, append(reqBytes, '\n'), 0o644); err != nil {
					t.Fatalf("write request golden: %v", err)
				}
				if err := os.WriteFile(respPath, norm, 0o644); err != nil {
					t.Fatalf("write response golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(respPath)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with -update to create it)", respPath, err)
			}
			if !bytes.Equal(bytes.TrimRight(want, "\n"), bytes.TrimRight(norm, "\n")) {
				t.Errorf("MCP response for %s diverged from golden.\n--- want ---\n%s\n--- got ---\n%s",
					tc.tool, want, norm)
			}
		})
	}
}

// Coverage gaps (honest record):
//   - Mutating tools (create_*, update_*, delete_*, build_docs) are excluded: their
//     side effects and/or generated IDs make byte-stable goldens fragile here.
//   - validate / query_dependencies / query_related_components / analyze_coupling
//     are exercised indirectly via unit tests in internal/mcp/tools and
//     internal/core/usecases; they are not in this smoke set.
//   - validate_diagram requires a live d2 DiagramRenderer and is covered by the
//     d2 adapter tests instead.
