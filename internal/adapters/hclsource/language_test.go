package hclsource

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestDecoderAcceptsTheLanguageTable holds the decoder to arch.Language: an
// unknown attribute in each block is refused with exactly the table's
// attribute names as the legal set. locals takes any name, so it is skipped.
func TestDecoderAcceptsTheLanguageTable(t *testing.T) {
	t.Parallel()
	for _, b := range arch.Language {
		if b.Name == "locals" {
			continue
		}
		t.Run(b.Name, func(t *testing.T) {
			t.Parallel()
			src := withUnknownAttr(t, b)
			_, diags := loadSample(t, map[string]string{"a.loko.hcl": src})
			want := "Supported attributes are: " + strings.Join(b.AttrNames(), ", ") + "."
			for _, d := range diags {
				if d.Code == arch.CodeUnknownAttribute && strings.Contains(d.Detail, `"zz_unknown"`) {
					if !strings.Contains(d.Detail, want) {
						t.Errorf("detail %q, want it to contain %q", d.Detail, want)
					}
					return
				}
			}
			t.Errorf("no unknown_attribute for zz_unknown in:\n%s\ngot %v", src, diags)
		})
	}
}

// TestLanguageExamplesDecode: every example parses and decodes without an
// error, and is canonically formatted.
func TestLanguageExamplesDecode(t *testing.T) {
	t.Parallel()
	for _, b := range arch.Language {
		t.Run(b.Name, func(t *testing.T) {
			t.Parallel()
			if got := string(hclwrite.Format([]byte(b.Example))); got != b.Example {
				t.Errorf("example is not canonical:\n%s\nwant:\n%s", b.Example, got)
			}
			_, diags := loadSample(t, map[string]string{"a.loko.hcl": b.Example})
			for _, d := range diags {
				if d.Severity == arch.SeverityError {
					t.Errorf("%s: %s", d.Code, d.Detail)
				}
			}
		})
	}
}

// withUnknownAttr inserts zz_unknown into the block's own body in its example.
func withUnknownAttr(t *testing.T, b arch.LanguageBlock) string {
	t.Helper()
	lines := strings.Split(b.Example, "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if (strings.HasPrefix(trimmed, b.Name+" ") || trimmed == b.Name+" {") && strings.HasSuffix(trimmed, "{") {
			indent := l[:len(l)-len(strings.TrimLeft(l, " "))] + "  "
			out := append(append(append([]string{}, lines[:i+1]...), indent+"zz_unknown = 1"), lines[i+1:]...)
			return strings.Join(out, "\n")
		}
	}
	t.Fatalf("no %q block line in example:\n%s", b.Name, b.Example)
	return ""
}
