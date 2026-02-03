// Package html provides HTML rendering capabilities.
package html

import (
	"fmt"
	"regexp"
	"strings"
)

// MarkdownRenderer converts markdown to HTML with styling.
type MarkdownRenderer struct {
	title       string
	description string
	styles      string
}

// NewMarkdownRenderer creates a new markdown renderer with defaults.
func NewMarkdownRenderer(title, description string) *MarkdownRenderer {
	return &MarkdownRenderer{
		title:       title,
		description: description,
		styles:      defaultStyles,
	}
}

// RenderMarkdownToHTML converts markdown string to styled HTML.
func (mr *MarkdownRenderer) RenderMarkdownToHTML(markdown string) string {
	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"en\">\n")
	html.WriteString("<head>\n")
	html.WriteString("<meta charset=\"UTF-8\">\n")
	html.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	html.WriteString(fmt.Sprintf("<title>%s</title>\n", escapeHTML(mr.title)))
	html.WriteString("<style>\n")
	html.WriteString(mr.styles)
	html.WriteString("</style>\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")
	html.WriteString("<div class=\"container\">\n")

	// Parse YAML frontmatter if present
	markdown = mr.stripFrontmatter(markdown)

	// Convert markdown to HTML
	htmlContent := mr.parseMarkdown(markdown)
	html.WriteString(htmlContent)

	html.WriteString("</div>\n")
	html.WriteString("</body>\n")
	html.WriteString("</html>\n")

	return html.String()
}

// stripFrontmatter removes YAML frontmatter from markdown.
func (mr *MarkdownRenderer) stripFrontmatter(markdown string) string {
	lines := strings.Split(markdown, "\n")
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "---") {
		return markdown
	}

	// Find closing ---
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "---") {
			// Return everything after the closing ---
			return strings.Join(lines[i+1:], "\n")
		}
	}

	return markdown
}

// parseMarkdown converts markdown text to HTML.
func (mr *MarkdownRenderer) parseMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var html strings.Builder
	var inCodeBlock bool
	var inList bool
	var inTable bool

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				html.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				lang := strings.TrimPrefix(strings.TrimPrefix(line, "```"), " ")
				html.WriteString(fmt.Sprintf("<pre><code class=\"language-%s\">\n", escapeHTML(lang)))
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			html.WriteString(escapeHTML(line))
			html.WriteString("\n")
			continue
		}

		// Handle tables
		if strings.Contains(line, "|") && i+1 < len(lines) && strings.Contains(lines[i+1], "|") && strings.Contains(lines[i+1], "-") {
			if !inTable {
				html.WriteString("<table>\n<thead>\n<tr>\n")
				inTable = true
			}

			cells := strings.Split(strings.Trim(line, "| "), "|")
			for _, cell := range cells {
				html.WriteString(fmt.Sprintf("<th>%s</th>\n", strings.TrimSpace(cell)))
			}
			html.WriteString("</tr>\n</thead>\n<tbody>\n")

			// Skip separator line
			i++
			continue
		}

		if inTable && strings.Contains(line, "|") && !strings.Contains(line, "-") {
			cells := strings.Split(strings.Trim(line, "| "), "|")
			html.WriteString("<tr>\n")
			for _, cell := range cells {
				html.WriteString(fmt.Sprintf("<td>%s</td>\n", strings.TrimSpace(cell)))
			}
			html.WriteString("</tr>\n")
			continue
		}

		if inTable && !strings.Contains(line, "|") {
			html.WriteString("</tbody>\n</table>\n")
			inTable = false
		}

		// Skip empty lines if already in list/table
		if strings.TrimSpace(line) == "" {
			if inList {
				html.WriteString("</ul>\n")
				inList = false
			}
			html.WriteString("\n")
			continue
		}

		// Handle headings
		if strings.HasPrefix(line, "# ") {
			html.WriteString(fmt.Sprintf("<h1>%s</h1>\n", escapeHTML(strings.TrimPrefix(line, "# "))))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			html.WriteString(fmt.Sprintf("<h2>%s</h2>\n", escapeHTML(strings.TrimPrefix(line, "## "))))
			continue
		}
		if strings.HasPrefix(line, "### ") {
			html.WriteString(fmt.Sprintf("<h3>%s</h3>\n", escapeHTML(strings.TrimPrefix(line, "### "))))
			continue
		}
		if strings.HasPrefix(line, "#### ") {
			html.WriteString(fmt.Sprintf("<h4>%s</h4>\n", escapeHTML(strings.TrimPrefix(line, "#### "))))
			continue
		}

		// Handle lists
		if strings.HasPrefix(line, "- ") {
			if !inList {
				html.WriteString("<ul>\n")
				inList = true
			}
			content := mr.renderInline(strings.TrimPrefix(line, "- "))
			html.WriteString(fmt.Sprintf("<li>%s</li>\n", content))
			continue
		}

		// Regular paragraph
		if strings.TrimSpace(line) != "" {
			content := mr.renderInline(line)
			html.WriteString(fmt.Sprintf("<p>%s</p>\n", content))
		}
	}

	// Close any open elements
	if inList {
		html.WriteString("</ul>\n")
	}
	if inTable {
		html.WriteString("</tbody>\n</table>\n")
	}

	return html.String()
}

