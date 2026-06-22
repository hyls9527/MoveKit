package winsyscall

// RealImpl is the production adapter that implements all 7 syscall interfaces
// by delegating each method call to the corresponding package-level Win32 API
// wrapper function.
type RealImpl struct{}

// NewRealLayer returns a SyscallLayer backed by real Win32 API calls.
// All 7 fields share the same *RealImpl instance since RealImpl implements
// every interface.
func NewRealLayer() SyscallLayer {
	r := &RealImpl{}
	return SyscallLayer{
		Registry: r,
		Process:  r,
		Volume:   r,
		Junction: r,
		Shell:    r,
		File:     r,
		Power:    r,
	}
}

// --- RegistryAccessor ---

func (r *RealImpl) ReadExpandString(keyPath, valueName string) (string, error) {
	return ReadExpandString(keyPath, valueName)
}

func (r *RealImpl) WriteExpandString(keyPath, valueName, value string) error {
	return WriteExpandString(keyPath, valueName, value)
}

func (r *RealImpl) WriteString(keyPath, valueName, value string) error {
	return WriteString(keyPath, valueName, value)
}

func (r *RealImpl) ReadInteger(keyPath, valueName string) (uint64, error) {
	return ReadInteger(keyPath, valueName)
}

func (r *RealImpl) ReadIntegerLM(keyPath, valueName string) (uint64, error) {
	return ReadIntegerLM(keyPath, valueName)
}

// --- ProcessChecker ---

func (r *RealImpl) IsProcessRunning(name string) bool {
	return IsProcessRunning(name)
}

// --- VolumeChecker ---

func (r *RealImpl) GetVolumeInfo(drive string) (VolumeInfo, error) {
	return GetVolumeInfo(drive)
}

func (r *RealImpl) GetDiskFreeSpace(drive string) (int64, int64, int64, error) {
	return GetDiskFreeSpace(drive)
}

func (r *RealImpl) IsCrossVolume(source, target string) bool {
	return IsCrossVolume(source, target)
}

// --- JunctionOps ---

func (r *RealImpl) IsJunction(path string) bool {
	return IsJunction(path)
}

func (r *RealImpl) CreateJunction(linkPath, targetPath string) error {
	return CreateJunction(linkPath, targetPath)
}

func (r *RealImpl) DeleteJunction(linkPath string) error {
	return DeleteJunction(linkPath)
}

// --- ShellNotifier ---

func (r *RealImpl) NotifyFolderRenamed(oldPath, newPath string) {
	NotifyFolderRenamed(oldPath, newPath)
}

func (r *RealImpl) NotifyDirUpdated(path string) {
	NotifyDirUpdated(path)
}

// --- FileOps ---

func (r *RealImpl) CopyFileWithACL(src, dst string, callback CopyProgressCallback) error {
	return CopyFileWithACL(src, dst, callback)
}

// --- PowerChecker ---

func (r *RealImpl) IsOnBattery() bool {
	return IsOnBattery()
}
