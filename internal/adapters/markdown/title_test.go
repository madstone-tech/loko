package markdown

import (
	"context"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

func TestMarkdownShowsTitlesAndTags(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	for i := range proj.Pages {
		pg := &proj.Pages[i]
		if pg.Address == "container.api" {
			pg.Title = "Orders API"
		}
		for j := range pg.Uses {
			if pg.Uses[j].Other.Address == "container.api" {
				pg.Uses[j].Other.Title = "Orders API"
				pg.Uses[j].Tags = []string{"https"}
			}
		}
	}
	as, err := New().Render(context.Background(), proj, usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	docs := map[string]string{}
	for _, a := range as {
		docs[a.Path] = string(a.Bytes)
	}
	if api := docs["md/element/container/api.md"]; !strings.Contains(api, "\n# Orders API\n\n`api`\n\n") {
		t.Errorf("titled heading:\n%s", api[:min(len(api), 200)])
	}
	web := docs["md/element/container/web.md"]
	if !strings.Contains(web, "[Orders API](api.md)") || !strings.Contains(web, "| Tags |") || !strings.Contains(web, "#https") {
		t.Errorf("web.md uses table:\n%s", web)
	}
}
