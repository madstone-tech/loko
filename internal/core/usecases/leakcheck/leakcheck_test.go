package leakcheck

import (
	"strings"
	"testing"
)

// TestLeakedReportsABlockedGoroutine plants a goroutine blocked forever on a
// channel nothing else can reach — the exact shape the goroutineleak profile
// detects — and checks the report names it. This binary deliberately does
// not use Main, since the planted goroutine stays leaked.
func TestLeakedReportsABlockedGoroutine(t *testing.T) {
	if got := Leaked(); got != "" {
		t.Fatalf("leak reported before any was planted:\n%s", got)
	}
	started := make(chan struct{})
	go func() {
		ch := make(chan int)
		close(started)
		<-ch // unreachable from any other goroutine: leaked
	}()
	<-started
	report := Leaked()
	if !strings.Contains(report, "TestLeakedReportsABlockedGoroutine") {
		t.Fatalf("the planted leak is not reported:\n%s", report)
	}
}
