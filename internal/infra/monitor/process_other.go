//go:build !windows

package monitor

import (
	"runtime"
	"time"

	domainmonitor "devenv/internal/domain/monitor"
)

type processSnapshot struct {
	cpuTime time.Duration
	metrics domainmonitor.ProcessMetrics
}

type systemSnapshot struct {
	metrics   domainmonitor.SystemMetrics
	idleTime  time.Duration
	totalTime time.Duration
}

func readProcess(_ int) (processSnapshot, bool) {
	return processSnapshot{}, false
}

func readSystem() systemSnapshot {
	return systemSnapshot{}
}

func cpuCount() int {
	count := runtime.NumCPU()
	if count < 1 {
		return 1
	}
	return count
}
