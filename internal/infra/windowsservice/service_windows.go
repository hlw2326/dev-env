//go:build windows

package windowsservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
)

func Run(configPath string) error {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read service definition: %w", err)
	}
	var definition Definition
	if err := json.Unmarshal(raw, &definition); err != nil {
		return fmt.Errorf("decode service definition: %w", err)
	}
	if definition.ServiceName == "" || definition.Binary == "" {
		return errors.New("service definition is incomplete")
	}
	return svc.Run(definition.ServiceName, &processService{definition: definition})
}

type processService struct {
	definition Definition
	command    *exec.Cmd
	logFile    *os.File
}

func (service *processService) Execute(_ []string, requests <-chan svc.ChangeRequest, statuses chan<- svc.Status) (bool, uint32) {
	accepted := svc.AcceptStop | svc.AcceptShutdown
	statuses <- svc.Status{State: svc.StartPending, Accepts: accepted, WaitHint: 10000}
	if err := service.start(); err != nil {
		service.appendLog("service process start failed: " + err.Error())
		statuses <- svc.Status{State: svc.Stopped, Win32ExitCode: 1}
		return false, 1
	}

	statuses <- svc.Status{State: svc.Running, Accepts: accepted}
	finished := make(chan error, 1)
	go func() {
		finished <- service.command.Wait()
	}()

	for {
		select {
		case waitErr := <-finished:
			service.closeLog()
			if waitErr != nil {
				service.appendLog("service process exited: " + waitErr.Error())
				statuses <- svc.Status{State: svc.Stopped, Win32ExitCode: 1}
				return false, 1
			}
			statuses <- svc.Status{State: svc.Stopped}
			return false, 0
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				statuses <- request.CurrentStatus
			case svc.Stop, svc.Shutdown:
				statuses <- svc.Status{State: svc.StopPending, Accepts: accepted, WaitHint: 10000}
				service.stop()
				waitErr := <-finished
				service.closeLog()
				if waitErr != nil && !errors.Is(waitErr, os.ErrProcessDone) {
					service.appendLog("service process stopped: " + waitErr.Error())
				}
				statuses <- svc.Status{State: svc.Stopped}
				return false, 0
			}
		}
	}
}

func (service *processService) start() error {
	if filepath.IsAbs(service.definition.Binary) {
		if info, err := os.Stat(service.definition.Binary); err != nil || info.IsDir() {
			if err == nil {
				err = errors.New("binary path is a directory")
			}
			return err
		}
	}
	if service.definition.Log != "" {
		if err := os.MkdirAll(filepath.Dir(service.definition.Log), 0o755); err != nil {
			return err
		}
		file, err := os.OpenFile(service.definition.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		service.logFile = file
	}
	service.command = exec.Command(service.definition.Binary, service.definition.Arguments...)
	service.command.Dir = service.definition.Directory
	if len(service.definition.Environment) > 0 {
		service.command.Env = append(os.Environ(), service.definition.Environment...)
	}
	service.command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if service.logFile != nil {
		service.command.Stdout = service.logFile
		service.command.Stderr = service.logFile
	}
	if err := service.command.Start(); err != nil {
		service.closeLog()
		return err
	}
	return nil
}

func (service *processService) stop() {
	if service.command == nil || service.command.Process == nil {
		return
	}
	killer := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(service.command.Process.Pid))
	killer.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := killer.Run(); err != nil {
		_ = service.command.Process.Kill()
	}
}

func (service *processService) closeLog() {
	if service.logFile == nil {
		return
	}
	_ = service.logFile.Close()
	service.logFile = nil
}

func (service *processService) appendLog(message string) {
	if service.definition.Log == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(service.definition.Log), 0o755); err != nil {
		return
	}
	file, err := os.OpenFile(service.definition.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	_, _ = io.WriteString(file, fmt.Sprintf("%s %s\r\n", time.Now().Format(time.RFC3339), message))
}
