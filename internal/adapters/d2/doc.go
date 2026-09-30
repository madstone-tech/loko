// Package d2 implements the "d2" and "svg" output backends. Both share one
// emitter from view model to D2 source; the svg backend then compiles, lays
// out and renders that source in-process with the d2 library and its embedded
// dagre layout. No process is ever started and no executable is looked up
// (FR-014).
package d2
