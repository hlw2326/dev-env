package network

import (
	"net"
	"testing"
)

func TestChecker(t *testing.T) {
	checker := NewChecker()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	if checker.Available(port) {
		t.Errorf("expected port %d to NOT be available", port)
	}
	if !checker.Listening(port) {
		t.Errorf("expected port %d to be listening", port)
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("failed to close listener: %v", err)
	}

	if !checker.Available(port) {
		t.Errorf("expected port %d to be available after close", port)
	}
}
