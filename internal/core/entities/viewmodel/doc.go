// Package viewmodel holds the intermediate value every output backend
// consumes: views, the view models they project to, element pages, output
// paths, and the generated-file notice.
//
// It is standard-library only. Every collection is an ordered slice with a
// documented sort key, never a map, so byte-stable output (FR-021) is a
// structural property rather than something each backend must remember.
package viewmodel
