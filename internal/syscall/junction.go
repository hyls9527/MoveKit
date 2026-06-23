package winsyscall

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	IO_REPARSE_TAG_MOUNT_POINT = 0xA0000003

	FSCTL_SET_REPARSE_POINT    = 0x000900A4
	FSCTL_GET_REPARSE_POINT    = 0x000900A8
	FSCTL_DELETE_REPARSE_POINT = 0x000900AC

	FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000
	FILE_FLAG_BACKUP_SEMANTICS   = 0x02000000

	MAX_REPARSE_SIZE = 16 * 1024
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procDeviceIoCtrl = modkernel32.NewProc("DeviceIoControl")
	procCreateFileW  = modkernel32.NewProc("CreateFileW")
	procCloseHandle  = modkernel32.NewProc("CloseHandle")
)

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

func closeHandle(h syscall.Handle) {
	procCloseHandle.Call(uintptr(h))
}

func buildReparseBuffer(substitute, print string) ([]byte, error) {
	subUTF16, err := syscall.UTF16FromString(substitute)
	if err != nil {
		return nil, fmt.Errorf("utf16 substitute: %w", err)
	}
	printUTF16, err := syscall.UTF16FromString(print)
	if err != nil {
		return nil, fmt.Errorf("utf16 print: %w", err)
	}

	subLen := len(subUTF16) * 2
	printLen := len(printUTF16) * 2

	reparseDataLen := 8 + subLen + printLen
	totalSize := 8 + reparseDataLen

	buf := make([]byte, totalSize)
	le := binary.LittleEndian

	le.PutUint32(buf[0:], IO_REPARSE_TAG_MOUNT_POINT)
	le.PutUint16(buf[4:], uint16(reparseDataLen))
	le.PutUint16(buf[6:], 0)

	le.PutUint16(buf[8:], 0)
	le.PutUint16(buf[10:], uint16(subLen))
	le.PutUint16(buf[12:], uint16(subLen))
	le.PutUint16(buf[14:], uint16(printLen))

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

func CreateJunction(linkPath, targetPath string) error {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	absLink, err := filepath.Abs(linkPath)
	if err != nil {
		return fmt.Errorf("resolve link: %w", err)
	}

	fi, err := os.Stat(absLink)
	if err != nil {
		return fmt.Errorf("stat link: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("link path %s is not a directory", absLink)
	}

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
	r1, _, _ := procDeviceIoCtrl.Call(
		uintptr(h),
		uintptr(FSCTL_SET_REPARSE_POINT),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, 0,
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if r1 == 0 {
		// Note: do NOT close handle here — defer closeHandle(h) above handles cleanup.
		// Double-closing a Windows handle is undefined behavior.
		return createJunctionViaMklink(absLink, absTarget)
	}

	return nil
}

func createJunctionViaMklink(linkPath, targetPath string) error {
	_ = os.RemoveAll(linkPath)
	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, targetPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink /J: %w\n%s", err, string(out))
	}
	return nil
}

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

	var deleteBuf [8]byte
	le := binary.LittleEndian
	le.PutUint32(deleteBuf[0:], IO_REPARSE_TAG_MOUNT_POINT)
	le.PutUint16(deleteBuf[4:], 0)
	le.PutUint16(deleteBuf[6:], 0)

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
