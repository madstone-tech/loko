package viewmodel

import (
	"fmt"
	"strings"
)

// Output paths are produced only by the functions in this file and never
// assembled at a call site (contracts/output-layout.md). Every path is
// relative to the output root and uses forward slashes on every platform.

// Segment escapes a name for use as one path segment (research R9, amended in
// feature 015).
//
// [A-Za-z0-9._-] pass through, so names such as orders_db stay readable in
// file names and URLs. Every other byte, including '~', becomes '~' followed
// by two lowercase hex digits; because '~' is itself escaped the mapping is
// injective, and two different names can never share a file name. A leading
// '.' is escaped too, so no segment can be "." or "..".
func Segment(name string) string {
	return escape(name, '~', func(c byte) bool { return c == '_' })
}

// Ident escapes a name for an identifier that must not contain '~': D2 node
// keys and CSS class names. '_' is the escape character here and is itself
// escaped, which is what lets NodeIDFor use "__" for '.' without collisions.
func Ident(name string) string {
	return escape(name, '_', func(byte) bool { return false })
}

func escape(name string, esc byte, keep func(byte) bool) string {
	var b strings.Builder
	b.Grow(len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', keep(c):
			b.WriteByte(c)
		case c == '.' && i > 0:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "%c%02x", esc, c)
		}
	}
	return b.String()
}

// NodeIDFor derives a node key from an address. D2 uses '.' for nesting, so
// dots become "__"; that cannot collide with an escaped name because Ident
// always escapes '_'.
func NodeIDFor(address string) string {
	return strings.ReplaceAll(Ident(address), ".", "__")
}

// DiagramFile is the path of a view's diagram in the given format extension.
func DiagramFile(id ViewID, ext string) string {
	return "diagrams/" + Segment(string(id)) + "." + ext
}

// ViewPage is a view's page in the site.
func ViewPage(id ViewID) string { return "view/" + Segment(string(id)) + ".html" }

// MarkdownViewPage is a view's markdown document.
func MarkdownViewPage(id ViewID) string { return "md/view/" + Segment(string(id)) + ".md" }

// IndexPage is the entry document for the "html" or "md" output.
func IndexPage(ext string) string {
	if ext == "md" {
		return "md/index.md"
	}
	return "index." + ext
}

// ElementPath is an element's page: element/<kind>/<name>.<ext>, placed under
// md/ for markdown. address is the element address, "<kind>.<name>".
func ElementPath(address, ext string) string {
	kind, name, _ := strings.Cut(address, ".")
	p := "element/" + Segment(kind) + "/" + Segment(name) + "." + ext
	if ext == "md" {
		return "md/" + p
	}
	return p
}

// Rel returns the relative href from the document at from to the file at to.
// It is the only way links between outputs are formed, which is what makes
// every internal link checkable (FR-027).
func Rel(from, to string) string {
	fromDir := dirSegments(from)
	toSegs := strings.Split(to, "/")
	common := 0
	for common < len(fromDir) && common < len(toSegs)-1 && fromDir[common] == toSegs[common] {
		common++
	}
	var b strings.Builder
	for i := common; i < len(fromDir); i++ {
		b.WriteString("../")
	}
	b.WriteString(strings.Join(toSegs[common:], "/"))
	return b.String()
}

func dirSegments(p string) []string {
	i := strings.LastIndexByte(p, '/')
	if i < 0 {
		return nil
	}
	return strings.Split(p[:i], "/")
}

// MarkdownPage converts a site page path (element/... or view/...) to the
// matching markdown document path under md/.
func MarkdownPage(htmlPath string) string {
	return "md/" + strings.TrimSuffix(htmlPath, ".html") + ".md"
}
