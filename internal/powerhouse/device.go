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

// ReportBatteryMetrics starts reporting battery metrics on the returned channel, until the context is
// canceled.
func (dev *Device) ReportBatteryMetrics(ctx context.Context) (<-chan *BatteryMetrics, error) {
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

	// Read initial battery info from device
	initRes, err := drc.ReadIORegistry("AppleSmartBattery", "")
	if err != nil {
		drc.Close()
		lds.Close()
		ldc.Close()
		return nil, fmt.Errorf("read initial battery info from device: %w", err)
	}

	// Create initial battery metrics
	initBm, err := batteryMetricsFromIDevice(initRes)
	if err != nil {
		drc.Close()
		lds.Close()
		ldc.Close()
		return nil, fmt.Errorf("create initial battery metrics: %w", err)
	}

	// Spawn Go routine
	mch := make(chan *BatteryMetrics)

	go func() {
		// Create ticker
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Send initial metric
		mch <- initBm

		// Event loop
		lastTime := initBm.Time

	loop:
		for {
			select {
			case <-ctx.Done():
				// Stop
				break loop

			case <-ticker.C:
				// Read battery info from device
				res, err := drc.ReadIORegistry("AppleSmartBattery", "")
				if err != nil {
					mch <- batteryMetricsFromError(fmt.Errorf("read battery info from device: %w", err))
					continue
				}

				// Create battery metrics
				bm, err := batteryMetricsFromIDevice(res)
				if err != nil {
					mch <- batteryMetricsFromError(fmt.Errorf("create battery metrics: %w", err))
					continue
				}

				// Skip duplicates
				if bm.Time.Equal(lastTime) {
					continue
				}

				lastTime = bm.Time

				// Send out
				mch <- bm
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
