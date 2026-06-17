package winsyscall

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	// Reparse tag for NTFS junction points (mount points).
	IO_REPARSE_TAG_MOUNT_POINT = 0xA0000003

	// DeviceIoControl control codes.
	FSCTL_SET_REPARSE_POINT    = 0x000900A4
	FSCTL_GET_REPARSE_POINT    = 0x000900A8
	FSCTL_DELETE_REPARSE_POINT = 0x000900AC

	// File attribute / creation flags.
	FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000
	FILE_FLAG_BACKUP_SEMANTICS   = 0x02000000

	// Maximum reparse data buffer size.
	MAX_REPARSE_SIZE = 16 * 1024
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procDeviceIoCtrl = modkernel32.NewProc("DeviceIoControl")
	procCreateFileW  = modkernel32.NewProc("CreateFileW")
	procCloseHandle  = modkernel32.NewProc("CloseHandle")
)

// openReparse opens a directory handle suitable for reparse point operations.
func openReparse(path string, access uint32, shareMode uint32) (syscall.Handle, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	flags := uint32(FILE_FLAG_OPEN_REPARSE_POINT | FILE_FLAG_BACKUP_SEMANTICS)
	h, _, e := procCreateFileW.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(access),
		uintptr(shareMode),
		0,
		uintptr(syscall.OPEN_EXISTING),
		uintptr(flags),
		0,
	)
	if h == uintptr(syscall.InvalidHandle) {
		return syscall.InvalidHandle, fmt.Errorf("CreateFileW(%s): %w", path, e)
	}
	return syscall.Handle(h), nil
}

// closeHandle is a thin wrapper around CloseHandle.
func closeHandle(h syscall.Handle) {
	procCloseHandle.Call(uintptr(h))
}

// buildReparseBuffer constructs the byte slice for FSCTL_SET_REPARSE_POINT
// given a substitute name (e.g. "\??\D:\Target") and a print name (e.g. "D:\Target").
//
// Layout (REPARSE_DATA_BUFFER for IO_REPARSE_TAG_MOUNT_POINT):
//
//	Offset  Size  Field
//	0       4     ReparseTag         (IO_REPARSE_TAG_MOUNT_POINT)
//	4       2     ReparseDataLength  (size of everything after this common header)
//	6       2     Reserved           (0)
//	-- MountPointReparseBuffer starts here (ReparseDataLength covers from here) --
//	8       2     SubstituteNameOffset   (0)
//	10      2     SubstituteNameLength   (in bytes, including NUL)
//	12      2     PrintNameOffset        (= SubstituteNameLength)
//	14      2     PrintNameLength        (in bytes, including NUL)
//	16      var   PathBuffer             (substitute UTF-16 + print UTF-16)
func buildReparseBuffer(substitute, print string) ([]byte, error) {
	subUTF16, err := syscall.UTF16FromString(substitute)
	if err != nil {
		return nil, fmt.Errorf("utf16 substitute: %w", err)
	}
	printUTF16, err := syscall.UTF16FromString(print)
	if err != nil {
		return nil, fmt.Errorf("utf16 print: %w", err)
	}

	subLen := (len(subUTF16) - 1) * 2 // bytes EXCLUDING NUL terminator
	printLen := (len(printUTF16) - 1) * 2

	// ReparseDataLength = 8 (mount-point header) + path buffer bytes
	reparseDataLen := 8 + subLen + 2 + printLen + 2 // +2 for each NUL terminator in path buffer
	totalSize := 8 + reparseDataLen // 8 = common header (tag + dataLen + reserved)

	buf := make([]byte, totalSize)
	le := binary.LittleEndian

	// Common header
	le.PutUint32(buf[0:], IO_REPARSE_TAG_MOUNT_POINT)
	le.PutUint16(buf[4:], uint16(reparseDataLen))
	le.PutUint16(buf[6:], 0) // Reserved

	// MountPointReparseBuffer header
	le.PutUint16(buf[8:], 0)              // SubstituteNameOffset
	le.PutUint16(buf[10:], uint16(subLen)) // SubstituteNameLength
	le.PutUint16(buf[12:], uint16(subLen+2)) // PrintNameOffset (= end of substitute + NUL)
	le.PutUint16(buf[14:], uint16(printLen)) // PrintNameLength

	// PathBuffer: substitute then print (already NUL-terminated by UTF16FromString)
	offset := 16
	for _, c := range subUTF16 {
		le.PutUint16(buf[offset:], c)
		offset += 2
	}
	for _, c := range printUTF16 {
		le.PutUint16(buf[offset:], c)
		offset += 2
	}

	return buf, nil
}

