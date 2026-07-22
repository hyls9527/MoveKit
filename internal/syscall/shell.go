package winsyscall

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// Shell change notification event IDs.
	SHCNE_RENAMEFOLDER = 0x00020000
	SHCNE_UPDATEDIR    = 0x00001000
	SHCNE_UPDATEITEM   = 0x00002000

	// Shell change notification flags.
	SHCNF_PATH        = 0x0005
	SHCNF_FLUSHNOWAIT = 0x2000
)

var (
	modshell32         = syscall.NewLazyDLL("shell32.dll")
	procSHChangeNotify = modshell32.NewProc("SHChangeNotify")
)

// SHChangeNotify calls the Win32 SHChangeNotify function to notify the shell
// that a change has occurred. This causes Explorer to refresh its view.
//
//	eventId  — e.g. SHCNE_RENAMEFOLDER
//	flags    — e.g. SHCNF_PATH
//	dwItem1  — path to the old item (or just item for non-rename events)
//	dwItem2  — path to the new item (for rename events, otherwise 0)
func SHChangeNotify(eventId int32, flags uint32, dwItem1, dwItem2 uintptr) {
	// SHChangeNotify returns void; no error to check.
	procSHChangeNotify.Call(
		uintptr(eventId),
		uintptr(flags),
		dwItem1,
		dwItem2,
	)
}

// NotifyFolderRenamed tells the shell that a folder has been renamed/moved.
// oldPath and newPath must be UTF-16 encoded paths; this function handles encoding.
func NotifyFolderRenamed(oldPath, newPath string) {
	oldPtr, err := syscall.UTF16PtrFromString(oldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "winsyscall: NotifyFolderRenamed utf16 oldPath %q: %v\n", oldPath, err)
		return
	}
	newPtr, err := syscall.UTF16PtrFromString(newPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "winsyscall: NotifyFolderRenamed utf16 newPath %q: %v\n", newPath, err)
		return
	}
	SHChangeNotify(
		SHCNE_RENAMEFOLDER,
		SHCNF_PATH,
		uintptr(unsafe.Pointer(oldPtr)),
		uintptr(unsafe.Pointer(newPtr)),
	)
}

// NotifyDirUpdated tells the shell that the contents of a directory have changed.
// This causes Explorer to re-enumerate the directory.
func NotifyDirUpdated(path string) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "winsyscall: NotifyDirUpdated utf16 path %q: %v\n", path, err)
		return
	}
	SHChangeNotify(
		SHCNE_UPDATEDIR,
		SHCNF_PATH|SHCNF_FLUSHNOWAIT,
		uintptr(unsafe.Pointer(p)),
		0,
	)
}
