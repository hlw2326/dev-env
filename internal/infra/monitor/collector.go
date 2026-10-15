package monitor

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	domainmonitor "devenv/internal/domain/monitor"
)

type processSample struct {
	when    time.Time
	CPUTime time.Duration
	metrics domainmonitor.ProcessMetrics
}

type systemSample struct {
	idle  time.Duration
	total time.Duration
}

// Collector keeps the previous process sample in memory so CPU usage can be
// calculated without writing high-frequency data to SQLite.
type Collector struct {
	mu       sync.Mutex
	previous map[int]processSample
	system   *systemSample
}

func NewCollector() *Collector {
	return &Collector{previous: make(map[int]processSample)}
}

func (c *Collector) Process(pid int) domainmonitor.ProcessMetrics {
	if pid <= 0 {
		return domainmonitor.ProcessMetrics{}
	}
	current, ok := readProcess(pid)
	if !ok {
		c.mu.Lock()
		delete(c.previous, pid)
		c.mu.Unlock()
		return domainmonitor.ProcessMetrics{}
	}

	now := time.Now()
	currentSample := processSample{when: now, CPUTime: current.cpuTime, metrics: current.metrics}
	c.mu.Lock()
	previous, found := c.previous[pid]
	c.previous[pid] = currentSample
	c.mu.Unlock()
	if found {
		elapsed := now.Sub(previous.when).Seconds()
		if elapsed > 0 {
			cpu := (current.cpuTime - previous.CPUTime).Seconds() / elapsed
			current.metrics.CPUPercent = clamp(cpu * 100 / float64(cpuCount()))
		}
	}
	return current.metrics
}

func (c *Collector) Traffic(path string) domainmonitor.ProjectTraffic {
	file, err := os.Open(path)
	if err != nil {
		return domainmonitor.ProjectTraffic{}
	}
	defer func() { _ = file.Close() }()

	traffic := domainmonitor.ProjectTraffic{}
	var totalResponseMs float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(strings.TrimSuffix(scanner.Text(), "\r"), "\t")
		if len(fields) < 4 {
			continue
		}
		status, statusErr := strconv.Atoi(fields[1])
		responseBytes, bytesErr := strconv.ParseInt(fields[2], 10, 64)
		responseMs, responseErr := strconv.ParseFloat(fields[3], 64)
		if statusErr != nil || bytesErr != nil || responseErr != nil {
			continue
		}
		traffic.RequestCount++
		traffic.ResponseBytes += responseBytes
		if status >= 400 {
			traffic.ErrorCount++
		}
		totalResponseMs += responseMs * 1000
		traffic.LastAccess = fields[0]
	}
	if traffic.RequestCount > 0 {
		traffic.AverageResponseMs = totalResponseMs / float64(traffic.RequestCount)
	}
	return traffic
}

func (c *Collector) System() domainmonitor.SystemMetrics {
	current := readSystem()
	c.mu.Lock()
	previous := c.system
	c.system = &systemSample{idle: current.idleTime, total: current.totalTime}
	c.mu.Unlock()
	if previous != nil {
		totalDelta := current.totalTime - previous.total
		idleDelta := current.idleTime - previous.idle
		if totalDelta > 0 {
			current.metrics.CPUPercent = clamp(float64(totalDelta-idleDelta) / float64(totalDelta) * 100)
		}
	}
	return current.metrics
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
