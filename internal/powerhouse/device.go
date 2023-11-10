package powerhouse

import (
	"fmt"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
	"github.com/mitchellh/mapstructure"
)

// Device ...
type Device struct {
	d giDevice.Device
}

// newDevice ...
func newDevice(d giDevice.Device) (*Device, error) {
	return &Device{d: d}, nil
}

// DeviceInfo ...
type DeviceInfo struct {
	DeviceName                string `mapstructure:"DeviceName,omitempty"`
	DeviceColor               string `mapstructure:"DeviceColor,omitempty"`
	DeviceClass               string `mapstructure:"DeviceClass,omitempty"`
	ProductVersion            string `mapstructure:"ProductVersion,omitempty"`
	ProductType               string `mapstructure:"ProductType,omitempty"`
	ProductName               string `mapstructure:"ProductName,omitempty"`
	ModelNumber               string `mapstructure:"ModelNumber,omitempty"`
	SerialNumber              string `mapstructure:"SerialNumber,omitempty"`
	SIMStatus                 string `mapstructure:"SIMStatus,omitempty"`
	PhoneNumber               string `mapstructure:"PhoneNumber,omitempty"`
	CPUArchitecture           string `mapstructure:"CPUArchitecture,omitempty"`
	ProtocolVersion           string `mapstructure:"ProtocolVersion,omitempty"`
	RegionInfo                string `mapstructure:"RegionInfo,omitempty"`
	TelephonyCapability       bool   `mapstructure:"TelephonyCapability,omitempty"`
	TimeZone                  string `mapstructure:"TimeZone,omitempty"`
	UniqueDeviceID            string `mapstructure:"UniqueDeviceID,omitempty"`
	WiFiAddress               string `mapstructure:"WiFiAddress,omitempty"`
	WirelessBoardSerialNumber string `mapstructure:"WirelessBoardSerialNumber,omitempty"`
	BluetoothAddress          string `mapstructure:"BluetoothAddress,omitempty"`
	BuildVersion              string `mapstructure:"BuildVersion,omitempty"`
}

// Info ...
func (d *Device) Info() (*DeviceInfo, error) {
	detail, err := d.d.GetValue("", "")
	if err != nil {
		return nil, fmt.Errorf("read lockdown device information: %w", err)
	}

	// Parse into device info
	var di DeviceInfo

	err = mapstructure.Decode(detail, &di)
	if err != nil {
		return nil, fmt.Errorf("parse lockdown device information: %w", err)
	}

	return &di, nil
}

// NewConnect ...
func (d *Device) NewConnect(port int) (libimobiledevice.InnerConn, error) {
	return d.d.NewConnect(port)
}

// ReadPairRecord ...
func (d *Device) ReadPairRecord() (*giDevice.PairRecord, error) {
	return d.d.ReadPairRecord()
}
