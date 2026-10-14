//go:build windows

package network

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	procQueryFullProcessImageName = modkernel32.NewProc("QueryFullProcessImageNameW")
)

func (c *Checker) FindProcess(port int) (int, string, error) {
	if port <= 0 || port > 65535 {
		return 0, "", fmt.Errorf("invalid port: %d", port)
	}

	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return 0, "", err
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	targetPortSuffix := fmt.Sprintf(":%d", port)

	var foundPID int
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "TCP") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		localAddr := fields[1]
		state := fields[3]
		pidStr := fields[4]

		if state == "LISTENING" && strings.HasSuffix(localAddr, targetPortSuffix) {
			pid, err := strconv.Atoi(pidStr)
			if err == nil && pid > 0 {
				foundPID = pid
				break
			}
		}
	}

	if foundPID == 0 {
		return 0, "", nil
	}

	exePath := getProcessPath(foundPID)
	return foundPID, exePath, nil
}

func getProcessPath(pid int) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)

	buf := make([]uint16, windows.MAX_PATH*2)
	size := uint32(len(buf))
	r1, _, _ := procQueryFullProcessImageName.Call(
		uintptr(h),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		return ""
	}
	return string(utf16.Decode(buf[:size]))
}
