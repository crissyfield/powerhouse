package powerhouse

import (
	"fmt"

	"github.com/crissyfield/powerhouse/internal/idevice"
)

// Powerhouse is the main object of the powerhouse package.
type Powerhouse struct {
	mux *idevice.USBMux // USBMux
}

// New creates a new Powerhouse object.
func New() (*Powerhouse, error) {
	// Create USB mux
	mux, err := idevice.NewUSBMux()
	if err != nil {
		return nil, fmt.Errorf("create USBmux: %w", err)
	}

	return &Powerhouse{mux: mux}, nil
}

// Devices ...
func (c *Powerhouse) Devices() ([]*Device, error) {
	// Get list of connected devices
	idevs, err := c.mux.Devices()
	if err != nil {
		return nil, fmt.Errorf("get list of connected devices: %w", err)
	}

	// ...
	devices := make([]*Device, 0, len(idevs))

	for _, idev := range idevs {
		// Create device
		device, err := newDevice(idev)
		if err != nil {
			return nil, fmt.Errorf("create device: %w", err)
		}

		// Append
		devices = append(devices, device)
	}

	return devices, nil
}
