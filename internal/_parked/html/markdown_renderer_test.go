package html

import (
	"strings"
	"testing"
)

func TestNewMarkdownRenderer(t *testing.T) {
	renderer := NewMarkdownRenderer("Test Title", "Test Description")
	if renderer == nil {
		t.Fatal("NewMarkdownRenderer returned nil")
	}
	if renderer.title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", renderer.title)
	}
	if renderer.description != "Test Description" {
		t.Errorf("expected description 'Test Description', got %q", renderer.description)
	}
}

func TestRenderMarkdownToHTML_Basic(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "Description")
	markdown := "# Heading 1\n\nThis is a paragraph."
	html := renderer.RenderMarkdownToHTML(markdown)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}
	if !strings.Contains(html, "<h1>Heading 1</h1>") {
		t.Error("HTML should contain h1 tag")
	}
	if !strings.Contains(html, "<p>This is a paragraph.</p>") {
		t.Error("HTML should contain paragraph tag")
	}
}

func TestRenderMarkdownToHTML_Headings(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := `# H1
## H2
### H3
#### H4`

	html := renderer.RenderMarkdownToHTML(markdown)

	tests := []struct {
		tag  string
		text string
	}{
		{"<h1>H1</h1>", "<h1>H1</h1>"},
		{"<h2>H2</h2>", "<h2>H2</h2>"},
		{"<h3>H3</h3>", "<h3>H3</h3>"},
		{"<h4>H4</h4>", "<h4>H4</h4>"},
	}

	for _, tt := range tests {
		if !strings.Contains(html, tt.text) {
			t.Errorf("expected HTML to contain %q", tt.text)
		}
	}
}

func TestRenderMarkdownToHTML_Lists(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := `- Item 1
- Item 2
- Item 3`

	html := renderer.RenderMarkdownToHTML(markdown)

	if !strings.Contains(html, "<ul>") {
		t.Error("HTML should contain ul tag")
	}
	if !strings.Contains(html, "<li>Item 1</li>") {
		t.Error("HTML should contain li tag for Item 1")
	}
	if !strings.Contains(html, "</ul>") {
		t.Error("HTML should contain closing ul tag")
	}
}

func TestRenderMarkdownToHTML_InlineFormatting(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := "This is **bold** and *italic* and `code` text."

	html := renderer.RenderMarkdownToHTML(markdown)

	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Error("HTML should contain strong tag for bold")
	}
	if !strings.Contains(html, "<em>italic</em>") {
		t.Error("HTML should contain em tag for italic")
	}
	if !strings.Contains(html, "<code>code</code>") {
		t.Error("HTML should contain code tag")
	}
}

func TestRenderMarkdownToHTML_Links(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := "Check out [this link](https://example.com)"

	html := renderer.RenderMarkdownToHTML(markdown)

	if !strings.Contains(html, `<a href="https://example.com">this link</a>`) {
		t.Error("HTML should contain link")
	}
}

func TestRenderMarkdownToHTML_Tables(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |`

	html := renderer.RenderMarkdownToHTML(markdown)

	if !strings.Contains(html, "<table>") {
		t.Error("HTML should contain table tag")
	}
	if !strings.Contains(html, "<thead>") {
		t.Error("HTML should contain thead tag")
	}
	if !strings.Contains(html, "<th>Header 1</th>") {
		t.Error("HTML should contain th tag for headers")
	}
	if !strings.Contains(html, "<td>Cell 1</td>") {
		t.Error("HTML should contain td tag for cells")
	}
}

func TestRenderMarkdownToHTML_StripsFrontmatter(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := `---
name: "Test"
description: "Test description"
---

# Content`

	html := renderer.RenderMarkdownToHTML(markdown)

	if !strings.Contains(html, "<h1>Content</h1>") {
		t.Error("HTML should contain content heading")
	}
	if strings.Contains(html, "name:") {
		t.Error("HTML should not contain frontmatter")
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"a & b", "a &amp; b"},
		{"it's", "it&#39;s"},
	}

	for _, tt := range tests {
		result := escapeHTML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRenderInline(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")

	tests := []struct {
		input    string
		contains string
		notEmpty bool
	}{
		{"**bold**", "<strong>bold</strong>", true},
		{"__bold__", "<strong>bold</strong>", true},
		{"*italic*", "<em>italic</em>", true},
		{"_italic_", "<em>italic</em>", true},
		{"`code`", "<code>code</code>", true},
		{"[link](url)", `<a href="url">link</a>`, true},
	}

	for _, tt := range tests {
		result := renderer.renderInline(tt.input)
		if tt.notEmpty && !strings.Contains(result, tt.contains) {
			t.Errorf("renderInline(%q) should contain %q, got %q", tt.input, tt.contains, result)
		}
	}
}

func TestStripFrontmatter(t *testing.T) {
	renderer := NewMarkdownRenderer("Test", "")

	tests := []struct {
		input    string
		expected string
	}{
		{
			"---\nname: test\n---\n# Content",
			"# Content",
		},
		{
			"No frontmatter\n# Content",
			"No frontmatter\n# Content",
		},
		{
			"---\n# Content",
			"---\n# Content",
		},
	}

	for _, tt := range tests {
		result := renderer.stripFrontmatter(tt.input)
		if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
			t.Errorf("stripFrontmatter(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func BenchmarkRenderMarkdownToHTML(b *testing.B) {
	renderer := NewMarkdownRenderer("Test", "")
	markdown := `# Introduction

## Features
- Feature 1
- Feature 2
- Feature 3

## Configuration

| Key | Value |
|-----|-------|
| name | test |
| version | 1.0 |

This is **bold** and *italic* text with [a link](https://example.com).

Code block example.

## Conclusion

Final paragraph.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.RenderMarkdownToHTML(markdown)
	}
}
