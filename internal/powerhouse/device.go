package powerhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/crissyfield/powerhouse/internal/idevice"
)

// Device wraps information of a specific iDevice.
type Device struct {
	ConnectionType string // Either "USB" or "Network"
	UDID           string // Unique device ID
	Name           string // Device name
	Type           string // Device type
	OSVersion      string // Version of the installed OS
	OSBuild        string // Build number of the installed OS
	WiFiAddress    string // MAC address of the device

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
		UniqueDeviceID string `mapstructure:"UniqueDeviceID"`
		DeviceName     string `mapstructure:"DeviceName"`
		ProductType    string `mapstructure:"ProductType"`
		ProductVersion string `mapstructure:"ProductVersion"`
		BuildVersion   string `mapstructure:"BuildVersion"`
		WiFiAddress    string `mapstructure:"WiFiAddress"`
	}

	err = mapstructure.Decode(info, &di)
	if err != nil {
		return nil, fmt.Errorf("parse device info: %w", err)
	}

	// Return device
	return &Device{
		ConnectionType: idev.ConnectionType(),
		UDID:           di.UniqueDeviceID,
		Name:           di.DeviceName,
		Type:           di.ProductType,
		OSVersion:      di.ProductVersion,
		OSBuild:        di.BuildVersion,
		WiFiAddress:    di.WiFiAddress,
		idev:           idev,
	}, nil
}

// Metrics ...
type Metrics struct {
	Err       error
	Battery   *BatteryMetrics
	Backlight *BacklightMetrics
}

// ReportMetrics starts reporting battery and backlight metrics on the returned channel, until the context is
// canceled.
func (dev *Device) ReportMetrics(ctx context.Context) (<-chan *Metrics, error) {
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

	// Read initial battery metrics
	initBattery, err := batteryMetricsFromDiagnosticRelayClient(drc)
	if err != nil {
		drc.Close()
		lds.Close()
		ldc.Close()
		return nil, fmt.Errorf("create initial battery metrics: %w", err)
	}

	// Read initial backlight metrics
	initBacklight, err := backlightMetricsFromDiagnosticRelayClient(drc)
	if err != nil {
		drc.Close()
		lds.Close()
		ldc.Close()
		return nil, fmt.Errorf("create initial backlight metrics: %w", err)
	}

	// Spawn Go routine
	metrics := make(chan *Metrics)

	go func() {
		// Create ticker
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Send initial metrics
		metrics <- &Metrics{Battery: initBattery, Backlight: initBacklight}

		// Event loop
		lastBatteryTime := initBattery.Time

	loop:
		for {
			select {
			case <-ctx.Done():
				// Stop
				break loop

			case <-ticker.C:
				// Read battery metrics
				battery, err := batteryMetricsFromDiagnosticRelayClient(drc)
				if err != nil {
					metrics <- &Metrics{Err: fmt.Errorf("read battery metrics: %w", err)}
					continue
				}

				// Skip duplicates
				if battery.Time.Equal(lastBatteryTime) {
					continue
				}

				lastBatteryTime = battery.Time

				// Read backlight metrics
				backlight, err := backlightMetricsFromDiagnosticRelayClient(drc)
				if err != nil {
					metrics <- &Metrics{Err: fmt.Errorf("read backlight info from device: %w", err)}
					continue
				}

				// Send out
				metrics <- &Metrics{Battery: battery, Backlight: backlight}
			}
		}

		// Clean up
		drc.Close()
		lds.Close()
		ldc.Close()

		// We're done
		close(metrics)
	}()

	return metrics, nil
}
