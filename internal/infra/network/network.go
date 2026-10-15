package network

import (
	"fmt"
	"net"
	"time"

	"devenv/internal/shared/paths"
)

const (
	MAX_PORT     = 65535
	DIAL_TIMEOUT = 250 * time.Millisecond
)

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) Available(port int) bool {
	if port <= 0 || port > MAX_PORT {
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
	if port <= 0 || port > MAX_PORT {
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
