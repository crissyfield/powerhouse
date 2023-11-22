package idevice

import (
	"encoding/binary"
	"fmt"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
	"howett.net/plist"
)

// DiagnosticRelayClient ...
type DiagnosticRelayClient struct {
	// Underlying diagnostic relay client
	drc *libimobiledevice.DiagnosticsRelayClient
}

// newDiagnosticRelayClient ...
func newDiagnosticRelayClient(drc *libimobiledevice.DiagnosticsRelayClient) *DiagnosticRelayClient {
	return &DiagnosticRelayClient{drc: drc}
}

// Close ...
func (*DiagnosticRelayClient) Close() {
	// TODO
}

// ReadIORegistry ...
func (drc *DiagnosticRelayClient) ReadIORegistry(name string, class string) (any, error) {
	// IORegistry protocol
	type Request struct {
		Request    string `plist:"Request"`
		EntryName  string `plist:"EntryName,omitempty"`
		EntryClass string `plist:"EntryClass,omitempty"`
	}

	type Response struct {
		libimobiledevice.LockdownBasicResponse
		Diagnostics struct {
			IORegistry any `plist:"IORegistry"`
		} `plist:"Diagnostics"`
	}

	// Read IORegistry entry
	var resp Response

	err := drc.send(
		&Request{
			Request:    "IORegistry",
			EntryName:  name,
			EntryClass: class,
		},
		&resp,
	)

	if err != nil {
		return nil, fmt.Errorf("read IORegistry entry: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("read IORegistry entry (server): %s", resp.Error)
	}

	return resp.Diagnostics.IORegistry, nil
}

// send ...
func (drc *DiagnosticRelayClient) send(req any, resp any) error {
	// Create request packet
	packetReq, err := drc.drc.NewXmlPacket(req)
	if err != nil {
		return fmt.Errorf("create request packet: %w", err)
	}

	// Send request packet
	if err := drc.drc.SendPacket(packetReq); err != nil {
		return fmt.Errorf("send request packet: %w", err)
	}

	// Receive response packet length
	responseLenRaw, err := drc.drc.InnerConn().Read(4)
	if err != nil {
		return fmt.Errorf("receive response length: %w", err)
	}

	responseLen := binary.BigEndian.Uint32(responseLenRaw)

	// Receive response packet response
	response, err := drc.drc.InnerConn().Read(int(responseLen))
	if err != nil {
		return fmt.Errorf("receive response body: %w", err)
	}

	// Parse response packet body
	if _, err := plist.Unmarshal(response, resp); err != nil {
		return fmt.Errorf("unmarshal response packet: %w", err)
	}

	return nil
}
