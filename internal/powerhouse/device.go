package powerhouse

import (
	"fmt"
	"sync"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
	"github.com/mitchellh/mapstructure"
)

// Device ...
type Device struct {
	d giDevice.Device

	// LockdownClient
	ldc     *libimobiledevice.LockdownClient
	ldcErr  error
	ldcOnce sync.Once
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

// LDC ...
func (d *Device) LDC() *libimobiledevice.LockdownClient {
	return d.ldc
}

// getLockdownClient ...
func (d *Device) getLockdownClient() (*libimobiledevice.LockdownClient, error) {
	// Deferred creation
	d.ldcOnce.Do(func() {
		// Connect to lockdown service
		ldConn, err := d.d.NewConnect(giDevice.LockdownPort)
		if err != nil {
			d.ldcErr = fmt.Errorf("create lockdown connection: %w", err)
			return
		}

		// Create client
		d.ldc = libimobiledevice.NewLockdownClient(ldConn)
	})

	return d.ldc, d.ldcErr
}

// LockdownSend ...
func (d *Device) LockdownSend(req any, resp any) error {
	// ...
	ldc, err := d.getLockdownClient()
	if err != nil {
		return fmt.Errorf("get lockdown client: %w", err)
	}

	// ...
	packetReq, err := ldc.NewXmlPacket(req)
	if err != nil {
		return fmt.Errorf("create request packet: %w", err)
	}

	// ...
	if err := ldc.SendPacket(packetReq); err != nil {
		return fmt.Errorf("send request packet: %w", err)
	}

	// ...
	packetResp, err := ldc.ReceivePacket()
	if err != nil {
		return fmt.Errorf("receive response packet: %w", err)
	}

	// ...
	if err := packetResp.Unmarshal(resp); err != nil {
		return fmt.Errorf("unmarshal response packet: %w", err)
	}

	return nil
}
