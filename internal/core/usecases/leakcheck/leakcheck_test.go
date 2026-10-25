package leakcheck

import (
	"strings"
	"testing"
	"time"
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
	// started only says the goroutine is running, not that it has reached
	// <-ch: until it blocks there it is not leaked. Poll rather than assume,
	// since a loaded runner under -race can leave a wide gap.
	var report string
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if report = Leaked(); strings.Contains(report, "TestLeakedReportsABlockedGoroutine") {
			return
		}
	}
	t.Fatalf("the planted leak is not reported:\n%s", report)
}
