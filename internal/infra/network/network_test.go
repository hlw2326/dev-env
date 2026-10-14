package network

import (
	"net"
	"testing"
)

func TestChecker(t *testing.T) {
	c := NewChecker()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	if c.Available(port) {
		t.Errorf("expected port %d to NOT be available", port)
	}
	if !c.Listening(port) {
		t.Errorf("expected port %d to be listening", port)
	}

	pid, _, err := c.FindProcess(port)
	if err != nil {
		t.Logf("FindProcess returned error: %v", err)
	} else if pid <= 0 {
		t.Logf("FindProcess pid was %d", pid)
	}
}
