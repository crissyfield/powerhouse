package powerhouse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/crissyfield/powerhouse/internal/idevice"
)

// Device wraps information of a specific iDevice.
type Device struct {
	UDID        string // Unique device ID
	Name        string // Device name
	Type        string // Device type
	OSVersion   string // Version of the installed OS
	OSBuild     string // Build number of the installed OS
	WiFiAddress string // MAC address of the device

	idev *idevice.Device
}

// newDevice creates a new iDevice.
func newDevice(idev *idevice.Device) (*Device, error) {
	// Get device info
	info, err := idev.Info()
	if err != nil {
		return nil, fmt.Errorf("get device info: %w", err)
	}

	// Parse device info
	var di struct {
		UniqueDeviceID string `mapstructure:"UniqueDeviceID,omitempty"`
		DeviceName     string `mapstructure:"DeviceName,omitempty"`
		ProductType    string `mapstructure:"ProductType,omitempty"`
		ProductVersion string `mapstructure:"ProductVersion,omitempty"`
		BuildVersion   string `mapstructure:"BuildVersion,omitempty"`
		WiFiAddress    string `mapstructure:"WiFiAddress,omitempty"`
	}

	err = mapstructure.Decode(info, &di)
	if err != nil {
		return nil, fmt.Errorf("parse device info: %w", err)
	}

	// Return device
	return &Device{
		UDID:        di.UniqueDeviceID,
		Name:        di.DeviceName,
		Type:        di.ProductType,
		OSVersion:   di.ProductVersion,
		OSBuild:     di.BuildVersion,
		WiFiAddress: di.WiFiAddress,
		idev:        idev,
	}, nil
}

// ReportMetrics starts reporting metrics on the returned channel, until the context is canceled.
func (dev *Device) ReportMetrics(ctx context.Context) (<-chan any, error) {
	// Create lockdown client
	ldc, err := idevice.NewLockdownClient(dev.idev)
	if err != nil {
		return nil, fmt.Errorf("create lockdown client: %w", err)
	}

	// Start lockdown lds
	lds, err := ldc.StartSession()
	if err != nil {
		ldc.Close()
		return nil, fmt.Errorf("start lockdown session: %w", err)
	}

	// Start diagnostic service
	drc, err := lds.StartDiagnosticRelayService()
	if err != nil {
		lds.Close()
		ldc.Close()
		return nil, fmt.Errorf("start diagnostic service: %w", err)
	}

	// Spawn Go routine
	mch := make(chan any)

	go func() {
		// Create ticker
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Event loop
	loop:
		for {
			select {
			case <-ctx.Done():
				// Requested to stop
				break loop

			case <-ticker.C:
				// Read battery battery info from device
				response, err := drc.ReadIORegistry("AppleSmartBattery", "")
				if err != nil {
					slog.Error("Unable to read battery info from device", slog.Any("error", err))
					continue
				}

				// Parse battery info
				var battery struct {
					UpdateTime              uint64 `mapstructure:"UpdateTime"`
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

				err = mapstructure.Decode(response, &battery)
				if err != nil {
					slog.Error("Unable to parse batter info", slog.Any("error", err))
					continue
				}

				// Send
				mch <- battery
			}
		}

		// Clean up
		drc.Close()
		lds.Close()
		ldc.Close()

		// Indicate that we're done
		close(mch)
	}()

	return mch, nil
}
