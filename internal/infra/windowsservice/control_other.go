//go:build !windows

package windowsservice

import (
	"errors"
	"time"
)

func Exists(string) (bool, error)           { return false, errors.New("Windows 服务仅支持 Windows") }
func IsStopped(string) (bool, error)        { return false, errors.New("Windows 服务仅支持 Windows") }
func Start(string) error                    { return errors.New("Windows 服务仅支持 Windows") }
func Stop(string, time.Duration) error      { return errors.New("Windows 服务仅支持 Windows") }
func OwnsProcess(string, int, string) bool  { return false }
func ExecutablePath(string) (string, error) { return "", errors.New("Windows 服务仅支持 Windows") }
