package markdown

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestMarkdownSteps: a dynamic view's markdown page lists its steps (feature
// 019, FR-011).
func TestMarkdownSteps(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "projects", "dynamic-views", "expected", "projection.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p vm.Projection
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	as, err := New().Render(context.Background(), &p, usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	pages := map[string]string{}
	for _, a := range as {
		pages[a.Path] = string(a.Bytes)
	}
	page := pages["md/view/checkout.md"]
	for _, want := range []string{"## Steps", "| 1 | [Customer](../element/person/customer.md) | [Web app](../element/container/web.md) |",
		"| 3 | [Order API](../element/container/api.md) | [Payment provider](../element/external/payments.md) | Charge card |"} {
		if !strings.Contains(page, want) {
			t.Errorf("checkout.md lacks %q\n%s", want, page)
		}
	}
	if !strings.Contains(pages["md/view/checkout_sequence.md"], "Payment confirmed (reply)") {
		t.Error("the reply step is not marked")
	}
	golden(t, filepath.Join("testdata", "golden", "dynamic-views", "checkout.md"), []byte(page))
}
