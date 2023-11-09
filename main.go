package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version will be set during build.
var Version = "(unknown)"

// CmdRoot defines the CLI root command.
var CmdRoot = &cobra.Command{
	Use:               "powerhouse",
	Long:              "...",
	Args:              cobra.NoArgs,
	Version:           Version,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	PersistentPreRunE: setup,
	Run:               runRoot,
}

// Initialize CLI options.
func init() {
	// Logging
	CmdRoot.PersistentFlags().String("log-level", "info", "verbosity of logging output")
	CmdRoot.PersistentFlags().Bool("log-as-json", false, "change logging format to JSON")
}

// setup will set up configuration management and logging.
//
// Configuration options can be set via the command line, via a configuration file (in the current folder, at
// "/etc/powerhouse/config.yaml" or at "~/.config/powerhouse/config.yaml"), and via environment variables
// (all uppercase and prefixed with "POWERHOUSE_").
func setup(cmd *cobra.Command, _ []string) error {
	// Connect all options to Viper
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return fmt.Errorf("bind command line flags: %w", err)
	}

	// Environment variables
	viper.SetEnvPrefix("POWERHOUSE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// Configuration file
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/powerhouse")
	viper.AddConfigPath(os.Getenv("HOME") + "/.config/powerhouse")
	viper.AddConfigPath(".")

	// Configuration file
	if err := viper.ReadInConfig(); err != nil {
		// Don't fail if config not found
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return fmt.Errorf("read config file: %w", err)
		}
	}

	// Logging
	var level slog.Level

	err = level.UnmarshalText([]byte(viper.GetString("log-level")))
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}

	var handler slog.Handler

	if viper.GetBool("log-as-json") {
		// Use JSON handler
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		// Use text handler
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(handler))

	return nil
}

// main is the main entry point of the command.
func main() {
	if err := CmdRoot.Execute(); err != nil {
		slog.Error("Unable to execute command", slog.Any("error", err))
	}
}

// runRoot is called when the "root" command is used.
func runRoot(_ *cobra.Command, _ []string) {
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

	for _, d := range devices {
		// Read full lockdown information
		detail, err := d.GetValue("", "")
		if err != nil {
			slog.Error("Unable to read lockdown information", slog.Any("error", err))
			continue
		}

		// Parse into device info
		var dd DeviceInfo

		err = mapstructure.Decode(detail, &dd)
		if err != nil {
			slog.Error("Unable to parse into device info", slog.Any("error", err))
			continue
		}

		slog.Info("Device", slog.Any("detail", dd))
	}
}
