// Package winsyscall provides low-level Win32 API wrappers for NTFS junction
// operations, Windows registry access, file operations, process enumeration,
// volume/disk queries, power status checks, and shell notifications.
//
// interfaces.go defines the testability abstraction layer — every function
// the migrate package calls through this package is represented as an interface
// method. Per D-03, existing package-level function signatures are NOT modified.
package winsyscall

// RegistryAccessor abstracts Windows registry read/write operations under HKCU and HKLM.
type RegistryAccessor interface {
	ReadExpandString(keyPath, valueName string) (string, error)
	WriteExpandString(keyPath, valueName, value string) error
	WriteString(keyPath, valueName, value string) error
	ReadInteger(keyPath, valueName string) (uint64, error)
	ReadIntegerLM(keyPath, valueName string) (uint64, error)
}

// ProcessChecker abstracts process enumeration via CreateToolhelp32Snapshot.
type ProcessChecker interface {
	IsProcessRunning(name string) bool
}

// VolumeChecker abstracts volume and disk space queries.
type VolumeChecker interface {
	GetVolumeInfo(drive string) (VolumeInfo, error)
	GetDiskFreeSpace(drive string) (freeBytes, totalBytes, totalFreeBytes int64, err error)
	IsCrossVolume(source, target string) bool
}

// JunctionOps abstracts NTFS junction (mount point) creation, detection, and deletion.
type JunctionOps interface {
	IsJunction(path string) bool
	CreateJunction(linkPath, targetPath string) error
	DeleteJunction(linkPath string) error
	ResolveTarget(path string) (string, error)
}

// ShellNotifier abstracts Windows shell change notifications via SHChangeNotify.
type ShellNotifier interface {
	NotifyFolderRenamed(oldPath, newPath string)
	NotifyDirUpdated(path string)
}

// FileOps abstracts cross-platform file copy operations with ACL preservation.
type FileOps interface {
	CopyFileWithACL(src, dst string, callback CopyProgressCallback) error
}

// PowerChecker abstracts system power status queries.
type PowerChecker interface {
	IsOnBattery() bool
}

// SyscallLayer bundles all syscall abstractions as public fields for
// dependency injection. In production, use NewRealLayer() to create a
// layer backed by real Win32 API calls. In tests, construct a SyscallLayer
// with mock implementations to avoid touching the real Windows API.
//
//go:generate mockgen -destination=mocks/mock_registry.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall RegistryAccessor
//go:generate mockgen -destination=mocks/mock_process.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall ProcessChecker
//go:generate mockgen -destination=mocks/mock_volume.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall VolumeChecker
//go:generate mockgen -destination=mocks/mock_junction.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall JunctionOps
//go:generate mockgen -destination=mocks/mock_shell.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall ShellNotifier
//go:generate mockgen -destination=mocks/mock_fileops.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall FileOps
//go:generate mockgen -destination=mocks/mock_power.go -package=mocks github.com/hyls9527/MoveKit/internal/syscall PowerChecker
type SyscallLayer struct {
	Registry RegistryAccessor
	Process  ProcessChecker
	Volume   VolumeChecker
	Junction JunctionOps
	Shell    ShellNotifier
	File     FileOps
	Power    PowerChecker
}
