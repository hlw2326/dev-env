package monitor

// ProcessMetrics describes the current resource usage of one process.
type ProcessMetrics struct {
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryBytes   uint64  `json:"memoryBytes"`
	ReadBytes     uint64  `json:"readBytes"`
	WriteBytes    uint64  `json:"writeBytes"`
	UptimeSeconds int64   `json:"uptimeSeconds"`
}

// ProjectTraffic describes traffic parsed from a project's access log.
type ProjectTraffic struct {
	RequestCount      int64   `json:"requestCount"`
	ResponseBytes     int64   `json:"responseBytes"`
	ErrorCount        int64   `json:"errorCount"`
	AverageResponseMs float64 `json:"averageResponseMs"`
	LastAccess        string  `json:"lastAccess"`
}

// SystemMetrics describes host-wide CPU and memory usage.
type SystemMetrics struct {
	CPUPercent       float64 `json:"cpuPercent"`
	MemoryTotalBytes uint64  `json:"memoryTotalBytes"`
	MemoryUsedBytes  uint64  `json:"memoryUsedBytes"`
	MemoryPercent    float64 `json:"memoryPercent"`
}
