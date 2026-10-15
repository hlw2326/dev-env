package network

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
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

// ListeningPID returns the process ID owning a listening TCP port.
// It is intentionally best-effort because the port can disappear while it is being queried.
func (c *Checker) ListeningPID(port int) int {
	if port <= 0 || port > ports.MAX_PORT {
		return 0
	}

	command := exec.Command("netstat", "-ano", "-p", "tcp")
	output, err := command.Output()
	if err != nil {
		return 0
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 || !strings.EqualFold(fields[0], "TCP") || !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}
		localAddress := fields[1]
		separator := strings.LastIndex(localAddress, ":")
		if separator < 0 {
			continue
		}
		listeningPort, err := strconv.Atoi(localAddress[separator+1:])
		if err != nil || listeningPort != port {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			return 0
		}
		return pid
	}
	return 0
}
