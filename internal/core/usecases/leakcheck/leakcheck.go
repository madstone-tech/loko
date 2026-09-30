// Package leakcheck fails a test binary that leaks goroutines, using the
// goroutineleak profile made generally available in Go 1.27.
//
// It is test support, imported only from _test.go files. It lives under
// internal/core/usecases because that is the one location the use-case,
// adapter and cmd layers may all already import, so no layer rule changes.
package leakcheck

import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
	"testing"
)

// Main runs the package's tests and then, if they passed, fails the binary
// when any goroutine has leaked. Use it from TestMain:
//
//	func TestMain(m *testing.M) { leakcheck.Main(m) }
func Main(m *testing.M) {
	code := m.Run()
	if code == 0 {
		if report := Leaked(); report != "" {
			fmt.Fprintf(os.Stderr, "FAIL: goroutine leak detected after all tests passed:\n%s", report)
			code = 1
		}
	}
	os.Exit(code)
}

// Leaked returns the goroutineleak profile when it lists any goroutine, and
// "" otherwise. Writing the profile first runs a leak-detecting GC cycle: a
// goroutine is leaked when it is blocked on a concurrency primitive that no
// runnable goroutine can reach, so it can never wake.
func Leaked() string {
	var b bytes.Buffer
	if err := pprof.Lookup("goroutineleak").WriteTo(&b, 1); err != nil {
		return "reading the goroutineleak profile: " + err.Error() + "\n"
	}
	first, _, _ := strings.Cut(b.String(), "\n")
	if strings.HasSuffix(first, "total 0") {
		return ""
	}
	return b.String()
}