// CreateJunction creates an NTFS junction (mount point) at `linkPath` that points to `targetPath`.
// The link directory must already exist. The target must be an absolute path.
func CreateJunction(linkPath, targetPath string) error {
	// Normalise to absolute.
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	absLink, err := filepath.Abs(linkPath)
	if err != nil {
		return fmt.Errorf("resolve link: %w", err)
	}

	// Ensure link directory exists.
	fi, err := os.Stat(absLink)
	if err != nil {
		return fmt.Errorf("stat link: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("link path %s is not a directory", absLink)
	}

	// Build substitute and print names.
	substituteName := `\??\` + absTarget
	printName := absTarget

	buf, err := buildReparseBuffer(substituteName, printName)
	if err != nil {
		return fmt.Errorf("build reparse buffer: %w", err)
	}

	h, err := openReparse(absLink, syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE)
	if err != nil {
		return fmt.Errorf("open link: %w", err)
	}
	defer closeHandle(h)

	var bytesReturned uint32
	r1, _, e := procDeviceIoCtrl.Call(
		uintptr(h),
		uintptr(FSCTL_SET_REPARSE_POINT),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, 0,
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if r1 == 0 {
		return fmt.Errorf("DeviceIoControl(FSCTL_SET_REPARSE_POINT): %w", e)
	}

	return nil
}

// IsJunction returns true if the given path is an NTFS junction point.
func IsJunction(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	h, err := openReparse(absPath, syscall.GENERIC_READ, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE)
	if err != nil {
		return false
	}
	defer closeHandle(h)

	buf := make([]byte, MAX_REPARSE_SIZE)
	var bytesReturned uint32
	r1, _, _ := procDeviceIoCtrl.Call(
		uintptr(h),
		uintptr(FSCTL_GET_REPARSE_POINT),
		0, 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if r1 == 0 {
		return false
	}

	tag := *(*uint32)(unsafe.Pointer(&buf[0]))
	return tag == IO_REPARSE_TAG_MOUNT_POINT
}

// DeleteJunction removes the NTFS junction point at `linkPath`.
// The directory itself is NOT deleted — only the reparse point attribute is cleared.
func DeleteJunction(linkPath string) error {
	absLink, err := filepath.Abs(linkPath)
	if err != nil {
		return fmt.Errorf("resolve link: %w", err)
	}

	h, err := openReparse(absLink, syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE)
	if err != nil {
		return fmt.Errorf("open link: %w", err)
	}
	defer closeHandle(h)

	// For delete, we pass a small buffer with the tag and zero data length.
	var deleteBuf [8]byte
	le := binary.LittleEndian
	le.PutUint32(deleteBuf[0:], IO_REPARSE_TAG_MOUNT_POINT)
	le.PutUint16(deleteBuf[4:], 0) // ReparseDataLength
	le.PutUint16(deleteBuf[6:], 0) // Reserved

	var bytesReturned uint32
	r1, _, e := procDeviceIoCtrl.Call(
		uintptr(h),
		uintptr(FSCTL_DELETE_REPARSE_POINT),
		uintptr(unsafe.Pointer(&deleteBuf[0])),
		uintptr(len(deleteBuf)),
		0, 0,
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if r1 == 0 {
		return fmt.Errorf("DeviceIoControl(FSCTL_DELETE_REPARSE_POINT): %w", e)
	}

	return nil
}
