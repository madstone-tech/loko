package html

import (
	"io/fs"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var tokenDecl = regexp.MustCompile(`(--[a-z0-9-]+):\s*(#[0-9A-Fa-f]{6})\s*;`)

// themeTokens reads the colour tokens of the light block (the first :root)
// and the dark block (:root[data-theme="dark"]) of the built-in stylesheet.
func themeTokens(t *testing.T) (light, dark map[string]string, css string) {
	t.Helper()
	data, err := fs.ReadFile(builtinTheme, "theme/style.css")
	if err != nil {
		t.Fatal(err)
	}
	css = string(data)
	block := func(start string) map[string]string {
		i := strings.Index(css, start)
		if i < 0 {
			t.Fatalf("style.css has no %q block", start)
		}
		body := css[i : i+strings.Index(css[i:], "\n}")]
		out := map[string]string{}
		for _, m := range tokenDecl.FindAllStringSubmatch(body, -1) {
			out[m[1]] = strings.ToUpper(m[2])
		}
		return out
	}
	light = block(":root {")
	dark = block(`:root[data-theme="dark"] {`)
	for k, v := range light { // dark inherits what it does not redefine
		if _, ok := dark[k]; !ok {
			dark[k] = v
		}
	}
	return light, dark, css
}

func luminance(hex string) float64 {
	channel := func(s string) float64 {
		v, _ := strconv.ParseUint(s, 16, 8)
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(hex[1:3]) + 0.7152*channel(hex[3:5]) + 0.0722*channel(hex[5:7])
}

func contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// TestThemeContrast is spec FR-032 and SC-004 (research R11): every text
// colour meets WCAG AA (4.5:1) against its background in both themes.
func TestThemeContrast(t *testing.T) {
	t.Parallel()
	light, dark, css := themeTokens(t)
	pairs := [][2]string{
		{"--txt-1", "--bg"}, {"--txt-2", "--bg"}, {"--txt-3", "--bg"},
		{"--txt-1", "--card"}, {"--txt-2", "--card"}, {"--txt-3", "--card"},
		{"--link", "--bg"}, {"--link", "--card"},
		{"--accent-text", "--bg"}, {"--accent-text", "--card"},
		{"--chip-text", "--chip-bg"},
		{"--side-txt", "--side-bg"}, {"--side-muted", "--side-bg"}, {"--side-link", "--side-bg"},
		{"--kind-person-fg", "--kind-person-bg"}, {"--kind-system-fg", "--kind-system-bg"},
		{"--kind-container-fg", "--kind-container-bg"}, {"--kind-component-fg", "--kind-component-bg"},
		{"--kind-external-fg", "--kind-external-bg"},
	}
	for name, tokens := range map[string]map[string]string{"light": light, "dark": dark} {
		for _, p := range pairs {
			fg, bg := tokens[p[0]], tokens[p[1]]
			if fg == "" || bg == "" {
				t.Errorf("%s: %s or %s is not a #RRGGBB token", name, p[0], p[1])
				continue
			}
			if c := contrast(fg, bg); c < 4.5 {
				t.Errorf("%s: %s %s on %s %s is %.2f:1, below 4.5:1", name, p[0], fg, p[1], bg, c)
			}
		}
	}
	if light["--link"] != "#2D66AE" || light["--accent-text"] != "#128200" {
		t.Errorf("light link %s / accent text %s; FR-032 fixes them at #2D66AE / #128200", light["--link"], light["--accent-text"])
	}
	if regexp.MustCompile(`(^|[;{\s])color:\s*var\(--txt-4\)`).MatchString(css) {
		t.Error("--txt-4 is decoration only and must not colour text")
	}
}
