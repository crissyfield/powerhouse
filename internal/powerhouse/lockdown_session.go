package powerhouse

import (
	"fmt"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

// LockdownSession ...
type LockdownSession struct {
	pr *libimobiledevice.PairRecord
	d  *Device
}

// StartService ...
func (ls *LockdownSession) StartService(serviceName string) (libimobiledevice.InnerConn, error) {
	// Get iOS version
	iosv, err := ls.d.getIOSVersion()
	if err != nil {
		return nil, fmt.Errorf("get iOS version: %w", err)
	}

	// Start lockdown service
	var startService libimobiledevice.LockdownStartServiceResponse

	err = ls.d.lockdownSend(
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
	innerConn, err := ls.d.d.NewConnect(startService.Port)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	// Optionally, enable SSH
	if startService.EnableServiceSSL {
		if err := innerConn.Handshake(iosv, ls.pr); err != nil {
			return nil, fmt.Errorf("enable SSL: %w", err)
		}
	}

	return innerConn, nil
}
