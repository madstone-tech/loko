package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
	"github.com/madstone-tech/loko/internal/core/usecases"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

const twoSystemsRoot = "../testdata/projects/two-systems"

func queryFor(t *testing.T, root, format string, req usecases.QueryRequest) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code, err := runQueryWith(t.Context(), QueryOptions{Root: root, Format: format, Request: req, Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatal(err)
	}
	return code, out.String(), errOut.String()
}

// TestQueryMatchesMCP is SC-006: for every query kind, `loko query --format
// json` prints exactly what the MCP query tool returns, plus a final newline.
func TestQueryMatchesMCP(t *testing.T) {
	t.Parallel()
	svc := usecases.NewAuthoringService(newAuthoringDeps(twoSystemsRoot))
	mcpQuery := tools.NewQueryTool(svc, encoding.NewEncoder())
	cases := []usecases.QueryRequest{
		{Kind: "dependents", Address: "container.orders_db"},
		{Kind: "dependents", Address: "container.ledger", Transitive: true},
		{Kind: "dependencies", Address: "container.web", Transitive: true},
		{Kind: "path", Address: "person.customer", To: "external.bank"},
		{Kind: "orphans"},
		{Kind: "coupling", Limit: 5},
	}
	for _, req := range cases {
		code, cli, _ := queryFor(t, twoSystemsRoot, "json", req)
		args := map[string]any{"kind": req.Kind, "format": "json"}
		if req.Address != "" {
			args["address"] = req.Address
		}
		if req.To != "" {
			args["to"] = req.To
		}
		if req.Transitive {
			args["transitive"] = true
		}
		if req.Limit > 0 {
			args["limit"] = float64(req.Limit)
		}
		mcp, err := mcpQuery.Call(t.Context(), args)
		if err != nil {
			t.Fatal(err)
		}
		if code != usecases.ExitSuccess || cli != mcp.(string)+"\n" {
			t.Errorf("%+v: CLI and MCP differ (exit %d)\ncli: %s\nmcp: %s", req, code, cli, mcp)
		}
	}
}

func TestQueryUnknownAddress(t *testing.T) {
	t.Parallel()
	code, out, errOut := queryFor(t, twoSystemsRoot, "text", usecases.QueryRequest{Kind: "dependents", Address: "container.ledgr"})
	if code != usecases.ExitErrors || out != "" ||
		!strings.Contains(errOut, `unknown address "container.ledgr": did you mean container.ledger?`) {
		t.Errorf("exit %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestQueryBrokenProject(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.loko.hcl"), []byte("container \"c\" {\n  system = system.gone\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := queryFor(t, root, "text", usecases.QueryRequest{Kind: "orphans"})
	if code != usecases.ExitErrors || out != "" || !strings.Contains(errOut, "system.gone") {
		t.Errorf("a project that does not compile prints its diagnostics and exits 1: %d %q %q", code, out, errOut)
	}
}

func TestQueryText(t *testing.T) {
	t.Parallel()
	_, deps, _ := queryFor(t, twoSystemsRoot, "text", usecases.QueryRequest{Kind: "dependents", Address: "container.orders_db"})
	if !strings.Contains(deps, "ADDRESS") || !strings.Contains(deps, "component.repo") || !strings.Contains(deps, "container.api") {
		t.Errorf("dependents table:\n%s", deps)
	}
	_, path, _ := queryFor(t, twoSystemsRoot, "text", usecases.QueryRequest{Kind: "path", Address: "person.customer", To: "external.bank"})
	if !strings.Contains(path, "person.customer → container.web") || !strings.Contains(path, "container.gateway.uses.acquire") {
		t.Errorf("path:\n%s", path)
	}
	_, none, _ := queryFor(t, twoSystemsRoot, "text", usecases.QueryRequest{Kind: "path", Address: "external.bank", To: "person.customer"})
	if !strings.Contains(none, "no path") {
		t.Errorf("no path:\n%s", none)
	}
	_, coupling, _ := queryFor(t, twoSystemsRoot, "text", usecases.QueryRequest{Kind: "coupling", Limit: 3})
	if !strings.Contains(coupling, "FAN-IN") || strings.Count(coupling, "\n") != 4 {
		t.Errorf("coupling table (header + 3 rows):\n%s", coupling)
	}
	_, toon, _ := queryFor(t, twoSystemsRoot, "toon", usecases.QueryRequest{Kind: "orphans"})
	if !strings.Contains(toon, "kind: orphans") {
		t.Errorf("toon:\n%s", toon)
	}
}
