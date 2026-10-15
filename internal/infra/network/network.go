package network

import (
	"fmt"
	"net"
	"time"

	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

const (
	DIAL_TIMEOUT = 250 * time.Millisecond
)

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) Available(port int) bool {
	if port <= 0 || port > ports.MAX_PORT {
		return false
	}
	if c.Listening(port) {
		return false
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", paths.LOOPBACK, port))
	if err != nil {
		return false
	}
	if err := listener.Close(); err != nil {
		return false
	}
	return true
}

func (c *Checker) Listening(port int) bool {
	if port <= 0 || port > ports.MAX_PORT {
		return false
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", paths.LOOPBACK, port), DIAL_TIMEOUT)
	if err != nil {
		return false
	}
	if err := conn.Close(); err != nil {
		return false
	}
	return true
}
