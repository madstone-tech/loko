// Package watch implements the ChangeWatcher port by polling: every tick it
// stats the watched set (architecture source, referenced prose, theme
// overrides) and signals once the set has settled. It uses only the standard
// library; fsnotify was removed in v1.0 (research R12).
package watch
