package winsyscall

import (
	"unsafe"
)

var (
	procGetSystemPowerStatus = modkernel32.NewProc("GetSystemPowerStatus")
)

// PowerStatus holds the system power/battery status.
type PowerStatus struct {
	ACLineStatus        uint8
	BatteryFlag         uint8
	BatteryLifePercent  uint8
}

// GetPowerStatus returns the current system power status.
func GetPowerStatus() PowerStatus {
	var status struct {
		ACLineStatus        byte
		BatteryFlag         byte
		BatteryLifePercent  byte
		SystemStatusFlag    byte
		BatteryLifeTime     uint32
		BatteryFullLifeTime uint32
	}

	r1, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&status)))
	if r1 == 0 {
		// If the call fails, return a default indicating unknown state.
		return PowerStatus{ACLineStatus: 255, BatteryFlag: 255, BatteryLifePercent: 255}
	}

	return PowerStatus{
		ACLineStatus:       status.ACLineStatus,
		BatteryFlag:        status.BatteryFlag,
		BatteryLifePercent: status.BatteryLifePercent,
	}
}

// IsOnBattery returns true if the system is running on battery power.
// ACLineStatus==0 means offline (battery), and BatteryFlag!=128 means a
// battery is present (128 = no battery).
func IsOnBattery() bool {
	s := GetPowerStatus()
	return s.ACLineStatus == 0 && s.BatteryFlag != 128
}
