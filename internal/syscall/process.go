package winsyscall

import (
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	procCreateMutexW             = modkernel32.NewProc("CreateMutexW") // CreateMutexW is in kernel32
	procCreateToolhelp32Snapshot = modkernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = modkernel32.NewProc("Process32FirstW")
	procProcess32NextW           = modkernel32.NewProc("Process32NextW")
)

const (
	TH32CS_SNAPPROCESS = 0x00000002
)

// EnsureSingleInstance creates a named mutex to enforce single-instance behavior.
// If another instance already holds the mutex, an error is returned.
// On success, the returned handle must be kept alive (do not close it) for the
// lifetime of the process.
func EnsureSingleInstance(name string) (syscall.Handle, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return syscall.InvalidHandle, fmt.Errorf("utf16 mutex name: %w", err)
	}

	h, _, e := procCreateMutexW.Call(
		0, // lpMutexAttributes (NULL = default security)
		0, // bInitialOwner
		uintptr(unsafe.Pointer(namePtr)),
	)
	if h == 0 {
		return syscall.InvalidHandle, fmt.Errorf("CreateMutexW(%q): %w", name, e)
	}

	// ERROR_ALREADY_EXISTS (183) means another instance already created this mutex.
	const ERROR_ALREADY_EXISTS = 183
	if errno, ok := e.(syscall.Errno); ok && errno == ERROR_ALREADY_EXISTS {
		// Close the duplicate handle before returning.
		closeHandle(syscall.Handle(h))
		return syscall.InvalidHandle, fmt.Errorf("another instance is already running (mutex %q)", name)
	}

	return syscall.Handle(h), nil
}

// CloseHandle closes a Windows handle. Exported so callers outside this
// package (e.g. main.go) can close the single-instance mutex handle.
func CloseHandle(h syscall.Handle) {
	closeHandle(h)
}

// processEntry32 mirrors the Win32 PROCESSENTRY32W structure.
type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriorityClass   int32
	Flags           uint32
	ExeFile         [260]uint16
}

// IsProcessRunning returns true if a process with the given executable name
// (e.g. "360sd.exe") is currently running. Uses CreateToolhelp32Snapshot.
func IsProcessRunning(name string) bool {
	snap, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snap == 0 || snap == ^uintptr(0) {
		return false
	}
	defer closeHandle(syscall.Handle(snap))

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	r1, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	if r1 == 0 {
		// ERROR_NO_MORE_FILES (18) is expected when no processes match.
		// Other errors (e.g. ERROR_ACCESS_DENIED) mean enumeration failed.
		_ = syscall.GetLastError()
		return false
	}

	nameLower := strings.ToLower(name)
	for {
		exeName := nullTerminatedUTF16(entry.ExeFile[:])
		if strings.ToLower(exeName) == nameLower {
			return true
		}
		r1, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&entry)))
		if r1 == 0 {
			// ERROR_NO_MORE_FILES (18): end of process list — expected.
			// Other errors indicate enumeration was interrupted.
			_ = syscall.GetLastError()
			break
		}
	}
	return false
}

// nullTerminatedUTF16 extracts a string from a null-terminated UTF-16 buffer.
func nullTerminatedUTF16(buf []uint16) string {
	for i, c := range buf {
		if c == 0 {
			return string(utf16.Decode(buf[:i]))
		}
	}
	return string(utf16.Decode(buf))
}
