package html

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func loadFixtureProjection(t *testing.T, name string) *vm.Projection {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "projects", name, "expected", "projection.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p vm.Projection
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	return &p
}

// TestDeploymentNodesTable: a deployment page lists the nodes that say what
// they are, by path; a deployment whose nodes say nothing has no table
// (feature 018, FR-002a).
func TestDeploymentNodesTable(t *testing.T) {
	t.Parallel()
	site := renderSite(t, loadFixtureProjection(t, "deployment-details"), nil)
	prod := site["view/deployment-prod.html"]
	for _, want := range []string{
		"<h2>Nodes</h2>",
		"<tr><td>aws</td><td>AWS account 1234</td><td>AWS</td><td></td><td></td></tr>",
		"<tr><td>aws / eu_west_1 / vpc</td><td>Production VPC</td><td>Amazon VPC</td><td>Private subnets in eu-west-1a and eu-west-1b</td><td>network, private</td></tr>",
	} {
		if !strings.Contains(prod, want) {
			t.Errorf("deployment-prod page lacks %q", want)
		}
	}
	if strings.Contains(prod, "<td>aws / eu_west_1</td>") {
		t.Error("a node without attributes is listed")
	}
	if dev := site["view/deployment-dev.html"]; strings.Contains(dev, "<h2>Nodes</h2>") {
		t.Error("a deployment whose nodes have no attributes has a Nodes table")
	}
	for _, page := range []string{"view/deployment-prod.html", "view/deployment-dev.html"} {
		golden(t, filepath.Join("testdata", "golden", "deployment-details", filepath.FromSlash(page)), []byte(site[page]))
	}
}

// TestViewTitleAndDescription: a titled view's page heading, browser title and
// navigation entry use the title, its description sits under the heading,
// and its address and file name keep the name (feature 018, FR-006).
func TestViewTitleAndDescription(t *testing.T) {
	t.Parallel()
	site := renderSite(t, loadFixtureProjection(t, "deployment-details"), nil)
	page, ok := site["view/request_path.html"]
	if !ok {
		t.Fatal("the view's page is not named after the view")
	}
	for _, want := range []string{
		"<h1>How an order is stored</h1>",
		"<title>How an order is stored · deployment-details</title>",
		`<p class="view-description">The order path only: receipts are left out.</p>`,
		`href="../view/request_path.html">How an order is stored</a>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("view page lacks %q", want)
		}
	}
	if strings.Contains(site["view/landscape.html"], `class="view-description"`) {
		t.Error("a view without a description has one")
	}
	golden(t, filepath.Join("testdata", "golden", "deployment-details", "view", "request_path.html"), []byte(page))
}
