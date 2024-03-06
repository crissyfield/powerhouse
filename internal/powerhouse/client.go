package powerhouse

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/crissyfield/powerhouse/internal/idevice"
	"github.com/mitchellh/mapstructure"
)

// Client ...
type Client struct {
	// USBMux
	mux *idevice.USBMux
}

// Device ...
type Device struct {
	UDID string
}

// NewClient ...
func NewClient() (*Client, error) {
	// Create USB mux
	mux, err := idevice.NewUSBMux()
	if err != nil {
		return nil, fmt.Errorf("create USBmux: %w", err)
	}

	return &Client{mux: mux}, nil
}

// Devices ...
func (c *Client) Devices() ([]*Device, error) {
	// Get list of connected devices
	devs, err := c.mux.Devices()
	if err != nil {
		return nil, fmt.Errorf("get list of connected devices: %w", err)
	}

	// ...
	devices := make([]*Device, 0, len(devs))

	for _, dev := range devs {
		// ...
		info, err := dev.Info()
		if err != nil {
			slog.Warn("Unable to get device info", slog.Any("error", err))
			continue
		}

		// ...
		var di struct {
			UniqueDeviceID            string `mapstructure:"UniqueDeviceID,omitempty"`
			ProductType               string `mapstructure:"ProductType,omitempty"`
			ProductVersion            string `mapstructure:"ProductVersion,omitempty"`
			ProductName               string `mapstructure:"ProductName,omitempty"`
			DeviceName                string `mapstructure:"DeviceName,omitempty"`
			DeviceClass               string `mapstructure:"DeviceClass,omitempty"`
			ModelNumber               string `mapstructure:"ModelNumber,omitempty"`
			SerialNumber              string `mapstructure:"SerialNumber,omitempty"`
			SIMStatus                 string `mapstructure:"SIMStatus,omitempty"`
			PhoneNumber               string `mapstructure:"PhoneNumber,omitempty"`
			CPUArchitecture           string `mapstructure:"CPUArchitecture,omitempty"`
			ProtocolVersion           string `mapstructure:"ProtocolVersion,omitempty"`
			RegionInfo                string `mapstructure:"RegionInfo,omitempty"`
			TelephonyCapability       bool   `mapstructure:"TelephonyCapability,omitempty"`
			TimeZone                  string `mapstructure:"TimeZone,omitempty"`
			WiFiAddress               string `mapstructure:"WiFiAddress,omitempty"`
			WirelessBoardSerialNumber string `mapstructure:"WirelessBoardSerialNumber,omitempty"`
			BluetoothAddress          string `mapstructure:"BluetoothAddress,omitempty"`
			BuildVersion              string `mapstructure:"BuildVersion,omitempty"`
		}

		err = mapstructure.Decode(info, &di)
		if err != nil {
			slog.Warn("Unable to parse info", slog.Any("error", err))
			continue
		}

		_ = json.NewEncoder(os.Stdout).Encode(di)

		// ...
		devices = append(devices, &Device{
			UDID: di.UniqueDeviceID,
		})
	}

	return devices, nil
}
