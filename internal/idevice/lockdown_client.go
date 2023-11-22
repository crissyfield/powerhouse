package idevice

import (
	"fmt"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

const (
	lockdownPort = 62078
)

// LockdownClient ...
type LockdownClient struct {
	// Underlying lockdown client
	ldc *libimobiledevice.LockdownClient

	// Related device
	dev *Device
}

// NewLockdownClient ...
func NewLockdownClient(dev *Device) (*LockdownClient, error) {
	// Create lockdown connection
	conn, err := dev.dev.NewConnect(lockdownPort)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	// Create client
	ldc := libimobiledevice.NewLockdownClient(conn)

	return &LockdownClient{ldc: ldc, dev: dev}, nil
}

// Close ...
func (*LockdownClient) Close() {
	// TODO
}

// Info ...
func (ldc *LockdownClient) Info() (any, error) {
	// Get lockdown product version
	var value libimobiledevice.LockdownValueResponse

	err := ldc.send(
		&libimobiledevice.LockdownValueRequest{
			LockdownBasicRequest: libimobiledevice.LockdownBasicRequest{
				Label:           libimobiledevice.BundleID,
				ProtocolVersion: libimobiledevice.ProtocolVersion,
				Request:         libimobiledevice.RequestTypeGetValue,
			},
			Domain: "",
			Key:    "",
		},
		&value,
	)

	if err != nil {
		return nil, fmt.Errorf("get lockdown information: %w", err)
	}

	return value.Value, nil
}

// StartSession ...
func (ldc *LockdownClient) StartSession() (*LockdownSession, error) {
	// Get iOS version
	ver, err := ldc.dev.iOSVersionFn()
	if err != nil {
		return nil, fmt.Errorf("get iOS version: %w", err)
	}

	// Get pair record
	pairRecord, err := ldc.dev.readPairRecordFn()
	if err != nil {
		return nil, fmt.Errorf("get pair record: %w", err)
	}

	// Start lockdown session
	var startSession libimobiledevice.LockdownStartSessionResponse

	err = ldc.send(
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
		// Enable SSL
		if err := ldc.ldc.EnableSSL(ver, pairRecord); err != nil {
			return nil, fmt.Errorf("enable SSL: %w", err)
		}
	}

	return newLockdownSession(ldc), nil
}

// send ...
func (ldc *LockdownClient) send(req any, resp any) error {
	// Create request packet
	packetReq, err := ldc.ldc.NewXmlPacket(req)
	if err != nil {
		return fmt.Errorf("create request packet: %w", err)
	}

	// Send request packet
	if err := ldc.ldc.SendPacket(packetReq); err != nil {
		return fmt.Errorf("send request packet: %w", err)
	}

	// Receive response packet
	packetResp, err := ldc.ldc.ReceivePacket()
	if err != nil {
		return fmt.Errorf("receive response packet: %w", err)
	}

	// Parse response packet
	if err := packetResp.Unmarshal(resp); err != nil {
		return fmt.Errorf("unmarshal response packet: %w", err)
	}

	return nil
}
