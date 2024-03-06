package powerhouse

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
)

// BatteryMetrics ...
type BatteryMetrics struct {
	// Err is set if there was an error.
	Err error

	// Time this metric was generated.
	Time time.Time

	// IsConnected is true if the device is connected, false otherwise.
	IsConnected bool

	// IsCharging is true if the device is charging, false otherwise.
	IsCharging bool

	// IsFullCharged is true if the device is fully charged, false otherwise.
	IsFullyCharged bool

	// Subject to change
	CycleCount              uint64
	DesignCapacity          uint64
	AppleRawMaxCapacity     uint64
	AppleRawCurrentCapacity uint64
	AppleRawBatteryVoltage  uint64
	InstantAmperage         int64
}

// batteryMetricsFromError creates a BatteryMetrics object from the given error.
func batteryMetricsFromError(err error) *BatteryMetrics {
	return &BatteryMetrics{Err: err}
}

// batteryMetricsFromIDevice creates a BatteryMetrics object from the
func batteryMetricsFromIDevice(in any) (*BatteryMetrics, error) {
	// Parse battery info
	var battery struct {
		UpdateTime              int64  `mapstructure:"UpdateTime"`
		ExternalConnected       bool   `mapstructure:"ExternalConnected"`
		IsCharging              bool   `mapstructure:"IsCharging"`
		FullyCharged            bool   `mapstructure:"FullyCharged"`
		CycleCount              uint64 `mapstructure:"CycleCount"`
		DesignCapacity          uint64 `mapstructure:"DesignCapacity"`
		AppleRawMaxCapacity     uint64 `mapstructure:"AppleRawMaxCapacity"`
		AppleRawCurrentCapacity uint64 `mapstructure:"AppleRawCurrentCapacity"`
		AppleRawBatteryVoltage  uint64 `mapstructure:"AppleRawBatteryVoltage"`
		InstantAmperage         int64  `mapstructure:"InstantAmperage"`
	}

	err := mapstructure.Decode(in, &battery)
	if err != nil {
		return nil, fmt.Errorf("parse battery info: %w", err)
	}

	// Send
	return &BatteryMetrics{
		Time:           time.Unix(battery.UpdateTime, 0),
		IsConnected:    battery.ExternalConnected,
		IsCharging:     battery.IsCharging,
		IsFullyCharged: battery.FullyCharged,

		CycleCount:              battery.CycleCount,
		DesignCapacity:          battery.DesignCapacity,
		AppleRawMaxCapacity:     battery.AppleRawMaxCapacity,
		AppleRawCurrentCapacity: battery.AppleRawCurrentCapacity,
		AppleRawBatteryVoltage:  battery.AppleRawBatteryVoltage,
		InstantAmperage:         battery.InstantAmperage,
	}, nil
}
