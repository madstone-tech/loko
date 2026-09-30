package devserver

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases/leakcheck"
)

// TestMain fails the package if any test leaks a goroutine (Go 1.27
// goroutineleak profile).
func TestMain(m *testing.M) { leakcheck.Main(m) }
