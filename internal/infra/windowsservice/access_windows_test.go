//go:build windows

package windowsservice

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestGrantDirectoryReadDoesNotChangeChildren(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child.txt")
	if err := os.WriteFile(child, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	readDescriptor := func(path string) string {
		t.Helper()
		descriptor, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		// Check the account being granted; unrelated user ACEs may be
		// re-canonicalized by Windows or another concurrently running process.
		entries := regexp.MustCompile(`\([^()]*;;;(?:LS|S-1-5-19)\)`).FindAllString(descriptor.String(), -1)
		sort.Strings(entries)
		return strings.Join(entries, "")
	}
	before := readDescriptor(child)
	if err := GrantDirectoryRead(root); err != nil {
		t.Fatal(err)
	}
	if after := readDescriptor(child); after != before {
		t.Fatalf("child permissions changed: before=%s after=%s", before, after)
	}
	if err := GrantDirectoryRead(root); err != nil {
		t.Fatalf("repeated grant failed: %v", err)
	}
}
