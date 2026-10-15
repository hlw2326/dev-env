//go:build windows

package monitor

import (
	"runtime"
	"time"
	"unsafe"

	domainmonitor "devenv/internal/domain/monitor"
	"golang.org/x/sys/windows"
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

type processMemoryCounters struct {
	cb                         uint32
	pageFaultCount             uint32
	peakWorkingSetSize         uintptr
	workingSetSize             uintptr
	quotaPeakPagedPoolUsage    uintptr
	quotaPagedPoolUsage        uintptr
	quotaPeakNonPagedPoolUsage uintptr
	quotaNonPagedPoolUsage     uintptr
	pagefileUsage              uintptr
	peakPagefileUsage          uintptr
}

var (
	psapi                    = windows.NewLazySystemDLL("psapi.dll")
	getProcessMemoryInfoProc = psapi.NewProc("GetProcessMemoryInfo")
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	globalMemoryStatusProc   = kernel32.NewProc("GlobalMemoryStatusEx")
	getSystemTimesProc       = kernel32.NewProc("GetSystemTimes")
	getProcessIoCountersProc = kernel32.NewProc("GetProcessIoCounters")
)

func readProcess(pid int) (processSnapshot, bool) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_VM_READ, false, uint32(pid))
	if err != nil {
		return processSnapshot{}, false
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	creation, exitTime, kernel, user := windows.Filetime{}, windows.Filetime{}, windows.Filetime{}, windows.Filetime{}
	if err := windows.GetProcessTimes(handle, &creation, &exitTime, &kernel, &user); err != nil {
		return processSnapshot{}, false
	}
	processTime := filetimeDuration(kernel) + filetimeDuration(user)
	creationTime := time.Unix(0, creation.Nanoseconds())
	metrics := domainmonitor.ProcessMetrics{UptimeSeconds: int64(time.Since(creationTime).Seconds())}

	counters := processMemoryCounters{cb: uint32(unsafe.Sizeof(processMemoryCounters{}))}
	result, _, _ := getProcessMemoryInfoProc.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&counters)),
		uintptr(unsafe.Sizeof(counters)),
	)
	if result != 0 {
		metrics.MemoryBytes = uint64(counters.workingSetSize)
	}

	var ioCounters windows.IO_COUNTERS
	result, _, _ = getProcessIoCountersProc.Call(uintptr(handle), uintptr(unsafe.Pointer(&ioCounters)))
	if result != 0 {
		metrics.ReadBytes = ioCounters.ReadTransferCount
		metrics.WriteBytes = ioCounters.WriteTransferCount
	}
	return processSnapshot{cpuTime: processTime, metrics: metrics}, true
}

type memoryStatusEx struct {
	Length                  uint32
	MemoryLoad              uint32
	TotalPhysicalMemory     uint64
	AvailablePhysicalMemory uint64
	TotalPageFile           uint64
	AvailablePageFile       uint64
	TotalVirtual            uint64
	AvailableVirtual        uint64
	AvailableExtended       uint64
}

func readSystem() systemSnapshot {
	status := memoryStatusEx{Length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	result, _, _ := globalMemoryStatusProc.Call(uintptr(unsafe.Pointer(&status)))
	if result == 0 {
		return systemSnapshot{}
	}
	used := status.TotalPhysicalMemory - status.AvailablePhysicalMemory
	percent := float64(status.MemoryLoad)
	snapshot := systemSnapshot{
		metrics: domainmonitor.SystemMetrics{
			MemoryTotalBytes: status.TotalPhysicalMemory,
			MemoryUsedBytes:  used,
			MemoryPercent:    percent,
		},
	}
	var idle windows.Filetime
	var kernel windows.Filetime
	var user windows.Filetime
	if result, _, _ := getSystemTimesProc.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	); result != 0 {
		snapshot.idleTime = filetimeDuration(idle)
		snapshot.totalTime = filetimeDuration(kernel) + filetimeDuration(user)
	}
	return snapshot
}

func cpuCount() int {
	count := runtime.NumCPU()
	if count < 1 {
		return 1
	}
	return count
}

func filetimeDuration(value windows.Filetime) time.Duration {
	units := (uint64(value.HighDateTime) << 32) | uint64(value.LowDateTime)
	return time.Duration(units * 100)
}
