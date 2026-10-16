//go:build windows

package windowsservice

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var setFileSecurity = windows.NewLazySystemDLL("advapi32.dll").NewProc("SetFileSecurityW")

// GrantDirectoryRead grants LocalService read access to this directory only.
// SetFileSecurity avoids propagating ACL changes through the entire user profile.
func GrantDirectoryRead(path string) error {
	descriptor, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return fmt.Errorf("读取服务目录权限 %s: %w", path, err)
	}
	existing, _, err := descriptor.DACL()
	if err != nil {
		return err
	}
	account, err := windows.CreateWellKnownSid(windows.WinLocalServiceSid)
	if err != nil {
		return err
	}
	var pinner runtime.Pinner
	pinner.Pin(account)
	defer pinner.Unpin()
	permissions, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.FILE_GENERIC_READ | windows.FILE_GENERIC_EXECUTE,
		AccessMode:        windows.GRANT_ACCESS,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
			TrusteeValue: windows.TrusteeValueFromSID(account),
		},
	}}, existing)
	if err != nil {
		return err
	}
	updated, err := windows.NewSecurityDescriptor()
	if err != nil {
		return err
	}
	if err := updated.SetDACL(permissions, true, false); err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	result, _, callError := setFileSecurity.Call(uintptr(unsafe.Pointer(name)), uintptr(windows.DACL_SECURITY_INFORMATION), uintptr(unsafe.Pointer(updated)))
	runtime.KeepAlive(permissions)
	runtime.KeepAlive(updated)
	runtime.KeepAlive(name)
	if result == 0 {
		return fmt.Errorf("授予服务目录读取权限 %s: %w", path, callError)
	}
	return nil
}
