// Package html implements the "html" output backend: a browsable site with a
// page per element and per view, rendered with html/template from embedded
// theme files and optional user overrides. It reads nothing but the projection
// and the theme files it is handed (FR-011) and imports no other backend
// (FR-012); diagrams are referenced by path, not embedded.
package html
