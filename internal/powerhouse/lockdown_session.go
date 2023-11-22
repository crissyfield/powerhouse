package powerhouse

import (
	"fmt"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

// LockdownSession ...
type LockdownSession struct {
	// Related lockdown client
	ldc *LockdownClient
}

// newLockdownSession ...
func newLockdownSession(ldc *LockdownClient) *LockdownSession {
	return &LockdownSession{ldc: ldc}
}

// Close ...
func (*LockdownSession) Close() {
	// TODO
}

// StartService ...
func (lds *LockdownSession) StartService(serviceName string) (libimobiledevice.InnerConn, error) {
	// Get iOS version
	ver, err := lds.ldc.dev.iOSVersionFn()
	if err != nil {
		return nil, fmt.Errorf("get iOS version: %w", err)
	}

	// Get pair record
	pairRecord, err := lds.ldc.dev.readPairRecordFn()
	if err != nil {
		return nil, fmt.Errorf("get pair record: %w", err)
	}

	// Start lockdown service
	var startService libimobiledevice.LockdownStartServiceResponse

	err = lds.ldc.send(
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
	conn, err := lds.ldc.dev.dev.NewConnect(startService.Port)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	// Optionally, enable SSH
	if startService.EnableServiceSSL {
		if err := conn.Handshake(ver, pairRecord); err != nil {
			return nil, fmt.Errorf("enable SSL: %w", err)
		}
	}

	return conn, nil
}
