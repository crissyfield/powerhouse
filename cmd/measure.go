package cmd

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
	"github.com/spf13/cobra"
	"howett.net/plist"

	"github.com/crissyfield/powerhouse/internal/powerhouse"
)

// CmdMeasure defines the CLI sub-command 'list'.
var CmdMeasure = &cobra.Command{
	Use:   "measure [flags]",
	Short: "...",
	Args:  cobra.NoArgs,
	Run:   runMeasure,
}

// Initialize CLI options.
func init() {
}

// runMeasure is called when the "test" command is used.
func runMeasure(_ *cobra.Command, _ []string) {
	// Create client
	client, err := powerhouse.NewClient()
	if err != nil {
		slog.Error("Unable to create client", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read list of connected devices
	devices, err := client.Devices()
	if err != nil {
		slog.Error("Unable to read list of connected devices", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if len(devices) == 0 {
		slog.Warn("No device connected. Exiting")
		os.Exit(1) //nolint
	}

	// Get device information
	info, err := devices[0].Info()
	if err != nil {
		slog.Error("Unable to get device information", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Extract product version
	pv := strings.Split(info.ProductVersion, ".")

	iOSVersion := make([]int, len(pv))
	for i, v := range pv {
		iOSVersion[i], _ = strconv.Atoi(v)
	}

	// Fetch lockdown query type
	var queryType libimobiledevice.LockdownTypeResponse

	err = devices[0].LockdownSend(
		&libimobiledevice.LockdownBasicRequest{
			Label:           libimobiledevice.BundleID,
			ProtocolVersion: libimobiledevice.ProtocolVersion,
			Request:         libimobiledevice.RequestTypeQueryType,
		},
		&queryType,
	)

	if err != nil {
		slog.Error("Unable to fetch lockdown query type", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read pair record
	pairRecord, err := devices[0].ReadPairRecord()
	if err != nil {
		slog.Error("Unable to read pair record", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Start lockdown session
	var startSession libimobiledevice.LockdownStartSessionResponse

	err = devices[0].LockdownSend(
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
		slog.Error("Unable to start lockdown session", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// TODO
	if startSession.EnableSessionSSL {
		if err := devices[0].LDC().EnableSSL(iOSVersion, pairRecord); err != nil {
			slog.Error("Unable to enable SSL", slog.Any("error", err))
			os.Exit(1) //nolint
		}
	}

	// Start lockdown service
	var startService libimobiledevice.LockdownStartServiceResponse

	err = devices[0].LockdownSend(
		&libimobiledevice.LockdownStartServiceRequest{
			LockdownBasicRequest: libimobiledevice.LockdownBasicRequest{
				Label:           libimobiledevice.BundleID,
				ProtocolVersion: libimobiledevice.ProtocolVersion,
				Request:         libimobiledevice.RequestTypeStartService,
			},
			Service: libimobiledevice.DiagnosticsRelayServiceName,
		},
		&startService,
	)

	if err != nil {
		slog.Error("Unable to start lockdown service", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if startService.Error != "" {
		slog.Error("Unable to start service", slog.String("error", startService.Error))
		os.Exit(1) //nolint
	}

	// TODO: Stop lockdown session

	// ...
	diagnosticRelayConn, err := devices[0].NewConnect(startService.Port)
	if err != nil {
		slog.Error("Unable to create diagnostic relay connection", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if startService.EnableServiceSSL {
		if err := diagnosticRelayConn.Handshake(iOSVersion, pairRecord); err != nil {
			slog.Error("Unable to enable SSL", slog.Any("error", err))
			os.Exit(1) //nolint
		}
	}

	drc := libimobiledevice.NewDiagnosticsRelayClient(diagnosticRelayConn)

	// ...
	type DiagnosticsRelayIORegistryRequest struct {
		Request    string `plist:"Request"`
		EntryName  string `plist:"EntryName,omitempty"`
		EntryClass string `plist:"EntryClass,omitempty"`
	}

	ioRegistryReq, err := drc.NewXmlPacket(&DiagnosticsRelayIORegistryRequest{
		Request:   "IORegistry",
		EntryName: "AppleSmartBattery",
	})

	if err != nil {
		slog.Error("Unable to create IORegistry lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if err = drc.SendPacket(ioRegistryReq); err != nil {
		slog.Error("Unable to send IORegistry lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Receive lockdown response
	bufLen, err := diagnosticRelayConn.Read(4)
	if err != nil {
		slog.Error("Unable to receive packet length", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	lenPkg := binary.BigEndian.Uint32(bufLen)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(bufLen)

	buf, err := diagnosticRelayConn.Read(int(lenPkg))
	if err != nil {
		slog.Error("Unable to receive packet body", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	buffer.Write(buf)

	type servicePacket struct {
		length uint32
		body   []byte
	}

	var respPkt servicePacket
	err = binary.Read(buffer, binary.BigEndian, &respPkt.length)
	if err != nil {
		slog.Error("Unable to unpack packet", slog.Any("error", err))
		os.Exit(1) //nolint
	}
	respPkt.body = buffer.Bytes()

	var reply struct {
		libimobiledevice.LockdownBasicResponse
		Diagnostics struct {
			IORegistry struct {
				UpdateTime              uint64 `plist:"UpdateTime"`
				ExternalConnected       bool   `plist:"ExternalConnected"`
				IsCharging              bool   `plist:"IsCharging"`
				FullyCharged            bool   `plist:"FullyCharged"`
				CycleCount              uint64 `plist:"CycleCount"`
				DesignCapacity          uint64 `plist:"DesignCapacity"`
				AppleRawMaxCapacity     uint64 `plist:"AppleRawMaxCapacity"`
				AppleRawCurrentCapacity uint64 `plist:"AppleRawCurrentCapacity"`
				AppleRawBatteryVoltage  uint64 `plist:"AppleRawBatteryVoltage"`
				InstantAmperage         uint64 `plist:"InstantAmperage"`
			} `plist:"IORegistry"`
		} `plist:"Diagnostics"`
	}

	if _, err := plist.Unmarshal(respPkt.body, &reply); err != nil {
		slog.Error("Unable to unmarshal", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if reply.Error != "" {
		slog.Error("???", slog.String("error", reply.Error))
		os.Exit(1) //nolint
	}

	// ...
	slog.Info("Battery", slog.Any("ioregistry", reply.Diagnostics.IORegistry))
}
