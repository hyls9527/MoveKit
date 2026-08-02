package winsyscall

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	// UserShellFoldersKey is the registry path for per-user known folder paths
	// stored as REG_EXPAND_SZ values (may contain %USERPROFILE% etc.).
	UserShellFoldersKey = `Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`

	// ShellFoldersKey is the registry path for per-user known folder paths
	// stored as REG_SZ values (fully expanded, no environment variables).
	ShellFoldersKey = `Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Folders`
)

// ReadExpandString reads a REG_EXPAND_SZ value from the given registry key path
// under HKEY_CURRENT_USER. Environment variables in the value are NOT expanded
// by this function — the caller gets the raw value (e.g. "%USERPROFILE%\Desktop").
func ReadExpandString(keyPath, valueName string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.READ)
	if err != nil {
		return "", fmt.Errorf("open key HKCU\\%s: %w", keyPath, err)
	}
	defer k.Close()

	val, valType, err := k.GetStringValue(valueName)
	if err != nil {
		return "", fmt.Errorf("read value %q: %w", valueName, err)
	}
	if valType != registry.EXPAND_SZ && valType != registry.SZ {
		return "", fmt.Errorf("unexpected registry type %d for %q (want REG_EXPAND_SZ or REG_SZ)", valType, valueName)
	}
	return val, nil
}

// WriteExpandString writes a REG_EXPAND_SZ value to the given registry key path
// under HKEY_CURRENT_USER. The value may contain %ENVVAR% references.
func WriteExpandString(keyPath, valueName, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE|registry.CREATE_SUB_KEY)
	if err != nil {
		return fmt.Errorf("open key HKCU\\%s: %w", keyPath, err)
	}
	defer k.Close()

	if err := k.SetExpandStringValue(valueName, value); err != nil {
		return fmt.Errorf("set expand string %q: %w", valueName, err)
	}
	return nil
}

// WriteString writes a REG_SZ value to the given registry key path
// under HKEY_CURRENT_USER.
func WriteString(keyPath, valueName, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE|registry.CREATE_SUB_KEY)
	if err != nil {
		return fmt.Errorf("open key HKCU\\%s: %w", keyPath, err)
	}
	defer k.Close()

	if err := k.SetStringValue(valueName, value); err != nil {
		return fmt.Errorf("set string %q: %w", valueName, err)
	}
	return nil
}

// ReadInteger reads a REG_DWORD value from the given registry key path
// under HKEY_CURRENT_USER.
func ReadInteger(keyPath, valueName string) (uint64, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.READ)
	if err != nil {
		return 0, fmt.Errorf("open key HKCU\\%s: %w", keyPath, err)
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue(valueName)
	if err != nil {
		return 0, fmt.Errorf("read integer %q: %w", valueName, err)
	}
	return val, nil
}

// ReadIntegerLM reads a REG_DWORD value from the given registry key path
// under HKEY_LOCAL_MACHINE.
func ReadIntegerLM(keyPath, valueName string) (uint64, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ)
	if err != nil {
		return 0, fmt.Errorf("open key HKLM\\%s: %w", keyPath, err)
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue(valueName)
	if err != nil {
		return 0, fmt.Errorf("read integer %q: %w", valueName, err)
	}
	return val, nil
}
