package exporter

import "testing"

func TestServeAddrBindsLocalhost(t *testing.T) {
	if got := serveAddr(3123); got != "127.0.0.1:3123" {
		t.Fatalf("expected localhost bind address, got %q", got)
	}
}
