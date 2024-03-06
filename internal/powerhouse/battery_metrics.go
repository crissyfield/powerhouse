package powerhouse

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
)

// BatteryMetricsAdapterDetails ...
type BatteryMetricsAdapterDetails struct {
	// Description could be "batt", "usb host", "baseline arcas", "pd charger", "usb charger", or "magsave acc".
	Description string

	// IsWireless is true if adapter support wireless charge.
	IsWireless bool

	// Current supported by the adapter (in A).
	Current float64

	// Watts supported by the adapter (in W).
	Watts float64
}

// BatteryMetrics ...
type BatteryMetrics struct {
	// Err is set if there was an error.
	Err error

	// Time this metric was generated.
	Time time.Time

	// Serial is the battery's serial number.
	Serial string

	// IsConnected is true if the device is connected to an external power source (false otherwise).
	IsConnected bool

	// IsExternalChargeCapable is true if the device is capable of being charged externally (false otherwise).
	IsExternalChargeCapable bool

	// IsCharging is true if the device's batery is charging (false otherwise).
	IsCharging bool

	// IsFullCharged is true if the device is fully charged (false otherwise).
	IsFullyCharged bool

	// Capacity remaining for this battery (0 - 100%).
	CurrentCapacity int

	// Number of times the battery has been charged already.
	CycleCount int

	// Full capacity the battery was designed for (in Ah).
	DesignCapacity float64

	// Max capacity the battery was designed for (in Ah).
	AppleRawMaxCapacity float64

	// Remaining charge capacity for this battery (in Ah). Use AppleRawMaxCapacity if this is 0.0.
	NominalChargeCapacity float64

	// Raw capacity remaining for this battery (in Ah).
	AppleRawCurrentCapacity float64

	// Raw battery voltage (in V).
	AppleRawBatteryVoltage float64

	// BootVoltage is the voltage during last boot (in V).
	BootVoltage float64

	// Voltage is the current voltage (in V).
	Voltage float64

	// Positive when charging, negative when discharging (in A).
	InstantAmperage float64

	// Temperature is the current temperature (in °C)
	Temperature float64

	// Adapter details
	AdapterDetails BatteryMetricsAdapterDetails
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
		Serial                  string `mapstructure:"Serial"`
		ExternalConnected       bool   `mapstructure:"ExternalConnected"`
		ExternalChargeCapable   bool   `mapstructure:"ExternalChargeCapable"`
		IsCharging              bool   `mapstructure:"IsCharging"`
		FullyCharged            bool   `mapstructure:"FullyCharged"`
		CurrentCapacity         int    `mapstructure:"CurrentCapacity"`
		CycleCount              uint64 `mapstructure:"CycleCount"`
		DesignCapacity          uint64 `mapstructure:"DesignCapacity"`
		AppleRawMaxCapacity     uint64 `mapstructure:"AppleRawMaxCapacity"`
		NominalChargeCapacity   uint64 `mapstructure:"NominalChargeCapacity"`
		AppleRawCurrentCapacity uint64 `mapstructure:"AppleRawCurrentCapacity"`
		AppleRawBatteryVoltage  uint64 `mapstructure:"AppleRawBatteryVoltage"`
		BootVoltage             uint64 `mapstructure:"BootVoltage"`
		Voltage                 uint64 `mapstructure:"Voltage"`
		InstantAmperage         int64  `mapstructure:"InstantAmperage"`
		Temperature             int64  `mapstructure:"Temperature"`

		AdapterDetails struct {
			Current     uint64 `mapstructure:"Current"`
			Description string `mapstructure:"Description"`
			IsWireless  bool   `mapstructure:"IsWireless"`
			Watts       uint64 `mapstructure:"Watts"`
		} `mapstructure:"AdapterDetails"`
	}

	err := mapstructure.Decode(in, &battery)
	if err != nil {
		return nil, fmt.Errorf("parse battery info: %w", err)
	}

	// Send
	return &BatteryMetrics{
		Time:                    time.Unix(battery.UpdateTime, 0),
		Serial:                  battery.Serial,
		IsConnected:             battery.ExternalConnected,
		IsExternalChargeCapable: battery.ExternalChargeCapable,
		IsCharging:              battery.IsCharging,
		IsFullyCharged:          battery.FullyCharged,
		CurrentCapacity:         battery.CurrentCapacity,
		CycleCount:              int(battery.CycleCount),
		DesignCapacity:          float64(battery.DesignCapacity) / 1000.0,
		AppleRawMaxCapacity:     float64(battery.AppleRawMaxCapacity) / 1000.0,
		AppleRawCurrentCapacity: float64(battery.AppleRawCurrentCapacity) / 1000.0,
		NominalChargeCapacity:   float64(battery.NominalChargeCapacity) / 1000.0,
		AppleRawBatteryVoltage:  float64(battery.AppleRawBatteryVoltage) / 1000.0,
		BootVoltage:             float64(battery.BootVoltage) / 1000.0,
		Voltage:                 float64(battery.Voltage) / 1000.0,
		InstantAmperage:         float64(battery.InstantAmperage) / 1000.0,
		Temperature:             float64(battery.Temperature)/100.0 + 30.0,
		AdapterDetails: BatteryMetricsAdapterDetails{
			Description: battery.AdapterDetails.Description,
			IsWireless:  battery.AdapterDetails.IsWireless,
			Current:     float64(battery.AdapterDetails.Current) / 1000.0,
			Watts:       float64(battery.AdapterDetails.Watts),
		},
	}, nil
}
