package powerhouse

import (
	"fmt"
	"strconv"
	"strings"
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

	// pair record
	pairRecord     *giDevice.PairRecord
	pairRecordErr  error
	pairRecordOnce sync.Once

	// iOS version
	iOSVersion     []int
	iOSVersionErr  error
	iOSVersionOnce sync.Once
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

// StartLockdownSession ...
func (d *Device) StartLockdownSession() (*giDevice.PairRecord, error) {
	// Get iOS version
	iosv, err := d.getIOSVersion()
	if err != nil {
		return nil, fmt.Errorf("get iOS version: %w", err)
	}

	// Get pair record
	pairRecord, err := d.getPairRecord()
	if err != nil {
		return nil, fmt.Errorf("get pair record: %w", err)
	}

	// Start lockdown session
	var startSession libimobiledevice.LockdownStartSessionResponse

	err = d.lockdownSend(
		&libimobiledevice.LockdownStartSessionRequest{
			LockdownBasicRequest: libimobiledevice.LockdownBasicRequest{
				Label:           libimobiledevice.BundleID,
				ProtocolVersion: libimobiledevice.ProtocolVersion,
				Request:         libimobiledevice.RequestTypeStartSession,
			},
			SystemBUID: pairRecord.SystemBUID,
			HostID:     pairRecord.HostID,
		},
		&startSession,
	)

	if err != nil {
		return nil, fmt.Errorf("start lockdown session: %w", err)
	}

	// Optionally enable SSL
	if startSession.EnableSessionSSL {
		// Get lockdown client
		ldc, err := d.getLockdownClient()
		if err != nil {
			return nil, fmt.Errorf("get lockdown cient: %w", err)
		}

		// Enable SSL
		if err := ldc.EnableSSL(iosv, pairRecord); err != nil {
			return nil, fmt.Errorf("enable SSL: %w", err)
		}
	}

	return pairRecord, nil
}

// StartLockdownService ...
func (d *Device) StartLockdownService(serviceName string, pairRecord *giDevice.PairRecord) (libimobiledevice.InnerConn, error) {
	// Get iOS version
	iosv, err := d.getIOSVersion()
	if err != nil {
		return nil, fmt.Errorf("get iOS version: %w", err)
	}

	// Start lockdown service
	var startService libimobiledevice.LockdownStartServiceResponse

	err = d.lockdownSend(
		&libimobiledevice.LockdownStartServiceRequest{
			LockdownBasicRequest: libimobiledevice.LockdownBasicRequest{
				Label:           libimobiledevice.BundleID,
				ProtocolVersion: libimobiledevice.ProtocolVersion,
				Request:         libimobiledevice.RequestTypeStartService,
			},
			Service: serviceName,
		},
		&startService,
	)

	if err != nil {
		return nil, fmt.Errorf("start lockdown service: %w", err)
	}

	if startService.Error != "" {
		return nil, fmt.Errorf("start lockdown service (server): %s", startService.Error)
	}

	// Create new connection
	innerConn, err := d.d.NewConnect(startService.Port)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	// Optionally, enable SSH
	if startService.EnableServiceSSL {
		if err := innerConn.Handshake(iosv, pairRecord); err != nil {
			return nil, fmt.Errorf("enable SSL: %w", err)
		}
	}

	return innerConn, nil
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

// gerPairRecord ...
func (d *Device) getPairRecord() (*giDevice.PairRecord, error) {
	// Deferred creation
	d.pairRecordOnce.Do(func() {
		// Read pair record
		d.pairRecord, d.pairRecordErr = d.d.ReadPairRecord()
	})

	return d.pairRecord, d.pairRecordErr
}

// getIOSVersion ...
func (d *Device) getIOSVersion() ([]int, error) {
	// Deferred fetch
	d.iOSVersionOnce.Do(func() {
		// Get lockdown product version
		var value libimobiledevice.LockdownValueResponse

		err := d.lockdownSend(
			&libimobiledevice.LockdownValueRequest{
				LockdownBasicRequest: libimobiledevice.LockdownBasicRequest{
					Label:           libimobiledevice.BundleID,
					ProtocolVersion: libimobiledevice.ProtocolVersion,
					Request:         libimobiledevice.RequestTypeGetValue,
				},
				Domain: "",
				Key:    "ProductVersion",
			},
			&value,
		)

		if err != nil {
			d.iOSVersionErr = fmt.Errorf("get lockdown product version: %w", err)
			return
		}

		productVersion, ok := value.Value.(string)
		if !ok {
			d.iOSVersionErr = fmt.Errorf("unexpected product version: %v", value.Value)
			return
		}

		// Extract product version
		pv := strings.Split(productVersion, ".")

		d.iOSVersion = make([]int, len(pv))
		for i, v := range pv {
			d.iOSVersion[i], _ = strconv.Atoi(v)
		}
	})

	return d.iOSVersion, d.iOSVersionErr
}

// lockdownSend ...
func (d *Device) lockdownSend(req any, resp any) error {
	// Get lockdown client
	ldc, err := d.getLockdownClient()
	if err != nil {
		return fmt.Errorf("get lockdown client: %w", err)
	}

	// Create request packet
	packetReq, err := ldc.NewXmlPacket(req)
	if err != nil {
		return fmt.Errorf("create request packet: %w", err)
	}

	// Send request packet
	if err := ldc.SendPacket(packetReq); err != nil {
		return fmt.Errorf("send request packet: %w", err)
	}

	// Receive response packet
	packetResp, err := ldc.ReceivePacket()
	if err != nil {
		return fmt.Errorf("receive response packet: %w", err)
	}

	// Parse response packet
	if err := packetResp.Unmarshal(resp); err != nil {
		return fmt.Errorf("unmarshal response packet: %w", err)
	}

	return nil
}
