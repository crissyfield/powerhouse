package powerhouse

import (
	"fmt"

	"github.com/crissyfield/powerhouse/internal/idevice"
	"github.com/mitchellh/mapstructure"
)

// Device wraps information of a specific iDevice.
type Device struct {
	UDID        string // Unique device ID
	Name        string // Device name
	Type        string // Device type
	OSVersion   string // Version of the installed OS
	OSBuild     string // Build number of the installed OS
	WiFiAddress string // MAC address of the device
}

// newDevice creates a new iDevice.
func newDevice(dev *idevice.Device) (*Device, error) {
	// Get device info
	info, err := dev.Info()
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
	}, nil
}
