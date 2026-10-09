package cmd

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

func TestWarnIfExposed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip   string
		warn bool
	}{
		{"127.0.0.1", false},
		{"::1", false},
		{"0.0.0.0", true},
		{"192.168.1.20", true},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		warnIfExposed(&buf, &net.TCPAddr{IP: net.ParseIP(tt.ip), Port: 8080})
		if got := strings.Contains(buf.String(), "not loopback"); got != tt.warn {
			t.Errorf("%s: warned = %v, want %v (%q)", tt.ip, got, tt.warn, buf.String())
		}
	}
}
