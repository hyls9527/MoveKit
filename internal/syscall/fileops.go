package winsyscall

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Security information constants for Get/SetNamedSecurityInfoW.
const (
	SE_FILE_OBJECT            = 1
	DACL_SECURITY_INFORMATION = 4
)

// ACL represents a Windows Access Control List structure.
type ACL struct {
	AclRevision byte
	Sbz1        byte
	AclSize     uint16
	AceCount    uint16
	Sbz2        uint16
}

var (
	procCopyFileExW = modkernel32.NewProc("CopyFileExW")
	procLocalFree   = modkernel32.NewProc("LocalFree")

	modadvapi32               = syscall.NewLazyDLL("advapi32.dll")
	procGetNamedSecurityInfoW = modadvapi32.NewProc("GetNamedSecurityInfoW")
	procSetNamedSecurityInfoW = modadvapi32.NewProc("SetNamedSecurityInfoW")
)

// CopyProgressRoutine callback type.
// totalFileSize, totalBytesTransferred are passed by Windows.
// Returns a non-zero value to continue, zero to cancel.
type CopyProgressCallback func(totalFileSize, totalBytesTransferred int64, streamSize, streamBytesTransferred int64, dwStreamNumber, dwCallbackReason uint32, hSourceFile, hDestinationFile syscall.Handle) uintptr

// copyProgressCallback is the actual Go function registered with Windows.
// We store the user callback in a global to avoid passing a Go pointer through
// the C callback boundary — only one copy can be active at a time.
var activeCopyCallback CopyProgressCallback

// copyProgressThunk is the C-callable function pointer that Windows invokes.
//
//go:uintptrescapes
func copyProgressThunk(totalFileSize int64, totalBytesTransferred int64, streamSize int64, streamBytesTransferred int64, dwStreamNumber uint32, dwCallbackReason uint32, hSourceFile syscall.Handle, hDestinationFile syscall.Handle) uintptr {
	if activeCopyCallback != nil {
		return activeCopyCallback(totalFileSize, totalBytesTransferred, streamSize, streamBytesTransferred, dwStreamNumber, dwCallbackReason, hSourceFile, hDestinationFile)
	}
	return 1 // PROGRESS_CONTINUE
}

// CopyFileEx copies a file from src to dst with an optional progress callback.
// Pass nil for callback if no progress reporting is needed.
func CopyFileEx(src, dst string, callback CopyProgressCallback) error {
	srcPtr, err := syscall.UTF16PtrFromString(src)
	if err != nil {
		return fmt.Errorf("utf16 src: %w", err)
	}
	dstPtr, err := syscall.UTF16PtrFromString(dst)
	if err != nil {
		return fmt.Errorf("utf16 dst: %w", err)
	}

	var progressRoutine uintptr
	if callback != nil {
		activeCopyCallback = callback
		progressRoutine = syscall.NewCallback(copyProgressThunk)
	}

	// Use COPY_FILE_NO_BUFFERING for better read performance on large files.
	// Note: FILE_FLAG_SEQUENTIAL_SCAN is a CreateFile flag, not valid for CopyFileExW.
	var flags uint32 = 0

	r1, _, e := procCopyFileExW.Call(
		uintptr(unsafe.Pointer(srcPtr)),
		uintptr(unsafe.Pointer(dstPtr)),
		progressRoutine,
		0, // lpData
		0, // pbCancel
		uintptr(flags),
	)
	if r1 == 0 {
		return fmt.Errorf("CopyFileExW(%s -> %s): %w", src, dst, e)
	}
	return nil
}

// CopyFileWithACL copies a file and preserves its NTFS ACL (DDAC).
// Uses GetNamedSecurityInfoW to read source ACL and SetNamedSecurityInfoW to apply to destination.
func CopyFileWithACL(src, dst string, callback CopyProgressCallback) error {
	// 1. Copy the file using CopyFileExW
	if err := CopyFileEx(src, dst, callback); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	// 2. Read source file's security descriptor
	srcPtr, _ := syscall.UTF16PtrFromString(src)
	var pSD unsafe.Pointer
	var pACL *ACL

	r1, _, e := procGetNamedSecurityInfoW.Call(
		uintptr(unsafe.Pointer(srcPtr)),
		SE_FILE_OBJECT,
		DACL_SECURITY_INFORMATION,
		0, 0, // owner, group
		uintptr(unsafe.Pointer(&pACL)),
		0, // sacl
		uintptr(unsafe.Pointer(&pSD)),
	)
	if r1 != 0 {
		return fmt.Errorf("GetNamedSecurityInfoW(%s): %v", src, e)
	}
	defer procLocalFree.Call(uintptr(pSD))

	// 3. Apply ACL to destination
	dstPtr, _ := syscall.UTF16PtrFromString(dst)
	r1, _, e = procSetNamedSecurityInfoW.Call(
		uintptr(unsafe.Pointer(dstPtr)),
		SE_FILE_OBJECT,
		DACL_SECURITY_INFORMATION,
		0, 0, // owner, group
		uintptr(unsafe.Pointer(pACL)),
		0, // sacl
	)
	if r1 != 0 {
		return fmt.Errorf("SetNamedSecurityInfoW(%s): %v", dst, e)
	}

	return nil
}

// ToLongPath adds the `\\?\` prefix to a path, enabling extended-length path
// support (up to ~32 767 characters) on Windows.
func ToLongPath(path string) string {
	if strings.HasPrefix(path, `\\?\`) {
		return path
	}
	// UNC paths: \\server\share -> \\?\UNC\server\share
	if strings.HasPrefix(path, `\\`) {
		return `\\?\UNC\` + path[2:]
	}
	// Drive-letter paths: C:\foo -> \\?\C:\foo
	return `\\?\` + path
}

// IsCrossVolume returns true if source and target paths reside on different
// drive volumes (e.g. C: vs D:), which means a move requires copy+delete
// rather than a simple rename.
func IsCrossVolume(source, target string) bool {
	srcVol := filepath.VolumeName(source)
	tgtVol := filepath.VolumeName(target)
	if srcVol == "" || tgtVol == "" {
		return false
	}
	return !strings.EqualFold(srcVol, tgtVol)
}
