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
	for _, address := range []string{
		fmt.Sprintf("%s:%d", paths.LOOPBACK, port),
		fmt.Sprintf("[::1]:%d", port),
	} {
		conn, err := net.DialTimeout("tcp", address, DIAL_TIMEOUT)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

// ListeningPID returns the process ID owning a listening TCP port.
// It is intentionally best-effort because the port can disappear while it is being queried.
func (c *Checker) ListeningPID(port int) int {
	return c.ListeningPIDs([]int{port})[port]
}

// ListeningPIDs returns listener process IDs for requested TCP ports using a
// single system snapshot. This avoids launching netstat once per service row.
func (c *Checker) ListeningPIDs(portsToFind []int) map[int]int {
	result := make(map[int]int, len(portsToFind))
	wanted := make(map[int]bool, len(portsToFind))
	for _, port := range portsToFind {
		if port > 0 && port <= ports.MAX_PORT {
			wanted[port] = true
		}
	}
	if len(wanted) == 0 {
		return result
	}
	command := exec.Command("netstat", "-ano", "-p", "tcp")
	output, err := command.Output()
	if err != nil {
		return result
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
		if err != nil || !wanted[listeningPort] || result[listeningPort] != 0 {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err == nil {
			result[listeningPort] = pid
		}
	}
	return result
}
