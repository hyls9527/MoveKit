package winsyscall

import (
	"testing"
)

func TestGetPowerStatus(t *testing.T) {
	status := GetPowerStatus()
	t.Logf("ACLineStatus: %d, BatteryFlag: %d, BatteryLifePercent: %d%%",
		status.ACLineStatus, status.BatteryFlag, status.BatteryLifePercent)

	// ACLineStatus: 0=offline, 1=online, 255=unknown
	if status.ACLineStatus > 1 && status.ACLineStatus != 255 {
		t.Errorf("unexpected ACLineStatus: %d", status.ACLineStatus)
	}

	// BatteryLifePercent should be 0-100 or 255 (unknown)
	if status.BatteryLifePercent > 100 && status.BatteryLifePercent != 255 {
		t.Errorf("unexpected BatteryLifePercent: %d", status.BatteryLifePercent)
	}
}

func TestIsOnBattery(t *testing.T) {
	onBattery := IsOnBattery()
	t.Logf("IsOnBattery: %v", onBattery)

	// We can't assert a specific value since it depends on whether
	// the test machine is plugged in. Just verify the function doesn't panic.
}
