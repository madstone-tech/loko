// Package d2 implements the "d2" and "svg" output backends. Both share one
// emitter from view model to D2 source; the svg backend then compiles, lays
// out and renders that source in-process with the d2 library and its embedded
// dagre layout. No process is ever started and no executable is looked up
// (FR-014).
//
// The emitter also returns a label plan: each element card's icon and the
// role of every label line. After rendering, the svg backend runs one pass
// over the SVG that adds the icons and a style per line from that plan
// (feature 020, ADR-0019). The .d2 output carries neither.
package d2
