package cmd

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"os"
	"strconv"
	"strings"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"howett.net/plist"
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
	// Create USBMUX connection
	usbmux, err := giDevice.NewUsbmux()
	if err != nil {
		slog.Error("Unable to create USBMUX connection", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read list of connected devices
	devices, err := usbmux.Devices()
	if err != nil {
		slog.Error("Unable to read list of connected devices", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// ...
	if len(devices) == 0 {
		slog.Warn("No device connected. Exiting")
		os.Exit(1) //nolint
	}

	// Dump device information
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

	// Read full lockdown information
	detail, err := devices[0].GetValue("", "")
	if err != nil {
		slog.Error("Unable to read lockdown information", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Parse into device info
	var dd DeviceInfo

	err = mapstructure.Decode(detail, &dd)
	if err != nil {
		slog.Error("Unable to parse into device info", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Extract product version
	pv := strings.Split(dd.ProductVersion, ".")

	iOSVersion := make([]int, len(pv))
	for i, v := range pv {
		iOSVersion[i], _ = strconv.Atoi(v)
	}

	// ...
	lockdownConn, err := devices[0].NewConnect(giDevice.LockdownPort)
	if err != nil {
		slog.Error("Unable to create lockdown connection", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	lc := libimobiledevice.NewLockdownClient(lockdownConn)

	// Create and send basic lockdown request
	queryTypeReq, err := lc.NewXmlPacket(lc.NewBasicRequest(libimobiledevice.RequestTypeQueryType))
	if err != nil {
		slog.Error("Unable to create basic lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if err := lc.SendPacket(queryTypeReq); err != nil {
		slog.Error("Unable to send basic lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Receive lockdown response
	queryTypeResp, err := lc.ReceivePacket()
	if err != nil {
		slog.Error("Unable to receive lockdown response", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	var queryType libimobiledevice.LockdownTypeResponse

	if err := queryTypeResp.Unmarshal(&queryType); err != nil {
		slog.Error("Unable to unmarshal", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read pair record
	pairRecord, err := devices[0].ReadPairRecord()
	if err != nil {
		slog.Error("Unable to read pair record", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Create and send start session lockdown request
	startSessionReq, err := lc.NewXmlPacket(lc.NewStartSessionRequest(pairRecord.SystemBUID, pairRecord.HostID))
	if err != nil {
		slog.Error("Unable to create start session lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if err = lc.SendPacket(startSessionReq); err != nil {
		slog.Error("Unable to send start session lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Receive lockdown response
	startSessionResp, err := lc.ReceivePacket()
	if err != nil {
		slog.Error("Unable to receive lockdown response", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	var startSession libimobiledevice.LockdownStartSessionResponse

	if err := startSessionResp.Unmarshal(&startSession); err != nil {
		slog.Error("Unable to unmarshal", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if startSession.EnableSessionSSL {
		if err := lc.EnableSSL(iOSVersion, pairRecord); err != nil {
			slog.Error("Unable to enable SSL", slog.Any("error", err))
			os.Exit(1) //nolint
		}
	}

	// Create and send start service lockdown request
	startServiceReq, err := lc.NewXmlPacket(lc.NewStartServiceRequest(libimobiledevice.DiagnosticsRelayServiceName))
	if err != nil {
		slog.Error("Unable to create start service lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if err = lc.SendPacket(startServiceReq); err != nil {
		slog.Error("Unable to send start service lockdown request", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Receive lockdown response
	startServiceResp, err := lc.ReceivePacket()
	if err != nil {
		slog.Error("Unable to receive lockdown response", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	var startService libimobiledevice.LockdownStartServiceResponse

	if err := startServiceResp.Unmarshal(&startService); err != nil {
		slog.Error("Unable to unmarshal", slog.Any("error", err))
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
