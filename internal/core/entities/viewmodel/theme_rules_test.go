package viewmodel

import (
	"strings"
	"testing"
)

func TestValidateTheme(t *testing.T) {
	t.Parallel()
	overridable := []string{"layout.gohtml", "style.css"}
	files := []ThemeFile{
		{Name: "sytle.css", Origin: "templates/sytle.css"},
		{Name: "layout.gohtml", Origin: "templates/layout.gohtml"},
		{Name: "extra.js", Origin: "templates/extra.js"},
	}
	got := ValidateTheme(files, overridable)
	if len(got) != 2 || got[0].Origin != "templates/extra.js" || got[1].Origin != "templates/sytle.css" {
		t.Fatalf("ValidateTheme = %+v", got)
	}
	if !strings.Contains(got[1].Message, "layout.gohtml, style.css") {
		t.Errorf("message must list the overridable names: %q", got[1].Message)
	}
	if ValidateTheme(files[1:2], overridable) != nil {
		t.Error("a known file is not a problem")
	}
}
