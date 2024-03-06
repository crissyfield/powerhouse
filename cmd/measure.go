package cmd

import (
	"log/slog"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/crissyfield/powerhouse/internal/idevice"
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
	// Create mux
	mux, err := idevice.NewUSBMux()
	if err != nil {
		slog.Error("Unable to create client", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read list of connected devices
	devices, err := mux.Devices()
	if err != nil {
		slog.Error("Unable to read list of connected devices", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	if len(devices) == 0 {
		slog.Warn("No device connected. Exiting")
		os.Exit(1) //nolint
	}

	// ...
	var deviceInfo struct {
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

	info, err := devices[0].Info()
	if err != nil {
		slog.Error("Unable to read device info", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	err = mapstructure.Decode(info, &deviceInfo)
	if err != nil {
		slog.Error("Unable to parse device info", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	slog.Info("Device Info", slog.Any("info", deviceInfo))

	// Create lockdown client
	ldc, err := idevice.NewLockdownClient(devices[0])
	if err != nil {
		slog.Error("Unable to create lockdown client", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	defer ldc.Close()

	// Start lockdown lds
	lds, err := ldc.StartSession()
	if err != nil {
		slog.Error("Unable to start lockdown session", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	defer lds.Close()

	// Start diagnostic service
	drc, err := lds.StartDiagnosticRelayService()
	if err != nil {
		slog.Error("Unable to start diagnostic relay service", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	defer drc.Close()

	// Read battery IORegistry entries
	response, err := drc.ReadIORegistry("AppleSmartBattery", "")
	if err != nil {
		slog.Error("Unable to read AppleSmartBattery IORegistry entry", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// ...
	var battery struct {
		UpdateTime              uint64 `mapstructure:"UpdateTime"`
		ExternalConnected       bool   `mapstructure:"ExternalConnected"`
		IsCharging              bool   `mapstructure:"IsCharging"`
		FullyCharged            bool   `mapstructure:"FullyCharged"`
		CycleCount              uint64 `mapstructure:"CycleCount"`
		DesignCapacity          uint64 `mapstructure:"DesignCapacity"`
		AppleRawMaxCapacity     uint64 `mapstructure:"AppleRawMaxCapacity"`
		AppleRawCurrentCapacity uint64 `mapstructure:"AppleRawCurrentCapacity"`
		AppleRawBatteryVoltage  uint64 `mapstructure:"AppleRawBatteryVoltage"`
		InstantAmperage         int64  `mapstructure:"InstantAmperage"`
	}

	err = mapstructure.Decode(response, &battery)
	if err != nil {
		slog.Error("Unable to unmarshal", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// ...
	slog.Info("Battery", slog.Any("ioregistry", battery))
}
