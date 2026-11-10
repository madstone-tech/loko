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

// TestValidateThemeAssets: an asset path must be safe, a regular file and
// unique without regard to case. Each failure names the file (contract T-4).
func TestValidateThemeAssets(t *testing.T) {
	t.Parallel()
	asset := func(rel string) ThemeFile { return ThemeFile{Name: "x", Rel: rel, Origin: "templates/" + rel} }
	tests := []struct {
		name  string
		files []ThemeFile
		want  string
	}{
		{"absolute", []ThemeFile{asset("/abs.svg")}, "templates//abs.svg"},
		{"dot dot", []ThemeFile{asset("assets/../x.svg")}, "templates/assets/../x.svg"},
		{"backslash", []ThemeFile{asset(`assets/a\b.svg`)}, `templates/assets/a\b.svg`},
		{"drive letter", []ThemeFile{asset("assets/C:x.svg")}, "templates/assets/C:x.svg"},
		{"control character", []ThemeFile{asset("assets/a\x01.svg")}, "templates/assets/a\x01.svg"},
		{"symbolic link", []ThemeFile{{Name: "l.svg", Rel: "assets/l.svg", Origin: "templates/assets/l.svg", Link: true}},
			"templates/assets/l.svg"},
		{"case collision", []ThemeFile{asset("assets/A.svg"), asset("assets/a.svg")}, "templates/assets/a.svg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateTheme(tt.files, OverridableThemeFiles)
			if len(got) != 1 || got[0].Origin != tt.want {
				t.Fatalf("ValidateTheme = %+v, want one problem naming %q", got, tt.want)
			}
			if strings.Contains(got[0].Message, ";") {
				t.Errorf("message has a semicolon: %q", got[0].Message)
			}
		})
	}
	ok := []ThemeFile{{Name: "fonts.css", Origin: "templates/fonts.css"}, asset("assets/img/mark.svg"), asset("assets/fonts.css")}
	if p := ValidateTheme(ok, OverridableThemeFiles); p != nil {
		t.Errorf("fonts.css and safe assets are valid, got %+v", p)
	}
}
