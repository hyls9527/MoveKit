package winsyscall

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"
)

func TestJunctionWithMklink(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")

	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir target: %v", err)
	}

	// mklink /J creates the link itself — do NOT pre-create it.
	cmd := exec.Command("cmd", "/c", "mklink", "/J", link, target)
	out, err := cmd.CombinedOutput()
	t.Logf("mklink output: %s", string(out))
	if err != nil {
		t.Fatalf("mklink /J: %v", err)
	}

	if !IsJunction(link) {
		t.Fatal("IsJunction returned false after mklink /J")
	}

	// Now delete it.
	if err := DeleteJunction(link); err != nil {
		t.Fatalf("DeleteJunction: %v", err)
	}

	if IsJunction(link) {
		t.Fatal("IsJunction returned true after DeleteJunction")
	}
	t.Log("mklink /J + DeleteJunction works")
}

func TestDebugReparseBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")

	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir target: %v", err)
	}

	absTarget, _ := filepath.Abs(target)

	// Create via mklink first (does not pre-create link dir)
	link := filepath.Join(tmpDir, "link")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", link, absTarget)
	out, err := cmd.CombinedOutput()
	t.Logf("mklink output: %s err=%v", string(out), err)
	if err != nil {
		t.Fatalf("mklink: %v", err)
	}

	absLink, _ := filepath.Abs(link)

	// Read the reparse data from the mklink-created junction.
	h, err := openReparse(absLink, syscall.GENERIC_READ, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE)
	if err != nil {
		t.Fatalf("openReparse read: %v", err)
	}
	defer closeHandle(h)

	readBuf := make([]byte, MAX_REPARSE_SIZE)
	var bytesRead uint32
	r1, _, e := procDeviceIoCtrl.Call(
		uintptr(h),
		uintptr(FSCTL_GET_REPARSE_POINT),
		0, 0,
		uintptr(unsafe.Pointer(&readBuf[0])),
		uintptr(len(readBuf)),
		uintptr(unsafe.Pointer(&bytesRead)),
		0,
	)
	t.Logf("Read returned: %d, error: %v, bytesRead: %d", r1, e, bytesRead)
	if r1 == 0 {
		t.Fatalf("FSCTL_GET_REPARSE_POINT failed: %v", e)
	}

	le := binary.LittleEndian
	readTag := le.Uint32(readBuf[0:])
	readDataLen := le.Uint16(readBuf[4:])
	t.Logf("mklink ReparseTag: 0x%08X, ReparseDataLength: %d", readTag, readDataLen)

	mpSubOff := le.Uint16(readBuf[8:])
	mpSubLen := le.Uint16(readBuf[10:])
	mpPrintOff := le.Uint16(readBuf[12:])
	mpPrintLen := le.Uint16(readBuf[14:])
	t.Logf("mklink MP: SubOff=%d SubLen=%d PrintOff=%d PrintLen=%d", mpSubOff, mpSubLen, mpPrintOff, mpPrintLen)

	if mpSubLen > 0 {
		subRaw := readBuf[16+int(mpSubOff) : 16+int(mpSubOff)+int(mpSubLen)]
		subStr := utf16BytesToString(subRaw)
		t.Logf("mklink SubName: %q", subStr)
	}
	if mpPrintLen > 0 {
		printRaw := readBuf[16+int(mpPrintOff) : 16+int(mpPrintOff)+int(mpPrintLen)]
		printStr := utf16BytesToString(printRaw)
		t.Logf("mklink PrintName: %q", printStr)
	}

	// Now try CreateJunction on a fresh link dir
	closeHandle(h)
	os.Remove(absLink) // remove the mklink junction
	os.Mkdir(absLink, 0o755)

	substituteName := `\??\` + absTarget
	printName := absTarget

	t.Logf("CreateJunction substituteName: %q", substituteName)
	t.Logf("CreateJunction printName: %q", printName)

	buf, err := buildReparseBuffer(substituteName, printName)
	if err != nil {
		t.Fatalf("buildReparseBuffer: %v", err)
	}

	// Compare with what mklink produced
	t.Logf("Our buffer size: %d, mklink buffer size: %d", len(buf), bytesRead)
	t.Logf("Our dataLen: %d, mklink dataLen: %d", le.Uint16(buf[4:]), readDataLen)

	ourTag := le.Uint32(buf[0:])
	t.Logf("Our tag: 0x%08X, mklink tag: 0x%08X", ourTag, readTag)

	ourSubLen := le.Uint16(buf[10:])
	ourPrintLen := le.Uint16(buf[14:])
	t.Logf("Our SubLen=%d PrintLen=%d vs mklink SubLen=%d PrintLen=%d", ourSubLen, ourPrintLen, mpSubLen, mpPrintLen)

	// Dump first 32 bytes of each
	t.Logf("Ours:     %x", buf[:min(32, len(buf))])
	t.Logf("Mklink:   %x", readBuf[:min(32, int(bytesRead))])

	// Create the junction with our code
	h2, err := openReparse(absLink, syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE)
	if err != nil {
		t.Fatalf("openReparse write: %v", err)
	}
	defer closeHandle(h2)

	var bytesReturned uint32
	r2, _, e2 := procDeviceIoCtrl.Call(
		uintptr(h2),
		uintptr(FSCTL_SET_REPARSE_POINT),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, 0,
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	t.Logf("SET returned: %d, error: %v", r2, e2)
	if r2 == 0 {
		t.Logf("SET failed: %v (Windows error %d)", e2, e2.(syscall.Errno))
	}
}

func utf16BytesToUint16(b []byte) []uint16 {
	n := len(b) / 2
	result := make([]uint16, n)
	for i := 0; i < n; i++ {
		result[i] = *(*uint16)(unsafe.Pointer(&b[i*2]))
	}
	return result
}

func utf16BytesToString(b []byte) string {
	u16 := utf16BytesToUint16(b)
	return syscall.UTF16ToString(u16)
}

func TestDebugRawJunction(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")

	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir target: %v", err)
	}
	if err := os.Mkdir(link, 0o755); err != nil {
		t.Fatalf("Mkdir link: %v", err)
	}

	absTarget, _ := filepath.Abs(target)
	absLink, _ := filepath.Abs(link)

	// Build buffer the same way Windows does
	subName := fmt.Sprintf(`\??\%s`, absTarget)
	printName := absTarget

	subRet, _ := syscall.UTF16FromString(subName)
	printRet, _ := syscall.UTF16FromString(printName)

	// ... but actually, let me try raw manual construction using little endian directly
	le := binary.LittleEndian

	subNameBytes := make([]byte, len(subRet)*2)
	for i, c := range subRet {
		le.PutUint16(subNameBytes[i*2:], c)
	}
	printNameBytes := make([]byte, len(printRet)*2)
	for i, c := range printRet {
		le.PutUint16(printNameBytes[i*2:], c)
	}

	subLen := len(subNameBytes)
	printLen := len(printNameBytes)
	dataLen := 8 + subLen + printLen
	totalLen := 8 + dataLen

	t.Logf("subName=%q subBytes=%d printName=%q printBytes=%d dataLen=%d totalLen=%d", subName, subLen, printName, printLen, dataLen, totalLen)

	rawBuf := make([]byte, totalLen)
	le.PutUint32(rawBuf[0:], IO_REPARSE_TAG_MOUNT_POINT)
	le.PutUint16(rawBuf[4:], uint16(dataLen))
	le.PutUint16(rawBuf[6:], 0)
	le.PutUint16(rawBuf[8:], 0)                 // SubstituteNameOffset
	le.PutUint16(rawBuf[10:], uint16(subLen))   // SubstituteNameLength
	le.PutUint16(rawBuf[12:], uint16(subLen))   // PrintNameOffset
	le.PutUint16(rawBuf[14:], uint16(printLen)) // PrintNameLength
	copy(rawBuf[16:], subNameBytes)
	copy(rawBuf[16+subLen:], printNameBytes)

	t.Logf("Raw buffer first 48 bytes: %x", rawBuf[:min(48, len(rawBuf))])

	h, err := openReparse(absLink, syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer closeHandle(h)

	var br uint32
	r, _, e := procDeviceIoCtrl.Call(
		uintptr(h),
		uintptr(FSCTL_SET_REPARSE_POINT),
		uintptr(unsafe.Pointer(&rawBuf[0])),
		uintptr(len(rawBuf)),
		0, 0,
		uintptr(unsafe.Pointer(&br)),
		0,
	)
	t.Logf("DeviceIoControl SET returned: %d, err: %v, errno: %d", r, e, e.(syscall.Errno))
	if r != 0 {
		t.Log("SUCCESS! Junction created")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
