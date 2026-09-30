//go:build race

package cmd

// raceEnabled reports whether the race detector is compiled in. It slows
// execution several-fold, so wall-clock budgets are not meaningful under it.
const raceEnabled = true
