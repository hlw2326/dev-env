package app

import (
	domainmonitor "devenv/internal/domain/monitor"
	inframonitor "devenv/internal/infra/monitor"
)

// MonitorService exposes host metrics without persisting high-frequency samples.
type MonitorService struct {
	collector *inframonitor.Collector
}

func NewMonitorService(collector *inframonitor.Collector) *MonitorService {
	return &MonitorService{collector: collector}
}

func (s *MonitorService) Metrics() domainmonitor.SystemMetrics {
	return s.collector.System()
}