// renderInline handles inline markdown formatting (bold, italic, code, links).
func (mr *MarkdownRenderer) renderInline(text string) string {
	// Bold: **text** or __text__
	re := regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
	text = re.ReplaceAllString(text, "<strong>$1$2</strong>")

	// Italic: *text* or _text_
	re = regexp.MustCompile(`\*(.*?)\*|_(.*?)_`)
	text = re.ReplaceAllString(text, "<em>$1$2</em>")

	// Inline code: `text`
	re = regexp.MustCompile("`(.*?)`")
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		code := strings.TrimPrefix(strings.TrimSuffix(match, "`"), "`")
		return fmt.Sprintf("<code>%s</code>", escapeHTML(code))
	})

	// Links: [text](url)
	re = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	text = re.ReplaceAllString(text, "<a href=\"$2\">$1</a>")

	return text
}

// escapeHTML escapes HTML special characters.
func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

const defaultStyles = `
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
	line-height: 1.6;
	color: #1f2937;
	background-color: #f9fafb;
	padding: 2rem 0;
}

.container {
	max-width: 900px;
	margin: 0 auto;
	padding: 2rem;
	background-color: white;
	border-radius: 0.375rem;
	box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

h1 {
	font-size: 2.25rem;
	font-weight: 700;
	margin: 2rem 0 1rem 0;
	color: #1f2937;
	border-bottom: 3px solid #2563eb;
	padding-bottom: 0.5rem;
}

h2 {
	font-size: 1.875rem;
	font-weight: 600;
	margin: 1.5rem 0 1rem 0;
	color: #1f2937;
	border-left: 4px solid #2563eb;
	padding-left: 1rem;
}

h3 {
	font-size: 1.25rem;
	font-weight: 600;
	margin: 1rem 0 0.5rem 0;
	color: #374151;
}

h4 {
	font-size: 1rem;
	font-weight: 600;
	margin: 0.75rem 0 0.5rem 0;
	color: #4b5563;
}

p {
	margin: 1rem 0;
	line-height: 1.8;
}

ul, ol {
	margin: 1rem 0 1rem 2rem;
}

li {
	margin: 0.5rem 0;
	line-height: 1.8;
}

code {
	background-color: #f3f4f6;
	padding: 0.2rem 0.5rem;
	border-radius: 0.25rem;
	font-family: "Menlo", "Monaco", "Courier New", monospace;
	font-size: 0.9rem;
	color: #dc2626;
}

pre {
	background-color: #1f2937;
	color: #e5e7eb;
	padding: 1.5rem;
	border-radius: 0.375rem;
	overflow-x: auto;
	margin: 1rem 0;
	line-height: 1.4;
}

pre code {
	background: none;
	color: #e5e7eb;
	padding: 0;
	border-radius: 0;
}

table {
	width: 100%;
	border-collapse: collapse;
	margin: 1.5rem 0;
	border: 1px solid #e5e7eb;
}

thead {
	background-color: #f3f4f6;
}

th {
	padding: 0.75rem 1rem;
	text-align: left;
	font-weight: 600;
	border-bottom: 2px solid #d1d5db;
	color: #1f2937;
}

td {
	padding: 0.75rem 1rem;
	border-bottom: 1px solid #e5e7eb;
}

tr:hover {
	background-color: #f9fafb;
}

a {
	color: #2563eb;
	text-decoration: none;
	border-bottom: 1px dotted #2563eb;
}

a:hover {
	color: #1d4ed8;
	border-bottom-style: solid;
}

strong {
	font-weight: 600;
	color: #111827;
}

em {
	font-style: italic;
	color: #4b5563;
}

blockquote {
	border-left: 4px solid #dbeafe;
	background-color: #f0f9ff;
	padding: 1rem;
	margin: 1.5rem 0;
	border-radius: 0.25rem;
}

@media (max-width: 768px) {
	.container {
		padding: 1rem;
	}

	h1 {
		font-size: 1.75rem;
	}

	h2 {
		font-size: 1.5rem;
	}

	table {
		font-size: 0.9rem;
	}

	th, td {
		padding: 0.5rem;
	}
}
`
