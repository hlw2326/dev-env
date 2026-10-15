//go:build windows

package process

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

type jobObject struct {
	handle windows.Handle
}

func newJob() *jobObject {
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return &jobObject{handle: 0}
	}

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}

	if _, err := windows.SetInformationJobObject(
		handle,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			return &jobObject{handle: 0}
		}
		return &jobObject{handle: 0}
	}

	return &jobObject{handle: handle}
}

func (j *jobObject) assign(proc *os.Process) {
	if j == nil || j.handle == 0 || proc == nil {
		return
	}
	handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(proc.Pid))
	if err != nil {
		return
	}
	defer func() {
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			return
		}
	}()
	if err := windows.AssignProcessToJobObject(j.handle, handle); err != nil {
		return
	}
}
